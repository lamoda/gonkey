package main

import (
	"log"
	"net/http"
)

func main() {
	initServer()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	http.HandleFunc("/do", Do)
}

func Do(w http.ResponseWriter, _ *http.Request) {

	w.Header().Add("Content-Type", "application/xml")
	_, err := w.Write([]byte("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\r\n      " +
		"<SOAP-ENV:Envelope xmlns:SOAP-ENV=\"http://schemas.xmlsoap.org/soap/envelope/\"\r\n" +
		"                         xmlns:ns1=\"http://app.example.com\">\r\n" +
		"      \t<SOAP-ENV:Body>\r\n      \t\t<ns1:notifyResponse/>\r\n      " +
		"\t</SOAP-ENV:Body>\r\n      </SOAP-ENV:Envelope>"))
	if err != nil {
		return
	}
}
