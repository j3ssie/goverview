package core

import (
	"github.com/j3ssie/goverview/libs"
	"testing"
)

func TestRodScreenshot(t *testing.T) {
	var opt libs.Options
	url := "http://httpbin.org/anything?q=123&id=11"
	result := DoScreenshot(opt, url)
	if result == "" {
		t.Errorf("Error RodScreenshot")
	}
}
