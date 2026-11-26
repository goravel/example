package feature

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goravel/framework/contracts/schedule"
	"github.com/goravel/framework/support/carbon"
	"github.com/stretchr/testify/suite"

	"goravel/app/console/commands"
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

// startWork runs schedule:work in the background and returns a stop func that
// shuts the loop down and waits for the command to return.
func (s *ScheduleTestSuite) startWork() (stop func()) {
	var callErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		callErr = facades.Artisan().Call("schedule:work")
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			// A deadline turns a wedged event into a failure instead of a hung run.
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			s.NoError(facades.Schedule().Shutdown(ctx))
			s.Require().Eventually(func() bool {
				select {
				case <-done:
					return true
				default:
					return false
				}
			}, 2*time.Second, 10*time.Millisecond, "Shutdown should make schedule:work return")
			s.NoError(callErr, "schedule:work should return without an error")
		})
	}
}

func (s *ScheduleTestSuite) TestRunExecutesEachDueEventOnce() {
	_, err := s.CaptureArtisanOutput("schedule:run")
	s.NoError(err)

	s.Equal(1, s.everyMinute.value(), "an event due this minute runs once")
	s.Equal(1, s.everySecond.value(), "an event that fires every second runs once per trigger")
	s.Equal(0, s.never.value(), "an expression that can never match is skipped")
	s.Equal(0, s.hourly.value(), "no unrelated event ran")
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
	s.Contains(output, "schedule:work", "the new command is registered on the app")
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

// hookLog records what ran, in the order it ran. schedule:run calls the event
// and its hooks on the calling goroutine, so the log needs no lock.
type hookLog struct {
	entries []string
}

func (l *hookLog) add(entry string) { l.entries = append(l.entries, entry) }

func (l *hookLog) all() []string { return l.entries }

// ScheduleHooksTestSuite covers framework#1625: Before, After, OnSuccess and
// OnFailure run around an event, and Timezone evaluates its cron expression in
// another IANA timezone.
type ScheduleHooksTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestScheduleHooksTestSuite(t *testing.T) {
	suite.Run(t, new(ScheduleHooksTestSuite))
}

// freshSchedule gives each test an empty schedule: OverrideConfig restarts the
// application, and a restart builds a new schedule that holds only the events
// the app registers at boot, so an event that fails or panics in one test never
// changes the result of a run in another.
func (s *ScheduleHooksTestSuite) freshSchedule() {
	scope, err := tests.OverrideConfig(map[string]any{
		"app.disabled_runners": []string{"goravel:schedule"},
	})
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.NoError(scope.Restore()) })
}

func (s *ScheduleHooksTestSuite) TestHooksRunAroundASuccessfulEvent() {
	s.freshSchedule()

	var log hookLog
	facades.Schedule().Register([]schedule.Event{
		facades.Schedule().Call(func() { log.add("event") }).
			Before(func() { log.add("before:1") }).
			Before(func() { log.add("before:2") }).
			After(func() { log.add("after:1") }).
			OnSuccess(func() { log.add("success") }).
			OnFailure(func(err error) { log.add("failure:" + err.Error()) }).
			After(func() { log.add("after:2") }).
			EveryMinute().
			Name("test:schedule-hook-success"),
	})

	_, err := s.CaptureArtisanOutput("schedule:run")
	s.NoError(err)

	s.Equal([]string{"before:1", "before:2", "event", "after:1", "success", "after:2"}, log.all(),
		"Before runs first, then the event, then After, OnSuccess and OnFailure in the order they were registered")
}

func (s *ScheduleHooksTestSuite) TestHooksRunAroundAPanickingEvent() {
	s.freshSchedule()

	var log hookLog
	var failures []error
	facades.Schedule().Register([]schedule.Event{
		facades.Schedule().Call(func() {
			log.add("event")
			panic("boom")
		}).
			Before(func() { log.add("before") }).
			OnSuccess(func() { log.add("success") }).
			OnFailure(func(err error) {
				log.add("failure")
				failures = append(failures, err)
			}).
			After(func() { log.add("after") }).
			EveryMinute().
			Name("test:schedule-hook-panicking-event"),
	})

	_, err := s.CaptureArtisanOutput("schedule:run")
	s.ErrorContains(err, "schedule event test:schedule-hook-panicking-event panicked: boom")

	s.Equal([]string{"before", "event", "failure", "after"}, log.all(),
		"OnFailure runs, OnSuccess does not, and After still runs")
	s.Require().Len(failures, 1)
	s.ErrorContains(failures[0], "schedule event test:schedule-hook-panicking-event panicked: boom",
		"OnFailure receives the panic of the event")
}

func (s *ScheduleHooksTestSuite) TestOnFailureReceivesTheCommandError() {
	s.freshSchedule()

	var log hookLog
	var failures []error
	facades.Schedule().Register([]schedule.Event{
		// --approved opens the command hook gate, so the command reaches Handle
		// and returns its own error.
		facades.Schedule().Command("test:console-hook --approved --fail").
			OnSuccess(func() { log.add("success") }).
			OnFailure(func(err error) {
				log.add("failure")
				failures = append(failures, err)
			}).
			After(func() { log.add("after") }).
			EveryMinute().
			Name("test:schedule-hook-failing-command"),
	})

	_, err := s.CaptureArtisanOutput("schedule:run")
	s.ErrorContains(err, "schedule event test:schedule-hook-failing-command failed")

	s.Equal([]string{"failure", "after"}, log.all())
	s.Require().Len(failures, 1)
	s.ErrorIs(failures[0], commands.ErrConsoleHookPublishFailed,
		"OnFailure receives the error the command returned")
}

func (s *ScheduleHooksTestSuite) TestPanickingAfterHookDoesNotStopTheOtherHooks() {
	s.freshSchedule()

	var log hookLog
	facades.Schedule().Register([]schedule.Event{
		facades.Schedule().Call(func() { log.add("event") }).
			After(func() { panic("after boom") }).
			OnSuccess(func() { log.add("success") }).
			After(func() { log.add("after:2") }).
			EveryMinute().
			Name("test:schedule-hook-panicking-after"),
	})

	_, err := s.CaptureArtisanOutput("schedule:run")
	s.NoError(err, "a panicking hook is logged and the event keeps its own result")

	s.Equal([]string{"event", "success", "after:2"}, log.all(),
		"the hooks registered after the panicking one still run")
}

func (s *ScheduleHooksTestSuite) TestPanickingBeforeHookStopsTheEvent() {
	s.freshSchedule()

	var log hookLog
	facades.Schedule().Register([]schedule.Event{
		facades.Schedule().Call(func() { log.add("event") }).
			Before(func() {
				log.add("before")
				panic("not now")
			}).
			OnSuccess(func() { log.add("success") }).
			OnFailure(func(err error) { log.add("failure") }).
			EveryMinute().
			Name("test:schedule-hook-panicking-before"),
	})

	_, err := s.CaptureArtisanOutput("schedule:run")
	s.ErrorContains(err, "schedule event test:schedule-hook-panicking-before panicked: not now")

	s.Equal([]string{"before"}, log.all(), "the event and the other hooks do not run")
}

func (s *ScheduleHooksTestSuite) TestTimezoneEvaluatesTheCronInItsZone() {
	s.freshSchedule()

	// A frozen clock makes the due check the same on every machine; the
	// schedule's own zone is app.timezone, which is UTC in config/app.go.
	carbon.SetTestNow(carbon.Parse("2026-11-24 09:00:00", carbon.UTC))
	// Cleanups run last in, first out: the clock is cleared before the restart
	// of freshSchedule runs.
	s.T().Cleanup(carbon.ClearTestNow)

	var log hookLog
	facades.Schedule().Register([]schedule.Event{
		// 09:00 UTC is 23:00 on the same day in Kiritimati (UTC+14)...
		facades.Schedule().Call(func() { log.add("kiritimati") }).
			Timezone("Pacific/Kiritimati").Cron("0 23 * * *").Name("test:schedule-hook-tz-kiritimati"),
		// ...and 22:00 on the previous day in Midway (UTC-11).
		facades.Schedule().Call(func() { log.add("midway") }).
			Timezone("Pacific/Midway").Cron("0 23 * * *").Name("test:schedule-hook-tz-midway"),
		// No Timezone: read in the schedule's own zone, UTC.
		facades.Schedule().Call(func() { log.add("utc") }).
			Cron("0 9 * * *").Name("test:schedule-hook-tz-utc"),
	})

	_, err := s.CaptureArtisanOutput("schedule:run")
	s.NoError(err)

	s.Equal([]string{"kiritimati", "utc"}, log.all(),
		"each expression is evaluated in the zone of its own event")
}

func (s *ScheduleHooksTestSuite) TestUnknownTimezoneIsNotRegistered() {
	s.freshSchedule()

	var log hookLog
	facades.Schedule().Register([]schedule.Event{
		facades.Schedule().Call(func() { log.add("mars") }).
			Timezone("Mars/Olympus_Mons").EveryMinute().Name("test:schedule-hook-unknown-timezone"),
	})

	var names []string
	for _, event := range facades.Schedule().Events() {
		names = append(names, event.GetName())
	}
	s.NotContains(names, "test:schedule-hook-unknown-timezone",
		"an event with an unknown timezone is not registered")

	_, err := s.CaptureArtisanOutput("schedule:run")
	s.NoError(err)
	s.Empty(log.all(), "the event never ran")
}
