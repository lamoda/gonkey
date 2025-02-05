package runner

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"slices"
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
	var boundary string

	if test.ContentType() != "" {
		contentType, params, err := mime.ParseMediaType(test.ContentType())
		if err != nil {
			return nil, err
		}
		if contentType != "multipart/form-data" {
			return nil, fmt.Errorf(
				"test has unexpected Content-Type: %s, expected: multipart/form-data",
				test.ContentType(),
			)
		}

		if b, ok := params["boundary"]; ok {
			boundary = b
		}
	}

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	if boundary != "" {
		if err := w.SetBoundary(boundary); err != nil {
			return nil, fmt.Errorf("SetBoundary : %w", err)
		}
	}

	err := addFields(test.GetForm().Fields, w)
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
func addFields(fields map[string]string, w *multipart.Writer) error {
	// TODO: sort fields better
	fieldNames := make([]string, 0, len(fields))
	for n, _ := range fields {
		fieldNames = append(fieldNames, n)
	}
	slices.Sort(fieldNames)

	for _, name := range fieldNames {
		n := name
		v := fields[n]
		if err := w.WriteField(n, v); err != nil {
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
		reqBody, _ := io.ReadAll(reqBodyStream)

		return string(reqBody)
	}

	return ""
}
