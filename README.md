# Go Task Scheduler

An application that takes in jobs and runs them on a given schedule.

## Requirements

- Go 1.26+

## Running

From the project root:

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
cmd/scheduler/              # entry point: wires the layers and starts the HTTP server
internal/
  domain/job/               # Job aggregate and Repository interface
  application/              # use cases (JobService)
  infrastructure/persistence/  # Repository implementations
  interfaces/http/          # HTTP handlers and routes
```
