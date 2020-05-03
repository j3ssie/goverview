package core

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-resty/resty"
	"github.com/sirupsen/logrus"
)

// JustSend just sending request
func JustSend(options Options, url string) (res Response, err error) {
	method := "GET"
	headers := map[string]string{
		"UserAgent":  "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36",
		"Accept":     "*/*",
		"AcceptLang": "en-US,en;q=0.8",
	}
	timeout := options.Timeout

	// disable log when retry
	logger := logrus.New()
	if !options.Debug {
		logger.Out = ioutil.Discard
	}

	client := resty.New()
	client.SetLogger(logger)
	client.SetTransport(&http.Transport{
		MaxIdleConns:          100,
		MaxConnsPerHost:       1000,
		IdleConnTimeout:       time.Duration(timeout) * time.Second,
		ExpectContinueTimeout: time.Duration(timeout) * time.Second,
		ResponseHeaderTimeout: time.Duration(timeout) * time.Second,
		TLSHandshakeTimeout:   time.Duration(timeout) * time.Second,
		DisableCompression:    true,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	})

	client.SetHeaders(headers)
	client.SetCloseConnection(true)
	if options.Retry > 0 {
		client.SetRetryCount(options.Retry)
	}
	client.SetTimeout(time.Duration(timeout) * time.Second)
	client.SetRetryWaitTime(time.Duration(timeout/2) * time.Second)
	client.SetRetryMaxWaitTime(time.Duration(timeout) * time.Second)
	timeStart := time.Now()

	// redirect policy
	if options.Redirect == false {
		client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			// keep the header the same
			res.StatusCode = req.Response.StatusCode
			res.Status = req.Response.Status
			resp := req.Response
			bodyBytes, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				ErrorF("%v", err)
			}
			bodyString := string(bodyBytes)
			resLength := len(bodyString)
			// format the headers
			var resHeaders []map[string]string
			for k, v := range resp.Header {
				element := make(map[string]string)
				element[k] = strings.Join(v[:], "")
				resLength += len(fmt.Sprintf("%s: %s\n", k, strings.Join(v[:], "")))
				resHeaders = append(resHeaders, element)
			}

			// response time in second
			resTime := time.Since(timeStart).Seconds()
			resHeaders = append(resHeaders,
				map[string]string{"Total Length": strconv.Itoa(resLength)},
				map[string]string{"Response Time": fmt.Sprintf("%f", resTime)},
			)

			// set some variable
			res.Headers = resHeaders
			res.StatusCode = resp.StatusCode
			res.Status = fmt.Sprintf("%v %v", resp.Status, resp.Proto)
			res.Body = bodyString
			res.ResponseTime = resTime
			res.Length = resLength
			// beautify
			res.Beautify = BeautifyResponse(res)
			res.BeautifyHeader = BeautifyHeaders(res)
			return errors.New("auto redirect is disabled")
		}))

		client.AddRetryCondition(
			func(r *resty.Response, err error) bool {
				return false
			},
		)
	} else {
		client.SetRedirectPolicy(resty.RedirectPolicyFunc(func(req *http.Request, via []*http.Request) error {
			// keep the header the same
			client.SetHeaders(headers)
			return nil
		}))
	}

	var resp *resty.Response
	// really sending things here
	method = strings.ToLower(strings.TrimSpace(method))
	switch method {
	case "get":
		resp, err = client.R().
			Get(url)
		break
	}

	// in case we want to get redirect stuff
	if res.StatusCode != 0 {
		return res, nil
	}

	if err != nil || resp == nil {
		ErrorF("%v %v", url, err)
		return Response{}, err
	}

	return ParseResponse(*resp), nil
}

// ParseResponse field to Response
func ParseResponse(resp resty.Response) (res Response) {
	// var res libs.Response
	resLength := len(string(resp.Body()))
	// format the headers
	var resHeaders []map[string]string
	for k, v := range resp.RawResponse.Header {
		element := make(map[string]string)
		element[k] = strings.Join(v[:], "")
		resLength += len(fmt.Sprintf("%s: %s\n", k, strings.Join(v[:], "")))
		resHeaders = append(resHeaders, element)
	}
	// response time in second
	resTime := float64(resp.Time()) / float64(time.Second)
	resHeaders = append(resHeaders,
		map[string]string{"Total Length": strconv.Itoa(resLength)},
		map[string]string{"Response Time": fmt.Sprintf("%f", resTime)},
	)

	// set some variable
	res.Headers = resHeaders
	res.StatusCode = resp.StatusCode()
	res.Status = fmt.Sprintf("%v %v", resp.Status(), resp.RawResponse.Proto)
	res.Body = string(resp.Body())
	res.ResponseTime = resTime
	res.Length = resLength
	// beautify
	res.Beautify = BeautifyResponse(res)
	res.BeautifyHeader = BeautifyHeaders(res)
	return res
}

// BeautifyRequest beautify request
func BeautifyRequest(req Request) string {
	var beautifyReq string
	// hardcoded HTTP/1.1 for now
	beautifyReq += fmt.Sprintf("%v %v HTTP/1.1\n", req.Method, req.URL)

	for _, header := range req.Headers {
		for key, value := range header {
			if key != "" && value != "" {
				beautifyReq += fmt.Sprintf("%v: %v\n", key, value)
			}
		}
	}
	if req.Body != "" {
		beautifyReq += fmt.Sprintf("\n%v\n", req.Body)
	}
	return beautifyReq
}

func BeautifyHeaders(res Response) string {
	beautifyHeader := fmt.Sprintf("< %v \n", res.Status)
	for _, header := range res.Headers {
		for key, value := range header {
			beautifyHeader += fmt.Sprintf("< %v: %v\n", key, value)
		}
	}
	return beautifyHeader
}

// BeautifyResponse beautify response
func BeautifyResponse(res Response) string {
	var beautifyRes string
	beautifyRes += fmt.Sprintf("%v \n", res.Status)

	for _, header := range res.Headers {
		for key, value := range header {
			beautifyRes += fmt.Sprintf("%v: %v\n", key, value)
		}
	}

	beautifyRes += fmt.Sprintf("\n%v\n", res.Body)
	return beautifyRes
}
