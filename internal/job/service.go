package job

import (
	"context"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// Service holds the job use cases. It works with any Task type.
type Service struct {
	repo     Repository
	listener ChangeListener
}

// NewService takes an optional listener (nil is allowed) that is notified
// after every successful change.
func NewService(repo Repository, listener ChangeListener) *Service {
	if listener == nil {
		listener = noopListener{}
	}
	return &Service{repo: repo, listener: listener}
}

type CreateInput struct {
	Name     string
	Schedule string
	Task     task.Task
}

type UpdateInput struct {
	Name     string
	Schedule string
	Task     task.Task
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Job, error) {
	j, err := NewJob(in.Name, in.Schedule, in.Task)
	if err != nil {
		return nil, err
	}
	if err := s.repo.Create(ctx, j); err != nil {
		return nil, err
	}
	s.listener.JobChanged(j)
	return j, nil
}

func (s *Service) Get(ctx context.Context, id string) (*Job, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) List(ctx context.Context, filter ListFilter) ([]*Job, error) {
	return s.repo.FindAll(ctx, filter)
}

func (s *Service) Update(ctx context.Context, id string, in UpdateInput) (*Job, error) {
	return s.modify(ctx, id, func(j *Job) error {
		return j.Update(in.Name, in.Schedule, in.Task)
	})
}

func (s *Service) Pause(ctx context.Context, id string) (*Job, error) {
	return s.modify(ctx, id, func(j *Job) error {
		j.Pause()
		return nil
	})
}

func (s *Service) Resume(ctx context.Context, id string) (*Job, error) {
	return s.modify(ctx, id, func(j *Job) error {
		j.Resume()
		return nil
	})
}

func (s *Service) Delete(ctx context.Context, id string) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.listener.JobDeleted(id)
	return nil
}

// modify loads a job, applies change, saves it and notifies the listener.
func (s *Service) modify(ctx context.Context, id string, change func(*Job) error) (*Job, error) {
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
	s.listener.JobChanged(j)
	return j, nil
}
