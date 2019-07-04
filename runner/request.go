package runner

import (
	"bytes"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/keyclaim/gonkey/models"
)

func newClient() *http.Client {
	return &http.Client{Transport: &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}}
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

	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Connection", "close")

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
