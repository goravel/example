package feature

import (
	"sync"
	"testing"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/jobs"
	"goravel/tests"
)

// QueueMultiQueueTestSuite covers framework#1564: a worker accepts a
// comma-separated queue list and polls it left to right (priority order),
// while dispatch without OnQueue targets the first configured queue name.
type QueueMultiQueueTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestQueueMultiQueueTestSuite(t *testing.T) {
	suite.Run(t, &QueueMultiQueueTestSuite{})
}

func (s *QueueMultiQueueTestSuite) SetupSuite() {
	// The sync driver processes jobs in-process and never pops, so priority
	// polling is only observable on the database driver. Disable the boot-time
	// workers so they cannot steal the jobs these tests reserve.
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default":                    "database",
		"queue.connections.database.queue": "high,default",
		// Pin concurrency to 1 so the no-args worker runs a single goroutine;
		// jobs.TestResult is an unlocked global shared by the handled jobs.
		"queue.connections.database.concurrent": 1,
		"app.disabled_runners":                  []string{"goravel:queue", "app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	s.T().Cleanup(func() { s.NoError(scope.Restore()) })
}

func (s *QueueMultiQueueTestSuite) SetupTest() {
	s.RefreshDatabase()

	jobs.TestResult = nil
	jobs.TestErrResult = nil
}

// startWorker runs a worker in the background and returns an idempotent stop
// func that shuts it down and waits for Run to return.
func (s *QueueMultiQueueTestSuite) startWorker(args ...contractsqueue.Args) (stop func()) {
	worker := facades.Queue().Worker(args...)
	errCh := make(chan error, 1)
	go func() { errCh <- worker.Run() }()

	var once sync.Once
	return func() {
		once.Do(func() {
			s.NoError(worker.Shutdown())
			s.NoError(<-errCh)
		})
	}
}

func (s *QueueMultiQueueTestSuite) TestWorkerPollsQueuesInPriorityOrder() {
	// Dispatch the lower-priority job first so its id is smaller; only queue
	// priority (not insertion order) can explain "high" running first.
	s.NoError(facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
		{Type: "string", Value: "default-job"},
	}).OnQueue("default").Dispatch())
	s.NoError(facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
		{Type: "string", Value: "high-job"},
	}).OnQueue("high").Dispatch())

	stop := s.startWorker(contractsqueue.Args{
		Connection: "database",
		Queue:      "high,default",
		Concurrent: 1,
	})
	defer stop()

	s.Require().Eventually(func() bool {
		count, err := facades.DB().Table("jobs").Count()
		return err == nil && count == 0
	}, 5*time.Second, 25*time.Millisecond)

	stop()
	s.Equal([]any{"high-job", "default-job"}, jobs.TestResult)
}

func (s *QueueMultiQueueTestSuite) TestWorkerWithoutArgsUsesConfiguredQueueList() {
	// Worker() with no args must consume the configured list
	// ("high,default"), not the literal name.
	s.NoError(facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
		{Type: "string", Value: "high-job"},
	}).OnQueue("high").Dispatch())
	s.NoError(facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
		{Type: "string", Value: "default-job"},
	}).OnQueue("default").Dispatch())

	stop := s.startWorker()
	defer stop()

	s.Require().Eventually(func() bool {
		count, err := facades.DB().Table("jobs").Count()
		return err == nil && count == 0
	}, 5*time.Second, 25*time.Millisecond)

	// Shutdown waits for in-flight jobs, so the handled results are settled
	// by the time the assertion runs.
	stop()
	s.ElementsMatch([]any{"high-job", "default-job"}, jobs.TestResult)
}

func (s *QueueMultiQueueTestSuite) TestDefaultDispatchUsesFirstConfiguredQueue() {
	// The connection queue is "high,default"; a job with no OnQueue must land
	// on "high" (the first valid name), not the literal "high,default".
	s.NoError(facades.Queue().Job(&jobs.Test{}).Dispatch())

	var rows []jobRow // defined in tests/feature/notification_test.go
	s.NoError(facades.DB().Table("jobs").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal("high", rows[0].Queue)
}

func (s *QueueMultiQueueTestSuite) TestFailedJobRecordsReservedQueue() {
	// The worker polls "high,default"; a job reserved from "default" must be
	// recorded against "default", not the whole configured list.
	s.NoError(facades.Queue().Job(&jobs.TestErr{}).OnQueue("default").Dispatch())

	stop := s.startWorker(contractsqueue.Args{
		Connection: "database",
		Queue:      "high,default",
		Concurrent: 1,
	})
	defer stop()

	s.Require().Eventually(func() bool {
		failedJobs, err := facades.Queue().Failer().All()
		return err == nil && len(failedJobs) == 1
	}, 5*time.Second, 25*time.Millisecond)

	failedJobs, err := facades.Queue().Failer().All()
	s.Require().NoError(err)
	s.Require().Len(failedJobs, 1)
	s.Equal("default", failedJobs[0].Queue())
	s.Equal("database", failedJobs[0].Connection())
}

func (s *QueueMultiQueueTestSuite) TestWorkerNormalizesQueueList() {
	// Whitespace is trimmed and empty entries are skipped; an empty list
	// falls back to "default". Passing the messy list straight to Worker()
	// exercises splitQueueNames without another app restart.
	cases := []struct {
		name        string
		workerQueue string
		targetQueue string
	}{
		{name: "trims whitespace and keeps priority order", workerQueue: " high , default ", targetQueue: "high"},
		{name: "whitespace-only list falls back to default", workerQueue: " , ", targetQueue: "default"},
	}

	for _, tt := range cases {
		s.Run(tt.name, func() {
			s.RefreshDatabase()
			// SetupTest does not run between subtests, so reset the globals
			// here to keep each case independent.
			jobs.TestResult = nil
			jobs.TestErrResult = nil

			s.NoError(facades.Queue().Job(&jobs.Test{}, []contractsqueue.Arg{
				{Type: "string", Value: tt.targetQueue + "-job"},
			}).OnQueue(tt.targetQueue).Dispatch())

			stop := s.startWorker(contractsqueue.Args{
				Connection: "database",
				Queue:      tt.workerQueue,
				Concurrent: 1,
			})
			defer stop()

			s.Require().Eventually(func() bool {
				count, err := facades.DB().Table("jobs").Count()
				return err == nil && count == 0
			}, 5*time.Second, 25*time.Millisecond)

			stop()
			s.Equal([]any{tt.targetQueue + "-job"}, jobs.TestResult)
		})
	}
}
