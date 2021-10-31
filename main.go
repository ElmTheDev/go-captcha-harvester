package main

import (
	"net/url"
	"time"
)

func main() {
	u, _ := url.Parse("https://www.yeezysupply.com/product/EG6463")
	harv := Harvester{
		CustomName: "my harvester",

		Url:  u,
		Type: "v3",
	}

	go harv.Initialize()
	println("hi")

	queueResult, err := harv.Harvest("6Lf34M8ZAAAAANgE72rhfideXH21Lab333mdd2d-", true, map[string]string{})
	if err != nil {
		panic(err)
	}

	println("got token",  queueResult)
	time.Sleep(15 * time.Second)
	harv.Destroy()
	println("destroyed")
	time.Sleep(5 * time.Second)
}
