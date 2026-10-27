package scheduler

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/robfig/cron/v3"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

const changeBuffer = 100

// Scheduler keeps an in-memory pool of active jobs and registers each one
// with cron. The pool is filled by Load and kept in sync by change
// notifications from job.Service.
type Scheduler struct {
	repo    job.Repository
	changes chan change
	done    chan struct{} // closed when Run returns
	cron    *cron.Cron

	// pool is owned by the Run goroutine (and by Load before Run starts),
	// so it needs no lock.
	pool map[string]*entry

	// Pause/pending state for -tags debug builds; empty otherwise. See debug.go.
	debugState
}

type entry struct {
	job     *job.Job
	entryID cron.EntryID // needed to remove the job from cron later
}

// change is either an updated job or the ID of a deleted one.
type change struct {
	job       *job.Job
	deletedID string
}

var _ job.ChangeListener = (*Scheduler)(nil)

func New(repo job.Repository) *Scheduler {
	logger := cron.PrintfLogger(log.Default())
	s := &Scheduler{
		repo:    repo,
		changes: make(chan change, changeBuffer),
		done:    make(chan struct{}),
		cron: cron.New(
			cron.WithLocation(time.UTC),
			cron.WithChain(
				cron.Recover(logger),            // a panicking run is logged, not fatal
				cron.SkipIfStillRunning(logger), // no overlapping runs of one job
			),
		),
		pool: make(map[string]*entry),
	}
	s.initDebug()
	return s
}

// Load fills the pool with every active job. Call it before Run; nothing
// fires until Run starts cron.
func (s *Scheduler) Load(ctx context.Context) error {
	jobs, err := s.repo.FindAll(ctx, job.ListFilter{Status: job.StatusActive})
	if err != nil {
		return fmt.Errorf("load active jobs: %w", err)
	}

	s.pool = make(map[string]*entry, len(jobs))
	for _, j := range jobs {
		s.pool[j.ID] = &entry{job: j, entryID: s.schedule(j)}
	}
	log.Printf("scheduler: loaded %d active jobs", len(s.pool))
	return nil
}

// Run starts cron and applies change notifications to the pool, one at a
// time, until ctx is cancelled. It then stops cron and waits for runs in
// progress to finish.
func (s *Scheduler) Run(ctx context.Context) {
	defer close(s.done)
	s.cron.Start()
	defer func() { <-s.cron.Stop().Done() }()
	for {
		select {
		case c := <-s.changes:
			s.handle(c)
		case fn := <-s.debugCalls(): // never ready in normal builds
			fn()
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
		// cron can't update an entry, so replace it. A run already in
		// progress finishes with the old snapshot.
		s.cron.Remove(e.entryID)
		e.job = j
		e.entryID = s.schedule(j)
		log.Printf("scheduler: updated job %s", j.ID)
		return
	}
	s.pool[j.ID] = &entry{job: j, entryID: s.schedule(j)}
	log.Printf("scheduler: added job %s", j.ID)
}

func (s *Scheduler) remove(id, reason string) {
	e, ok := s.pool[id]
	if !ok {
		return
	}
	s.cron.Remove(e.entryID)
	delete(s.pool, id)
	log.Printf("scheduler: removed job %s (%s)", id, reason)
}

// schedule registers j with cron and returns its entry ID (0 if the
// schedule is invalid, which Job.Validate should already have prevented).
func (s *Scheduler) schedule(j *job.Job) cron.EntryID {
	id, err := s.cron.AddFunc(j.Schedule, func() { s.fire(j) })
	if err != nil {
		log.Printf("scheduler: job %s has invalid schedule %q: %v", j.ID, j.Schedule, err)
		return 0
	}
	return id
}

// fire runs on a cron goroutine, not on Run, so it must not touch the pool.
// j is the snapshot captured when the job was scheduled and is never modified.
// It passes the job to the handler for its task type (see handler.go).
func (s *Scheduler) fire(j *job.Job) {
	switch t := j.Task.(type) {
	case *task.Webhook:
		handleWebhook(j, t)
	default:
		log.Printf("scheduler: job %s has no handler for task type %q", j.ID, j.Task.Type())
	}
}
