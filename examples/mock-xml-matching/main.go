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
	w.Write([]byte("{\"status\": \"ok\"}"))
}

func BackendPost() error {
	body := `
	<Person>
		<FullName>Harry Potter</FullName>
		<Company>Hogwarts School of Witchcraft and Wizardry</Company>
		<Email where="home">hpotter@gmail.com</Email>
		<Email where="work">hpotter@hog.gb</Email>
		<Group>
			<Value>Jinxes</Value>
			<Value>Hexes</Value>
		</Group>
		<Addr>4 Privet Drive</Addr>
	</Person>
	`
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
