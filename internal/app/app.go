package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/application"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/config"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/infrastructure/persistence/postgres"
	httpapi "github.com/phuongnamvu1004/Go-Task-Scheduler/internal/interfaces/http"
)

// App wires all components together and manages their lifecycle.
type App struct {
	cfg    config.Config
	db     *pgxpool.Pool
	server *http.Server
}

// New builds every component and registers the HTTP routes.
func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	jobRepo := postgres.NewJobRepository(db)
	jobService := application.NewJobService(jobRepo)

	router := chi.NewRouter()
	router.Use(middleware.Logger, middleware.Recoverer)
	httpapi.NewHealthHandler().RegisterRoutes(router)
	httpapi.NewJobHandler(jobService).RegisterRoutes(router)

	return &App{
		cfg: cfg,
		db:  db,
		server: &http.Server{
			Addr:              ":" + cfg.HTTP.Port,
			Handler:           router,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}, nil
}

// Run serves until ctx is cancelled or the server fails, then shuts down
// gracefully.
func (a *App) Run(ctx context.Context) error {
	defer a.db.Close()

	// TODO: start the scheduler loop here.

	errCh := make(chan error, 1)
	go func() {
		log.Printf("task scheduler listening on %s", a.server.Addr)
		errCh <- a.server.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
	}

	log.Println("shutting down")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return nil
}
