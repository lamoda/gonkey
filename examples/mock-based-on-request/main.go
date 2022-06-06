package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	urlpkg "net/url"
	"os"
	"sync/atomic"

	"golang.org/x/sync/errgroup"
)

func main() {
	initServer()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	http.HandleFunc("/do", Do)
}

func Do(w http.ResponseWriter, r *http.Request) {
	params1 := urlpkg.Values{"key": []string{"value1"}}.Encode()
	params2 := urlpkg.Values{"key": []string{"value2"}}.Encode()
	params3 := urlpkg.Values{"value": []string{"3"}}.Encode()
	params4 := urlpkg.Values{"value": []string{"4"}}.Encode()

	doRequest := func(params string, method string, reqBody []byte) (int64, error) {
		url := fmt.Sprintf("http://%s/request?%s", os.Getenv("BACKEND_ADDR"), params)

		req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
		if err != nil {
			return 0, fmt.Errorf("failed to build request: %w", err)
		}

		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return 0, err
		}
		if res.StatusCode != http.StatusOK {
			return 0, fmt.Errorf("backend response status code %d", res.StatusCode)
		}
		body, err := ioutil.ReadAll(res.Body)
		_ = res.Body.Close()
		if err != nil {
			return 0, fmt.Errorf("cannot read response body %w", err)
		}
		var resp struct {
			Value int64 `json:"value"`
		}
		err = json.Unmarshal(body, &resp)
		if err != nil {
			return 0, fmt.Errorf("cannot unmarshal response body %w", err)
		}
		return resp.Value, nil
	}

	var total int64
	errg := errgroup.Group{}
	errg.Go(func() error {
		v, err := doRequest(params1, http.MethodGet, nil)
		atomic.AddInt64(&total, v)
		return err
	})
	errg.Go(func() error {
		v, err := doRequest(params2, http.MethodGet, nil)
		atomic.AddInt64(&total, v)
		return err
	})
	errg.Go(func() error {
		v, err := doRequest(params3, http.MethodGet, nil)
		atomic.AddInt64(&total, v)
		return err
	})
	errg.Go(func() error {
		v, err := doRequest(params4, http.MethodPost, []byte(`{"data":{"value": 10}}`))
		atomic.AddInt64(&total, v)
		return err
	})
	err := errg.Wait()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = io.WriteString(w, err.Error())
		return
	}
	_, _ = fmt.Fprintf(w, `{"total":%v}`, total)
}
