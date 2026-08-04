package jobs

import (
	"errors"
	"sync"
	"time"
)

var (
	// TestRetryableResult records the arguments received by TestRetryable.Handle.
	// TestRetryableFailUntil is the attempt count up to which Handle fails.
	// TestRetryableNeverSucceed makes Handle always fail, so ShouldRetry is the
	// sole terminator (used to test the retry-exhausted path).
	//
	// All are package globals (like TestResult/TestErrResult) because queue
	// dispatch resolves jobs by signature: the worker always constructs a fresh
	// &TestRetryable{}, so instance state set at dispatch time never reaches it.
	TestRetryableResult       []any
	TestRetryableFailUntil    int
	TestRetryableNeverSucceed bool

	testRetryableMu sync.Mutex
)

type TestRetryable struct {
}

// NewTestRetryable returns a TestRetryable job that fails until the
// failUntil-th attempt, then succeeds. The threshold is stored in
// TestRetryableFailUntil so the worker-side instance can read it.
func NewTestRetryable(failUntil int) *TestRetryable {
	TestRetryableFailUntil = failUntil

	return &TestRetryable{}
}

// ResetTestRetryable resets the package state so the job can be dispatched
// again from a clean slate.
func ResetTestRetryable() {
	testRetryableMu.Lock()
	defer testRetryableMu.Unlock()

	TestRetryableResult = nil
	TestRetryableFailUntil = 0
	TestRetryableNeverSucceed = false
}

// TestRetryableResultLen returns the current number of records in
// TestRetryableResult. It exists so tests can poll for completion without
// racing the worker goroutine that appends to the slice in Handle.
func TestRetryableResultLen() int {
	testRetryableMu.Lock()
	defer testRetryableMu.Unlock()

	return len(TestRetryableResult)
}

// Signature returns the unique signature of the job.
func (r *TestRetryable) Signature() string {
	return "test_retryable"
}

// Handle executes the job, recording args into TestRetryableResult. It
// fails when TestRetryableNeverSucceed is true, or while
// len(TestRetryableResult) <= TestRetryableFailUntil.
func (r *TestRetryable) Handle(args ...any) error {
	// args is a per-invocation parameter, so len(args) needs no
	// synchronization and is checked before acquiring the mutex.
	if len(args) > 0 {
		testRetryableMu.Lock()
		TestRetryableResult = append(TestRetryableResult, args...)
		testRetryableMu.Unlock()
	}

	testRetryableMu.Lock()
	defer testRetryableMu.Unlock()

	if TestRetryableNeverSucceed || len(TestRetryableResult) <= TestRetryableFailUntil {
		return errors.New("test retryable error")
	}

	return nil
}

// ShouldRetry implements queue.JobWithShouldRetry. It retries while the
// attempt count is within TestRetryableFailUntil, and gives up afterwards,
// matching Handle's failure window. The 100ms delay is preserved by the
// database queue driver (time.Time precision).
func (r *TestRetryable) ShouldRetry(err error, attempt int) (bool, time.Duration) {
	// TestRetryableFailUntil is set once at dispatch time and never mutated
	// during the run, so a plain read without the mutex is safe here.
	if attempt <= TestRetryableFailUntil {
		return true, 100 * time.Millisecond
	}

	return false, 0
}
