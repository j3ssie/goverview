package core

import (
	"fmt"
	"testing"
)

func TestJustSend(t *testing.T) {
	var options Options
	options.Level = 5
	url := "http://httpbin.org/anything?q=123&id=11"
	res, err := JustSend(options, url)
	fmt.Println(res.Beautify)
	fmt.Println(res.BeautifyHeader)
	if err != nil {
		ErrorF("Error sending: %v", url)
	}
}
