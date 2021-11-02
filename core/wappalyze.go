package core

import (
	"bytes"
	"fmt"
	"github.com/markbates/pkger"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// WappalyzerURL is the link to the latest apps.json file in the Wappalyzer repo
const WappalyzerURL = "https://raw.githubusercontent.com/AliasIO/wappalyzer/master/src/technologies.json"

var WA *WebAnalyzer

// WebAnalyzer types holds an analyzation job
type WebAnalyzer struct {
	AppDefs *AppsDefinition
}

// StringArray type is a wrapper for []string for use in unmarshalling the apps.json
type StringArray []string

// App type encapsulates all the data about an App from apps.json
type App struct {
	Cats     StringArray       `json:"cats"`
	CatNames []string          `json:"category_names"`
	Cookies  map[string]string `json:"cookies"`
	Headers  map[string]string `json:"headers"`
	Meta     map[string]string `json:"meta"`
	JS       map[string]string `json:"js"`
	HTML     StringArray       `json:"html"`
	Script   StringArray       `json:"script"`
	URL      StringArray       `json:"url"`
	Website  string            `json:"website"`
	Implies  StringArray       `json:"implies"`

	HTMLRegex   []AppRegexp `json:"-"`
	ScriptRegex []AppRegexp `json:"-"`
	URLRegex    []AppRegexp `json:"-"`
	HeaderRegex []AppRegexp `json:"-"`
	MetaRegex   []AppRegexp `json:"-"`
	CookieRegex []AppRegexp `json:"-"`
	JSRegex     []AppRegexp `json:"-"`
}

// Category names defined by wappalyzer
type Category struct {
	Name string `json:"name"`
}

// AppsDefinition type encapsulates the json encoding of the whole apps.json file
type AppsDefinition struct {
	Apps map[string]App      `json:"technologies"`
	Cats map[string]Category `json:"categories"`
}

type AppRegexp struct {
	Name    string
	Regexp  *regexp.Regexp
	Version string
}

func (app *App) FindInHeaders(headers http.Header) (matches [][]string, version string) {
	var v string

	for _, hre := range app.HeaderRegex {
		if headers.Get(hre.Name) == "" {
			continue
		}
		hk := http.CanonicalHeaderKey(hre.Name)
		for _, headerValue := range headers[hk] {
			if headerValue == "" {
				continue
			}
			if m, version := FindMatches(headerValue, []AppRegexp{hre}); len(m) > 0 {
				matches = append(matches, m...)
				v = version
			}
		}
	}
	return matches, v
}

// UnmarshalJSON is a custom unmarshaler for handling bogus apps.json types from wappalyzer
func (t *StringArray) UnmarshalJSON(data []byte) error {
	var s string
	var sa []string
	var na []int

	if err := jsoniter.Unmarshal(data, &s); err != nil {
		if err := jsoniter.Unmarshal(data, &na); err == nil {
			// not a string, so maybe []int?
			*t = make(StringArray, len(na))

			for i, number := range na {
				(*t)[i] = fmt.Sprintf("%d", number)
			}

			return nil
		} else if err := jsoniter.Unmarshal(data, &sa); err == nil {
			// not a string, so maybe []string?
			*t = sa
			return nil
		}
		fmt.Println(string(data))
		return err
	}
	*t = StringArray{s}
	return nil
}

// DownloadFile pulls the latest apps.json file from the Wappalyzer github
func DownloadFile(from, to string) error {
	resp, err := http.Get(from)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(to)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, resp.Body)
	return err
}

// LoadApps load apps from file
func (wa *WebAnalyzer) LoadApps(filename string) error {
	if filename == "" {
		f, err := pkger.Open("/static/technologies.json")
		if err != nil {
			return fmt.Errorf("blank tech file")
		}

		buf := new(bytes.Buffer)
		buf.ReadFrom(f)

		if err = jsoniter.UnmarshalFromString(buf.String(), &wa.AppDefs); err != nil {
			return err
		}
	} else {
		f, err := os.Open(filename)
		if err != nil {
			return err
		}
		defer f.Close()
		dec := jsoniter.NewDecoder(f)
		if err = dec.Decode(&wa.AppDefs); err != nil {
			return err
		}
	}

	for key, value := range wa.AppDefs.Apps {

		app := wa.AppDefs.Apps[key]

		app.HTMLRegex = compileRegexes(value.HTML)
		app.ScriptRegex = compileRegexes(value.Script)
		app.URLRegex = compileRegexes(value.URL)

		app.HeaderRegex = compileNamedRegexes(app.Headers)
		app.MetaRegex = compileNamedRegexes(app.Meta)
		app.CookieRegex = compileNamedRegexes(app.Cookies)
		app.JSRegex = compileNamedRegexes(app.JS)

		app.CatNames = make([]string, 0)

		for _, cid := range app.Cats {
			if category, ok := wa.AppDefs.Cats[string(cid)]; ok && category.Name != "" {
				app.CatNames = append(app.CatNames, category.Name)
			}
		}
		wa.AppDefs.Apps[key] = app
	}
	return nil
}

func (wa *WebAnalyzer) CategoryById(cid string) string {
	if _, ok := wa.AppDefs.Cats[cid]; !ok {
		return ""
	}
	return wa.AppDefs.Cats[cid].Name
}

func compileNamedRegexes(from map[string]string) []AppRegexp {
	var list []AppRegexp
	for key, value := range from {
		h := AppRegexp{
			Name: key,
		}
		if value == "" {
			value = ".*"
		}

		// Filter out webapplyzer attributes from regular expression
		splitted := strings.Split(value, "\\;")
		r, err := regexp.Compile(splitted[0])
		if err != nil {
			continue
		}

		if len(splitted) > 1 && strings.HasPrefix(splitted[1], "version:") {
			h.Version = splitted[1][8:]
		}
		h.Regexp = r
		list = append(list, h)
	}
	return list
}

func compileRegexes(s StringArray) []AppRegexp {
	var list []AppRegexp
	for _, regexString := range s {
		// Split version detection
		splitted := strings.Split(regexString, "\\;")
		regex, err := regexp.Compile(splitted[0])
		if err != nil {
			// ignore failed compiling for now
			// log.Printf("warning: compiling regexp for failed: %v", regexString, err)
		} else {
			rv := AppRegexp{
				Regexp: regex,
			}
			if len(splitted) > 1 && strings.HasPrefix(splitted[0], "version") {
				rv.Version = splitted[1][8:]
			}
			list = append(list, rv)
		}
	}

	return list
}

// runs a list of regexes on content
func FindMatches(content string, regexes []AppRegexp) ([][]string, string) {
	var m [][]string
	var version string

	for _, r := range regexes {
		matches := r.Regexp.FindAllStringSubmatch(content, -1)
		if matches == nil {
			continue
		}
		m = append(m, matches...)

		if r.Version != "" {
			version = FindVersion(m, r.Version)
		}
	}
	return m, version
}

// parses a version against matches
func FindVersion(matches [][]string, version string) string {
	var v string

	for _, matchPair := range matches {
		// replace backtraces (max: 3)
		for i := 1; i <= 3; i++ {
			bt := fmt.Sprintf("\\%v", i)
			if strings.Contains(version, bt) && len(matchPair) >= i {
				v = strings.Replace(version, bt, matchPair[i], 1)
			}
		}

		// return first found version
		if v != "" {
			return v
		}
	}
	return ""
}
