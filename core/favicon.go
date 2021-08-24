package core

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/j3ssie/goverview/utils"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/twmb/murmur3"
)

func GetFavHash(URL string) string {
	u, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	hashURL := fmt.Sprintf("%v://%v/favicon.ico", u.Scheme, u.Host)
	utils.DebugF("Get favicon at %v", hashURL)
	data := BigResponseReq(hashURL)
	if data == "" {
		return ""
	}
	hashedFav := Mmh3Hash32(StandBase64([]byte(data)))
	return hashedFav
}

func Mmh3Hash32(raw []byte) string {
	h32 := murmur3.New32()
	_, err := h32.Write(raw)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%d", int32(h32.Sum32()))
}

// StandBase64 base64 from bytes
func StandBase64(data []byte) []byte {
	raw := base64.StdEncoding.EncodeToString(data)
	var buffer bytes.Buffer
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		buffer.WriteByte(ch)
		if (i+1)%76 == 0 {
			buffer.WriteByte('\n')
		}
	}
	buffer.WriteByte('\n')
	return buffer.Bytes()
}

func BigResponseReq(baseUrl string) string {
	client := &http.Client{
		Timeout: time.Duration(10*3) * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: time.Second * 60,
			}).DialContext,
			MaxIdleConns:        1000,
			MaxIdleConnsPerHost: 500,
			MaxConnsPerHost:     500,
			TLSClientConfig:     &tls.Config{InsecureSkipVerify: true, Renegotiation: tls.RenegotiateOnceAsClient},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error { return http.ErrUseLastResponse },
	}

	req, _ := http.NewRequest("GET", baseUrl, nil)
	req.Header.Add("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/80.0.3987.132 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return ""
	}
	return string(content)
}
