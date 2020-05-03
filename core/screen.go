package core

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

func DoScreenshot(options Options, raw string) string {
	imageName := strings.Replace(raw, "://", "___", -1)
	imageScreen := path.Join(options.ScreenOutput, fmt.Sprintf("%v.png", strings.Replace(imageName, "/", "_", -1)))

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
	// create context
	allocCtx, bcancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer bcancel()
	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	ctx, cancel = context.WithTimeout(ctx, time.Duration(options.Timeout)*time.Second)
	defer cancel()

	// capture screenshot of an element
	var buf []byte
	err := chromedp.Run(ctx, fullScreenshot(options, raw, 90, &buf))
	// clean chromedp-runner folder
	cleanUp()
	if err != nil {
		ErrorF("screen err: %v", raw)
		return ""
	}

	// write image
	if err := ioutil.WriteFile(imageScreen, buf, 0644); err != nil {
		ErrorF("screen err: %v", raw)
	}
	if options.AbsPath {
		return imageScreen
	}
	return path.Base(imageScreen)
}

// fullScreenshot takes a screenshot of the entire browser viewport.
// Liberally copied from puppeteer's source.
// Note: this will override the viewport emulation settings.
func fullScreenshot(options Options, urlstr string, quality int64, res *[]byte) chromedp.Tasks {
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
			if options.ImgWidth != 0 && options.ImgHeight != 0 {
				width = int64(options.ImgWidth)
				height = int64(options.ImgHeight)
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
