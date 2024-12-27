package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	initServer()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	http.HandleFunc("/proxy", ProxyRequest)
}

func ProxyRequest(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
		w.Write([]byte("{\"status\": \"error\"}"))
		return
	}

	if err := BackendPost(string(body)); err != nil {
		log.Print(err)
		w.Write([]byte("{\"status\": \"error\"}"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("{\"status\": \"ok\"}"))
}

type BackendParams struct {
	ClientCode    string `json:"client_code"`
	OriginRequest struct {
		Body string `json:"body"`
	} `json:"origin_request"`
}

func BackendPost(originBody string) error {
	params := BackendParams{
		ClientCode: "proxy",
	}
	params.OriginRequest.Body = originBody

	jsonParams, err := json.Marshal(params)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("http://%s/process", os.Getenv("BACKEND_ADDR"))
	res, err := http.Post(url, "application/json", bytes.NewReader(jsonParams))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("backend response status code %d", res.StatusCode)
	}

	return nil
}
