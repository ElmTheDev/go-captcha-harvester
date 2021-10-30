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

	IsReady bool
	IsSolving     bool
	SolvedCount   int
	Type string
	Url  *url.URL

	Done chan bool

	Browser *rod.Browser
	Page *rod.Page
}

// Initializes harvester and prepares it for solving
func (e *Harvester) Initialize() {
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

		ctx.Response.SetBody(constants.V3Html)
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
	page.MustEval(constants.V3Loader)
	
	e.IsReady = true
	<- e.Done
}