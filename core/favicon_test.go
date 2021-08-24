package core

import (
	"fmt"
	"testing"
)

func TestGenFavHash(t *testing.T) {
	data := GetFavHash("https://1.1.1.212/favicon.ico")
	fmt.Println(data)
}
