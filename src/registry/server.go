package registry

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"pluralsight-go-building-distributed-apps/pkg/util"
	"sync"
)

var ServerPort string
var ServicesURL string

func init() {

	ServerPort = util.StringOr(os.Getenv("REGISTRY_SERVICE_PORT"), "3000")
	serviceHost := util.StringOr(os.Getenv("REGISTRY_SERVICE_HOST"), "localhost")

	ServicesURL = fmt.Sprintf("http://%s:%s/services", serviceHost, ServerPort)
}

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

func (r *registry) remove(url string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := range r.registrations {
		if r.registrations[i].ServiceURL == url {
			// why locking here?? and not outside???
			r.registrations = append(r.registrations[:i], r.registrations[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("service at URL %s not found", url)
}

var reg = registry{registrations: make([]Registration, 0)}

type RegistryService struct{}

func (s RegistryService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("received request: HTTP %s %s", r.Method, r.URL)
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

	case http.MethodDelete:
		payload, err := io.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		url := string(payload)
		log.Printf("removing service with URL: %s", url)
		if err := reg.remove(url); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}
