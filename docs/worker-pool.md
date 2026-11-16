# Worker pool

How due jobs get executed. The code is in `internal/scheduler/worker.go`.

## Flow

```
cron (one tick per due job)
  └─ fire(job) ── queue the run, never wait
        │
   runs channel (buffered, 100 slots)
        │
   ┌────┼──────────┬───────────┐
worker 1   worker 2   ...   worker N      N = WORKERS (default 4)
   └─ execute(run) → dispatch by task type → handle<Type> (handler.go)
```

- **cron** decides *when* a job is due. Its only work is `fire`, which puts the run on the queue
  and returns at once, so a slow job can never delay other schedules.
- **Workers** are a fixed number of goroutines started with the scheduler. Each takes one run at a
  time from the queue, so at most `WORKERS` jobs execute at the same time and the rest wait in the
  queue.

## Rules

| Situation | What happens |
|---|---|
| Job is due while its previous run is still queued or running | This run is skipped and logged. One run per job at a time, also across an update of the job. |
| Queue is full (100 runs waiting) | The run is skipped and logged. Usually means `WORKERS` is too low for the load. |
| A handler panics | The panic is logged; the worker carries on with the next run. |
| Shutdown | Cron stops first, then the workers finish the runs already queued and exit. |

Skipped runs are not retried; the job runs again at its next scheduled time.

## Configuration

| Variable | Default | Description |
|---|---|---|
| `WORKERS` | `4` | Number of workers; must be a positive integer, or the app refuses to start. |

The queue size is a constant (`runQueueSize` in `worker.go`). Raise `WORKERS` if you see
`run queue full` in the logs, or if many jobs are due at the same time and take long to run.

## Logs

| Log line | Build |
|---|---|
| `worker N started job … waited … in queue` / `worker N finished job … in …` | debug only (`make run-debug`) |
| `job … skipped: previous run still queued or running` | always |
| `job … skipped: run queue full (100)` | always |
| `worker N: job … panicked: …` | always |

## Go building blocks

| Need | Go feature |
|---|---|
| Fixed set of workers | goroutines started in a loop |
| Queue between cron and workers | buffered channel |
| Queue without blocking cron | `select` with `default` |
| One run per job | `sync.Map.LoadOrStore` (check and claim in one step) |
| Survive a panicking job | `defer` + `recover()` |
| Wait for workers on shutdown | `close(channel)` + `sync.WaitGroup` |

## Known limits

- `app.Run` doesn't yet wait for the scheduler to stop, so on `Ctrl+C` the process may exit before
  queued runs finish. Fixed by graceful shutdown (SCRUM-9).
- Workers run in a single process; several instances would each run every job.
