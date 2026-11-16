# Setup

How to configure, run and develop the scheduler locally.

## Configuration

Configuration is read from environment variables. A `.env` file in the project root is loaded
automatically if present (variables already set in the environment win).

```sh
cp .env.example .env
# then edit .env with your Postgres credentials
```

| Variable      | Default             | Description                                            |
|---------------|---------------------|--------------------------------------------------------|
| `PORT`        | `8080`              | HTTP listening port                                    |
| `DB_HOST`     | `localhost`         | Postgres host                                          |
| `DB_PORT`     | `5432`              | Postgres port                                          |
| `DB_USER`     | —                   | Postgres user (required)                               |
| `DB_PASSWORD` | —                   | Postgres password                                      |
| `DB_NAME`     | —                   | Postgres database (required)                           |
| `DB_SCHEMA`   | `go_task_scheduler` | Schema used by the app (created on startup if missing) |
| `DB_SSLMODE`  | `disable`           | Postgres SSL mode                                      |
| `WORKERS`     | `4`                 | Workers that execute due jobs in parallel (see [worker-pool.md](worker-pool.md)) |

## Make targets

Common tasks are in the `Makefile`; `make` lists them:

| Target | Does |
|---|---|
| `make run` / `make run-debug` | Run the app (debug adds `/debug/scheduler/*` and run logs) |
| `make check` | Build and vet, with and without the `debug` tag |
| `make swag` | Regenerate the Swagger docs |
| `make db-up` / `make db-down` | Start / stop local Postgres |
| `make migrate-create name=...` | Create the next migration file |
| `make migrate-up` / `make migrate-status` | Apply pending migrations / show which are applied |
| `make migrate-down` | Undo the latest migration (prints a warning, 5s to cancel) |

The `migrate-*` targets read the same `DB_*` settings from `.env` as the app.

## Local Postgres

A Postgres server for development is defined in `docker-compose.yaml`. It uses the same `.env`
credentials as the app. `make db-up` / `make db-down` wrap the first and third commands:

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
make migrate-create name=add_something    # creates e.g. 00003_add_something.sql
```

Write the SQL under `-- +goose Up` (and the undo under `-- +goose Down`), update the repository
queries in `internal/job/postgres/`, then restart the app or run `make migrate-up`.
`make migrate-status` shows which migrations a database has applied.

If the database is reset (e.g. `docker compose down -v`), the migrations rebuild every table on the
next start. The data itself is not recovered.

## Running

Make sure Postgres is running, then from the project root:

```sh
make run                          # or: go run ./cmd/scheduler
PORT=9090 go run ./cmd/scheduler  # different port
```

Check that it's up:

```sh
curl localhost:8080/health    # status of each service (503 if any is down)
# {"status":"up","services":{"postgres":{"status":"up"}}}
```

Stop it with `Ctrl+C`; the server shuts down gracefully.

## Building

```sh
go build -o bin/scheduler ./cmd/scheduler
./bin/scheduler
```

## API docs (Swagger)

With the app running, open the Swagger UI at
[http://localhost:8080/swagger](http://localhost:8080/swagger) (same port as the API, `PORT`).

The spec in `docs/swagger/` is generated from the `@Summary`, `@Param`, `@Router`, ... comments on
the handlers; don't edit it by hand. After changing an endpoint or its request/response types,
regenerate it with `make swag`.

## Debug build

Scheduler debug tools are compiled in only with the `debug` build tag (`make run-debug`, or
`go run -tags debug ./cmd/scheduler`); normal builds don't contain them.

| Endpoint | What it does |
|---|---|
| `POST /debug/scheduler/pause` | Stop applying job changes; new changes are parked in a pending list |
| `GET /debug/scheduler/pending` | Parked changes, in arrival order |
| `GET /debug/scheduler/pool` | Jobs currently in the scheduler pool, with their cron entry ID and next run time |
| `POST /debug/scheduler/resume` | Apply the parked changes in order and stop pausing |

Debug builds also log every job run: which worker took it, how long it waited in the queue, and how
long it took. Normal builds log only skipped runs and panics.

The code is in `internal/scheduler/debug.go` (`//go:build debug`) with no-op stand-ins in
`debug_off.go`. To have your editor analyze `debug.go`, add the tag: GoLand → *Settings › Go ›
Build Tags*, VS Code → `"gopls": {"buildFlags": ["-tags=debug"]}`.
