package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/application"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

// JobHandler owns the /jobs route group. Each task type mounts its own
// sub-group from a separate file.
type JobHandler struct {
	service *application.JobService
}

func NewJobHandler(service *application.JobService) *JobHandler {
	return &JobHandler{service: service}
}

func (h *JobHandler) RegisterRoutes(r chi.Router) {
	r.Route("/jobs", func(r chi.Router) {
		r.Get("/", h.List)
		r.Route("/webhook", NewWebhookJobHandler(h.service).RegisterRoutes)
	})
}

// List returns jobs of every task type.
func (h *JobHandler) List(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.service.ListJobs(r.Context(), job.ListFilter{})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toJobResponses(jobs))
}
