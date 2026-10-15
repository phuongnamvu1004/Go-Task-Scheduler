package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/domain/job"
)

// encodeTask turns a task into the JSON stored in task_config.
func encodeTask(t job.Task) ([]byte, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("encode %s task: %w", t.Type(), err)
	}
	return data, nil
}

// decodeTask rebuilds the concrete task from task_type. Add a case for
// each new task type.
func decodeTask(taskType job.TaskType, data []byte) (job.Task, error) {
	var t job.Task
	switch taskType {
	case job.TaskWebhook:
		t = &job.WebhookTask{}
	default:
		return nil, fmt.Errorf("unknown task type %q", taskType)
	}

	if err := json.Unmarshal(data, t); err != nil {
		return nil, fmt.Errorf("decode %s task: %w", taskType, err)
	}
	return t, nil
}
