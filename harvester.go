package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ElmTheDev/go-captcha-harvester/constants"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/phf/go-queue/queue"
)

// Initializes harvester and prepares it for solving
func (e *Harvester) Initialize() {
	if e.Proxy != "" {
		splitProxy := strings.Split(e.Proxy, ":")
		if len(splitProxy) == 4 {
			e.ParsedProxy = harvesterProxy{
				Ip: splitProxy[0],
				Port: splitProxy[1],
				Username: splitProxy[2],
				Password: splitProxy[3],
			}
		} else {
			e.ParsedProxy = harvesterProxy{
				Ip: splitProxy[0],
				Port: splitProxy[1],
			}
		}
	}

	e.Queue = queue.New().Init()

	e.setupHtml()
	e.Done = make(chan bool)

	// Start browser
	browserLauncher := launcher.New().
		Set("window-size", "200,700").
		Headless(false)

	if e.ParsedProxy.Ip != "" {
		browserLauncher.Proxy(fmt.Sprintf("%s:%s", e.ParsedProxy.Ip, e.ParsedProxy.Port))
	}

	url := browserLauncher.MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()
	

	if e.ParsedProxy.Username != "" && e.ParsedProxy.Password != "" {
		go browser.MustHandleAuth(e.ParsedProxy.Username, e.ParsedProxy.Password)()
	}

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
