package feature

import (
	"context"
	"errors"
	"sync"
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

	timeoutDriver    *receiveTimeoutDriver
	savedConnections map[string]any
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, &QueueTestSuite{})
}

const timeoutTestConnection = "timeout_receive"

var (
	_ contractsqueue.Driver            = (*receiveTimeoutDriver)(nil)
	_ contractsqueue.DriverWithReceive = (*receiveTimeoutDriver)(nil)
)

// receiveCall records one Receive invocation: its arguments and the deadline of
// the context the worker passed in.
type receiveCall struct {
	queue       string
	count       int
	entered     time.Time
	deadline    time.Time
	hasDeadline bool
}

// receiveTimeoutDriver implements Driver + DriverWithReceive and records every
// Receive call, so the suite can prove `queue.connections.<connection>.timeout`
// reaches the blocking receive call and bounds it.
type receiveTimeoutDriver struct {
	mu            sync.Mutex
	receivedCalls []receiveCall
	popQueues     []string
	deliver       bool // when true, a non-blocked Receive hands back one job
	block         bool // when true, the first Receive waits until the context deadline
	deliveredJob  *timeoutReservedJob
	firstReceive  chan struct{}
	secondReceive chan struct{}
	popSignal     chan struct{}
	firstOnce     sync.Once
	secondOnce    sync.Once
	popOnce       sync.Once
}

func newReceiveTimeoutDriver() *receiveTimeoutDriver {
	return &receiveTimeoutDriver{
		firstReceive:  make(chan struct{}),
		secondReceive: make(chan struct{}),
		popSignal:     make(chan struct{}),
	}
}

func (d *receiveTimeoutDriver) Driver() string                         { return contractsqueue.DriverCustom }
func (d *receiveTimeoutDriver) Push(contractsqueue.Task, string) error { return nil }
func (d *receiveTimeoutDriver) Pop(queue string) (contractsqueue.ReservedJob, error) {
	d.mu.Lock()
	d.popQueues = append(d.popQueues, queue)
	d.mu.Unlock()

	d.popOnce.Do(func() { close(d.popSignal) })
	return nil, nil
}

func (d *receiveTimeoutDriver) Receive(ctx context.Context, queue string, count int) ([]contractsqueue.ReservedJob, error) {
	deadline, hasDeadline := ctx.Deadline()

	d.mu.Lock()
	d.receivedCalls = append(d.receivedCalls, receiveCall{
		queue:       queue,
		count:       count,
		entered:     time.Now(),
		deadline:    deadline,
		hasDeadline: hasDeadline,
	})
	callNumber := len(d.receivedCalls)
	// Only the first call blocks; later calls take the delivery path so the test
	// can prove a job is still handled after a receive timeout.
	block := d.block && callNumber == 1
	deliver := d.deliver && !block
	if deliver {
		d.deliver = false
		d.deliveredJob = &timeoutReservedJob{}
	}
	job := d.deliveredJob
	d.mu.Unlock()

	d.firstOnce.Do(func() { close(d.firstReceive) })
	if callNumber >= 2 {
		d.secondOnce.Do(func() { close(d.secondReceive) })
	}

	if block {
		<-ctx.Done()
		return nil, nil
	}
	if deliver {
		return []contractsqueue.ReservedJob{job}, nil
	}
	return nil, nil
}

// calls returns a copy of every recorded Receive invocation. Returning a copy
// lets tests assert on all calls without racing the worker goroutine.
func (d *receiveTimeoutDriver) calls() []receiveCall {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]receiveCall(nil), d.receivedCalls...)
}

// pops returns a copy of every queue name passed to Pop.
func (d *receiveTimeoutDriver) pops() []string {
	d.mu.Lock()
	defer d.mu.Unlock()
	return append([]string(nil), d.popQueues...)
}

// delivered returns the job the fake handed back, if any. Reading it under the
// driver mutex keeps the pointer access race-free.
func (d *receiveTimeoutDriver) delivered() *timeoutReservedJob {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.deliveredJob
}

// timeoutReservedJob is the minimal ReservedJob the fake driver returns to prove
// a job is still processed and acknowledged when a custom timeout is configured.
type timeoutReservedJob struct {
	mu          sync.Mutex
	deleteCount int
}

func (j *timeoutReservedJob) Attempts() int { return 1 }

func (j *timeoutReservedJob) Delete() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.deleteCount++
	return nil
}

func (j *timeoutReservedJob) Release(time.Duration) error { return nil }
func (j *timeoutReservedJob) Task() contractsqueue.Task {
	return contractsqueue.Task{
		ChainJob: contractsqueue.ChainJob{
			Job:  &jobs.Test{},
			Args: []contractsqueue.Arg{{Type: "string", Value: "timeout-job"}},
		},
	}
}

// deletes returns how many times the reservation was acknowledged.
func (j *timeoutReservedJob) deletes() int {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.deleteCount
}

// startWorker runs worker in the background and returns an idempotent stop func
// that shuts it down and waits for Run to return.
//
// Callers must observe a first Receive/Pop before calling stop; stopping before
// the worker has started can race its startup. stop bounds both Shutdown and the
// Run wait with timeouts so a wedged worker is reported instead of hanging.
func (s *QueueTestSuite) startWorker(worker contractsqueue.Worker) (stop func()) {
	s.T().Helper()

	errCh := make(chan error, 1)
	go func() { errCh <- worker.Run() }()

	var once sync.Once
	return func() {
		once.Do(func() {
			shutdownCh := make(chan error, 1)
			go func() { shutdownCh <- worker.Shutdown() }()

			select {
			case err := <-shutdownCh:
				s.NoError(err, "worker shutdown")
			case <-time.After(2 * time.Second):
				s.Fail("worker shutdown did not return; worker may still be running")
			}

			select {
			case err := <-errCh:
				s.NoError(err, "worker run")
			case <-time.After(2 * time.Second):
				s.Fail("worker did not stop")
			}
		})
	}
}

// SetupSuite snapshots a copy of the queue connections so TearDownSuite can
// restore them without mutating config-owned state.
func (s *QueueTestSuite) SetupSuite() {
	connections, ok := facades.Config().Get("queue.connections").(map[string]any)
	s.Require().True(ok, "queue.connections is not a map")

	saved := make(map[string]any, len(connections))
	for name, connection := range connections {
		saved[name] = connection
	}
	s.savedConnections = saved
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

// TearDownSuite restores the snapshotted queue connections so the test-only
// connection does not leak into later suites.
func (s *QueueTestSuite) TearDownSuite() {
	if s.savedConnections == nil {
		return
	}

	facades.Config().Add("queue.connections", s.savedConnections)
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

// TestDatabaseConnectionTimeoutScaffold asserts the documented scaffold value on
// the database connection, which the timeout tests otherwise never read.
func (s *QueueTestSuite) TestDatabaseConnectionTimeoutScaffold() {
	s.Equal(5, facades.Config().GetInt("queue.connections.database.timeout"))
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

func (s *QueueTestSuite) TestConfiguredTimeoutBoundsReceive() {
	cases := []struct {
		name    string
		include bool
		value   any
		expect  time.Duration
	}{
		{name: "integer seconds", include: true, value: 10, expect: 10 * time.Second},
		{name: "fractional seconds", include: true, value: 2.5, expect: 2500 * time.Millisecond},
		{name: "numeric string", include: true, value: "3", expect: 3 * time.Second},
		{name: "duration string", include: true, value: "1500ms", expect: 1500 * time.Millisecond},
		{name: "missing key falls back to default", include: false, value: nil, expect: contractsqueue.DefaultReceiveTimeout},
		{name: "zero falls back to default", include: true, value: 0, expect: contractsqueue.DefaultReceiveTimeout},
		{name: "negative falls back to default", include: true, value: -3, expect: contractsqueue.DefaultReceiveTimeout},
		{name: "unparsable string falls back to default", include: true, value: "not-a-duration", expect: contractsqueue.DefaultReceiveTimeout},
	}

	for _, tt := range cases {
		s.Run(tt.name, func() {
			// testify does not run SetupTest between subtests; allocate a fresh
			// driver and vary the connection map directly.
			s.timeoutDriver = newReceiveTimeoutDriver()
			s.registerTimeoutConnection(tt.include, tt.value)

			s.runTimeoutWorker()

			calls := s.timeoutDriver.calls()
			s.Require().NotEmpty(calls, "Receive should have been called")
			s.Equal("critical", calls[0].queue)
			s.Equal(3, calls[0].count)
			s.Require().True(calls[0].hasDeadline, "Receive context should carry a deadline")
			s.WithinDuration(calls[0].entered.Add(tt.expect), calls[0].deadline, 200*time.Millisecond)
		})
	}
}

func (s *QueueTestSuite) TestReceiveProcessesJobWithConfiguredTimeout() {
	s.timeoutDriver = newReceiveTimeoutDriver()
	s.registerTimeoutConnection(true, "1s")
	s.timeoutDriver.deliver = true

	s.runTimeoutWorker() // stop waits for Run, which waits for Handle/Delete to complete

	s.Equal([]any{"timeout-job"}, jobs.TestResult)
	job := s.timeoutDriver.delivered()
	s.Require().NotNil(job)
	s.Equal(1, job.deletes())
}

func (s *QueueTestSuite) TestBlockingReceiveIsBoundedByConfiguredTimeout() {
	s.timeoutDriver = newReceiveTimeoutDriver()
	s.registerTimeoutConnection(true, "300ms")
	s.timeoutDriver.block = true
	s.timeoutDriver.deliver = true

	stop := s.startTimeoutWorker()
	defer stop()

	select {
	case <-s.timeoutDriver.firstReceive:
	case <-time.After(2 * time.Second):
		s.FailNow("Receive was not called")
	}

	// The first Receive blocks until the 300ms deadline, so a second Receive
	// only happens if the configured timeout actually bounded the first one.
	select {
	case <-s.timeoutDriver.secondReceive:
	case <-time.After(2 * time.Second):
		s.FailNow("worker did not issue a second Receive")
	}

	// Stop explicitly so the post-timeout delivery has been processed and
	// acknowledged before asserting; the deferred stop is then a no-op.
	stop()

	s.Equal([]any{"timeout-job"}, jobs.TestResult)
	job := s.timeoutDriver.delivered()
	s.Require().NotNil(job)
	s.Equal(1, job.deletes())

	// Asserting both calls guards against a regression that builds the timeout
	// context once and reuses it across loop iterations.
	calls := s.timeoutDriver.calls()
	s.Require().GreaterOrEqual(len(calls), 2)
	s.Equal("critical", calls[0].queue)
	s.Equal(3, calls[0].count)
	s.Require().True(calls[0].hasDeadline, "Receive context should carry a deadline")
	s.WithinDuration(calls[0].entered.Add(300*time.Millisecond), calls[0].deadline, 100*time.Millisecond)
	s.Equal("critical", calls[1].queue)
	s.Equal(3, calls[1].count)
	s.Require().True(calls[1].hasDeadline, "Receive context should carry a deadline")
	s.WithinDuration(calls[1].entered.Add(300*time.Millisecond), calls[1].deadline, 100*time.Millisecond)
}

func (s *QueueTestSuite) TestShutdownCancelsBlockedReceive() {
	s.timeoutDriver = newReceiveTimeoutDriver()
	s.registerTimeoutConnection(true, "30s")
	s.timeoutDriver.block = true

	stop := s.startTimeoutWorker()
	defer stop()

	select {
	case <-s.timeoutDriver.firstReceive:
	case <-time.After(2 * time.Second):
		s.FailNow("Receive was not called")
	}

	start := time.Now()
	stop()
	s.Less(time.Since(start), time.Second, "Shutdown should cancel the in-flight Receive")
}

func (s *QueueTestSuite) TestReceiveCapableDriverUsesPopForQueueList() {
	s.timeoutDriver = newReceiveTimeoutDriver()
	s.registerTimeoutConnection(false, nil)

	// A comma-separated queue list is consumed via Pop to preserve priority
	// order, so the receive path must not be taken even though the driver
	// implements DriverWithReceive.
	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: timeoutTestConnection,
		Queue:      "default,high",
		Concurrent: 1,
	})
	stop := s.startWorker(worker)
	defer stop()

	select {
	case <-s.timeoutDriver.popSignal:
	case <-time.After(2 * time.Second):
		s.FailNow("Pop was not called")
	}

	s.Require().Eventually(func() bool {
		return len(s.timeoutDriver.pops()) >= 2
	}, 2*time.Second, 10*time.Millisecond, "expected both queues to be polled")

	s.Equal([]string{"default", "high"}, s.timeoutDriver.pops()[:2])
	s.Empty(s.timeoutDriver.calls(), "Receive must not be called for a multi-queue worker")
}

// registerTimeoutConnection installs the test-only receive connection for the
// current case. The worker reads `timeout` and `via` once at construction, so
// the connection must be registered before facades.Queue().Worker(...).
func (s *QueueTestSuite) registerTimeoutConnection(includeTimeout bool, timeout any) {
	driver := s.timeoutDriver
	connection := map[string]any{
		"driver":     "custom",
		"connection": "default",
		"queue":      "default",
		"concurrent": 1,
		"via": func() (contractsqueue.Driver, error) {
			return driver, nil
		},
	}
	if includeTimeout {
		connection["timeout"] = timeout
	}
	facades.Config().Add("queue.connections."+timeoutTestConnection, connection)
}

// startTimeoutWorker starts a worker on the test connection. The queue and
// concurrency differ from the connection defaults so the receive assertions
// prove the worker args are forwarded rather than read from config.
func (s *QueueTestSuite) startTimeoutWorker() (stop func()) {
	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: timeoutTestConnection,
		Queue:      "critical",
		Concurrent: 3,
	})

	return s.startWorker(worker)
}

// runTimeoutWorker starts a worker on the test connection, waits for the first
// Receive, then stops the worker.
func (s *QueueTestSuite) runTimeoutWorker() {
	stop := s.startTimeoutWorker()
	defer stop()

	select {
	case <-s.timeoutDriver.firstReceive:
	case <-time.After(2 * time.Second):
		s.FailNow("Receive was not called")
	}
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
