package persistence

import (
	"context"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

// InMemoryJobRepository is a placeholder implementation of job.Repository.
type InMemoryJobRepository struct{}

var _ job.Repository = (*InMemoryJobRepository)(nil)

func NewInMemoryJobRepository() *InMemoryJobRepository {
	return &InMemoryJobRepository{}
}

func (r *InMemoryJobRepository) Save(ctx context.Context, j *job.Job) error {
	// TODO: implement
	return nil
}

func (r *InMemoryJobRepository) FindByID(ctx context.Context, id string) (*job.Job, error) {
	// TODO: implement
	return nil, nil
}

func (r *InMemoryJobRepository) FindAll(ctx context.Context) ([]*job.Job, error) {
	// TODO: implement
	return nil, nil
}

func (r *InMemoryJobRepository) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}
