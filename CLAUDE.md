# CLAUDE.md

Go service that stores jobs (a task + a cron schedule) in Postgres and runs them in the background.
README.md covers setup and running; CHANGELOG.md records every change and rename.

## Working agreements

- Discuss first. When asked a question or for a take, answer without editing code until told to go ahead.
- Ask before structural choices (new folders/packages, where code lives, naming schemes).
- The user commits. Don't commit or push; leave changes in the working tree.
- Temporary tests are fine for verification, but delete them afterwards. Don't write to the user's
  local database to test; use fakes, or read-only requests against the running app.
- After each change, update README.md (if usage/structure changed) and add a CHANGELOG.md entry.
- The user is new to Go: explain Go-specific trade-offs briefly when they matter.

## Layout (feature-first)

- `internal/job/` – Job, errors, Repository interface, Service, ChangeListener (package `job`)
  - `task/` – Task interface + one file per task type (`webhook.go`)
  - `postgres/` – repository; tasks stored as `task_type` + `task_config` JSONB (`task_codec.go`)
  - `http/` – `/jobs` routes: `<type>_handler.go` (routes + handlers only), `<type>_dto.go`
    (request/response types + mapping such as `apply`), `job_dto.go`, `helpers.go`
- `internal/scheduler/` – in-memory pool of active jobs registered with robfig/cron; `fire` dispatches
  by task type (`scheduler.go`); `handler.go` has one `handle<Type>` per task type; `debug.go` /
  `debug_off.go` build-tag pair
- `internal/platform/` – shared setup with no business logic (`postgres/` pool + `schema.sql`, `http/` health + Swagger UI)
- `internal/app/app.go` – the only place that wires packages together and starts/stops them

## Conventions

- Request/response types end in `Request` / `Response` (e.g. `webhookJobPatchRequest`, `jobResponse`).
- No stutter: `job.Service`, `task.Webhook`, `postgres.Repository`.
- Interfaces live in the package that uses them (`job.Repository`, `job.ChangeListener`), never next to
  the implementation (it would also create an import cycle).
- Comments: minimal but clear; explain why, not what.
- Domain validation errors wrap `job.ErrInvalid` (→ 400); missing jobs return `job.ErrNotFound` (→ 404).
- Updates are partial (`PATCH` with a `<type>JobPatchRequest` of pointer fields); there is no `PUT`.
- Router is chi with route groups. Each task type mounts its own group under `/jobs/<type>`; handlers
  of other task types report `404` for jobs that aren't theirs.
- Scheduler: only the `Run` goroutine touches the pool. Other goroutines send changes over a channel
  or run work on `Run` (see `onRun` in `debug.go`). `job.Service` notifies the scheduler, not handlers.

## Commands

```sh
docker compose up -d                       # local Postgres (docker-compose.yaml, reads .env)
go run ./cmd/scheduler                     # app on PORT (default 8080), Swagger at /swagger
go run -tags debug ./cmd/scheduler         # adds /debug/scheduler/{pause,resume,pending,pool}


# Check both builds: a plain build never compiles debug.go, so it can break silently.
go build ./... && go vet ./... && go build -tags debug ./... && go vet -tags debug ./...

# Regenerate Swagger after changing any handler annotation or request/response type.
go tool swag init -g cmd/scheduler/main.go --parseInternal -o docs
```

## Gotchas

- Renaming a type used in a `@Param`/`@Success` annotation compiles fine but breaks `swag init`
  ("cannot find type definition"). Rerun the swag command and check for `ParseComment error`.
- `schema.sql` runs on every startup and only creates missing tables; column changes need a
  migration tool (goose planned).
- zsh doesn't word-split unquoted variables; spell out curl calls instead of `set -- $var` loops.

## Roadmap

1. ~~robfig/cron: entry IDs in the pool, `cron.ParseStandard` in `Job.Validate`~~ (done, UTC).
2. Periodic resync from Postgres as a safety net for missed notifications.
3. Execution: runs channel + fixed worker pool; webhook timeouts. Hook-in point: `handleWebhook`
   in `scheduler/handler.go` (TODO).
4. Run history in its own feature (`internal/run/`).

Open decisions: single instance vs. several (duplicate runs need locking), missed runs while down
(currently skipped; robfig/cron doesn't catch up). Possible later: switch `job/postgres` to GORM,
a production scheduler status endpoint, authentication.
