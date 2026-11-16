# Schedules

A job's `schedule` is a standard 5-field cron expression (minute, hour, day of month, month, day of
week) or a descriptor. It is evaluated in the job's `timezone`, an IANA name such as `Europe/Berlin`
or `America/New_York`; leave it out (or send `""` in a PATCH) for **UTC**.

```json
{ "name": "Morning report", "schedule": "0 9 * * 1-5", "timezone": "Europe/Berlin", "task": { ... } }
```

| Schedule | Runs |
| --- | --- |
| `*/5 * * * *` | every 5 minutes |
| `0 9 * * 1-5` | 09:00 in the job's time zone on weekdays |
| `@every 30s` | every 30 seconds |
| `@daily` / `@hourly` | midnight in the job's time zone / on the hour |

## Daylight saving time

DST follows the zone's rules, which has side effects:

- A run inside the hour that is skipped when clocks go forward (e.g. `30 2 * * *` in `Europe/Berlin`
  on the last Sunday of March) does not happen that day.
- A run inside the hour that repeats when clocks go back may happen twice.
- `@every` intervals ignore time zones and DST; they count real elapsed time.

Use `UTC`, or times outside 02:00–03:00, to avoid both.

## Rules and limits

- An unknown zone, `Local`, a `CRON_TZ=` prefix in the schedule, or an invalid schedule is rejected
  with `400`. Seconds fields are not supported.
- A run is skipped if the previous run of the same job is still queued or running (see
  [worker-pool.md](worker-pool.md)).
- Runs missed while the app was down are not caught up.
- Jobs currently only log when they fire; webhook requests come later.
