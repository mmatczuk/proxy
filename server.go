package proxy

import (
	"net/http"

	"github.com/gorilla/mux"
)

type server struct {
	service Service
}

// NewServer creates HTTP handler exposing service via JSON/REST interface.
func NewServer(service Service) http.Handler {
	if service == nil {
		panic("Missing service")
	}

	s := &server{
		service: service,
	}

	return router(s)
}

func router(s *server) http.Handler {
	r := mux.NewRouter()

	api := r.PathPrefix("/v1").Subrouter()

	api.
		Path("/task").
		Methods(http.MethodPost).
		HandlerFunc(s.createTask)

	api.
		Path("/task/{id}/status").
		Methods(http.MethodGet).
		HandlerFunc(s.taskStatus)

	api.
		Path("/task/{id}/kill").
		Methods(http.MethodGet).
		HandlerFunc(s.killTask)

	return r
}

func (s *server) createTask(w http.ResponseWriter, r *http.Request) {
	var c TaskConfig
	if err := readJSON(&c, r.Body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.service.CreateTask(r.Context(), &c)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, id)
}

func (s *server) taskStatus(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	t, err := s.service.TaskStatus(r.Context(), TaskID(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if t == nil {
		http.NotFound(w, r)
		return
	}

	writeJSON(w, http.StatusOK, t.Results)
}

func (s *server) killTask(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	t, err := s.service.KillTask(r.Context(), TaskID(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if t == nil {
		http.NotFound(w, r)
		return
	}

	var killed []Result
	for _, v := range t.Results {
		if v.Status == Killed {
			killed = append(killed, v)
		}
	}

	writeJSON(w, http.StatusOK, killed)
}
