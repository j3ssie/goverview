package core

import (
	"fmt"
	"github.com/j3ssie/goverview/libs"
	"testing"
)

func TestFingerprint(t *testing.T) {
	var opt libs.Options
	opt.Fin.TechFile = "/tmp/technologies.json"
	filename := "/tmp/uu"

	result := LocalFingerPrint(opt, filename)
	fmt.Println("finalTech --> ", result)

	if result == "" {
		t.Errorf("Error TestFingerprint")
	}
}
