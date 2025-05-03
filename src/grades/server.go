package grades

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func RegisterHandlers() {
	handler := new(studentsHandler)
	http.Handle("/students", handler)
	http.Handle("/students/", handler)
}

type studentsHandler struct{}

func (sh studentsHandler) toJSON(obj interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	if err := enc.Encode(obj); err != nil {
		return nil, fmt.Errorf("failed to serialize objects: %q", err)
	}

	return buf.Bytes(), nil
}

func (sh studentsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// GET /students - all students
	// GET /students/{id} - get a student by id
	// POST /students/{id}/grades - add grade for a single students
	pathSegments := strings.Split(r.URL.Path, "/")

	switch len(pathSegments) {
	case 2:
		sh.getAll(w, r)
	case 3:
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		sh.getOne(w, r, id)
	case 4:
		id, err := strconv.Atoi(pathSegments[2])
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		sh.addGrade(w, r, id)
	default:
		w.WriteHeader(http.StatusNotFound)

	}
}

func (sh studentsHandler) getAll(w http.ResponseWriter, r *http.Request) {
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	data, err := sh.toJSON(students)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (sh studentsHandler) getOne(w http.ResponseWriter, r *http.Request, id int) {
	// TODO pass id in context
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	data, err := sh.toJSON(student)
	if err != nil {
		log.Println(fmt.Errorf("failed to serialize student: %q", err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}

func (sh studentsHandler) addGrade(w http.ResponseWriter, r *http.Request, id int) {
	// TODO pass id in context
	studentsMutex.Lock()
	defer studentsMutex.Unlock()

	student, err := students.GetByID(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var g Grade
	dec := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err = dec.Decode(&g); err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return

	}

	student.Grades = append(student.Grades, g)

	w.WriteHeader(http.StatusCreated)
	data, err := sh.toJSON(g)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}
