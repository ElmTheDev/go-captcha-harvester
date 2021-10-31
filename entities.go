package main

import (
	"net/url"

	"github.com/go-rod/rod"
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