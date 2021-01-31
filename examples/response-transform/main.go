package main

import (
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	initServer()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	http.HandleFunc("/echo", Echo)
}

func Echo(w http.ResponseWriter, r *http.Request) {

	request, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(request)
}
