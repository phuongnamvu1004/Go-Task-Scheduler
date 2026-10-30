# CLAUDE.md

Go service that stores jobs (a task + a cron schedule) in Postgres and runs them in the background.
README.md covers setup and running; CHANGELOG.md records every change and rename.

## Working agreements

- Discuss first. When asked a question or for a take, answer without editing code until told to go ahead.
- Ask before structural choices (new folders/packages, where code lives, naming schemes).
- Git actions (commit, merge, squash, push, branch deletion) only when the user explicitly asks for
  them in that request. Otherwise just make the changes and leave them uncommitted in the working tree,
  so the user can review the diff. This applies inside Jira tickets too. Never push unless asked.
- Temporary tests are fine for verification, but delete them afterwards. Don't write to the user's
  local database to test; use fakes, or read-only requests against the running app.
- After each change, update README.md (if usage/structure changed) and add a CHANGELOG.md entry.
- The user is new to Go: explain Go-specific trade-offs briefly when they matter.

## Jira ticket workflow

Jira site `phuongnamvu1004.atlassian.net`, project **SCRUM** (Go Task Scheduler), board "SCRUM board".
Use the active sprint ("SCRUM Sprint 0", ends 2027-01-01) until told otherwise. Statuses:
To Do → In Progress → In Review → Done.

For every ticket:
1. Start from an up-to-date `main` with a clean working tree. Create a branch with Jira's default
   name, `<KEY>-<summary in kebab-case>` (e.g. `SCRUM-1-add-goose-database-migrations`). Move the
   ticket to In Progress.
2. Do the work. Usually the user goes through it step by step with you, and "discuss first" still
   applies. If told to "do the whole ticket", run all steps to the end without stopping; still ask
   when the ticket says "to confirm" or a structural choice isn't settled.
3. Leave changes uncommitted for the user to review. When asked to commit, prefix every message with
   the key (`SCRUM-1: ...`).
4. Finish (only when asked to finish/merge, or as part of "do the whole ticket"): both builds and
   vets pass, README/CHANGELOG updated, subtasks done. Tidy the branch history: squash commits that
   are noise (fixups, WIP), but keep separate commits when each is a meaningful step. Merge into
   `main` (fast-forward) and move the ticket and its subtasks to Done. Leave the branch in place
   unless told to delete it.

Keep subtasks few and meaningful; don't split a ticket into many tiny ones. UI work is not
ticketed yet.

## Layout (feature-first)

- `internal/job/` – Job, errors, Repository interface, Service, ChangeListener (package `job`)
  - `task/` – Task interface + one file per task type (`webhook.go`)
  - `postgres/` – repository; tasks stored as `task_type` + `task_config` JSONB (`task_codec.go`)
  - `http/` – `/jobs` routes: `<type>_handler.go` (routes + handlers only), `<type>_dto.go`
    (request/response types + mapping such as `apply`), `job_dto.go`, `helpers.go`
- `internal/scheduler/` – in-memory pool of active jobs registered with robfig/cron; `fire` dispatches
  by task type (`scheduler.go`); `handler.go` has one `handle<Type>` per task type; `debug.go` /
  `debug_off.go` build-tag pair
- `internal/platform/` – shared setup with no business logic (`postgres/` pool + goose `migrations/`,
  `http/` health + Swagger UI)
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

Prefer the `Makefile` targets (`make` lists them); the raw commands are in its recipes.

```sh
make db-up                                 # local Postgres (docker-compose.yaml, reads .env)
make run / make run-debug                  # app on PORT (default 8080); debug adds /debug/scheduler/*
make check                                 # both builds + vets: a plain build never compiles debug.go
make swag                                  # regenerate Swagger after changing annotations or DTOs
make migrate-create name=<name>            # next goose migration; the app applies it on startup
make migrate-status / migrate-up           # status (creates goose_db_version if missing) / apply without the app
```

## Gotchas

- Renaming a type used in a `@Param`/`@Success` annotation compiles fine but breaks `swag init`
  ("cannot find type definition"). Rerun the swag command and check for `ParseComment error`.
- Schema changes go in a new `internal/platform/postgres/migrations/NNNNN_name.sql` (goose, applied on
  startup). Never edit an applied migration; existing databases won't rerun it.
- zsh doesn't word-split unquoted variables; spell out curl calls instead of `set -- $var` loops.

## Roadmap

Done: robfig/cron (entry IDs in the pool, `cron.ParseStandard` in `Job.Validate`, UTC).

Order agreed with the user (Jira keys in brackets):
1. ~~goose migrations + Makefile [SCRUM-1]~~ (done)
2. Per-job time zone: `timezone` column, `Job.CronSpec()` adds the `CRON_TZ=` prefix [SCRUM-2, in progress]
3. Execution: worker pool, webhook requests with timeouts, SSRF guard, graceful shutdown [SCRUM-5,
   subtasks SCRUM-6..9]. Hook-in point: `handleWebhook` in `scheduler/handler.go` (TODO).
4. Single API token auth (`SCHEDULER_API_TOKEN`, required unless `AUTH_DISABLED=true`) [SCRUM-10]
5. Run history in its own feature (`internal/run/`) [SCRUM-11, subtasks SCRUM-12..13]
6. UI (later, not ticketed)

Not ticketed yet: periodic resync from Postgres as a safety net for missed notifications.

Goal: ship as a Docker image that anyone self-hosts (one owner per instance, so no multi-user auth).

Open decisions: single instance vs. several (duplicate runs need locking), missed runs while down
(currently skipped; robfig/cron doesn't catch up). Possible later: switch `job/postgres` to GORM,
a production scheduler status endpoint, multi-user auth for a hosted version.
