package jobs

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/application"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

// WebhookHandler serves /jobs/webhook. Jobs of other task types are
// reported as not found.
type WebhookHandler struct {
	service *application.JobService
}

func NewWebhookHandler(service *application.JobService) *WebhookHandler {
	return &WebhookHandler{service: service}
}

// RegisterRoutes registers routes relative to /jobs/webhook.
func (h *WebhookHandler) RegisterRoutes(r chi.Router) {
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

// Create godoc
//
//	@Summary	Create a webhook job
//	@Tags		webhook jobs
//	@Accept		json
//	@Produce	json
//	@Param		job	body		webhookJobRequest	true	"Webhook job"
//	@Success	201	{object}	jobResponse{task=job.WebhookTask}
//	@Failure	400	{object}	errorResponse
//	@Router		/jobs/webhook [post]
func (h *WebhookHandler) Create(w http.ResponseWriter, r *http.Request) {
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

// List godoc
//
//	@Summary	List webhook jobs
//	@Tags		webhook jobs
//	@Produce	json
//	@Success	200	{array}		jobResponse{task=job.WebhookTask}
//	@Failure	500	{object}	errorResponse
//	@Router		/jobs/webhook [get]
func (h *WebhookHandler) List(w http.ResponseWriter, r *http.Request) {
	jobs, err := h.service.ListJobs(r.Context(), job.ListFilter{TaskType: job.TaskWebhook})
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toJobResponses(jobs))
}

// Get godoc
//
//	@Summary	Get a webhook job
//	@Tags		webhook jobs
//	@Produce	json
//	@Param		id	path		string	true	"Job ID"
//	@Success	200	{object}	jobResponse{task=job.WebhookTask}
//	@Failure	404	{object}	errorResponse
//	@Router		/jobs/webhook/{id} [get]
func (h *WebhookHandler) Get(w http.ResponseWriter, r *http.Request) {
	j, err := h.find(r.Context(), r.PathValue("id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, toJobResponse(j))
}

// Update godoc
//
//	@Summary	Replace a webhook job
//	@Tags		webhook jobs
//	@Accept		json
//	@Produce	json
//	@Param		id	path		string				true	"Job ID"
//	@Param		job	body		webhookJobRequest	true	"Webhook job"
//	@Success	200	{object}	jobResponse{task=job.WebhookTask}
//	@Failure	400	{object}	errorResponse
//	@Failure	404	{object}	errorResponse
//	@Router		/jobs/webhook/{id} [put]
func (h *WebhookHandler) Update(w http.ResponseWriter, r *http.Request) {
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

// Delete godoc
//
//	@Summary	Delete a webhook job
//	@Tags		webhook jobs
//	@Param		id	path	string	true	"Job ID"
//	@Success	204
//	@Failure	404	{object}	errorResponse
//	@Router		/jobs/webhook/{id} [delete]
func (h *WebhookHandler) Delete(w http.ResponseWriter, r *http.Request) {
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

// Pause godoc
//
//	@Summary	Pause a webhook job
//	@Tags		webhook jobs
//	@Produce	json
//	@Param		id	path		string	true	"Job ID"
//	@Success	200	{object}	jobResponse{task=job.WebhookTask}
//	@Failure	404	{object}	errorResponse
//	@Router		/jobs/webhook/{id}/pause [post]
func (h *WebhookHandler) Pause(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.service.PauseJob)
}

// Resume godoc
//
//	@Summary	Resume a paused webhook job
//	@Tags		webhook jobs
//	@Produce	json
//	@Param		id	path		string	true	"Job ID"
//	@Success	200	{object}	jobResponse{task=job.WebhookTask}
//	@Failure	404	{object}	errorResponse
//	@Router		/jobs/webhook/{id}/resume [post]
func (h *WebhookHandler) Resume(w http.ResponseWriter, r *http.Request) {
	h.changeStatus(w, r, h.service.ResumeJob)
}

func (h *WebhookHandler) changeStatus(w http.ResponseWriter, r *http.Request, change func(context.Context, string) (*job.Job, error)) {
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
func (h *WebhookHandler) find(ctx context.Context, id string) (*job.Job, error) {
	j, err := h.service.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}
	if j.Task.Type() != job.TaskWebhook {
		return nil, job.ErrNotFound
	}
	return j, nil
}
