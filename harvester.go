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

	// Avoid bot detection
	page := browser.MustPage()
	page.Emulate(fakeDevice)
	e.Browser = browser
	e.Page = page

	if e.Email != "" && e.Password != "" {
		e.Login()
	}

	// Load the page and prepare it for solving
	page.MustNavigate(e.Url.String())

	// Wait for page to load before continuing
	page.MustElement("title").MustText()

	// Initialize stuct

	// Load harvesting code into the page
	// TODO: Support multiple
	println(e.Loader)
	page.MustEval(e.Loader)

	e.IsReady = true
	go e.clearQueue()
	<-e.Done

	browser.MustClose()
}

// Enqueues captcha for solving
func (e *Harvester) Harvest(siteKey string, isInvisible bool, isEnterprise bool, renderParams map[string]string) (string, error) {
	// Ensures we keep track of captchas each harvester got to solve.
	e.SolvedCount++

	renderParamsObject, err := json.Marshal(renderParams)
	if err != nil {
		return "", err
	}

	resultChannel := make(chan queueResult)
	e.Queue.PushBack(queueEntry{
		SiteKey: siteKey,
		RenderParams: string(renderParamsObject),
		IsEnterprise: isEnterprise,
		IsInvisible: isInvisible,
		Channel: resultChannel,
	})

	resultParsed := <-resultChannel

	if resultParsed.Error != nil {
		return "", resultParsed.Error
	}

	return resultParsed.Token, nil
}

// Execute solving on browser
func (e *Harvester) executeHarvest(entry queueEntry) queueResult {
	result := e.Page.MustEval(e.getJsCallString(entry))
	if result.Get("harvest_error").String() != "<nil>" {
		return queueResult{Error: errors.New(result.Get("harvest_error").String())}
	}

	return queueResult{Token: result.String(), Error: nil}
}

// CLoses browser
func (e *Harvester) Destroy() {
	e.Done <- true
}

// Sets up initial values for captcha harvesting request interception
func (e *Harvester) setupHtml() {
	switch e.Type {
	case "v3":
		{
			e.Loader = constants.V3Loader
			e.HTML = constants.V3Html
			break
		}

	case "v2":
		{
			e.Loader = constants.V2Loader
			e.HTML = constants.V2Html
			break
		}	

	default:
		e.Type = "v3"
		e.Loader = constants.V3Loader
		e.HTML = constants.V3Html
		println(fmt.Sprintf("harvester type \"%s\" isn't supported defaulting to v3", e.Type))
	}
}

// Generates JavaScript snippet to pass to browser
func (e *Harvester) getJsCallString(entry queueEntry) string {
	switch e.Type {
	case "v3":
			return fmt.Sprintf(`document.harv.harvest("%s", %t, %s)`, entry.SiteKey, entry.IsEnterprise, entry.RenderParams)
	case "v2":
			return fmt.Sprintf(`document.harv.harvest("%s", %t, %s)`, entry.SiteKey, entry.IsInvisible, entry.RenderParams)
	}

	return "alert('Not supported');"
}

// Loops through queue solving captchas one by one
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

			result := e.executeHarvest(parsedElement)
			parsedElement.Channel <- result
		} else {
			time.Sleep(250 * time.Millisecond)
			continue
		}
	}
}

func (e *Harvester) Login() {
	e.Page.MustNavigate("https://accounts.google.com/signin")
	time.Sleep(1 * time.Second)
	emailElement := e.Page.MustElement("[type=\"email\"]")
	time.Sleep(1 * time.Second)
	emailElement.MustInput(e.Email)
	time.Sleep(1 * time.Second)
	e.Page.MustElement("[type=\"button\"][jscontroller]").MustClick()
	time.Sleep(1 * time.Second)
	passwordElement := e.Page.MustElement("[type=\"password\"]")
	time.Sleep(1 * time.Second)
	passwordElement.MustInput(e.Password)
	time.Sleep(1 * time.Second)
	e.Page.MustElement("[type=\"button\"][jscontroller]").MustClick()

	e.Page.MustElement("[href*=\"SignOutOptions\"]")
}
