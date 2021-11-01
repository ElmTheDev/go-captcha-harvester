package main

import (
	"net/url"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
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

	Queue       *queue.Queue
	ParsedProxy harvesterProxy
}

type queueEntry struct {
	SiteKey      string
	IsEnterprise bool
	IsInvisible bool
	RenderParams string

	Channel chan queueResult
}

type queueResult struct {
	Error error
	Token string
}

type harvesterProxy struct {
	Ip       string
	Port     string
	Username string
	Password string
}

var fakeDevice = devices.Device{
	Title:          "iPhone X",
	Capabilities:   []string{"touch", "mobile"},
	UserAgent:      "Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1",
	AcceptLanguage: "en",
	Screen: devices.Screen{
		DevicePixelRatio: 1,
		Horizontal: devices.ScreenSize{
			Width:  1280,
			Height: 720,
		},
		Vertical: devices.ScreenSize{
			Width:  720,
			Height: 1280,
		},
	},
}