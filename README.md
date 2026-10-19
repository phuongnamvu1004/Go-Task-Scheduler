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

## API docs (Swagger)

With the app running, open the Swagger UI at
[http://localhost:8080/swagger](http://localhost:8080/swagger) (same port as the API, `PORT`).

The spec in `docs/` is generated from the `@Summary`, `@Param`, `@Router`, ... comments on the handlers.
After changing an endpoint or its request/response types, regenerate it:

```sh
go tool swag init -g cmd/scheduler/main.go --parseInternal -o docs
```

## Building

```sh
go build -o bin/scheduler ./cmd/scheduler
./bin/scheduler
```

## Project structure

Code is organized by feature: each feature folder holds its domain model, use cases and adapters.
Shared technical setup with no business logic lives in `platform/`.

```
cmd/scheduler/              # entry point: load config, handle signals, run the app
docs/                       # generated Swagger spec (do not edit by hand)
internal/
  app/                      # wires all components and manages their lifecycle
  config/                   # Config struct, loaded from env / .env
  job/                      # job feature: Job, Repository interface, Service
    task/                   # Task interface + one file per task type (webhook.go, ...)
    postgres/               # Postgres implementation of job.Repository
    http/                   # /jobs routes: one handler file per task type
  platform/
    postgres/               # connection pool, schema.sql, seed data
    http/                   # health check and Swagger UI routes
```
