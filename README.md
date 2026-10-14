# Go Task Scheduler

An application that takes in jobs and runs them on a given schedule.

## Requirements

- Go 1.26+
- PostgreSQL

## Configuration

Configuration is read from environment variables. A `.env` file in the project root is loaded automatically if present (variables already set in the environment win).

```sh
cp .env.example .env
# then edit .env with your Postgres credentials
```

| Variable      | Default             | Description                          |
|---------------|---------------------|--------------------------------------|
| `PORT`        | `8080`              | HTTP listening port                  |
| `DB_HOST`     | `localhost`         | Postgres host                        |
| `DB_PORT`     | `5432`              | Postgres port                        |
| `DB_USER`     | —                   | Postgres user (required)             |
| `DB_PASSWORD` | —                   | Postgres password                    |
| `DB_NAME`     | —                   | Postgres database (required)         |
| `DB_SCHEMA`   | `go_task_scheduler` | Schema used by the app (created on startup if missing) |
| `DB_SSLMODE`  | `disable`           | Postgres SSL mode                    |

## Local Postgres

A Postgres server for development is defined in `compose.yaml`. It uses the same `.env` credentials as the app:

```sh
docker compose up -d        # start Postgres in the background
docker compose ps           # check it's running
docker compose down         # stop it (data is kept in a volume)
docker compose down -v      # stop it and delete all data
```

## Running

Make sure Postgres is running, then from the project root:

```sh
go run ./cmd/scheduler
```

The server listens on port `8080` by default. To use a different port, set `PORT`:

```sh
PORT=9090 go run ./cmd/scheduler
```

Check that it's up:

```sh
curl localhost:8080/healthz
# ok
```

Stop it with `Ctrl+C`; the server shuts down gracefully.

## Building

```sh
go build -o bin/scheduler ./cmd/scheduler
./bin/scheduler
```

## Project structure

The project follows a Domain-Driven Design layout:

```
cmd/scheduler/                        # entry point: load config, handle signals, run the app
internal/
  app/                                # wires all components and manages their lifecycle
  config/                             # Config struct, loaded from env / .env
  domain/job/                         # Job aggregate and Repository interface
  application/                        # use cases (JobService)
  infrastructure/persistence/postgres/  # Postgres connection pool and Repository implementation
  interfaces/http/                    # HTTP handlers and routes (health, jobs)
```
