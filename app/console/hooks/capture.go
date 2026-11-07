package hooks

import "sync"

// RunCapture records what the command hooks saw during one run of
// test:console-hook. The suite asserts the whole struct.
type RunCapture struct {
	// Events holds the markers in the order they ran: "trace:before",
	// "approval:before", "audit:before", "handle", "audit:after", "trace:after".
	Events []string
	// Target is the --target value the audit hook read in Before. It stays empty
	// when an outer hook skips the run, because the audit hook never gets a Before.
	Target string
	// Approved is the --approved value the approval hook read in Before.
	Approved bool
	// HandleErr is the error the audit hook received from Handle in After.
	HandleErr error
}

var (
	captureMutex sync.RWMutex
	latestRun    *RunCapture
)

// Record appends an event marker to the current run.
func Record(event string) {
	update(func(c *RunCapture) {
		c.Events = append(c.Events, event)
	})
}

// SetTarget records the --target value the audit hook read.
func SetTarget(target string) {
	update(func(c *RunCapture) { c.Target = target })
}

// SetApproved records the --approved value the approval hook read.
func SetApproved(approved bool) {
	update(func(c *RunCapture) { c.Approved = approved })
}

// SetHandleErr records the error the audit hook received from Handle.
func SetHandleErr(err error) {
	update(func(c *RunCapture) { c.HandleErr = err })
}

// Reset clears the capture; the suite calls it before and after every test.
func Reset() {
	captureMutex.Lock()
	defer captureMutex.Unlock()

	latestRun = nil
}

// GetLatest returns the capture of the last run, or nil when nothing ran.
func GetLatest() *RunCapture {
	captureMutex.RLock()
	defer captureMutex.RUnlock()

	return latestRun
}

func update(fn func(*RunCapture)) {
	captureMutex.Lock()
	defer captureMutex.Unlock()

	if latestRun == nil {
		latestRun = &RunCapture{}
	}
	fn(latestRun)
}
