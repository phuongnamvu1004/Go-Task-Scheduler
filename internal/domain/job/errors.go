package job

import "errors"

var (
	ErrNotFound = errors.New("job not found")
	// ErrInvalid wraps every validation error; check it with errors.Is.
	ErrInvalid = errors.New("invalid job")
)
