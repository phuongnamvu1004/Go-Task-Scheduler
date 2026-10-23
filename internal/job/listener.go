package job

// ChangeListener is told about jobs after a change has been saved, e.g. so
// the scheduler can keep its pool in sync. Calls must return quickly.
type ChangeListener interface {
	JobChanged(j *Job)
	JobDeleted(id string)
}

type noopListener struct{}

func (noopListener) JobChanged(*Job)   {}
func (noopListener) JobDeleted(string) {}
