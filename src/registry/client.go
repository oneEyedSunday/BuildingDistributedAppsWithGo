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
