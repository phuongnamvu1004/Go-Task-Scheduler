package job

import (
	"fmt"
	"time"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// Job is the aggregate root: a Task that runs on a Schedule.
type Job struct {
	ID        string
	Name      string
	Schedule  string // e.g. a cron expression
	Task      task.Task
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Status string

const (
	StatusActive Status = "active"
	StatusPaused Status = "paused"
)

func (s Status) Valid() bool {
	return s == StatusActive || s == StatusPaused
}

// NewJob returns a validated, active Job. Its ID is set when it is saved.
func NewJob(name, schedule string, t task.Task) (*Job, error) {
	now := time.Now().UTC()
	j := &Job{
		Name:      name,
		Schedule:  schedule,
		Task:      t,
		Status:    StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := j.Validate(); err != nil {
		return nil, err
	}
	return j, nil
}

func (j *Job) Validate() error {
	if j.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalid)
	}
	if j.Schedule == "" {
		return fmt.Errorf("%w: schedule is required", ErrInvalid)
	}
	// TODO: parse and validate the schedule expression.
	if !j.Status.Valid() {
		return fmt.Errorf("%w: unknown status %q", ErrInvalid, j.Status)
	}
	if j.Task == nil {
		return fmt.Errorf("%w: task is required", ErrInvalid)
	}
	if err := j.Task.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalid, err)
	}
	return nil
}

// Update replaces the name, schedule and task, keeping the same task type.
// The job is unchanged if validation fails.
func (j *Job) Update(name, schedule string, t task.Task) error {
	if t != nil && j.Task != nil && t.Type() != j.Task.Type() {
		return fmt.Errorf("%w: task type cannot change from %q to %q", ErrInvalid, j.Task.Type(), t.Type())
	}

	updated := *j
	updated.Name = name
	updated.Schedule = schedule
	updated.Task = t
	if err := updated.Validate(); err != nil {
		return err
	}

	updated.UpdatedAt = time.Now().UTC()
	*j = updated
	return nil
}

// Pause stops the job from being scheduled.
func (j *Job) Pause() {
	j.Status = StatusPaused
	j.UpdatedAt = time.Now().UTC()
}

// Resume makes a paused job schedulable again.
func (j *Job) Resume() {
	j.Status = StatusActive
	j.UpdatedAt = time.Now().UTC()
}
