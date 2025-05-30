package log

import (
	"io"
	stlog "log"
	"net/http"
	"os"
)

var log *stlog.Logger

type fileLog string

func (flh fileLog) Write(data []byte) (int, error) {
	f, err := os.OpenFile(string(flh), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return 0, err
	}

	defer f.Close()
	return f.Write(data)
}

func Run(dst string) {
	log = stlog.New(fileLog(dst), "", stlog.LstdFlags)
}

func RegisterHandlers() {
	http.HandleFunc("/log", func(w http.ResponseWriter, r *http.Request) {
		msg, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil || len(msg) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		write(string(msg))

	})
}

func write(msg string) {
	log.Printf("%v\n", msg)
}
