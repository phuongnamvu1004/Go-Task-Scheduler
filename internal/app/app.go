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

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/config"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job"
	jobhttp "github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/http"
	jobpostgres "github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/postgres"
	platformhttp "github.com/phuongnamvu1004/Go-Task-Scheduler/internal/platform/http"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/platform/postgres"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/scheduler"
)

// App wires all components together and manages their lifecycle.
type App struct {
	cfg       config.Config
	db        *pgxpool.Pool
	scheduler *scheduler.Scheduler
	server    *http.Server
}

// New builds every component and registers the HTTP routes.
func New(ctx context.Context, cfg config.Config) (*App, error) {
	db, err := postgres.NewPool(ctx, cfg.Database)
	if err != nil {
		return nil, err
	}

	jobRepo := jobpostgres.NewRepository(db)
	sched := scheduler.New(jobRepo)
	jobService := job.NewService(jobRepo, sched)

	router := chi.NewRouter()
	router.Use(middleware.Logger, middleware.Recoverer)
	platformhttp.NewHealthHandler(map[string]platformhttp.CheckFunc{
		"postgres": db.Ping,
	}).RegisterRoutes(router)
	platformhttp.NewDocsHandler().RegisterRoutes(router)
	jobhttp.NewHandler(jobService).RegisterRoutes(router)
	sched.RegisterDebugRoutes(router) // no-op unless built with -tags debug

	return &App{
		cfg:       cfg,
		db:        db,
		scheduler: sched,
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

	// Load before serving, so no change can arrive before the pool is filled.
	if err := a.scheduler.Load(ctx); err != nil {
		return err
	}
	go a.scheduler.Run(ctx)

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
