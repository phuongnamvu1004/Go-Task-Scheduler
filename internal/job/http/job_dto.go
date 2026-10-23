package http

import (
	"time"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// jobResponse is the JSON shape of a job for every task type.
type jobResponse struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Schedule  string     `json:"schedule"`
	Status    job.Status `json:"status"`
	TaskType  task.Type  `json:"task_type"`
	Task      task.Task  `json:"task"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
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
