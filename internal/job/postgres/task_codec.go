package postgres

import (
	"encoding/json"
	"fmt"

	"github.com/phuongnamvu1004/Go-Task-Scheduler/internal/job/task"
)

// encodeTask turns a task into the JSON stored in task_config.
func encodeTask(t task.Task) ([]byte, error) {
	data, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("encode %s task: %w", t.Type(), err)
	}
	return data, nil
}

// decodeTask rebuilds the concrete task from task_type. Add a case for
// each new task type.
func decodeTask(taskType task.Type, data []byte) (task.Task, error) {
	var t task.Task
	switch taskType {
	case task.TypeWebhook:
		t = &task.Webhook{}
	default:
		return nil, fmt.Errorf("unknown task type %q", taskType)
	}

	if err := json.Unmarshal(data, t); err != nil {
		return nil, fmt.Errorf("decode %s task: %w", taskType, err)
	}
	return t, nil
}
