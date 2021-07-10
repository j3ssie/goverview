package core

import (
	"bytes"
	"fmt"
	"github.com/j3ssie/goverview/libs"
	"github.com/j3ssie/goverview/utils"
	jsoniter "github.com/json-iterator/go"
	"github.com/markbates/pkger"
	"html/template"
	"path"
	"strings"
)

type Content struct {
	ImgPath    string
	URL        string
	Title      string
	Tech       string
	ScreenPath string
	Header     string
	Status     string
	Length     string
	Checksum   string
}

type ReportData struct {
	Contents []Content
}

func RenderReport(options libs.Options) {
	var contents []Content

	if strings.TrimRight(options.Output, "/") == path.Dir(options.ReportFile) {
		options.AbsPath = false
	} else {
		options.AbsPath = true
	}

	// options.ContentFile
	if !utils.FileExists(options.ScreenShotFile) {
		utils.ErrorF("screenshot summary not found: %v", options.ScreenShotFile)
		return
	}

	utils.InforF("reading report file from: %v", options.ScreenShotFile)
	data := utils.ReadingLines(options.ScreenShotFile)
	for _, line := range data {
		var screen Screen
		err := jsoniter.UnmarshalFromString(line, &screen)
		if err == nil {
			header := "blank content"
			var length string
			if utils.FileExists(screen.ContentFile) {
				raw := utils.GetFileContent(screen.ContentFile)
				if strings.Contains(raw, "\n\n") {
					header = strings.Split(raw, "\n\n")[0]
				}

				if len(header) > 2000 {
					header = header[0:2000]
				}
				length = fmt.Sprintf("%d", len(raw))
			}

			if !options.AbsPath {
				screen.Image = strings.ReplaceAll(screen.Image, options.Output, "")
			}

			content := Content{
				Title:      screen.Title,
				Tech:       screen.Technologies,
				ScreenPath: screen.ContentFile,
				Checksum:   utils.GenHash(screen.Image),
				Status:     screen.Status,
				Header:     header,
				Length:     length,
				ImgPath:    screen.Image,
				URL:        screen.URL,
			}
			contents = append(contents, content)

		}

	}

	GenerateReport(options, contents)
	// options.ScreenShotFile
}

// GenerateReport generate report file
func GenerateReport(options libs.Options, contents []Content) error {
	if len(contents) == 0 {
		return fmt.Errorf("blank content")
	}

	data := struct {
		Contents    []Content
		CurrentDay  string
		Version     string
		ReportTitle string
	}{
		ReportTitle: "Goverview Report",
		Contents:    contents,
		CurrentDay:  utils.GetCurrentDay(),
		Version:     libs.VERSION,
	}

	f, err := pkger.Open("/static/index.html")
	if err != nil {
		return fmt.Errorf("blank template file")
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(f)
	tmpl := buf.String()

	// render template
	t := template.Must(template.New("").Parse(tmpl))
	tbuf := &bytes.Buffer{}
	err = t.Execute(tbuf, data)
	if err != nil {
		return err
	}
	result := tbuf.String()

	_, err = utils.WriteToFile(options.ReportFile, result)
	utils.GoodF("Writing report to: %v", options.ReportFile)
	return err
}
