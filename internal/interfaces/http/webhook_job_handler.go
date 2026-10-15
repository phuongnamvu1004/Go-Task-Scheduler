package http

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/application"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

// WebhookJobHandler serves /jobs/webhook. Jobs of other task types are
// reported as not found.
type WebhookJobHandler struct {
	service *application.JobService
}

func NewWebhookJobHandler(service *application.JobService) *WebhookJobHandler {
	return &WebhookJobHandler{service: service}
}

// RegisterRoutes registers routes relative to /jobs/webhook.
func (h *WebhookJobHandler) RegisterRoutes(r chi.Router) {
	r.Post("/", h.Create)
	r.Get("/", h.List)
	r.Get("/{id}", h.Get)
	r.Put("/{id}", h.Update)
	r.Delete("/{id}", h.Delete)
	r.Post("/{id}/pause", h.Pause)
	r.Post("/{id}/resume", h.Resume)
}

type webhookJobRequest struct {
	Name     string          `json:"name"`
	Schedule string          `json:"schedule"`
	Task     job.WebhookTask `json:"task"`
}

func (h *WebhookJobHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req webhookJobRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeBadRequest(w, err)
		return
	}

	j, err := h.service.CreateJob(r.Context(), application.CreateJobInput{
		Name:     req.Name,
		Schedule: req.Schedule,
		Task:     &req.Task,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}

	w.Header().Set("Location", "/jobs/webhook/"+j.ID)
	writeJSON(w, http.StatusCreated, toJobResponse(j))
}

func (h *WebhookJobHandler) List(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.service.ListJobs(r.Context(), job.ListFilter{TaskType: job.TaskWebhook})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toJobResponses(jobs))
}

func (h *WebhookJobHandler) Get(w http.ResponseWriter, r *http.Request) {
	j, err := h.find(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toJobResponse(j))
}

func (h *WebhookJobHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req webhookJobRequest
	if err := decodeJSON(w, r, &req); err != nil {
		writeBadRequest(w, err)
		return
	}
	if _, err := h.find(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}

	j, err := h.service.UpdateJob(r.Context(), id, application.UpdateJobInput{
		Name:     req.Name,
		Schedule: req.Schedule,
		Task:     &req.Task,
	})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toJobResponse(j))
}

func (h *WebhookJobHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if _, err := h.find(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	if err := h.service.DeleteJob(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WebhookJobHandler) Pause(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.service.PauseJob)
}

func (h *WebhookJobHandler) Resume(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.service.ResumeJob)
}

func (h *WebhookJobHandler) changeStatus(w http.ResponseWriter, r *http.Request, change func(context.Context, string) (*job.Job, error)) {
	id := r.PathValue("id")
	if _, err := h.find(r.Context(), id); err != nil {
		writeError(w, r, err)
		return
	}
	j, err := change(r.Context(), id)
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toJobResponse(j))
}

// find loads a job, returning job.ErrNotFound if it isn't a webhook job.
func (h *WebhookJobHandler) find(ctx context.Context, id string) (*job.Job, error) {
	j, err := h.service.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if j.Task.Type() != job.TaskWebhook {
		return nil, job.ErrNotFound
	}
	return j, nil
}
