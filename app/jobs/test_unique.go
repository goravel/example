package jobs

import (
	"fmt"
	"sync"
	"time"
)

var (
	// TestUniqueResult records the arguments received by TestUnique.Handle.
	TestUniqueResult []any
	// TestUniqueUniqueFor is what TestUnique.UniqueFor reports back to the
	// dispatcher; the worker builds its own instance, so only a package
	// variable carries the value set at dispatch time.
	TestUniqueUniqueFor time.Duration

	testUniqueMu sync.Mutex
)

type TestUnique struct {
}

// NewTestUnique returns a TestUnique job whose lock expires after uniqueFor,
// zero meaning it is held until the job is processed.
func NewTestUnique(uniqueFor time.Duration) *TestUnique {
	TestUniqueUniqueFor = uniqueFor

	return &TestUnique{}
}

// ResetTestUnique resets the package state between tests.
func ResetTestUnique() {
	testUniqueMu.Lock()
	defer testUniqueMu.Unlock()

	TestUniqueResult = nil
	TestUniqueUniqueFor = 0
}

// TestUniqueResultLen reports how many arguments were recorded, so tests can
// poll without racing the worker goroutine that appends them.
func TestUniqueResultLen() int {
	testUniqueMu.Lock()
	defer testUniqueMu.Unlock()

	return len(TestUniqueResult)
}

// Signature The name and signature of the job.
func (r *TestUnique) Signature() string {
	return "test_unique"
}

// Handle Execute the job.
func (r *TestUnique) Handle(args ...any) error {
	if len(args) > 0 {
		testUniqueMu.Lock()
		TestUniqueResult = append(TestUniqueResult, args...)
		testUniqueMu.Unlock()
	}

	return nil
}

// UniqueID tells jobs apart by their first argument; an empty ID falls back to
// the signature alone.
func (r *TestUnique) UniqueID(args ...any) string {
	if len(args) == 0 {
		return ""
	}

	return fmt.Sprint(args[0])
}

// UniqueFor bounds how long the unique lock is held.
func (r *TestUnique) UniqueFor() time.Duration {
	return TestUniqueUniqueFor
}
