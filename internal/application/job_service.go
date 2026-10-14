package application

import "github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"

// JobService orchestrates use cases around Jobs (create, schedule, run, ...).
type JobService struct {
	repo job.Repository
}

func NewJobService(repo job.Repository) *JobService {
	return &JobService{repo: repo}
}

// TODO: add use cases, e.g. CreateJob, PauseJob, ListJobs.
