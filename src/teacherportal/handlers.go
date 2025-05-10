package teacherportal

import "net/http"

func RegisterHandlers() {
	http.Handle("/", http.RedirectHandler("/students", http.StatusPermanentRedirect))
	h := new(studentsHandler)
	http.Handle("/students", h)
	http.Handle("/students/", h)
}

type studentsHandler struct{}

func (studentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {}
