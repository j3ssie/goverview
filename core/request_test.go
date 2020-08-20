package core

import (
	"fmt"
	"testing"

	"github.com/j3ssie/goverview/libs"
	"github.com/j3ssie/goverview/utils"
)

func TestJustSend(t *testing.T) {
	var options libs.Options
	options.Level = 5
	url := "http://httpbin.org/anything?q=123&id=11"
	res, err := JustSend(options, url)
	fmt.Println(res.Beautify)
	fmt.Println(res.BeautifyHeader)
	if err != nil {
		utils.ErrorF("Error sending: %v", url)
	}
}
