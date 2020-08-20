package core

import (
	"fmt"
	"testing"

	"github.com/j3ssie/goverview/libs"
)

func TestCalcCheckSum(t *testing.T) {
	var options libs.Options
	options.Level = 5
	url := "http://httpbin.org/anything?q=123&id=11"
	result := CalcCheckSum(options, url)
	fmt.Printf("Level %v ->  Hash: %v \n", options.Level, result)
	if result == "" {
		t.Errorf("Error CalcCheckSum")
	}

}

func TestCalcCheckSum0(t *testing.T) {
	var options libs.Options
	// level 0
	options.Level = 0
	url := "https://www.wappalyzer.com"
	result := CalcCheckSum(options, url)
	fmt.Printf("Level %v ->  Hash: %v \n", options.Level, result)
	if result == "" {
		t.Errorf("Error CalcCheckSum")
	}
}

func TestCalcCheckSum1(t *testing.T) {
	var options libs.Options
	// level 0
	options.Level = 1
	url := "https://www.wappalyzer.com"
	result := CalcCheckSum(options, url)
	fmt.Printf("Level %v ->  Hash: %v \n", options.Level, result)
	if result == "" {
		t.Errorf("Error CalcCheckSum")
	}
}

//
//func TestParseHTMLStructure(t *testing.T) {
//	// level 0
//	body := `
//<!doctype html>
//<html>
//  <head>
//    <title>Wappalyzer</title><meta data-n-head="1" charset="utf-8"><meta data-n-head="1" theme_color="#4608ad"><meta data-n-head="1" name="viewport" content="width=device-width,initial-scale=1"><script data-n-head="1" src="//js.stripe.com/v3/" defer async></script><link rel="preload" href="/_nuxt/559ffb9abf992bea00c8.js" as="script"><link rel="preload" href="/_nuxt/44700d7812058a7d0b81.js" as="script"><link rel="preload" href="/_nuxt/566931ec9823abd4e2ec.js" as="script"><link rel="preload" href="/_nuxt/4f6173bf628715241cac.js" as="script">
//  </head>
//  <body>
//    <div id="__nuxt"><style>#nuxt-loading{visibility:hidden;opacity:0;position:absolute;left:0;right:0;top:0;bottom:0;display:flex;justify-content:center;align-items:center;flex-direction:column;animation:nuxtLoadingIn 10s ease;-webkit-animation:nuxtLoadingIn 10s ease;animation-fill-mode:forwards;overflow:hidden}@keyframes nuxtLoadingIn{0%{visibility:hidden;opacity:0}20%{visibility:visible;opacity:0}100%{visibility:visible;opacity:1}}@-webkit-keyframes nuxtLoadingIn{0%{visibility:hidden;opacity:0}20%{visibility:visible;opacity:0}100%{visibility:visible;opacity:1}}#nuxt-loading>div,#nuxt-loading>div:after{border-radius:50%;width:5rem;height:5rem}#nuxt-loading>div{font-size:10px;position:relative;text-indent:-9999em;border:.5rem solid #f5f5f5;border-left:.5rem solid #fff;-webkit-transform:translateZ(0);-ms-transform:translateZ(0);transform:translateZ(0);-webkit-animation:nuxtLoading 1.1s infinite linear;animation:nuxtLoading 1.1s infinite linear}#nuxt-loading.error>div{border-left:.5rem solid #ff4500;animation-duration:5s}@-webkit-keyframes nuxtLoading{0%{-webkit-transform:rotate(0);transform:rotate(0)}100%{-webkit-transform:rotate(360deg);transform:rotate(360deg)}}@keyframes nuxtLoading{0%{-webkit-transform:rotate(0);transform:rotate(0)}100%{-webkit-transform:rotate(360deg);transform:rotate(360deg)}}</style><script>window.addEventListener("error",function(){var e=document.getElementById("nuxt-loading");e&&(e.className+=" error")})</script><div id="nuxt-loading" aria-live="polite" role="status"><div>Loading...</div></div></div>
//  <script type="text/javascript" src="/_nuxt/559ffb9abf992bea00c8.js"></script><script type="text/javascript" src="/_nuxt/44700d7812058a7d0b81.js"></script><script type="text/javascript" src="/_nuxt/566931ec9823abd4e2ec.js"></script><script type="text/javascript" src="/_nuxt/4f6173bf628715241cac.js"></script></body>
//</html>
//`
//
//	//body = `echo 'yahoo.com' |metabigor net -x --org`
//	result := ParseHTMLStructure(body)
//	fmt.Printf("HTML Parse ->  Hash: %v \n", result)
//	if result == "" {
//		t.Errorf("Error CalcCheckSum")
//	}
//}
