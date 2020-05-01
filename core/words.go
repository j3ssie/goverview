package core

import (
	"github.com/PuerkitoBio/goquery"
	"github.com/j3ssie/metabigor/core"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

// CleanWords clean wordlists
func CleanWords(filename string) {
	var cleaned []string
	words := core.ReadingFileUnique(filename)
	if len(words) <= 0 {
		return
	}
	var IsLetter = regexp.MustCompile(`^[a-zA-Z_0-9\[\]]+$`).MatchString

	for _, word := range words {
		if strings.Contains(word, ".") {
			continue
		}
		if !IsLetter(word) {
			continue
		}
		cleaned = append(cleaned, word)
	}
	sort.Sort(sort.StringSlice(cleaned))
	core.WriteToFile(filename, strings.Join(cleaned, "\n"))
}

// BuildWordlists based on HTML content
func BuildWordlists(options Options, doc *goquery.Document) {
	var result []string
	result = append(result, ParseID(doc)...)
	result = append(result, ParseLinks(doc)...)
	result = append(result, ParseInput(doc)...)
	if len(result) <= 0 {
		return
	}
	content := strings.Join(result, "\n")
	AppendTo(options.WordList, content)
}

// ParseInput parse input tag
func ParseInput(doc *goquery.Document) []string {
	var result []string
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("name")
		if src != "" {
			result = append(result, src)
		}
	})
	return result
}

// ParseLinks get words from link urls
func ParseLinks(doc *goquery.Document) []string {
	var result, links []string
	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		tag := goquery.NodeName(s)
		//result = append(result, tag)
		if tag == "script" || tag == "img" {
			src, _ := s.Attr("src")
			if src != "" {
				links = append(links, src)
			}
		}

		if tag == "a" {
			src, _ := s.Attr("href")
			if src != "" {
				links = append(links, src)
			}
		}
	})

	// parse all links found
	for _, link := range links {
		u, err := url.Parse(link)
		if err == nil {
			link = u.Path
			for k := range u.Query() {
				result = append(result, k)
			}
		}

		if strings.Contains(link, "/") {
			items := strings.Split(link, "/")
			for _, item := range items {
				result = append(result, strings.TrimSpace(item))
			}
		}
	}

	return result
}

// ParseID parse id attr
func ParseID(doc *goquery.Document) []string {
	var result []string

	doc.Find("*").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		if id != "" {
			result = append(result, id)
		}
	})
	return result
}
