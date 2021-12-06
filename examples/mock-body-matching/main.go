package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	initServer()
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	http.HandleFunc("/do", Do)
}

func Do(w http.ResponseWriter, r *http.Request) {
	if err := BackendPost(); err != nil {
		log.Print(err)
		w.Write([]byte("{\"status\": \"error\"}"))
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte("{\n  \"data\": {\n    \"hero\": {\n      \"name\": \"R2-D2\",\n   " +
		"   \"friends\": [\n        {\n          \"name\": \"Luke Skywalker\"\n        },\n        " +
		"{\n          \"name\": \"Han Solo\"\n        },\n        {\n          \"name\": \"Leia Organa\"\n   " +
		"     }\n      ]\n    }\n  }\n}"))
}

func BackendPost() error {
	body := `query HeroNameAndFriends {
      hero {
        name
        friends {
          name
        }
      }
    }`
	url := fmt.Sprintf("http://%s/process", os.Getenv("BACKEND_ADDR"))
	res, err := http.Post(url, "application/json", strings.NewReader(body))
	if err != nil {
		return err
	}
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("backend response status code %d", res.StatusCode)
	}

	return nil
}
