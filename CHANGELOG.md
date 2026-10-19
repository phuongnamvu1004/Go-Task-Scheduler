# Changelog

All notable changes to this project are documented here.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

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
