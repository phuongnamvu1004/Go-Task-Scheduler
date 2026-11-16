package scheduler

import (
	"log"
	"time"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// runQueueSize is how many due runs can wait for a free worker. When it is
// full, new runs are skipped so cron's goroutine never blocks.
const runQueueSize = 100

// run is one queued execution of a job.
type run struct {
	job      *job.Job
	queuedAt time.Time // to log how long the run waited for a free worker
}

// startWorkers starts the fixed worker pool. Each worker takes runs until
// the runs channel is closed (see stopWorkers).
func (s *Scheduler) startWorkers() {
	for i := range s.workers {
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			for r := range s.runs {
				s.execute(i+1, r) // numbered from 1 in logs
			}
		}()
	}
}

// stopWorkers lets the workers finish the queued runs and waits for them.
// Call it only after cron has stopped: a send on a closed channel panics.
func (s *Scheduler) stopWorkers() {
	close(s.runs)
	s.wg.Wait()
}

// fire is called by cron when j is due. It runs on a cron goroutine, so it
// must not touch the pool; it only queues the run. j is the snapshot
// captured when the job was scheduled and is never modified.
func (s *Scheduler) fire(j *job.Job) {
	// One run per job at a time. Keyed by job ID rather than cron entry, so
	// a run of the old version also blocks the updated job.
	if _, busy := s.inFlight.LoadOrStore(j.ID, struct{}{}); busy {
		log.Printf("scheduler: job %s skipped: previous run still queued or running", j.ID)
		return
	}
	select {
	case s.runs <- run{job: j, queuedAt: time.Now()}:
	default:
		s.inFlight.Delete(j.ID)
		log.Printf("scheduler: job %s skipped: run queue full (%d)", j.ID, runQueueSize)
	}
}

// execute runs one job on worker w. A panic in a handler is logged;
// unrecovered, it would crash the whole process.
func (s *Scheduler) execute(w int, r run) {
	j := r.job
	start := time.Now()
	debugf("scheduler: worker %d started job %s (%s), waited %s in queue",
		w, j.ID, j.Name, start.Sub(r.queuedAt).Round(time.Millisecond))

	// Deferred in reverse order of running: recover, then log, then release.
	defer s.inFlight.Delete(j.ID)
	defer func() {
		debugf("scheduler: worker %d finished job %s in %s",
			w, j.ID, time.Since(start).Round(time.Millisecond))
	}()
	defer func() {
		if p := recover(); p != nil {
			log.Printf("scheduler: worker %d: job %s panicked: %v", w, j.ID, p)
		}
	}()
	dispatch(j)
}

// dispatch passes the job to the handler for its task type (see handler.go).
func dispatch(j *job.Job) {
	switch t := j.Task.(type) {
	case *task.Webhook:
		handleWebhook(j, t)
	default:
		log.Printf("scheduler: job %s has no handler for task type %q", j.ID, j.Task.Type())
	}
}
