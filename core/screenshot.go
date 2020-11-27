package core

import (
	"context"
	"fmt"
	"net/url"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/j3ssie/goverview/libs"
	"github.com/j3ssie/goverview/utils"
	jsoniter "github.com/json-iterator/go"

	"io/ioutil"
	"log"
	"math"
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

	screen := Screen{
		URL: raw,
	}

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("no-first-run", true),
		chromedp.Flag("no-default-browser-check", true),
		chromedp.Flag("single-process", true),
		chromedp.Flag("no-zygote", true),
	)

	// create context
	allocCtx, bcancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer bcancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	ctx, cancel = context.WithTimeout(ctx, time.Duration(options.Screen.ScreenTimeout)*time.Second)
	defer cancel()

	// capture screenshot of an element
	var buf []byte
	err := chromedp.Run(ctx, fullScreenshot(options, raw, 90, &buf))
	// clean chromedp-runner folder
	cleanUp()
	if err != nil {
		utils.ErrorF("screen err: %v - %v", raw, err)
		return PrintScreen(options, screen)
	}

	// write image
	if err := ioutil.WriteFile(imageScreen, buf, 0644); err != nil {
		utils.ErrorF("write screen err: %v - %v", raw, err)
		return PrintScreen(options, screen)
	}
	screen.Image = imageScreen
	return PrintScreen(options, screen)
}

// fullScreenshot takes a screenshot of the entire browser viewport.
// Liberally copied from puppeteer's source.
// Note: this will override the viewport emulation settings.
func fullScreenshot(options libs.Options, urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))
			//imgWidth       int
			//imgHeight      int
			if options.Screen.ImgWidth != 0 && options.Screen.ImgHeight != 0 {
				width = int64(options.Screen.ImgWidth)
				height = int64(options.Screen.ImgHeight)
			}

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  float64(width),
					Height: float64(height),
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
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

	page := rod.New().MustConnect().MustIgnoreCertErrors(true).MustPage("")
	ctx, cancel := context.WithCancel(context.Background())
	browser := page.Context(ctx)

	go func() {
		time.Sleep(time.Duration(options.Screen.ScreenTimeout) * time.Second)
		cancel()
	}()
	err = rod.Try(func() {
		browser.MustNavigate(raw)
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
	go browser.EachEvent(func(e *proto.NetworkResponseReceived) {
		// only get event match base URL
		if strings.HasPrefix(e.Response.URL, raw) {
			//if e.Response.URL == raw {
			screen.Status = e.Response.StatusText
			for k, v := range e.Response.Headers {
				content += fmt.Sprintf("< %s: %s\n", k, v)
			}
		}
		//
		//spew.Dump(e.Response.RequestHeaders)
		//fmt.Println("Status: ", e.Response.Status, e.Response.URL, e.Response.Headers)
		//spew.Dump(e.Response)
	})()

	err = rod.Try(func() {
		browser.MustWaitLoad()
	})
	if err != nil {
		utils.ErrorF("error screenshot")
		return PrintScreen(options, screen)
	}

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
	techs := LocalFingerPrint(options, contentFile)
	screen.Technologies = techs

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
