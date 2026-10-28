package postgres

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/lock"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/config"
)

// Embedded so the binary (and Docker image) carries its own migrations.
//
//go:embed migrations/*.sql
var migrations embed.FS

// NewPool connects to Postgres, uses cfg.Schema as the search_path, creates
// the schema if missing and applies pending migrations.
func NewPool(ctx context.Context, cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse database config: %w", err)
	}
	poolCfg.ConnConfig.RuntimeParams["search_path"] = cfg.Schema

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	schema := pgx.Identifier{cfg.Schema}.Sanitize()
	if _, err := pool.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+schema); err != nil {
		pool.Close()
		return nil, fmt.Errorf("create schema %s: %w", schema, err)
	}

	if err := migrate(ctx, pool); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

// migrate applies pending migrations. goose records applied versions in
// goose_db_version, inside cfg.Schema because of the search_path.
func migrate(ctx context.Context, pool *pgxpool.Pool) error {
	fsys, err := fs.Sub(migrations, "migrations")
	if err != nil {
		return err
	}
	// goose needs database/sql; this wraps the pool, and closing it leaves
	// the pool open.
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	// The advisory lock stops two instances starting at once from running
	// the same migration twice.
	locker, err := lock.NewPostgresSessionLocker()
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, fsys, goose.WithSessionLocker(locker))
	if err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}

	results, err := provider.Up(ctx)
	if err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	for _, r := range results {
		log.Printf("postgres: applied migration %s", r.Source.Path)
	}
	return nil
}
