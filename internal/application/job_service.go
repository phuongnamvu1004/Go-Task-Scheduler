package application

import (
	"context"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

// JobService holds the job use cases. It works with any Task type.
type JobService struct {
	repo job.Repository
}

func NewJobService(repo job.Repository) *JobService {
	return &JobService{repo: repo}
}

type CreateJobInput struct {
	Name     string
	Schedule string
	Task     job.Task
}

type UpdateJobInput struct {
	Name     string
	Schedule string
	Task     job.Task
}

func (s *JobService) CreateJob(ctx context.Context, in CreateJobInput) (*job.Job, error) {
	j, err := job.NewJob(in.Name, in.Schedule, in.Task)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, j); err != nil {
		return nil, err
	}
	return j, nil
}

func (s *JobService) GetJob(ctx context.Context, id string) (*job.Job, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *JobService) ListJobs(ctx context.Context, filter job.ListFilter) ([]*job.Job, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *JobService) UpdateJob(ctx context.Context, id string, in UpdateJobInput) (*job.Job, error) {
	return s.modify(ctx, id, func(j *job.Job) error {
		return j.Update(in.Name, in.Schedule, in.Task)
	})
}

func (s *JobService) PauseJob(ctx context.Context, id string) (*job.Job, error) {
	return s.modify(ctx, id, func(j *job.Job) error {
		j.Pause()
		return nil
	})
}

func (s *JobService) ResumeJob(ctx context.Context, id string) (*job.Job, error) {
	return s.modify(ctx, id, func(j *job.Job) error {
		j.Resume()
		return nil
	})
}

func (s *JobService) DeleteJob(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

// modify loads a job, applies change, and saves it.
func (s *JobService) modify(ctx context.Context, id string, change func(*job.Job) error) (*job.Job, error) {
	j, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if err := change(j); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, j); err != nil {
		return nil, err
	}
	return j, nil
}
