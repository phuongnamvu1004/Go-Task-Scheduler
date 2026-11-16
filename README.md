# Go Task Scheduler

A service that stores jobs (a task plus a cron schedule) in Postgres and runs them in the background
with a fixed pool of workers. Jobs are managed through a REST API.

## Requirements

- Go 1.26+
- PostgreSQL (or Docker, to run the bundled `docker-compose.yaml`)

## Quick start

```sh
cp .env.example .env    # then set your Postgres credentials
make db-up              # start local Postgres
make run                # app on http://localhost:8080, API docs at /swagger
```

## Documentation

| Guide | Covers |
|---|---|
| [Setup](docs/setup.md) | Configuration, local Postgres, migrations, running, building, Swagger, debug build |
| [Schedules](docs/schedules.md) | Cron syntax, time zones, daylight saving time |
| [Worker pool](docs/worker-pool.md) | How due jobs are queued and executed, `WORKERS`, logs |
| [CHANGELOG](CHANGELOG.md) | Every change and rename |

## Project structure

Code is organized by feature: each feature folder holds its domain model, use cases and adapters.
Shared technical setup with no business logic lives in `platform/`.

```
cmd/scheduler/              # entry point: load config, handle signals, run the app
docs/                       # guides (*.md)
  swagger/                  # generated Swagger spec (do not edit by hand)
internal/
  app/                      # wires all components and manages their lifecycle
  config/                   # Config struct, loaded from env / .env
  job/                      # job feature: Job, Repository interface, Service, ChangeListener
    task/                   # Task interface + one file per task type (webhook.go, ...)
    postgres/               # Postgres implementation of job.Repository
    http/                   # /jobs routes: <type>_handler.go + <type>_dto.go per task type, helpers.go
  scheduler/                # pool of active jobs registered with cron (scheduler.go),
                            # worker pool that executes due jobs (worker.go),
                            # one handler per task type (handler.go)
  platform/
    postgres/               # connection pool, migrations/ (goose), seed data
    http/                   # health check and Swagger UI routes
```
