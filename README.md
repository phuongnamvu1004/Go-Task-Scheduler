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

## Make targets

Common tasks are in the `Makefile`; `make` lists them:

| Target | Does |
|---|---|
| `make run` / `make run-debug` | Run the app (debug adds `/debug/scheduler/*`) |
| `make check` | Build and vet, with and without the `debug` tag |
| `make swag` | Regenerate the Swagger docs |
| `make db-up` / `make db-down` | Start / stop local Postgres |
| `make migrate-create name=...` | Create the next migration file |
| `make migrate-up` / `make migrate-status` | Apply pending migrations / show which are applied |
| `make migrate-down` | Undo the latest migration (prints a warning, 5s to cancel) |

The `migrate-*` targets read the same `DB_*` settings from `.env` as the app.

## Local Postgres

A Postgres server for development is defined in `compose.yaml`. It uses the same `.env` credentials as the app:

```sh
docker compose up -d        # start Postgres in the background
docker compose ps           # check it's running
docker compose down         # stop it (data is kept in a volume)
docker compose down -v      # stop it and delete all data
```

## Database migrations

The schema is managed with [goose](https://github.com/pressly/goose). Migrations are SQL files in
`internal/platform/postgres/migrations/`, embedded into the binary and applied automatically on
startup, before the app loads any jobs. goose records applied versions in a `goose_db_version` table
in `DB_SCHEMA`, so each migration runs once per database.

goose does not compare your Go structs with the database: each change is a SQL file you write.
To change the schema, add a new file with the next number (never edit one that has been applied):

```sh
make migrate-create name=add_something    # creates e.g. 00002_add_something.sql
```

Write the SQL under `-- +goose Up` (and the undo under `-- +goose Down`), update the repository
queries in `internal/job/postgres/`, then restart the app or run `make migrate-up`.
`make migrate-status` shows which migrations a database has applied.

If the database is reset (e.g. `docker compose down -v`), the migrations rebuild every table on the
next start. The data itself is not recovered.

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
or a descriptor. It is evaluated in the job's `timezone`, an IANA name such as `Europe/Berlin` or
`America/New_York`; leave it out (or send `""` in a PATCH) for **UTC**.

```json
{ "name": "Morning report", "schedule": "0 9 * * 1-5", "timezone": "Europe/Berlin", "task": { ... } }
```

| Schedule | Runs |
| --- | --- |
| `*/5 * * * *` | every 5 minutes |
| `0 9 * * 1-5` | 09:00 in the job's time zone on weekdays |
| `@every 30s` | every 30 seconds |
| `@daily` / `@hourly` | midnight in the job's time zone / on the hour |

Daylight saving time follows the zone's rules, which has side effects:

- A run inside the hour that is skipped when clocks go forward (e.g. `30 2 * * *` in `Europe/Berlin` on
  the last Sunday of March) does not happen that day.
- A run inside the hour that repeats when clocks go back may happen twice.
- `@every` intervals ignore time zones and DST; they count real elapsed time.

Use `UTC`, or times outside 02:00–03:00, to avoid both. An unknown zone, `Local`, or a `CRON_TZ=`
prefix in the schedule is rejected with `400`; so is an invalid schedule. Seconds fields are not supported. A run is skipped if the
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
    postgres/               # connection pool, migrations/ (goose), seed data
    http/                   # health check and Swagger UI routes
```
