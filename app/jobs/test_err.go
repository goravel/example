package jobs

import (
	"errors"
)

var TestErrResult []any

type TestErr struct {
}

// Signature The name and signature of the job.
func (r *TestErr) Signature() string {
	return "test_err"
}

// Handle Execute the job.
func (r *TestErr) Handle(args ...any) error {
	if len(args) > 0 {
		TestErrResult = append(TestErrResult, args...)
	}

	return errors.New("test error")
}
