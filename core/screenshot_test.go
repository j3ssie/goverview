package core

import (
	"fmt"
	"github.com/j3ssie/goverview/libs"
	"testing"
)

func TestRodScreenshot(t *testing.T) {
	var opt libs.Options

	opt.ScreenOutput = "/tmp/"
	url := "https://github.com"
	result := NewDoScreenshot(opt, url)
	fmt.Println("Screen: ", url, "--", result)
	if result == "" {
		t.Errorf("Error RodScreenshot")
	}

	url = "https://35.184.252.145/"
	result = NewDoScreenshot(opt, url)
	fmt.Println("Screen: ", url, "--", result)
	if result == "" {
		t.Errorf("Error RodScreenshot")
	}
}
