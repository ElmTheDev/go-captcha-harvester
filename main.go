package main

import (
	"fmt"
	"net/http"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"

	"github.com/ElmTheDev/go-captcha-harvester/constants"
)

func main() {
	url := launcher.New().
	//	Proxy("127.0.0.1:8080").     // set flag "--proxy-server=127.0.0.1:8080"
	//	Delete("use-mock-keychain"). // delete flag "--use-mock-keychain"
	Set("window-size", "200,650").
	Headless(false).
		MustLaunch()

	browser := rod.New().ControlURL(url).MustConnect()
	defer browser.MustClose()


	browser.DefaultDevice(devices.Device{
		Title:          "iPhone 4",
		Capabilities:   devices.LaptopWithTouch.Capabilities,
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/94.0.4606.81 Safari/537.36",
		AcceptLanguage: "en",
		Screen: devices.Screen{
		  DevicePixelRatio: 2,
		  Horizontal: devices.ScreenSize{
			Width:  600,
			Height: 400,
		  },
		  Vertical: devices.ScreenSize{
			Width:  400,
			Height: 600,
		  },
		},
	  })

	router := browser.HijackRequests()
	defer router.MustStop()

	router.MustAdd("*yeezysupply.com*", func(ctx *rod.Hijack) {
		_ = ctx.LoadResponse(http.DefaultClient, true)

		ctx.Response.SetBody(constants.V3Html)
	})

	go router.Run()

	page := browser.MustPage("https://www.yeezysupply.com/product/EG6463")

	page.MustElement("title").MustText() // Wait for page to load

	fmt.Println(page.MustEval(constants.V3Loader))
	fmt.Println(page.MustEval(`document.harv.harvest('6Lf34M8ZAAAAANgE72rhfideXH21Lab333mdd2d-')`))
	for{}
}