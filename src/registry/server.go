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
	"time"
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
	r.notify(patch{
		Added:   []patchEntry{{Name: reg.ServiceName, URL: reg.ServiceURL}},
		Removed: []patchEntry{},
	})

	return err
}

// notify serves as a general purpose notification method
func (r *registry) notify(pat patch) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, reg := range r.registrations {
		go func(reg Registration) {
			for _, requiredService := range reg.RequiredServices {
				p := patch{Added: []patchEntry{}, Removed: []patchEntry{}}

				// signal so we can skip sending updates if not necessary
				sendUpdate := false

				for _, added := range pat.Added {
					if added.Name == requiredService {
						p.Added = append(p.Added, added)
						sendUpdate = true
					}
				}

				for _, removed := range pat.Removed {
					if removed.Name == requiredService {
						p.Removed = append(p.Removed, removed)
						sendUpdate = true
					}
				}

				log.Printf("notifying registration %+v of patch %+v\n, sendingUpdate: %t", reg, p, sendUpdate)

				if sendUpdate {
					if err := r.sendPatch(p, reg.ServiceUpdateURL); err != nil {
						log.Println(err)
						return
					}
				}
			}
		}(reg)
	}
}

func (r *registry) heartbeat(freq time.Duration, retryDuration time.Duration) {
	heartbeatTick := time.NewTicker(freq)
	defer heartbeatTick.Stop()
	for {
		log.Printf("heartbeat cycle started, watching %d services", len(r.registrations))
		var wg sync.WaitGroup
		for _, reg := range r.registrations {
			wg.Add(1)
			log.Println("running heartbeat check for registration", reg.ServiceName)

			go func(reg Registration) {
				defer wg.Done()

				success := true
				ticker := time.NewTicker(retryDuration)
				defer ticker.Stop()

				for attempts := 0; attempts < 3; attempts++ {

					res, err := http.Get(reg.HeartbeatURL)
					if err != nil {
						log.Println(err)
					} else if res.StatusCode == http.StatusOK {
						log.Printf("heartbeat checked passed for %v\n", reg.ServiceName)
						if !success {
							r.add(reg)
						}
						break
					}

					log.Printf("heartbeat check failed for %v", reg.ServiceName)
					if success {
						success = false
						r.remove(reg.ServiceURL)
					}

					<-ticker.C
				}
			}(reg)
		}

		wg.Wait()
		<-heartbeatTick.C
	}
}

var once sync.Once

func SetupRegistryService() {
	once.Do(func() {
		go reg.heartbeat(5*time.Second, 2*time.Second)
	})
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

	log.Printf("sending patch %+v to %s.\n", p, url)
	// we dont care about response here (for now)
	res, err := http.Post(url, "application/json", bytes.NewBuffer(d))
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("received result %d from sending patch %+v to %s\n", res.StatusCode, p, url)

	return nil
}

func (r *registry) remove(url string) error {
	// NOTE: bug here, this assumes the urls are unique (which indeed is likely the case in a non dockerized environment)
	// but with containers we can seemingly use the same ports on different containers
	// we should have an ID also, like in otoole's hraftd demo
	// so ID + url

	for i := range r.registrations {
		if r.registrations[i].ServiceURL == url {
			r.notify(patch{
				Added:   []patchEntry{},
				Removed: []patchEntry{{Name: r.registrations[i].ServiceName, URL: r.registrations[i].ServiceURL}},
			})
			// why locking here?? and not outside???
			// alot more code eventually happen, so for the sake of speed, do it here
			// although, the loop target of r.registrations may be updated elsewhere
			r.mu.Lock()
			r.registrations = append(r.registrations[:i], r.registrations[i+1:]...)
			r.mu.Unlock()
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
