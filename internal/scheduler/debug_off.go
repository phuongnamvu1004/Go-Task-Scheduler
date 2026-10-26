//go:build !debug

package scheduler

import "github.com/go-chi/chi/v5"

// Normal builds: no debug state, no debug routes. See debug.go.

type debugState struct{}

func (s *Scheduler) initDebug() {}

func (s *Scheduler) handle(c change) { s.apply(c) }

// debugCalls returns a nil channel, which is never ready in a select.
func (s *Scheduler) debugCalls() <-chan func() { return nil }

func (s *Scheduler) RegisterDebugRoutes(chi.Router) {}
