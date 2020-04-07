package runner

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/lamoda/gonkey/models"
)

func newClient() (*http.Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	if os.Getenv("HTTP_PROXY") != "" {
		proxyUrl, err := url.Parse(os.Getenv("HTTP_PROXY"))
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}, nil
}

func newRequest(host string, test models.TestInterface) (*http.Request, error) {
	body, err := test.ToJSON()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(
		strings.ToUpper(test.GetMethod()),
		host+test.Path()+test.ToQuery(),
		bytes.NewBuffer(body),
	)
	if err != nil {
		return nil, err
	}
	for k, v := range test.Headers() {
		request.Header.Add(k, v)
	}
	for k, v := range test.Cookies() {
		request.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	if request.Header.Get("Content-Type") == "" {
		request.Header.Set("Content-Type", "application/json")
	}

	return request, nil
}

func actualRequestBody(req *http.Request) string {
	if req.Body != nil {
		reqBodyStream, _ := req.GetBody()
		reqBody, _ := ioutil.ReadAll(reqBodyStream)
		return string(reqBody)
	}
	return ""
}
