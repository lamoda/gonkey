package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	Num int `json:"num"`
}

func main() {
	initServer()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	http.HandleFunc("/do", Do)
}

func Do(w http.ResponseWriter, r *http.Request) {
	var response Response

	if r.Method != "POST" {
		response.Num = 0
		fmt.Fprint(w, buildResponse(response))
		return
	}

	jsonRequest, _ := io.ReadAll(r.Body)
	request := buildRequest(jsonRequest)

	if request.Name == "a" {
		response.Num = 1
		fmt.Fprint(w, buildResponse(response))
		return
	}

	response.Num = 2
	fmt.Fprint(w, buildResponse(response))
}

func buildRequest(jsonRequest []byte) Request {
	var request Request

	json.Unmarshal(jsonRequest, &request)

	return request
}

func buildResponse(response Response) string {
	jsonResponse, _ := json.Marshal(response)

	return string(jsonResponse)
}
