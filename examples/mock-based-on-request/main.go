package main

import (
	"encoding/json"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	urlpkg "net/url"
	"os"
	"sync/atomic"
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
	doRequest := func(params string) (int64, error) {
		url := fmt.Sprintf("http://%s/request?%s", os.Getenv("BACKEND_ADDR"), params)
		res, err := http.Get(url)
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
		v, err := doRequest(params1)
		atomic.AddInt64(&total, v)
		return err
	})
	errg.Go(func() error {
		v, err := doRequest(params2)
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
