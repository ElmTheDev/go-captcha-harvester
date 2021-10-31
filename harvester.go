package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ElmTheDev/go-captcha-harvester/constants"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

type Harvester struct {
	CustomName string

	Email    string
	Password string
	Proxy string

	IsReady bool
	IsSolving     bool
	SolvedCount   int
	Type string
	Url  *url.URL

	Done chan bool

	Browser *rod.Browser
	Page *rod.Page

	Loader string
	HTML string
}

// Initializes harvester and prepares it for solving
func (e *Harvester) Initialize() {
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
	<- e.Done

	browser.MustClose()
}

func (e *Harvester) Destroy() {
	e.Done <- true
}

func (e *Harvester) setupHtml() {
	switch e.Type {
	case "v3": {
		e.Loader = constants.V3Loader
		e.HTML = constants.V3Html
		break
	}

	default:
		println(fmt.Sprintf("harvester type \"%s\" isn't supported defaulting to v3", e.Type))
		break
	}
}