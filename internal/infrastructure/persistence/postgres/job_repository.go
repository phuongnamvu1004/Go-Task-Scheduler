package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

// JobRepository is the Postgres implementation of job.Repository.
type JobRepository struct {
	pool *pgxpool.Pool
}

var _ job.Repository = (*JobRepository)(nil)

func NewJobRepository(pool *pgxpool.Pool) *JobRepository {
	return &JobRepository{pool: pool}
}

func (r *JobRepository) Save(ctx context.Context, j *job.Job) error {
	// TODO: implement
	return nil
}

func (r *JobRepository) FindByID(ctx context.Context, id string) (*job.Job, error) {
	// TODO: implement
	return nil, nil
}

func (r *JobRepository) FindAll(ctx context.Context) ([]*job.Job, error) {
	// TODO: implement
	return nil, nil
}

func (r *JobRepository) Delete(ctx context.Context, id string) error {
	// TODO: implement
	return nil
}
