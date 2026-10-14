package job

import "context"

// Repository defines persistence operations for Jobs.
// Implementations live in the infrastructure layer.
type Repository interface {
	Save(ctx context.Context, j *Job) error
	FindByID(ctx context.Context, id string) (*Job, error)
	FindAll(ctx context.Context) ([]*Job, error)
	Delete(ctx context.Context, id string) error
}
