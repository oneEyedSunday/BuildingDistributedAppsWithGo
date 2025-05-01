package registry

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const ServerPort = ":3000"
const ServicesURL = "http://localhost" + ServerPort + "/services"

type registry struct {
	registrations []Registration
	// mu is not declared as a pointer here because i find it more readable
	mu sync.Mutex
}

func (r *registry) add(reg Registration) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.registrations = append(r.registrations, reg)

	return nil
}

var reg = registry{registrations: make([]Registration, 0)}

type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Println("received request")
	switch r.Method {
	case http.MethodPost:
		dec := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var registration Registration

		if err := dec.Decode(&registration); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		log.Printf("Adding service: %v with URL: %s\n", registration.ServiceName, registration.ServiceURL)
		if err := reg.add(registration); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}
