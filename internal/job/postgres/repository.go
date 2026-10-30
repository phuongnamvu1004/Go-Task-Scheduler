package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// Repository is the Postgres implementation of job.Repository.
// The task is stored as task_type + task_config (JSONB).
type Repository struct {
	pool *pgxpool.Pool
}

var _ job.Repository = (*Repository)(nil)

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const jobColumns = `id::text, name, schedule, timezone, status, task_type, task_config, created_at, updated_at`

func (r *Repository) Create(ctx context.Context, j *job.Job) error {
	config, err := encodeTask(j.Task)
	if err != nil {
		return err
	}

	err = r.pool.QueryRow(ctx, `
		INSERT INTO jobs (name, schedule, timezone, status, task_type, task_config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text`,
		j.Name, j.Schedule, j.Timezone, j.Status, j.Task.Type(), config, j.CreatedAt, j.UpdatedAt,
	).Scan(&j.ID)
	if err != nil {
		return fmt.Errorf("insert job: %w", err)
	}
	return nil
}

func (r *Repository) Update(ctx context.Context, j *job.Job) error {
	config, err := encodeTask(j.Task)
	if err != nil {
		return err
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE jobs
		SET name = $2, schedule = $3, timezone = $4, status = $5, task_type = $6, task_config = $7, updated_at = $8
		WHERE id = $1`,
		j.ID, j.Name, j.Schedule, j.Timezone, j.Status, j.Task.Type(), config, j.UpdatedAt,
	)
	if err != nil {
		return notFoundOr(fmt.Errorf("update job: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return job.ErrNotFound
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*job.Job, error) {
	row := r.pool.QueryRow(ctx, `SELECT `+jobColumns+` FROM jobs WHERE id = $1`, id)
	j, err := scanJob(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, job.ErrNotFound
	}
	if err != nil {
		return nil, notFoundOr(fmt.Errorf("find job: %w", err))
	}
	return j, nil
}

func (r *Repository) FindAll(ctx context.Context, filter job.ListFilter) ([]*job.Job, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT `+jobColumns+` FROM jobs
		WHERE ($1 = '' OR task_type = $1)
		  AND ($2 = '' OR status = $2)
		ORDER BY created_at`,
		string(filter.TaskType), string(filter.Status),
	)
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	defer rows.Close()

	jobs := []*job.Job{}
	for rows.Next() {
		j, err := scanJob(rows)
		if err != nil {
			return nil, fmt.Errorf("list jobs: %w", err)
		}
		jobs = append(jobs, j)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	return jobs, nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM jobs WHERE id = $1`, id)
	if err != nil {
		return notFoundOr(fmt.Errorf("delete job: %w", err))
	}
	if tag.RowsAffected() == 0 {
		return job.ErrNotFound
	}
	return nil
}

func scanJob(row pgx.Row) (*job.Job, error) {
	var (
		j        job.Job
		taskType task.Type
		config   []byte
	)
	err := row.Scan(&j.ID, &j.Name, &j.Schedule, &j.Timezone, &j.Status, &taskType, &config, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}

	j.Task, err = decodeTask(taskType, config)
	if err != nil {
		return nil, err
	}
	return &j, nil
}

// notFoundOr treats a malformed (non-UUID) ID as job.ErrNotFound.
func notFoundOr(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "22P02" { // invalid_text_representation
		return job.ErrNotFound
	}
	return err
}
