package core

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"path"
	"sort"
	"strings"
)

func CalcCheckSum(options Options, url string) string {
	var result string
	title := "No-Title"
	hash := "No-CheckSum"
	contentFile := "No-Content"
	res, err := JustSend(options, url)
	DebugF("Headers: \n%v", res.BeautifyHeader)
	DebugF("Body: \n%v", res.Beautify)
	if err != nil && res.StatusCode == 0 {
		ErrorF("Error sending: %v", url)
		return fmt.Sprintf("%v ;; %v ;; %v ;; %v", url, title, hash, contentFile)
	}

	// store response
	content := res.BeautifyHeader
	if options.SaveReponse {
		content += "\n\n" + res.Body
	}
	if strings.TrimSpace(content) != "" {
		contentFile = fmt.Sprintf("%v.txt", strings.Replace(url, "://", "___", -1))
		contentFile = strings.Replace(contentFile, "?", "_", -1)
		contentFile = strings.Replace(contentFile, "/", "_", -1)
		content = fmt.Sprintf("> GET %v\n%v", url, content)
		contentFile = path.Join(options.ContentOutput, contentFile)
		DebugF("contentFile: %v", contentFile)
		_, err = WriteToFile(contentFile, content)
		if err != nil {
			ErrorF("WriteToFile: ", err)
			contentFile = "No-Content"
		}
	}

	// in case response is raw JSON
	result = GenHash(res.Body)
	if !strings.Contains(res.ContentType, "html") && !strings.Contains(res.ContentType, "xml") {
		if !strings.Contains(res.Body, "<html>") && !strings.Contains(res.Body, "<a>") {
			hash = GenHash(fmt.Sprintf("%v-%v", title, result))
			return fmt.Sprintf("%v ;; %v ;; %v ;; %v", url, title, GenHash(res.Body), contentFile)
		}
	}

	// parse body
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(res.Body))
	if err != nil {
		ErrorF("Error Parsing Body: %v", url)
		return fmt.Sprintf("%v ;; %v ;; %v ;; %v", url, title, GenHash(res.Body), contentFile)
	}
	title = GetTitle(doc)
	hash = GenHash(fmt.Sprintf("%v-%v", title, result))

	// wordlist builder
	BuildWordlists(options, url, doc)

	// calculate Hash based on level
	switch options.Level {
	case 0:
		result = ParseDocLevel0(options, doc)
	case 1:
		result = ParseDocLevel1(options, doc)
	case 2:
		result = ParseDocLevel2(options, doc)
	}
	if result != "" {
		hash = GenHash(fmt.Sprintf("%v-%v", title, result))
	}

	DebugF("Checksum-lv-%v: %v \n", options.Level, result)
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
func ParseDocLevel0(options Options, doc *goquery.Document) string {
	var result []string
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		if src != "" {
			result = append(result, src)
		}
	})

	if options.SortTag {
		sort.Strings(result)
	}
	return strings.Join(result, "-")
}

// ParseDocLevel1 calculate Hash based on src in scripts
func ParseDocLevel1(options Options, doc *goquery.Document) string {
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

	if options.SortTag {
		sort.Strings(result)
	}
	return strings.Join(result, "-")
}

// ParseDocLevel2 calculate Hash based on src in scripts
func ParseDocLevel2(options Options, doc *goquery.Document) string {
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
	if options.SortTag {
		sort.Strings(result)
	}
	return strings.Join(result, "-")
}
