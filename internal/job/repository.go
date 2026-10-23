package job

import (
	"context"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// ListFilter narrows FindAll results. Zero-value fields are ignored.
type ListFilter struct {
	TaskType task.Type
	Status   Status
}

// Repository persists Jobs. Update, FindByID and Delete return ErrNotFound
// for a missing job.
type Repository interface {
	// Create stores a new job and sets its ID.
	Create(ctx context.Context, j *Job) error
	Update(ctx context.Context, j *Job) error
	FindByID(ctx context.Context, id string) (*Job, error)
	FindAll(ctx context.Context, filter ListFilter) ([]*Job, error)
	Delete(ctx context.Context, id string) error
}
