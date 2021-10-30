package main

import "net/url"

func main() {
	u, _ := url.Parse("https://www.yeezysupply.com/product/EG6463")
	harv := Harvester{
		CustomName: "my harvester",

		Url: u,
		Type: "v3",
	}

	go harv.Initialize()
	println("hi")
}