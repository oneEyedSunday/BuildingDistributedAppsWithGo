package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"sync"

	stLog "log"
)

func RegisterService(r Registration) error {

	if r.ServiceUpdateURL != "" {
		serviceUpdateURL, err := url.Parse(r.ServiceUpdateURL)
		if err != nil {
			return err
		}

		stLog.Printf("handling service updates for %s on %s\n", r.ServiceName, serviceUpdateURL.Path)
		http.Handle(serviceUpdateURL.Path, &serviceUpdateHandler{})
	}

	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(r); err != nil {
		return err
	}

	stLog.Printf("attempting to register on servicesUrl %s: %v", ServicesURL, r)
	res, err := http.Post(ServicesURL, "application/json", buf)
	if err != nil {
		return err
	}
	// TODO accept 200 < 299
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register service. registry service responsed with code %v", res.StatusCode)
	}

	return nil
}

type serviceUpdateHandler struct{}

func (s *serviceUpdateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()
	var p patch
	if err := dec.Decode(&p); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// TODO confirm this actually registers for say running multiple instances of a service
	prov.Update(p)
}

func ShutdownService(serviceUrl string) error {
	req, err := http.NewRequest(http.MethodDelete, ServicesURL, bytes.NewBuffer([]byte(serviceUrl)))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "text/plain")
	res, err := http.DefaultClient.Do(req)
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to deregister service. Registry service responded with code %v", res.StatusCode)
	}
	return err
}

type providers struct {
	services map[ServiceName][]string
	mu       *sync.RWMutex
}

func (p *providers) Update(pat patch) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, patchEntry := range pat.Added {
		if _, ok := p.services[patchEntry.Name]; !ok {
			p.services[patchEntry.Name] = make([]string, 0)
		}

		// This should ideally be a set (for uniqueness)
		p.services[patchEntry.Name] = append(p.services[patchEntry.Name], patchEntry.URL)
	}

	for _, patchEntry := range pat.Removed {
		if providerURLs, ok := p.services[patchEntry.Name]; ok {
			for i := range providerURLs {
				if providerURLs[i] == patchEntry.URL {
					p.services[patchEntry.Name] = append(providerURLs[:i], providerURLs[i+1:]...)
				}
			}
		}
	}
}

func (p providers) get(name ServiceName) (string, error) {
	providers, ok := p.services[name]
	if !ok {
		return "", fmt.Errorf("no providers available for service %s", name)
	}

	idx := int(rand.Float32() * float32(len(providers)))
	return providers[idx], nil
}

func GetProvider(name ServiceName) (string, error) {
	return prov.get(name)

}

var prov = providers{
	services: make(map[ServiceName][]string),
	mu:       new(sync.RWMutex),
}
