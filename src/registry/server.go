package registry

import (
	"bytes"
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
	mu sync.RWMutex
}

func (r *registry) add(reg Registration) error {
	r.mu.Lock()
	r.registrations = append(r.registrations, reg)
	r.mu.Unlock()

	err := r.sendRequiredServices(reg)

	return err
}

func (r *registry) sendRequiredServices(reg Registration) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var p patch

	// TODO refactor this loop
	// If we loop first across all needed services, we can have a shorter loop
	for _, serviceReg := range r.registrations {
		for _, reqService := range reg.RequiredServices {
			if serviceReg.ServiceName == reqService {
				p.Added = append(p.Added, patchEntry{
					Name: serviceReg.ServiceName,
					URL:  serviceReg.ServiceURL,
				})
			}
		}
	}

	if err := r.sendPatch(p, reg.ServiceUpdateURL); err != nil {
		return err
	}

	return nil
}

func (r *registry) sendPatch(p patch, url string) error {
	if url == "" {
		log.Println("service doesnt have a serviceUpdateURL, ignoring patch", p)
		return nil
	}

	if p.IsEmpty() {
		log.Println("empty patch, skipping")
		return nil
	}

	d, err := json.Marshal(p)
	if err != nil {
		return err
	}

	// we dont care about response here (for now)
	_, err = http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		return err
	}

	return nil
}

func (r *registry) remove(url string) error {
	// NOTE: bug here, this assumes the urls are unique (which indeed is likely the case in a non dockerized environment)
	// but with containers we can seemingly use the same ports on different containers
	// we should have an ID also, like in otoole's hraftd demo
	// so ID + url
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

		log.Printf("Adding service: %v to registry with URL: %s %v\n", registration.ServiceName, registration.ServiceURL, registration)
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
