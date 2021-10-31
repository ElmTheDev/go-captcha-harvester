package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/ElmTheDev/go-captcha-harvester/constants"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/phf/go-queue/queue"
)

type Harvester struct {
	CustomName string

	Email    string
	Password string
	Proxy    string

	Type        string
	IsReady     bool
	IsSolving   bool
	SolvedCount int
	Url         *url.URL

	Done chan bool

	Browser *rod.Browser
	Page    *rod.Page

	Loader string
	HTML   string

	Queue *queue.Queue
}

type queueEntry struct {
	SiteKey      string
	IsEnterprise bool
	RenderParams string

	Channel chan queueResult
}

type queueResult struct {
	Error error
	Token string
}

type harvesterError struct {
	Error string `json:"harvest_error"`
}

// Initializes harvester and prepares it for solving
func (e *Harvester) Initialize() {
	e.Queue = queue.New().Init()

	e.setupHtml()
	e.Done = make(chan bool)

	// Start browser
	url := launcher.New().
		Set("window-size", "200,700").
		Headless(false).
		MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()

	// Setup request interception on the URL
	router := browser.HijackRequests()
	defer router.MustStop()
	router.MustAdd(fmt.Sprintf("*%s*", e.Url.Host), func(ctx *rod.Hijack) {
		_ = ctx.LoadResponse(http.DefaultClient, true)

		ctx.Response.SetBody(e.HTML)
	})

	go router.Run()

	// Load the page and prepare it for solving
	page := browser.MustPage(e.Url.String())

	// Wait for page to load before continuing
	page.MustElement("title").MustText()

	// Initialize stuct
	e.Browser = browser
	e.Page = page

	// Load harvesting code into the page
	// TODO: Support multiple
	page.MustEval(e.Loader)

	e.IsReady = true
	go e.clearQueue()
	<-e.Done

	browser.MustClose()
}

func (e *Harvester) Harvest(siteKey string, isEnterprise bool, renderParams map[string]string) (string, error) {
	renderParamsObject, err := json.Marshal(renderParams)
	if err != nil {
		return "", err
	}

	resultChannel := make(chan queueResult)
	e.Queue.PushBack(queueEntry{
		SiteKey: siteKey,
		RenderParams: string(renderParamsObject),
		IsEnterprise: isEnterprise,
		Channel: resultChannel,
	})

	resultParsed := <-resultChannel

	if resultParsed.Error != nil {
		return "", resultParsed.Error
	}

	return resultParsed.Token, nil
}

func (e *Harvester) executeHarvest(siteKey string, isEnterprise bool, renderParams string) queueResult {
	println(e.CustomName, e.Page)
	result := e.Page.MustEval(e.getJsCallString(siteKey, isEnterprise, renderParams))
	if result.Get("harvest_error").String() != "<nil>" {
		return queueResult{Error: errors.New(result.Get("harvest_error").String())}
	}

	return queueResult{Token: result.String(), Error: nil}
}

func (e *Harvester) Destroy() {
	e.Done <- true
}

func (e *Harvester) setupHtml() {
	switch e.Type {
	case "v3":
		{
			e.Loader = constants.V3Loader
			e.HTML = constants.V3Html
			break
		}

	default:
		e.Type = "v3"
		e.Loader = constants.V3Loader
		e.HTML = constants.V3Html
		println(fmt.Sprintf("harvester type \"%s\" isn't supported defaulting to v3", e.Type))
		break
	}
}

func (e *Harvester) getJsCallString(siteKey string, isEnterprise bool, renderParams string) string {
	switch e.Type {
	case "v3":
		{
			return fmt.Sprintf(`document.harv.harvest("%s", %t, %s)`, siteKey, isEnterprise, renderParams)
		}
	}

	return "alert('Not supported');"
}

func (e *Harvester) clearQueue() {
	for {
		if e.Queue.Len() != 0 && e.Page != nil {
			firstElement := e.Queue.PopFront()
			if firstElement == nil {
				time.Sleep(250 * time.Millisecond)
				continue
			}

			parsedElement := firstElement.(queueEntry)
			if parsedElement.SiteKey == "" {
				time.Sleep(250 * time.Millisecond)
				continue
			}

			result := e.executeHarvest(parsedElement.SiteKey, parsedElement.IsEnterprise, parsedElement.RenderParams)
			parsedElement.Channel <- result
		} else {
			time.Sleep(250 * time.Millisecond)
			continue
		}
	}
}
