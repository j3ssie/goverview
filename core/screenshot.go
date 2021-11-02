package core

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"net/url"

	"github.com/chromedp/chromedp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/j3ssie/goverview/libs"
	"github.com/j3ssie/goverview/utils"
	jsoniter "github.com/json-iterator/go"

	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Screen overview struct
type Screen struct {
	URL          string `json:"url"`
	Image        string `json:"image"`
	ContentFile  string `json:"content_file"`
	Technologies string `json:"tech"`
	// with check sum
	Title    string `json:"title"`
	CheckSum string `json:"checksum"`
	Status   string `json:"status"`
	//External []string `json:"external"`
}

// PrintScreen print probe string
func PrintScreen(options libs.Options, screen Screen) string {
	if screen.URL == "" || screen.Image == "" {
		return ""
	}
	if options.AbsPath {
		screen.Image = path.Base(screen.Image)
	}

	if options.JsonOutput {
		if data, err := jsoniter.MarshalToString(screen); err == nil {
			return data
		}
	}
	return fmt.Sprintf("%v ;; %v", screen.URL, screen.Image)
}

func DoScreenshot(options libs.Options, raw string) string {
	imageName := strings.Replace(raw, "://", "___", -1)
	imageScreen := path.Join(options.Screen.ScreenOutput, fmt.Sprintf("%v.png", strings.Replace(imageName, "/", "_", -1)))

	contentFile := fmt.Sprintf("%s.txt", strings.Replace(raw, "://", "___", -1))
	contentFile = strings.Replace(contentFile, "?", "_", -1)
	contentFile = strings.Replace(contentFile, "/", "_", -1)
	contentFile = path.Join(options.Screen.ScreenOutput, contentFile)
	content := fmt.Sprintf("> GET %s\n", raw)

	screen := Screen{
		URL:         raw,
		ContentFile: contentFile,
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
	)

	if options.Proxy != "" {
		opts = append(opts, chromedp.ProxyServer(options.Proxy))
	}

	// create context
	allocCtx, bcancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer bcancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	ctx, cancel = context.WithTimeout(ctx, time.Duration(options.Screen.ScreenTimeout)*time.Second)
	defer cancel()

	// capture screenshot of an element
	var buf []byte
	var res libs.Response

	err := chromedp.Run(ctx,
		fullScreenshot(ctx, options, raw, 90, &buf, &res),
		fetch.Enable().WithPatterns([]*fetch.RequestPattern{{RequestStage: fetch.RequestStageResponse}}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				utils.ErrorF("Error get content: %v", err)
				return err
			}
			res.Body, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	)

	// clean chromedp-runner folder
	cleanUp()
	if err != nil {
		utils.ErrorF("screen err: %v - %v", raw, err)
		return PrintScreen(options, screen)
	}

	if res.StatusCode != 0 || len(res.Body) > 0 {
		content = fmt.Sprintf("< HTTP/1.1 %v\n", res.Status)
		for _, head := range res.Headers {
			for k, v := range head {
				content += fmt.Sprintf("< %s: %s\n", k, v)
			}
		}

		content += "\n\n"
		content += res.Body

		if options.Output != "" {
			WriteToFile(contentFile, content)
		}

		if options.Fin.Loaded {
			techs := LocalFingerPrint(options, contentFile)
			screen.Technologies = techs
		}
	}

	// write image
	if options.Output != "" {
		if err := ioutil.WriteFile(imageScreen, buf, 0644); err != nil {
			utils.ErrorF("write screen err: %v - %v", raw, err)
			return PrintScreen(options, screen)
		}
	}

	screen.Image = imageScreen
	screen.Status = res.Status
	overview := CalcCheckSum(options, raw, res)
	screen.Title = overview.Title
	screen.CheckSum = overview.CheckSum
	return PrintScreen(options, screen)
}

// fullScreenshot takes a screenshot of the entire browser viewport.
// Liberally copied from puppeteer's source.
// Note: this will override the viewport emulation settings.
func fullScreenshot(chromeContext context.Context, options libs.Options, urlstr string, quality int64, imgContent *[]byte, res *libs.Response) chromedp.Tasks {
	// setup a listener for events
	//var requestHeaders map[string]interface{}
	var uu string

	chromedp.ListenTarget(chromeContext, func(event interface{}) {
		// get which type of event it is
		switch msg := event.(type) {
		// just before request sent
		case *network.EventRequestWillBeSent:
			request := msg.Request
			// see if we have been redirected
			// if so, change the URL that we are tracking
			if msg.RedirectResponse != nil {
				uu = request.URL
			}

		// once we have the full response
		case *network.EventResponseReceived:
			response := msg.Response
			// is the request we want the status/headers on?
			if response.URL == uu {
				res.StatusCode = int(response.Status)
				res.Status = response.StatusText
				//fmt.Printf(" status code: %d\n", res.StatusCode)
				for k, v := range response.Headers {
					header := make(map[string]string)
					// fmt.Println(k, v)
					header[k] = v.(string)
					res.Headers = append(res.Headers, header)
				}
			}
		}
	})

	//var imageContent *[]byte
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(imgContent, int(quality)),
		network.Enable(),
	}
}

func cleanUp() {
	tmpFolder := path.Join(os.TempDir(), "chromedp-runner*")
	if _, err := os.Stat("/tmp/"); !os.IsNotExist(err) {
		tmpFolder = path.Join("/tmp/", "chromedp-runner*")
	}
	junks, err := filepath.Glob(tmpFolder)
	if err != nil {
		return
	}
	for _, junk := range junks {
		os.RemoveAll(junk)
	}
}

/* Start using new lib */

// NewDoScreenshot new do screenshot based on rod
func NewDoScreenshot(options libs.Options, raw string) string {
	_, err := url.ParseRequestURI(raw)
	if err != nil {
		utils.ErrorF("invalid input: %v", raw)
		return ""
	}

	imageName := strings.Replace(raw, "://", "___", -1)
	imageScreen := path.Join(options.Screen.ScreenOutput, fmt.Sprintf("%v.png", strings.Replace(imageName, "/", "_", -1)))

	contentFile := fmt.Sprintf("%s.txt", strings.Replace(raw, "://", "___", -1))
	contentFile = strings.Replace(contentFile, "?", "_", -1)
	contentFile = strings.Replace(contentFile, "/", "_", -1)
	contentFile = path.Join(options.Screen.ScreenOutput, contentFile)
	content := fmt.Sprintf("> GET %s\n", raw)

	screen := Screen{
		URL:         raw,
		ContentFile: contentFile,
	}
	if options.Screen.ImgWidth == 0 {
		options.Screen.ImgWidth = 2500
	}
	if options.Screen.ImgHeight == 0 {
		options.Screen.ImgHeight = 1400
	}

	browser := rod.New().MustConnect().MustIgnoreCertErrors(true).MustPage("")
	err = rod.Try(func() {
		browser.MustNavigate(raw)

		browser.Timeout(time.Duration(options.Screen.ScreenTimeout) * time.Second)
		browser.MustWaitLoad()

		go browser.EachEvent(func(e *proto.NetworkResponseReceived) {
			// only get event match base URL
			if strings.HasPrefix(e.Response.URL, raw) {
				screen.Status = e.Response.StatusText
				content += fmt.Sprintf("< HTTP/1.1 %v", e.Response.StatusText)
				for k, v := range e.Response.Headers {
					content += fmt.Sprintf("< %s: %s\n", k, v)
				}
			}

		})()

	})
	if err != nil {
		utils.ErrorF("error screenshot")
		return PrintScreen(options, screen)
	}

	//browser.MustNavigate(raw)

	//browser := rod.New().MustConnect().MustIgnoreCertErrors(true).MustPage(raw)
	//browser := rod.New().MustConnect().MustIgnoreCertErrors(true).MustPage(raw)
	//defer browser.MustClose()
	//
	//err = browser.WaitLoad()
	//if err != nil {
	//	utils.ErrorF("error screenshot")
	//	return PrintScreen(options, screen)
	//}
	//browser.Timeout(time.Duration(options.Screen.ScreenTimeout) * time.Second)

	// get headers here

	// capture entire browser viewport, returning jpg with quality=90
	buf, err := browser.Screenshot(true, &proto.PageCaptureScreenshot{
		Format:  proto.PageCaptureScreenshotFormatJpeg,
		Quality: 90,
		//Clip: &proto.PageViewport{
		//	X:      0,
		//	Y:      0,
		//	Width:  float64(options.Screen.ImgHeight),
		//	Height: float64(options.Screen.ImgHeight),
		//	Scale:  1,
		//},
		FromSurface: true,
	})

	// store HTML data too in case we miss with probing
	html := browser.MustElement("html").MustHTML()
	content += html
	_, err = WriteToFile(contentFile, content)
	if options.Fin.Loaded {
		techs := LocalFingerPrint(options, contentFile)
		screen.Technologies = techs
	}

	if err != nil {
		utils.ErrorF("write screen err: %v - %v", raw, err)
		return PrintScreen(options, screen)
	}
	err = ioutil.WriteFile(imageScreen, buf, 0644)

	// write image
	if err != nil {
		utils.ErrorF("write screen err: %v - %v", raw, err)
		return PrintScreen(options, screen)
	}
	screen.Image = imageScreen
	return PrintScreen(options, screen)
}
