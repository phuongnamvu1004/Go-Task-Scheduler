package http

import (
	"net/http"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/application"
)

// JobHandler exposes Job use cases over HTTP.
type JobHandler struct {
	service *application.JobService
}

func NewJobHandler(service *application.JobService) *JobHandler {
	return &JobHandler{service: service}
}

// RegisterRoutes attaches the job endpoints to the given mux.
func (h *JobHandler) RegisterRoutes(mux *http.ServeMux) {
	// TODO: add job routes, e.g.
	// mux.HandleFunc("POST /jobs", h.CreateJob)
	// mux.HandleFunc("GET /jobs", h.ListJobs)
}
