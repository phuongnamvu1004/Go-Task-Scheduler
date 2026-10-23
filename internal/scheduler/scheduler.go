package scheduler

import (
	"context"
	"fmt"
	"log"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
)

const changeBuffer = 100

// Scheduler keeps an in-memory pool of active jobs. The pool is filled by
// Load and kept in sync by change notifications from job.Service.
type Scheduler struct {
	repo    job.Repository
	changes chan change
	done    chan struct{} // closed when Run returns

	// pool is owned by the Run goroutine (and by Load before Run starts),
	// so it needs no lock.
	pool map[string]*entry
}

type entry struct {
	job *job.Job
	// TODO: cron entry ID once robfig/cron is added.
}

// change is either an updated job or the ID of a deleted one.
type change struct {
	job       *job.Job
	deletedID string
}

var _ job.ChangeListener = (*Scheduler)(nil)

func New(repo job.Repository) *Scheduler {
	return &Scheduler{
		repo:    repo,
		changes: make(chan change, changeBuffer),
		done:    make(chan struct{}),
		pool:    make(map[string]*entry),
	}
}

// Load fills the pool with every active job. Call it before Run.
func (s *Scheduler) Load(ctx context.Context) error {
	jobs, err := s.repo.FindAll(ctx, job.ListFilter{Status: job.StatusActive})
	if err != nil {
		return fmt.Errorf("load active jobs: %w", err)
	}

	s.pool = make(map[string]*entry, len(jobs))
	for _, j := range jobs {
		s.pool[j.ID] = &entry{job: j}
	}
	log.Printf("scheduler: loaded %d active jobs", len(s.pool))
	return nil
}

// Run applies change notifications to the pool, one at a time, until ctx
// is cancelled.
func (s *Scheduler) Run(ctx context.Context) {
	defer close(s.done)
	for {
		select {
		case c := <-s.changes:
			s.apply(c)
		case <-ctx.Done():
			return
		}
	}
}

// JobChanged queues an updated job. It is called from HTTP request
// goroutines, so it only hands off a copy and never touches the pool.
func (s *Scheduler) JobChanged(j *job.Job) {
	snapshot := *j
	s.send(change{job: &snapshot})
}

func (s *Scheduler) JobDeleted(id string) {
	s.send(change{deletedID: id})
}

// send waits for room in the queue, or gives up once Run has stopped so
// requests can't hang during shutdown.
func (s *Scheduler) send(c change) {
	select {
	case s.changes <- c:
	case <-s.done:
	}
}

func (s *Scheduler) apply(c change) {
	if c.deletedID != "" {
		s.remove(c.deletedID, "deleted")
		return
	}
	if c.job.Status != job.StatusActive {
		s.remove(c.job.ID, string(c.job.Status))
		return
	}
	s.upsert(c.job)
}

func (s *Scheduler) upsert(j *job.Job) {
	if e, ok := s.pool[j.ID]; ok {
		// Two requests can finish out of order; keep the newer version.
		if e.job.UpdatedAt.After(j.UpdatedAt) {
			return
		}
		e.job = j
		log.Printf("scheduler: updated job %s", j.ID)
		return
	}
	s.pool[j.ID] = &entry{job: j}
	log.Printf("scheduler: added job %s", j.ID)
}

func (s *Scheduler) remove(id, reason string) {
	if _, ok := s.pool[id]; !ok {
		return
	}
	delete(s.pool, id)
	log.Printf("scheduler: removed job %s (%s)", id, reason)
}
