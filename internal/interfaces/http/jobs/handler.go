package jobs

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/application"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

// Handler owns the /jobs route group. Each task type has its own handler
// file in this package (e.g. webhook_handler.go) mounted as a sub-group.
type Handler struct {
	service *application.JobService
}

func NewHandler(service *application.JobService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Route("/jobs", func(r chi.Router) {
		r.Get("/", h.List)
		r.Route("/webhook", NewWebhookHandler(h.service).RegisterRoutes)
	})
}

// List godoc
//
//	@Summary	List jobs of every task type
//	@Tags		jobs
//	@Produce	json
//	@Success	200	{array}		jobResponse
//	@Failure	500	{object}	errorResponse
//	@Router		/jobs [get]
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.service.ListJobs(r.Context(), job.ListFilter{})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toJobResponses(jobs))
}
