package jobs

var TestResult []any

type Test struct {
}

// Signature The name and signature of the job.
func (r *Test) Signature() string {
	return "test"
}

// Handle Execute the job.
func (r *Test) Handle(args ...any) error {
	if len(args) > 0 {
		TestResult = append(TestResult, args...)
	}
	return nil
}
