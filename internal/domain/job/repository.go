package job

import "context"

// ListFilter narrows FindAll results. Zero-value fields are ignored.
type ListFilter struct {
	TaskType TaskType
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
