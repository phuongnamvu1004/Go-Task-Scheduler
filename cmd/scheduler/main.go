package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/app"
	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/config"
)

// @title       Go Task Scheduler API
// @version     0.1.0
// @description Create and manage jobs that run tasks on a schedule.
// @BasePath    /
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx, cfg)
	if err != nil {
		log.Fatalf("start app: %v", err)
	}

	if err := a.Run(ctx); err != nil {
		log.Fatalf("run app: %v", err)
	}
}
