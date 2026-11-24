package jobs

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// TestUniqueRetryableResult records the arguments received by
	// TestUniqueRetryable.Handle.
	TestUniqueRetryableResult []any
	// TestUniqueRetryableFailUntil is the attempt count up to which Handle
	// fails.
	TestUniqueRetryableFailUntil int
	// TestUniqueRetryableNeverSucceed makes Handle always fail, so ShouldRetry
	// is the only terminator.
	TestUniqueRetryableNeverSucceed bool
	// TestUniqueRetryableDelay is the delay ShouldRetry asks the worker to wait
	// before the next attempt.
	TestUniqueRetryableDelay time.Duration

	testUniqueRetryableMu sync.Mutex
)

type TestUniqueRetryable struct {
}

// ResetTestUniqueRetryable resets the package state between tests.
func ResetTestUniqueRetryable() {
	testUniqueRetryableMu.Lock()
	defer testUniqueRetryableMu.Unlock()

	TestUniqueRetryableResult = nil
	TestUniqueRetryableFailUntil = 0
	TestUniqueRetryableNeverSucceed = false
	TestUniqueRetryableDelay = 0
}

// TestUniqueRetryableResultLen reports how many arguments were recorded, so
// tests can poll without racing the worker goroutine that appends them.
func TestUniqueRetryableResultLen() int {
	testUniqueRetryableMu.Lock()
	defer testUniqueRetryableMu.Unlock()

	return len(TestUniqueRetryableResult)
}

// Signature The name and signature of the job.
func (r *TestUniqueRetryable) Signature() string {
	return "test_unique_retryable"
}

// Handle Execute the job.
func (r *TestUniqueRetryable) Handle(args ...any) error {
	if len(args) > 0 {
		testUniqueRetryableMu.Lock()
		TestUniqueRetryableResult = append(TestUniqueRetryableResult, args...)
		testUniqueRetryableMu.Unlock()
	}

	testUniqueRetryableMu.Lock()
	defer testUniqueRetryableMu.Unlock()

	if TestUniqueRetryableNeverSucceed || len(TestUniqueRetryableResult) <= TestUniqueRetryableFailUntil {
		return errors.New("test unique retryable error")
	}

	return nil
}

// UniqueID tells jobs apart by their first argument.
func (r *TestUniqueRetryable) UniqueID(args ...any) string {
	if len(args) == 0 {
		return ""
	}

	return fmt.Sprint(args[0])
}

// UniqueFor holds the lock until the job is processed.
func (r *TestUniqueRetryable) UniqueFor() time.Duration {
	return 0
}

// ShouldRetry retries while the attempt count is within
// TestUniqueRetryableFailUntil and gives up afterwards.
func (r *TestUniqueRetryable) ShouldRetry(err error, attempt, maxTries int) (bool, time.Duration) {
	return attempt <= TestUniqueRetryableFailUntil, TestUniqueRetryableDelay
}
