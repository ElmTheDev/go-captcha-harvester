# go-captcha-harvester

GoLang based captcha solver

## Example

```
package main

import (
	"net/url"
	"time"
)

func main() {
	u, _ := url.Parse("https://patrickhlauke.github.io/recaptcha/")
	harv := Harvester{
		CustomName: "my harvester",

		Url:  u,
		Type: "v2",
	}

	go harv.Initialize()

	queueResult, err := harv.Harvest("6Ld2sf4SAAAAAKSgzs0Q13IZhY02Pyo31S2jgOB5", true, false, map[string]string{})
	if err != nil {
		panic(err)
	}

	println("got token",  queueResult)
	time.Sleep(15 * time.Second)
	harv.Destroy()
	println("destroyed")
	time.Sleep(5 * time.Second)
}
```
