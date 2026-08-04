package feature

import (
	"errors"
	"testing"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/queue/utils"
	"github.com/goravel/framework/support/carbon"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/jobs"
	"goravel/tests"
)

type QueueTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, &QueueTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *QueueTestSuite) SetupTest() {
	jobs.TestResult = nil
	jobs.TestErrResult = nil
	jobs.ResetTestRetryable()
}

// TearDownTest will run after each test in the suite.
func (s *QueueTestSuite) TearDownTest() {
}

func (s *QueueTestSuite) TestDispatch() {
	s.NoError(facades.Queue().Job(&jobs.Test{}, testQueueArgs).Dispatch())

	time.Sleep(1 * time.Second)

	s.Equal(utils.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchWithDelay() {
	s.NoError(facades.Queue().Job(&jobs.Test{}, testQueueArgs).Delay(time.Now().Add(1 * time.Second)).Dispatch())

	time.Sleep(2 * time.Second)

	s.Equal(utils.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchChain() {
	s.NoError(facades.Queue().Chain([]contractsqueue.ChainJob{
		{
			Job:  &jobs.Test{},
			Args: testQueueArgs,
		},
		{
			Job:  &jobs.Test{},
			Args: testQueueArgs,
		},
	}).Dispatch())

	time.Sleep(1 * time.Second)

	var args []any
	for i := 0; i < 2; i++ {
		args = append(args, utils.ConvertArgs(testQueueArgs)...)
	}

	s.Equal(args, jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchWithQueue() {
	s.NoError(facades.Queue().Job(&jobs.Test{}, testQueueArgs).OnQueue("test").Dispatch())

	time.Sleep(1 * time.Second)

	s.Equal(utils.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchWithConnectionAndQueue() {
	if facades.Config().GetString("queue.default") == "sync" {
		s.T().Skip("skip test due to only for redis")
	}

	s.NoError(facades.Queue().Job(&jobs.Test{}, testQueueArgs).OnConnection("redis1").OnQueue("test").Dispatch())

	time.Sleep(1 * time.Second)

	s.Equal(utils.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestSyncFailedJob() {
	if facades.Config().GetString("queue.default") != "sync" {
		s.T().Skip("skip test due to only for sync")
	}

	s.Equal(errors.New("test error"), facades.Queue().Job(&jobs.TestErr{}).Dispatch())
}

func (s *QueueTestSuite) TestFailedJobAndRetry() {
	if facades.Config().GetString("queue.default") == "sync" {
		s.T().Skip("skip test due to only for non-sync")
	}

	carbon.SetTestNow(carbon.Now())
	defer carbon.ClearTestNow()

	testErr := &jobs.TestErr{}
	s.NoError(facades.Queue().Job(testErr, []contractsqueue.Arg{
		{
			Type:  "string",
			Value: "test",
		},
	}).Dispatch())

	time.Sleep(2 * time.Second)

	s.Equal([]any{"test"}, jobs.TestErrResult)

	failedJobs, err := facades.Queue().Failer().All()

	s.Require().NoError(err)
	s.Require().Equal(1, len(failedJobs))
	s.Equal("default", failedJobs[0].Queue())
	s.Equal(facades.Config().GetString("queue.default"), failedJobs[0].Connection())
	s.Equal(carbon.NewDateTime(carbon.Now()), failedJobs[0].FailedAt())
	s.Equal(testErr.Signature(), failedJobs[0].Signature())
	s.NotEmpty(failedJobs[0].UUID())

	s.NoError(facades.Artisan().Call("queue:retry"))

	time.Sleep(1 * time.Second)

	s.Equal([]any{"test", "test"}, jobs.TestErrResult)
}

func (s *QueueTestSuite) TestReleaseBasedRetry() {
	if facades.Config().GetString("queue.default") == "sync" {
		s.T().Skip("skip test due to only for non-sync")
	}

	worker := facades.Queue().Worker(contractsqueue.Args{
		Queue:      "default",
		Concurrent: 1,
	})
	go func() { _ = worker.Run() }()
	defer func() { _ = worker.Shutdown() }()

	// The job fails the first two attempts and succeeds on the third, proving
	// the reserved job is released back to the queue (Release(delay)) with its
	// attempt count preserved instead of being retried purely in-process.
	s.NoError(facades.Queue().Job(jobs.NewTestRetryable(2), []contractsqueue.Arg{
		{
			Type:  "string",
			Value: "retryable",
		},
	}).Dispatch())

	// Wait for all three attempts. Eventually fails loudly on timeout, and
	// TestRetryableResultLen reads the shared slice under the mutex so the
	// poll does not race Handle's append.
	s.Require().Eventually(func() bool {
		return jobs.TestRetryableResultLen() >= 3
	}, 5*time.Second, 25*time.Millisecond)

	s.Equal([]any{"retryable", "retryable", "retryable"}, jobs.TestRetryableResult)
}

func (s *QueueTestSuite) TestReleaseBasedRetryExhausted() {
	if facades.Config().GetString("queue.default") == "sync" {
		s.T().Skip("skip test due to only for non-sync")
	}

	// neverSucceed=true forces Handle to always fail, so ShouldRetry is the
	// only terminator. With failUntil=2, ShouldRetry returns true for
	// attempts 1-2 and false for attempt 3, causing the job to land in
	// failed_jobs after 3 handle calls.
	jobs.TestRetryableNeverSucceed = true
	s.NoError(facades.Queue().Job(jobs.NewTestRetryable(2), []contractsqueue.Arg{
		{Type: "string", Value: "exhausted"},
	}).Dispatch())

	worker := facades.Queue().Worker(contractsqueue.Args{
		Queue:      "default",
		Concurrent: 1,
	})
	go func() { _ = worker.Run() }()
	defer func() { _ = worker.Shutdown() }()

	// The job exhausts retries and lands in failed_jobs. Poll the failer
	// until the signature appears.
	s.Require().Eventually(func() bool {
		failedJobs, err := facades.Queue().Failer().All()
		if err != nil {
			return false
		}
		for _, fj := range failedJobs {
			if fj.Signature() == "test_retryable" {
				return true
			}
		}
		return false
	}, 10*time.Second, 25*time.Millisecond, "expected test_retryable to land in failed_jobs")

	// 3 Handle calls (attempts 1,2,3 → all fail), then ShouldRetry gives up.
	s.Equal([]any{"exhausted", "exhausted", "exhausted"}, jobs.TestRetryableResult)
}

var (
	testQueueArgs = []contractsqueue.Arg{
		{
			Type:  "bool",
			Value: true,
		},
		{
			Type:  "int",
			Value: 1,
		},
		{
			Type:  "int8",
			Value: int8(1),
		},
		{
			Type:  "int16",
			Value: int16(1),
		},
		{
			Type:  "int32",
			Value: int32(1),
		},
		{
			Type:  "int64",
			Value: int64(1),
		},
		{
			Type:  "uint",
			Value: uint(1),
		},
		{
			Type:  "uint8",
			Value: uint8(1),
		},
		{
			Type:  "uint16",
			Value: uint16(1),
		},
		{
			Type:  "uint32",
			Value: uint32(1),
		},
		{
			Type:  "uint64",
			Value: uint64(1),
		},
		{
			Type:  "float32",
			Value: float32(1.1),
		},
		{
			Type:  "float64",
			Value: float64(1.2),
		},
		{
			Type:  "string",
			Value: "test",
		},
		{
			Type:  "[]bool",
			Value: []bool{true, false},
		},
		{
			Type:  "[]int",
			Value: []int{1, 2, 3},
		},
		{
			Type:  "[]int8",
			Value: []int8{1, 2, 3},
		},
		{
			Type:  "[]int16",
			Value: []int16{1, 2, 3},
		},
		{
			Type:  "[]int32",
			Value: []int32{1, 2, 3},
		},
		{
			Type:  "[]int64",
			Value: []int64{1, 2, 3},
		},
		{
			Type:  "[]uint",
			Value: []uint{1, 2, 3},
		},
		{
			Type:  "[]uint8",
			Value: []uint8{1, 2, 3},
		},
		{
			Type:  "[]uint16",
			Value: []uint16{1, 2, 3},
		},
		{
			Type:  "[]uint32",
			Value: []uint32{1, 2, 3},
		},
		{
			Type:  "[]uint64",
			Value: []uint64{1, 2, 3},
		},
		{
			Type:  "[]float32",
			Value: []float32{1.1, 1.2, 1.3},
		},
		{
			Type:  "[]float64",
			Value: []float64{1.1, 1.2, 1.3},
		},
		{
			Type:  "[]string",
			Value: []string{"test", "test2", "test3"},
		},
	}
)
