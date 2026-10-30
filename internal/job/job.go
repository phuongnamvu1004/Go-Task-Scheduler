package job

import (
	"fmt"
	"strings"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// Job is the aggregate root: a Task that runs on a Schedule.
type Job struct {
	ID        string
	Name      string
	Schedule  string // standard 5-field cron expression or descriptor, e.g. "*/5 * * * *", "@every 1m"
	Timezone  string // IANA name the schedule is evaluated in, e.g. "Europe/Berlin"
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

// DefaultTimezone is used when a job is created or updated without one.
const DefaultTimezone = "UTC"

func timezoneOrDefault(tz string) string {
	if tz == "" {
		return DefaultTimezone
	}
	return tz
}

// NewJob returns a validated, active Job. An empty timezone means
// DefaultTimezone. Its ID is set when it is saved.
func NewJob(name, schedule, timezone string, t task.Task) (*Job, error) {
	now := time.Now().UTC()
	j := &Job{
		Name:      name,
		Schedule:  schedule,
		Timezone:  timezoneOrDefault(timezone),
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
	// The zone has its own field; a prefix in the schedule would conflict with it.
	if strings.HasPrefix(j.Schedule, "CRON_TZ=") || strings.HasPrefix(j.Schedule, "TZ=") {
		return fmt.Errorf("%w: set the time zone with the timezone field, not in the schedule", ErrInvalid)
	}
	// "Local" would depend on the server's settings, and LoadLocation treats "" as UTC.
	if j.Timezone == "" || j.Timezone == "Local" {
		return fmt.Errorf("%w: timezone must be an IANA name such as \"UTC\" or \"Europe/Berlin\"", ErrInvalid)
	}
	if _, err := time.LoadLocation(j.Timezone); err != nil {
		return fmt.Errorf("%w: unknown timezone %q", ErrInvalid, j.Timezone)
	}
	if _, err := cron.ParseStandard(j.CronSpec()); err != nil {
		return fmt.Errorf("%w: invalid schedule %q: %w", ErrInvalid, j.Schedule, err)
	}
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

// CronSpec is the schedule in robfig/cron's format, with the job's time zone.
func (j *Job) CronSpec() string {
	return "CRON_TZ=" + j.Timezone + " " + j.Schedule
}

// Update replaces the name, schedule, timezone and task, keeping the same
// task type. An empty timezone means DefaultTimezone. The job is unchanged
// if validation fails.
func (j *Job) Update(name, schedule, timezone string, t task.Task) error {
	if t != nil && j.Task != nil && t.Type() != j.Task.Type() {
		return fmt.Errorf("%w: task type cannot change from %q to %q", ErrInvalid, j.Task.Type(), t.Type())
	}

	updated := *j
	updated.Name = name
	updated.Schedule = schedule
	updated.Timezone = timezoneOrDefault(timezone)
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
