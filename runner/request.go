package runner

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/lamoda/gonkey/models"
)

func newClient(proxyURL *url.URL) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec // Client is only used for testing.
		Proxy:           http.ProxyURL(proxyURL),
	}

	return &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func newRequest(host string, test models.TestInterface) (req *http.Request, err error) {
	if test.GetForm() != nil {
		req, err = newMultipartRequest(host, test)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = newCommonRequest(host, test)
		if err != nil {
			return nil, err
		}
	}

	for k, v := range test.Cookies() {
		req.AddCookie(&http.Cookie{Name: k, Value: v})
	}

	return req, nil
}

func newMultipartRequest(host string, test models.TestInterface) (*http.Request, error) {
	if test.ContentType() != "" && test.ContentType() != "multipart/form-data" {
		return nil, fmt.Errorf(
			"test has unexpected Content-Type: %s, expected: multipart/form-data",
			test.ContentType(),
		)
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	params, err := url.ParseQuery(test.GetRequest())
	if err != nil {
		return nil, err
	}

	err = addFields(params, w)
	if err != nil {
		return nil, err
	}

	err = addFiles(test.GetForm().Files, w)
	if err != nil {
		return nil, err
	}

	_ = w.Close()

	req, err := request(test, &b, host)
	if err != nil {
		return nil, err
	}

	// this is necessary, it will contain boundary
	req.Header.Set("Content-Type", w.FormDataContentType())

	return req, nil
}

func addFiles(files map[string]string, w *multipart.Writer) error {
	for name, path := range files {
		err := addFile(path, w, name)
		if err != nil {
			return err
		}
	}

	return nil
}

func addFile(path string, w *multipart.Writer, name string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	fw, err := w.CreateFormFile(name, filepath.Base(f.Name()))
	if err != nil {
		return err
	}

	if _, err = io.Copy(fw, f); err != nil {
		return err
	}

	return nil
}

func addFields(params url.Values, w *multipart.Writer) error {
	for k, vv := range params {
		for _, v := range vv {
			fw, err := w.CreateFormField(k)
			if err != nil {
				return err
			}

			_, err = fw.Write([]byte(v))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func newCommonRequest(host string, test models.TestInterface) (*http.Request, error) {
	body, err := test.ToJSON()
	if err != nil {
		return nil, err
	}

	req, err := request(test, bytes.NewBuffer(body), host)
	if err != nil {
		return nil, err
	}

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func request(test models.TestInterface, b *bytes.Buffer, host string) (*http.Request, error) {
	req, err := http.NewRequest(
		strings.ToUpper(test.GetMethod()),
		host+test.Path()+test.ToQuery(),
		b,
	)
	if err != nil {
		return nil, err
	}

	for k, v := range test.Headers() {
		if strings.EqualFold(k, "host") {
			req.Host = v
		} else {
			req.Header.Add(k, v)
		}
	}

	return req, nil
}

func actualRequestBody(req *http.Request) string {
	if req.Body != nil {
		reqBodyStream, _ := req.GetBody()
		reqBody, _ := ioutil.ReadAll(reqBodyStream)

		return string(reqBody)
	}

	return ""
}
