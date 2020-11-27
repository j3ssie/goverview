package core

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/emirpasic/gods/maps/treemap"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/extensions"
	"github.com/j3ssie/goverview/libs"
	"github.com/j3ssie/goverview/utils"
)

// Result type encapsulates the result information from a given host
type Result struct {
	Host    string  `json:"host"`
	Matches []Match `json:"matches"`
}

type SecretRegex struct {
	R      *regexp.Regexp
	Reason string `json:"Reason"`
	Rule   string `json:"Rule"`
}

func (s *SecretRegex) SetRegex(regex string) {
	s.R = regexp.MustCompile(regex)
}

type SecretResult struct {
	Url     string
	Matches string
}

// Match type encapsulates the App information from a match on a document
type Match struct {
	App     `json:"app"`
	AppName string     `json:"app_name"`
	Matches [][]string `json:"matches"`
	Version string     `json:"version"`
}

func (m *Match) updateVersion(version string) {
	if version != "" {
		m.Version = version
	}
}

// LoadTechs
func LoadTechs(options libs.Options) error {
	WA = new(WebAnalyzer)
	if err := WA.LoadApps(options.Fin.TechFile); err != nil {
		utils.ErrorF("Technology file not found: %s", options.Fin.TechFile)
		return err
	}
	utils.DebugF("Loaded %v of tech fingerprint", len(WA.AppDefs.Apps))
	return nil
}

// LocalFingerPrint do fingerprint but from local file
func LocalFingerPrint(options libs.Options, filename string) string {
	utils.DebugF("Fingerprint tech from: %s", filename)

	if !utils.FileExists(filename) {
		utils.ErrorF("content file not found: %s", options.Fin.TechFile)
		return ""
	}

	if !options.Fin.Loaded {
		utils.ErrorF("error loading technology from: %s", options.Fin.TechFile)
		return ""
	}

	var err error
	var results []Result
	var mux sync.Mutex
	c := colly.NewCollector(
		colly.Async(true),
		colly.IgnoreRobotsTxt(),
		colly.MaxDepth(options.Fin.Depth),
	)

	t := &http.Transport{}
	filename, _ = filepath.Abs(filename)

	t.RegisterProtocol("file", http.NewFileTransport(http.Dir(path.Dir(filename))))
	c.WithTransport(t)

	htmlFile := path.Base(filename)
	utils.DebugF("Fingerprint at reading: %v", htmlFile)
	c.Visit(fmt.Sprintf("file:///%s", htmlFile))

	extensions.RandomMobileUserAgent(c)
	extensions.Referer(c)

	// Handle url
	c.OnHTML("[href]", func(e *colly.HTMLElement) {
		urlString := e.Request.AbsoluteURL(e.Attr("href"))
		if urlString == "" {
			return
		}
		_ = e.Request.Visit(urlString)
	})

	siteMap := treemap.NewWithStringComparator()
	c.OnHTML("[src]", func(e *colly.HTMLElement) {
		srcURL := e.Request.AbsoluteURL(e.Attr("src"))
		if _, ok := siteMap.Get(srcURL); !ok {
			siteMap.Put(srcURL, e.Request.URL.String())
		}

		urlString := e.Request.AbsoluteURL(srcURL)
		if urlString == "" {
			return
		}
		_ = e.Request.Visit(urlString)
	})

	// Setup app technologies detector handle
	c.OnResponse(func(response *colly.Response) {
		var apps = make([]Match, 0)
		var doc *goquery.Document
		htmlResponse := false
		jsFile := false

		//fmt.Println(string(response.Body))

		contentType := http.DetectContentType(response.Body)
		if strings.Contains(contentType, "html") {
			doc, err = goquery.NewDocumentFromReader(bytes.NewReader(response.Body))
			if err == nil {
				htmlResponse = true
			}
		} else {
			if path.Ext(response.Request.URL.EscapedPath()) == ".js" {
				jsFile = true
			}
		}

		var scripts []string
		if htmlResponse == true {
			doc.Find("script").Each(func(i int, s *goquery.Selection) {
				if script, exists := s.Attr("src"); exists {
					scripts = append(scripts, script)
				}
			})
		}

		// load Cookie info map
		var cookiesMap = make(map[string]string)
		for k, v := range *response.Headers {
			hk := http.CanonicalHeaderKey(k)
			if hk != "Set-Cookie" {
				continue
			}
			for _, cookie := range v {
				keyValues := strings.Split(cookie, ";")
				keyValueSlice := strings.Split(keyValues[0], "=")
				if len(keyValueSlice) > 1 {
					key, value := keyValueSlice[0], keyValueSlice[1]
					cookiesMap[key] = value
				}
			}
		}

		for appname, app := range WA.AppDefs.Apps {
			findings := Match{
				App:     app,
				AppName: appname,
				Matches: make([][]string, 0),
			}
			// check raw html
			if m, v := FindMatches(string(response.Body), app.HTMLRegex); len(m) > 0 {
				findings.Matches = append(findings.Matches, m...)
				findings.updateVersion(v)
			}

			// check response header
			headerFindings, version := app.FindInHeaders(*response.Headers)
			findings.Matches = append(findings.Matches, headerFindings...)
			findings.updateVersion(version)

			// check url
			if m, v := FindMatches(response.Request.URL.String(), app.URLRegex); len(m) > 0 {
				findings.Matches = append(findings.Matches, m...)
				findings.updateVersion(v)
			}

			if htmlResponse == true {
				// check script tags
				for _, script := range scripts {
					if m, v := FindMatches(script, app.ScriptRegex); len(m) > 0 {
						findings.Matches = append(findings.Matches, m...)
						findings.updateVersion(v)
					}
				}

				// check meta tags
				for _, h := range app.MetaRegex {
					selector := fmt.Sprintf("meta[name='%s']", h.Name)
					doc.Find(selector).Each(func(i int, s *goquery.Selection) {
						content, _ := s.Attr("content")
						if m, v := FindMatches(content, []AppRegexp{h}); len(m) > 0 {
							findings.Matches = append(findings.Matches, m...)
							findings.updateVersion(v)
						}
						selector := fmt.Sprintf("meta[property='%s']", h.Name)
						doc.Find(selector).Each(func(i int, s *goquery.Selection) {
							content, _ := s.Attr("content")
							if m, v := FindMatches(content, []AppRegexp{h}); len(m) > 0 {
								findings.Matches = append(findings.Matches, m...)
								findings.updateVersion(v)
							}
						})
					})
				}
			}

			if jsFile {
				// check JS
				for _, j := range app.JSRegex {
					if j.Regexp != nil {
						if strings.Contains(string(response.Body), j.Name) {
							findings.Matches = append(findings.Matches, []string{j.Name})
						}
					}
				}
			}

			// check cookies
			for _, c := range app.CookieRegex {
				if _, ok := cookiesMap[c.Name]; ok {
					// if there is a regexp set, ensure it matches.
					// otherwise just add this as a match
					if c.Regexp != nil {
						// only match single AppRegexp on this specific cookie
						if m, v := FindMatches(cookiesMap[c.Name], []AppRegexp{c}); len(m) > 0 {
							findings.Matches = append(findings.Matches, m...)
							findings.updateVersion(v)
						}
					} else {
						findings.Matches = append(findings.Matches, []string{c.Name})
					}
				}
			}

			if len(findings.Matches) > 0 {
				apps = append(apps, findings)

				// handle implies
				for _, implies := range app.Implies {
					for implyAppname, implyApp := range WA.AppDefs.Apps {
						if implies != implyAppname {
							continue
						}
						f2 := Match{
							App:     implyApp,
							AppName: implyAppname,
							Matches: make([][]string, 0),
						}
						apps = append(apps, f2)
					}
				}
			}
		}
		var result Result
		if jsFile {
			if v, ok := siteMap.Get(response.Request.URL.String()); ok {
				result = Result{
					Host:    v.(string),
					Matches: apps,
				}
			}
		}
		if result.Host == "" {
			result = Result{
				Host:    response.Request.URL.String(),
				Matches: apps,
			}
		}

		mux.Lock()
		results = append(results, result)
		mux.Unlock()
	})

	c.Wait()

	var finalTech string
	// set := treeset.NewWithStringComparator()
	for _, result := range results {
		for _, match := range result.Matches {
			sort.Strings(match.CatNames)
			app := fmt.Sprintf("%s", match.AppName)
			if match.Version != "" {
				app = fmt.Sprintf("%s/%s", match.AppName, match.Version)
			}
			finalTech += fmt.Sprintf("%s,", app)
			// row := []string{
			// 	app,
			// 	strings.Join(match.CatNames, ","),
			// }

			// tech := strings.Join(row, "|")
			// if !set.Contains(tech) {
			// 	set.Add(tech)
			// }
		}
	}

	if finalTech == "" {
		utils.ErrorF("no tech found from: %s", filename)
	}
	finalTech = strings.TrimRight(finalTech, ",")
	return finalTech
}
