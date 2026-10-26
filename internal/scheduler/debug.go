//go:build debug

// Debug-only tools for watching the scheduler, compiled in only with
// `-tags debug`. Delete this file and debug_off.go to remove them.

package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"sort"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// debugState lives on the Run goroutine like the pool.
type debugState struct {
	calls   chan func() // functions to run on the Run goroutine
	paused  bool
	pending []change
}

func (s *Scheduler) initDebug() { s.calls = make(chan func()) }

func (s *Scheduler) debugCalls() <-chan func() { return s.calls }

// handle parks changes while paused instead of applying them.
func (s *Scheduler) handle(c change) {
	if s.paused {
		s.pending = append(s.pending, c)
		return
	}
	s.apply(c)
}

var errStopped = errors.New("scheduler is not running")

// onRun runs fn on the Run goroutine and waits for it, so fn can safely
// read and modify the pool and debug state.
func (s *Scheduler) onRun(ctx context.Context, fn func()) error {
	finished := make(chan struct{})
	select {
	case s.calls <- func() { fn(); close(finished) }:
	case <-s.done:
		return errStopped
	case <-ctx.Done():
		return ctx.Err()
	}
	<-finished
	return nil
}

func (s *Scheduler) RegisterDebugRoutes(r chi.Router) {
	log.Println("scheduler: debug routes enabled at /debug/scheduler")
	r.Route("/debug/scheduler", func(r chi.Router) {
		r.Post("/pause", s.debugPause)
		r.Post("/resume", s.debugResume)
		r.Get("/pending", s.debugPending)
		r.Get("/pool", s.debugPool)
	})
}

type debugStatusResponse struct {
	Paused    bool `json:"paused"`
	Pending   int  `json:"pending"`
	Applied   int  `json:"applied,omitempty"`
	InChannel int  `json:"in_channel"`
	PoolSize  int  `json:"pool_size"`
}

type debugPendingResponse struct {
	Paused    bool                  `json:"paused"`
	InChannel int                   `json:"in_channel"` // received by nobody yet; should be 0
	Pending   []debugChangeResponse `json:"pending"`
}

type debugChangeResponse struct {
	Action    string     `json:"action"` // "changed" or "deleted"
	JobID     string     `json:"job_id"`
	Name      string     `json:"name,omitempty"`
	Status    job.Status `json:"status,omitempty"`
	Schedule  string     `json:"schedule,omitempty"`
	Task      task.Task  `json:"task,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type debugPoolResponse struct {
	Size int                      `json:"size"`
	Jobs []debugPoolEntryResponse `json:"jobs"`
}

type debugPoolEntryResponse struct {
	JobID     string     `json:"job_id"`
	Name      string     `json:"name"`
	Status    job.Status `json:"status"`
	Schedule  string     `json:"schedule"`
	Task      task.Task  `json:"task"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (s *Scheduler) debugPause(w http.ResponseWriter, r *http.Request) {
	var resp debugStatusResponse
	err := s.onRun(r.Context(), func() {
		s.paused = true
		resp = s.status()
	})
	writeDebug(w, resp, err)
}

// debugResume applies the parked changes in arrival order.
func (s *Scheduler) debugResume(w http.ResponseWriter, r *http.Request) {
	var resp debugStatusResponse
	err := s.onRun(r.Context(), func() {
		for _, c := range s.pending {
			s.apply(c)
		}
		applied := len(s.pending)
		s.pending = nil
		s.paused = false
		resp = s.status()
		resp.Applied = applied
	})
	writeDebug(w, resp, err)
}

func (s *Scheduler) debugPending(w http.ResponseWriter, r *http.Request) {
	var resp debugPendingResponse
	err := s.onRun(r.Context(), func() {
		resp = debugPendingResponse{
			Paused:    s.paused,
			InChannel: len(s.changes),
			Pending:   make([]debugChangeResponse, 0, len(s.pending)),
		}
		for _, c := range s.pending {
			resp.Pending = append(resp.Pending, toDebugChange(c))
		}
	})
	writeDebug(w, resp, err)
}

func (s *Scheduler) debugPool(w http.ResponseWriter, r *http.Request) {
	var resp debugPoolResponse
	err := s.onRun(r.Context(), func() {
		resp = debugPoolResponse{Size: len(s.pool), Jobs: make([]debugPoolEntryResponse, 0, len(s.pool))}
		for _, e := range s.pool {
			j := e.job
			resp.Jobs = append(resp.Jobs, debugPoolEntryResponse{
				JobID: j.ID, Name: j.Name, Status: j.Status, Schedule: j.Schedule, Task: j.Task, UpdatedAt: j.UpdatedAt,
			})
		}
	})
	sort.Slice(resp.Jobs, func(i, k int) bool { return resp.Jobs[i].Name < resp.Jobs[k].Name })
	writeDebug(w, resp, err)
}

// status must run on the Run goroutine.
func (s *Scheduler) status() debugStatusResponse {
	return debugStatusResponse{Paused: s.paused, Pending: len(s.pending), InChannel: len(s.changes), PoolSize: len(s.pool)}
}

func toDebugChange(c change) debugChangeResponse {
	if c.deletedID != "" {
		return debugChangeResponse{Action: "deleted", JobID: c.deletedID}
	}
	j := c.job
	return debugChangeResponse{
		Action: "changed", JobID: j.ID, Name: j.Name, Status: j.Status,
		Schedule: j.Schedule, Task: j.Task, UpdatedAt: &j.UpdatedAt,
	}
}

func writeDebug(w http.ResponseWriter, v any, err error) {
	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		v = map[string]string{"error": err.Error()}
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write response: %v", err)
	}
}
