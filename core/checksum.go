package core

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"path"
	"strings"
)

func CalcCheckSum(options Options, url string) string {
	var result string
	title := "No-Title"
	hash := "No-CheckSum"
	contentFile := "No-Content"
	res, err := JustSend(options, url)
	DebugF("Headers: ", res.BeautifyHeader)
	DebugF("Body: ", res.Beautify)
	if err != nil && res.StatusCode == 0 {
		ErrorF("Error sending: %v", url)
		return fmt.Sprintf("%v ;; %v ;; %v ;; %v", url, title, hash, contentFile)
	}

	// store response
	content := res.BeautifyHeader
	if options.SaveReponse {
		content += "\n" + res.Beautify
	}
	if strings.TrimSpace(content) != "" {
		contentFile = fmt.Sprintf("%v.txt", strings.Replace(url, "://", "___", -1))
		content = fmt.Sprintf("> GET %v\n%v", url, content)
		WriteToFile(path.Join(options.ContentOutput, contentFile), content)
	}

	// parse body
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Body))
	if err != nil {
		ErrorF("Error Parsing Body: %v", url)
		return fmt.Sprintf("%v ;; %v ;; %v", url, title, GenHash(res.Body))
	}
	result = GenHash(res.Body)
	title = GetTitle(doc)

	// wordlist builder
	BuildWordlists(options, doc)

	// calculate Hash based on level
	switch options.Level {
	case 0:
		result = ParseDocLevel0(doc)
	case 1:
		result = ParseDocLevel1(doc)
	case 2:
		result = ParseDocLevel2(doc)
	}

	if result != "" {
		hash = GenHash(res.Body)
	}

	return fmt.Sprintf("%v ;; %v ;; %v ;; %v", url, title, hash, contentFile)
}

// GetTitle get title of response
func GetTitle(doc *goquery.Document) string {
	var title string
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		title = s.Text()
	})
	if title == "" {
		title = "Blank Title"
	}
	return title
}

// ParseDocLevel0 calculate Hash based on src in scripts
func ParseDocLevel0(doc *goquery.Document) string {
	var result []string
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		if src != "" {
			result = append(result, src)
		}
	})

	DebugF("Checksum-lv-0: %v \n", strings.Join(result, "-"))
	return strings.Join(result, "-")
}

// ParseDocLevel1 calculate Hash based on src in scripts
func ParseDocLevel1(doc *goquery.Document) string {
	var result []string
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		result = append(result, tag)
		if tag == "script" {
			src, _ := s.Attr("src")
			if src != "" {
				result = append(result, src)
			}
		}
	})

	DebugF("Checksum-lv-1: %v \n", strings.Join(result, "-"))
	return strings.Join(result, "-")
}

// ParseDocLevel2 calculate Hash based on src in scripts
func ParseDocLevel2(doc *goquery.Document) string {
	var result []string
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		result = append(result, tag)
		if tag == "script" || tag == "img" {
			src, _ := s.Attr("src")
			if src != "" {
				result = append(result, src)
			}
		}

		if tag == "a" {
			src, _ := s.Attr("href")
			if src != "" {
				result = append(result, src)
			}
		}
	})

	DebugF("Checksum-lv-2: %v \n", strings.Join(result, "-"))
	return strings.Join(result, "-")
}
