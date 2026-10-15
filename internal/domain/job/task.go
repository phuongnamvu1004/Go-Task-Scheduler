package job

// TaskType identifies the kind of work a Task performs.
type TaskType string

const (
	TaskWebhook TaskType = "webhook"
)

// Task is the work a Job runs. Each task type lives in its own file
// (e.g. webhook_task.go) with only the fields it needs.
type Task interface {
	Type() TaskType
	Validate() error
}
