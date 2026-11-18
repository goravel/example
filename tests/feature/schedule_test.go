package feature

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goravel/framework/contracts/schedule"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

// scheduleCounter counts the runs of one event. The cron loop records from its
// own goroutine while the test reads the count.
type scheduleCounter struct {
	count atomic.Int32
}

func (c *scheduleCounter) record() { c.count.Add(1) }

func (c *scheduleCounter) reset() { c.count.Store(0) }

func (c *scheduleCounter) value() int {
	return int(c.count.Load())
}

// ScheduleTestSuite covers framework#1572: schedule:run runs the events due in
// the current minute once and exits, and schedule:work keeps running them until
// it is stopped.
type ScheduleTestSuite struct {
	suite.Suite
	tests.TestCase

	everyMinute scheduleCounter
	everySecond scheduleCounter
	never       scheduleCounter
	hourly      scheduleCounter
}

func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, &ScheduleTestSuite{})
}

// SetupSuite disables the boot-time scheduler so this suite owns the only cron
// loop in the process, then registers the events every test observes.
func (s *ScheduleTestSuite) SetupSuite() {
	scope, err := tests.OverrideConfig(map[string]any{
		"app.disabled_runners": []string{"goravel:schedule"},
	})
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.NoError(scope.Restore()) })

	facades.Schedule().Register([]schedule.Event{
		facades.Schedule().Call(s.everyMinute.record).EveryMinute().Name("test:schedule-every-minute"),
		// Fires every second: schedule:run must run it once per trigger, and
		// schedule:work must run it again on every following tick.
		facades.Schedule().Call(s.everySecond.record).EverySecond().Name("test:schedule-every-second"),
		// February 30 never happens: the expression can never be due.
		facades.Schedule().Call(s.never.record).Cron("0 0 30 2 *").Name("test:schedule-never"),
		// Only schedule:work can follow a delay longer than a minute.
		facades.Schedule().Call(s.hourly.record).Cron("@every 1h").Name("test:schedule-hourly"),
	})
}

func (s *ScheduleTestSuite) SetupTest() {
	s.everyMinute.reset()
	s.everySecond.reset()
	s.never.reset()
	s.hourly.reset()
}

// startWork runs the worker in the background and returns a stop func that
// shuts the loop down and waits for Work to return.
func (s *ScheduleTestSuite) startWork() (stop func()) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		facades.Schedule().Work()
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			s.NoError(facades.Schedule().Shutdown())
			s.Require().Eventually(func() bool {
				select {
				case <-done:
					return true
				default:
					return false
				}
			}, 2*time.Second, 10*time.Millisecond, "Shutdown should make Work return")
		})
	}
}

func (s *ScheduleTestSuite) TestRunExecutesEachDueEventOnce() {
	_, err := s.CaptureArtisanOutput("schedule:run")
	s.NoError(err)

	s.Equal(1, s.everyMinute.value(), "an event due this minute runs once")
	s.Equal(1, s.everySecond.value(), "an event that fires every second runs once per trigger")
	s.Equal(0, s.never.value(), "an expression that can never match is skipped")
	s.Equal(0, s.hourly.value(), "@every longer than a minute belongs to schedule:work")
}

func (s *ScheduleTestSuite) TestRunRunsAgainOnTheNextTrigger() {
	_, err := s.CaptureArtisanOutput("schedule:run")
	s.NoError(err)
	_, err = s.CaptureArtisanOutput("schedule:run")
	s.NoError(err)

	s.Equal(2, s.everyMinute.value(), "each external trigger runs the due events once")
	s.Equal(2, s.everySecond.value(), "counts accumulate across triggers instead of being deduplicated")
}

// TestWorkCommandIsRegistered probes the command instead of reading `list`: the
// help text is written to the process stdout, which CaptureArtisanOutput leaves
// alone, while the not-found notice goes through the captured logger.
func (s *ScheduleTestSuite) TestWorkCommandIsRegistered() {
	output, err := s.CaptureArtisanOutput("--no-ansi schedule:work --help")
	s.NoError(err)
	s.NotContains(output, "is not defined", "the new command is registered on the app")

	output, err = s.CaptureArtisanOutput("--no-ansi schedule:missing")
	s.NoError(err)
	s.Contains(output, "Command 'schedule:missing' is not defined.")
}

func (s *ScheduleTestSuite) TestWorkKeepsRunningEventsUntilShutdown() {
	stop := s.startWork()
	defer stop()

	// A one-tick Run can only reach 1; 2 or more runs proves the worker keeps
	// following the sub-minute schedule until Shutdown stops it.
	s.Require().Eventually(func() bool {
		return s.everySecond.value() >= 2
	}, 5*time.Second, 20*time.Millisecond, "schedule:work should run the event again on the next tick")
}
