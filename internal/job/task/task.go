package task

// Type identifies the kind of work a Task performs.
type Type string

const (
	TypeWebhook Type = "webhook"
)

// Task is the work a Job runs. Each task type lives in its own file in this
// package (e.g. webhook.go) with only the fields it needs.
type Task interface {
	Type() Type
	Validate() error
}
