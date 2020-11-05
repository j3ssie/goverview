package core

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	rodutils "github.com/go-rod/rod/lib/utils"
	"github.com/j3ssie/goverview/libs"
	"github.com/j3ssie/goverview/utils"
	jsoniter "github.com/json-iterator/go"
	"net/url"

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
	URL   string `json:"url"`
	Image string `json:"image"`
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
	_, err := url.Parse(raw)
	if err != nil {
		utils.ErrorF("invalid input: %v", raw)
		return ""
	}

	imageName := strings.Replace(raw, "://", "___", -1)
	imageScreen := path.Join(options.Screen.ScreenOutput, fmt.Sprintf("%v.png", strings.Replace(imageName, "/", "_", -1)))
	screen := Screen{
		URL: raw,
	}
	if options.Screen.ImgWidth == 0.0 {
		 options.Screen.ImgWidth = 1440
	}
		if options.Screen.ImgWidth == 0.0 {
		 options.Screen.ImgHeight = 900
	}
	//if options.Debug {
	//	isHeadless = false
	//}
	//
	//
	//rodChrome := rod.New().ControlURL(raw).MustConnect()
	//defer rodChrome.MustClose()
	//rodChrome.Timeout(time.Duration(options.Screen.ScreenTimeout)*time.Second)
	//rodChrome.MustIgnoreCertErrors(true)

	//browser.MustNavigate(raw)
	rodChrome := rod.New().MustConnect().MustPage(raw)
	defer rodChrome.MustClose()

	rodChrome.Timeout(time.Duration(options.Screen.ScreenTimeout)*time.Second)
	wait := rodChrome.MustWaitNavigation()
	utils.DebugF("Doing screenshot on: %v", raw)

	//rodChrome.MustNavigate(raw)
	rodChrome.MustNavigate(raw)
	//if err != nil {
	//	utils.ErrorF("error input: %v", raw)
	//	return ""
	//}
	wait() // until the navigation to settle down

	// simple version
	rodChrome.MustScreenshot(imageScreen)

	// customization version
	img, _ := rodChrome.Screenshot(true, &proto.PageCaptureScreenshot{
		Format:  proto.PageCaptureScreenshotFormatJpeg,
		Quality: 90,
		Clip: &proto.PageViewport{
			X:      0,
			Y:      0,
			Width:  float64(options.Screen.ImgHeight),
			Height: float64(options.Screen.ImgHeight),
			Scale:  1,
		},
		FromSurface: true,
	})
	err = rodutils.OutputFile(imageScreen, img)

	// write image
	if err != nil {
		utils.ErrorF("write screen err: %v - %v", raw, err)
		return PrintScreen(options, screen)
	}
	screen.Image = imageScreen
	return PrintScreen(options, screen)
}
