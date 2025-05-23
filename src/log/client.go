package log

import (
	"bytes"
	"fmt"
	stLog "log"
	"net/http"
	"pluralsight-go-building-distributed-apps/registry"
)

func SetClientLogger(serviceURL string, clientService registry.ServiceName) {
	stLog.SetPrefix(fmt.Sprintf("[%v] - ", clientService))
	stLog.SetFlags(0)
	stLog.SetOutput(&clientLogger{url: serviceURL})
}

type clientLogger struct {
	url string
}

func (cl clientLogger) Write(data []byte) (int, error) {
	b := bytes.NewBuffer([]byte(data))
	res, err := http.Post(cl.url+"/log", "text/plain", b)
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to send log message. Service responded with %v", res.StatusCode)
	}

	return len(data), nil
}
