package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// CheckFunc reports whether a dependency is reachable, e.g. pgxpool.Pool.Ping.
type CheckFunc func(ctx context.Context) error

const (
	checkTimeout = 2 * time.Second
	statusUp     = "up"
	statusDown   = "down"
)

type HealthHandler struct {
	checks map[string]CheckFunc
}

// NewHealthHandler takes one check per dependency, keyed by service name.
func NewHealthHandler(checks map[string]CheckFunc) *HealthHandler {
	return &HealthHandler{checks: checks}
}

func (h *HealthHandler) RegisterRoutes(r chi.Router) {
	r.Get("/health", h.Readiness)
}

type readinessResponse struct {
	Status   string                           `json:"status"`
	Services map[string]serviceStatusResponse `json:"services"`
}

type serviceStatusResponse struct {
	Status string `json:"status"`
}

// Readiness godoc
//
//	@Summary		Health status of each service
//	@Description	Returns 503 if any service is down. Error details are logged, not returned.
//	@Tags			health
//	@Produce		json
//	@Success		200	{object}	readinessResponse
//	@Failure		503	{object}	readinessResponse
//	@Router			/health [get]
func (h *HealthHandler) Readiness(w http.ResponseWriter, r *http.Request) {
	resp := readinessResponse{Status: statusUp, Services: make(map[string]serviceStatusResponse, len(h.checks))}
	code := http.StatusOK

	for name, check := range h.checks {
		ctx, cancel := context.WithTimeout(r.Context(), checkTimeout)
		err := check(ctx)
		cancel()

		if err != nil {
			log.Printf("readiness: %s is down: %v", name, err)
			resp.Services[name] = serviceStatusResponse{Status: statusDown}
			resp.Status = statusDown
			code = http.StatusServiceUnavailable
			continue
		}
		resp.Services[name] = serviceStatusResponse{Status: statusUp}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("write response: %v", err)
	}
}
