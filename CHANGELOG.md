# Changelog

All notable changes to this project are documented here.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## 2026-10-26

### Added

- **Scheduler debug endpoints**, only in builds with `-tags debug` (`go run -tags debug ./cmd/scheduler`):
  `POST /debug/scheduler/pause`, `GET /debug/scheduler/pending`, `GET /debug/scheduler/pool`,
  `POST /debug/scheduler/resume`. While paused, the scheduler keeps receiving changes but parks them
  in a pending list instead of applying them; resume applies them in arrival order.
  - Code is in `internal/scheduler/debug.go`; `debug_off.go` holds empty stand-ins for normal builds,
    where the debug state is 0 bytes and the routes don't exist (404).
  - The endpoints run their work on the scheduler's `Run` goroutine, so the pool still has a single owner.
  - Not in the Swagger docs, since they don't exist in normal builds.

## 2026-10-22

### Added

- **Scheduler, part 1: the job pool** (`internal/scheduler/`). Jobs are not executed yet.
  - On startup, before the HTTP server listens, `Load` fills an in-memory pool
    (`map[jobID]*entry`) with all active jobs.
  - `job.Service` notifies a `job.ChangeListener` after every successful create, update, pause,
    resume and delete. The scheduler implements it and queues each change on a channel.
  - `Run` (a background goroutine) applies the queued changes one at a time: adds or updates
    active jobs, removes paused or deleted ones, and ignores an update older than the one it holds.
  - Only the `Run` goroutine touches the pool, so it needs no lock.
- `job.ListFilter.Status`, used to load only active jobs.

- `GET /health` replaces `GET /healthz`: per-service status (currently `postgres`). Returns `503` if any
  service is down; error details are logged, not returned. Each check has a 2s timeout.
  New dependencies are added as one entry in the check map in `internal/app/app.go`.

### Changed

- Split `internal/job/http/` so handler files only hold routes and handlers. No behavior or API changes.
  - `job_dto.go`: `jobResponse`, `errorResponse`, `toJobResponse(s)`.
  - `webhook_dto.go`: `webhookJobRequest`, `webhookJobPatchRequest`, `webhookTaskPatchRequest`, `apply`.

### Fixed

- Swagger generation failed after the patch DTOs were renamed to `webhookJobPatchRequest` /
  `webhookTaskPatchRequest`: the `@Param` annotation on `Update` still referenced the old name.
  - `helpers.go` (was `respond.go`): `decodeJSON`, `writeJSON`, `writeError`, `writeBadRequest`.

## 2026-10-19

> **Status: not tested yet.** The endpoints themselves are still untested (see 2026-10-15).

### Added

- Swagger UI at `/swagger` on the API port, using [swaggo](https://github.com/swaggo/swag).
  The spec in `docs/` is generated from handler annotations with
  `go tool swag init -g cmd/scheduler/main.go --parseInternal -o docs`.
- `swag` CLI pinned as a Go tool dependency in `go.mod` (`go tool swag`).

### Changed

- Job handlers moved to `internal/interfaces/http/jobs/`, one file per task type.
  `JobHandler` → `jobs.Handler`, `WebhookJobHandler` → `jobs.WebhookHandler`.
- Error responses use an `errorResponse` type (`{"error": "..."}`) so they appear in the API docs.
- `PUT /jobs/webhook/{id}` is now a partial update: only the fields in the body change.
- **Restructured by feature instead of by layer.** No behavior or API changes.
  - `internal/job/`: domain model and use cases (was `domain/job/` + `application/`).
    `application.JobService` → `job.Service`, with methods `Create`, `Get`, `List`, `Update`,
    `Pause`, `Resume`, `Delete` (was `CreateJob`, `GetJob`, ...).
  - `internal/job/task/`: `Task` interface and one file per task type (was `domain/job/task.go`,
    `webhook_task.go`). `job.TaskType` → `task.Type`, `job.TaskWebhook` → `task.TypeWebhook`,
    `job.WebhookTask` → `task.Webhook`, `job.HTTPMethod` → `task.HTTPMethod`.
  - `internal/job/postgres/`: job repository (was `infrastructure/persistence/postgres/`).
    `JobRepository` → `postgres.Repository`.
  - `internal/job/http/`: `/jobs` handlers (was `interfaces/http/jobs/`).
  - `internal/platform/postgres/`: connection pool, `schema.sql`, seed data.
  - `internal/platform/http/`: health check and Swagger UI handlers.

## 2026-10-15

> **Status: not tested yet.** The HTTP endpoints and the Postgres repository
> have not been run against a live server and database, and there are no
> automated tests yet.

### Added

- **Job domain model** (`internal/domain/job`)
  - `Task` interface with `TaskType`, implemented by `WebhookTask` (URL, method, headers, body).
  - `Job` now holds a `Task`, with `NewJob`, `Validate`, `Update`, `Pause` and `Resume`.
  - `ErrNotFound` and `ErrInvalid` errors; all validation errors wrap `ErrInvalid`.
  - `ListFilter` for listing jobs by task type.
- **Job service** (`internal/application`): `CreateJob`, `GetJob`, `ListJobs`, `UpdateJob`,
  `PauseJob`, `ResumeJob`, `DeleteJob`. Works with any task type.
- **Postgres persistence** (`internal/infrastructure/persistence/postgres`)
  - `jobs` table in `schema.sql`, created on startup. Tasks are stored as `task_type` + `task_config` (JSONB).
  - `JobRepository` implementing create, update, find by ID, list and delete.
  - Task encoding/decoding by task type in `task_codec.go`.
- **HTTP endpoints** (`internal/interfaces/http`)
  - `GET /jobs` lists jobs of all task types.
  - Webhook jobs under `/jobs/webhook`: `POST /`, `GET /`, `GET /{id}`, `PUT /{id}`,
    `DELETE /{id}`, `POST /{id}/pause`, `POST /{id}/resume`.
  - Shared JSON request/response helpers with error mapping (`400` invalid input, `404` not found, `500` otherwise).
- Request logging and panic recovery middleware.

### Changed

- Switched the HTTP router from `net/http`'s `ServeMux` to [chi](https://github.com/go-chi/chi) for route groups.
  Each task type registers its routes in its own file under the `/jobs` group.
- `job.Repository`: `Save` split into `Create` and `Update`; `FindAll` takes a `ListFilter`.
- Code comments trimmed to be shorter and clearer.

### Known limitations

- Schedules are only checked for being non-empty; cron expressions are not validated yet.
- `schema.sql` only creates missing tables; column changes need a migration tool (e.g. goose).
- The scheduler loop that actually runs jobs is not implemented yet.
