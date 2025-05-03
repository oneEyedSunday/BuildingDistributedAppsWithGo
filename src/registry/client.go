package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	stLog "log"
)

func RegisterService(r Registration) error {
	buf := new(bytes.Buffer)
	enc := json.NewEncoder(buf)
	if err := enc.Encode(r); err != nil {
		return err
	}

	stLog.Printf("attempting to register on servicesUrl %s", ServicesURL)
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
