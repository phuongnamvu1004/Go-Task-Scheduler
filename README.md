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
curl localhost:8080/health    # status of each service (503 if any is down)
# {"status":"up","services":{"postgres":{"status":"up"}}}
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

## Debug build

Scheduler debug endpoints are compiled in only with the `debug` build tag; normal builds don't contain them.

```sh
go run -tags debug ./cmd/scheduler
```

| Endpoint | What it does |
|---|---|
| `POST /debug/scheduler/pause` | Stop applying job changes; new changes are parked in a pending list |
| `GET /debug/scheduler/pending` | Parked changes, in arrival order |
| `GET /debug/scheduler/pool` | Jobs currently in the scheduler pool, with their cron entry ID and next run time |
| `POST /debug/scheduler/resume` | Apply the parked changes in order and stop pausing |

The code is in `internal/scheduler/debug.go` (`//go:build debug`) with no-op stand-ins in `debug_off.go`.
To have your editor analyze `debug.go`, add the tag: GoLand → *Settings › Go › Build Tags*,
VS Code → `"gopls": {"buildFlags": ["-tags=debug"]}`.

## Building

```sh
go build -o bin/scheduler ./cmd/scheduler
./bin/scheduler
```

## Schedules

A job's `schedule` is a standard 5-field cron expression (minute, hour, day of month, month, day of week)
or a descriptor, evaluated in **UTC**:

| Schedule | Runs |
| --- | --- |
| `*/5 * * * *` | every 5 minutes |
| `0 9 * * 1-5` | 09:00 UTC on weekdays |
| `@every 30s` | every 30 seconds |
| `@daily` / `@hourly` | midnight UTC / on the hour |

An invalid schedule is rejected with `400`. Seconds fields are not supported. A run is skipped if the
previous run of the same job is still in progress, and runs missed while the app was down are not
caught up. Jobs currently only log when they fire; execution comes later.

## Project structure

Code is organized by feature: each feature folder holds its domain model, use cases and adapters.
Shared technical setup with no business logic lives in `platform/`.

```
cmd/scheduler/              # entry point: load config, handle signals, run the app
docs/                       # generated Swagger spec (do not edit by hand)
internal/
  app/                      # wires all components and manages their lifecycle
  config/                   # Config struct, loaded from env / .env
  job/                      # job feature: Job, Repository interface, Service, ChangeListener
    task/                   # Task interface + one file per task type (webhook.go, ...)
    postgres/               # Postgres implementation of job.Repository
    http/                   # /jobs routes: <type>_handler.go + <type>_dto.go per task type, helpers.go
  scheduler/                # pool of active jobs registered with cron (scheduler.go), one handler per task type (handler.go)
  platform/
    postgres/               # connection pool, schema.sql, seed data
    http/                   # health check and Swagger UI routes
```
