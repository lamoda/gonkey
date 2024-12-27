package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

// возможные состояния светофора
const (
	lightRed    = "red"
	lightYellow = "yellow"
	lightGreen  = "green"
)

// структура для хранения состояния светофора
type trafficLights struct {
	CurrentLight string       `json:"currentLight"`
	mutex        sync.RWMutex `json:"-"`
}

// экземпляр светофора
var lights = trafficLights{
	CurrentLight: lightRed,
}

func main() {
	initServer()

	// запуск сервера (блокирующий)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func initServer() {
	// метод для получения текущего состояния светофора
	http.HandleFunc("/light/get", func(w http.ResponseWriter, r *http.Request) {
		lights.mutex.RLock()
		defer lights.mutex.RUnlock()

		resp, err := json.Marshal(lights)
		if err != nil {
			log.Fatal(err)
		}

		w.Header().Add("Content-Type", "application/json")
		w.Write(resp)
	})

	// метод для установки нового состояния светофора
	http.HandleFunc("/light/set", func(w http.ResponseWriter, r *http.Request) {
		lights.mutex.Lock()
		defer lights.mutex.Unlock()

		request, err := io.ReadAll(r.Body)
		if err != nil {
			log.Fatal(err)
		}

		var newTrafficLights trafficLights
		if err := json.Unmarshal(request, &newTrafficLights); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := validateRequest(&newTrafficLights); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		lights.CurrentLight = newTrafficLights.CurrentLight
	})
}

func validateRequest(lights *trafficLights) error {
	if lights.CurrentLight != lightRed &&
		lights.CurrentLight != lightYellow &&
		lights.CurrentLight != lightGreen {
		return fmt.Errorf("incorrect current light: %s", lights.CurrentLight)
	}
	return nil
}
