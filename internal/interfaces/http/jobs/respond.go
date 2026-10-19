package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

const maxBodyBytes = 1 << 20 // 1 MiB

// jobResponse is the JSON shape of a job for every task type.
type jobResponse struct {
	ID        string       `json:"id"`
	Name      string       `json:"name"`
	Schedule  string       `json:"schedule"`
	Status    job.Status   `json:"status"`
	TaskType  job.TaskType `json:"task_type"`
	Task      job.Task     `json:"task"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func toJobResponse(j *job.Job) jobResponse {
	return jobResponse{
		ID:        j.ID,
		Name:      j.Name,
		Schedule:  j.Schedule,
		Status:    j.Status,
		TaskType:  j.Task.Type(),
		Task:      j.Task,
		CreatedAt: j.CreatedAt,
		UpdatedAt: j.UpdatedAt,
	}
}

func toJobResponses(jobs []*job.Job) []jobResponse {
	out := make([]jobResponse, 0, len(jobs))
	for _, j := range jobs {
		out = append(out, toJobResponse(j))
	}
	return out
}

// decodeJSON rejects unknown fields and bodies over maxBodyBytes.
func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("invalid JSON body: %w", err)
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}

// writeError maps domain errors to status codes. Unexpected errors are
// logged and hidden from the client.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, job.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
	case errors.Is(err, job.ErrNotFound):
		writeJSON(w, http.StatusNotFound, errorResponse{Error: err.Error()})
	default:
		log.Printf("%s %s: %v", r.Method, r.RequestURI, err)
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal server error"})
	}
}

func writeBadRequest(w http.ResponseWriter, err error) {
	writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
}
