package scheduler

import (
	"log"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// One handler per task type, called by fire when a job is due. A new task
// type needs a handler here and a case in fire.

func handleWebhook(j *job.Job, t *task.Webhook) {
	log.Printf("scheduler: job %s (%s) fired: webhook %s %s", j.ID, j.Name, t.Method, t.URL)
	// TODO: send the request (roadmap step 3).
}
