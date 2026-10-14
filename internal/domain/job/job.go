package job

import "time"

// Job is the aggregate root representing a unit of work run on a schedule.
type Job struct {
	ID        string
	Name      string
	Schedule  string // e.g. a cron expression
	Status    Status
	CreatedAt time.Time
	// TODO: add payload, next run time, last run result, etc.
}

// Status describes the lifecycle state of a Job.
type Status string

const (
	StatusActive Status = "active"
	StatusPaused Status = "paused"
)
