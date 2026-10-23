package http

import (
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

type webhookJobRequest struct {
	Name     string       `json:"name"`
	Schedule string       `json:"schedule"`
	Task     task.Webhook `json:"task"`
}

// webhookJobPatch is a partial update: fields left out of the body keep
// their current value.
type webhookJobPatchRequest struct {
	Name     *string                  `json:"name"`
	Schedule *string                  `json:"schedule"`
	Task     *webhookTaskPatchRequest `json:"task"`
}

type webhookTaskPatchRequest struct {
	URL     *string            `json:"url"`
	Method  *task.HTTPMethod   `json:"method"`
	Headers *map[string]string `json:"headers"`
	Body    *string            `json:"body"`
}

// apply returns j's current values with the patch laid over them.
func (p webhookJobPatchRequest) apply(j *job.Job) job.UpdateInput {
	in := job.UpdateInput{Name: j.Name, Schedule: j.Schedule}
	if p.Name != nil {
		in.Name = *p.Name
	}
	if p.Schedule != nil {
		in.Schedule = *p.Schedule
	}

	wt := *j.Task.(*task.Webhook)
	if t := p.Task; t != nil {
		if t.URL != nil {
			wt.URL = *t.URL
		}
		if t.Method != nil {
			wt.Method = *t.Method
		}
		if t.Headers != nil {
			wt.Headers = *t.Headers
		}
		if t.Body != nil {
			wt.Body = *t.Body
		}
	}
	in.Task = &wt
	return in
}
