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

// QueueUniqueTestSuite covers a job that implements queue.JobWithUnique: it
// takes a cache lock on dispatch, so a second push with the same unique ID is
// skipped while the first one is pending or running. The lock is released when
// the job succeeds or fails for good, not while it waits for a retry; UniqueFor
// bounds how long it is held, and DispatchSync ignores it.
type QueueUniqueTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestQueueUniqueTestSuite(t *testing.T) {
	suite.Run(t, &QueueUniqueTestSuite{})
}

func (s *QueueUniqueTestSuite) SetupSuite() {
	// The app's own database queue runners consume the very rows these tests
	// count, so turn them off and drive the worker by hand. The lock lives in
	// cache.default, and redis is the store a real worker shares with the web
	// process.
	scope, err := tests.OverrideConfig(map[string]any{
		"cache.default":        "redis",
		"app.disabled_runners": []string{"goravel:queue", "app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	s.Equal("redis", facades.Config().GetString("cache.default"))
	s.T().Cleanup(func() { s.NoError(scope.Restore()) })
}

func (s *QueueUniqueTestSuite) SetupTest() {
	s.RefreshDatabase()

	jobs.ResetTestUnique()
	jobs.ResetTestUniqueRetryable()
}

// startWorker runs a database worker in the background and returns an
// idempotent stop func that shuts it down and waits for Run to return.
func (s *QueueUniqueTestSuite) startWorker() (stop func()) {
	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "default",
		Concurrent: 1,
	})
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

// uniqueArgs builds the dispatch args that feed UniqueID.
func uniqueArgs(id string) []contractsqueue.Arg {
	return []contractsqueue.Arg{{Type: "string", Value: id}}
}

// pendingJobs returns the number of rows waiting in the jobs table.
func (s *QueueUniqueTestSuite) pendingJobs() int64 {
	count, err := facades.DB().Table("jobs").Count()
	s.Require().NoError(err)

	return count
}

// releaseLock drops the unique lock of a job, so a lock left behind by a failed
// assertion cannot outlive the suite and break the next -count run.
func (s *QueueUniqueTestSuite) releaseLock(signature, id string) {
	key := "goravel:queue-unique:" + signature
	if id != "" {
		key += ":" + id
	}

	facades.Cache().Lock(key).ForceRelease()
}

func (s *QueueUniqueTestSuite) TestUniqueJobIsSkippedWhilePending() {
	s.T().Cleanup(func() {
		s.releaseLock("test_unique", "pending")
		s.releaseLock("test_unique", "other")
	})

	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("pending")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	// The same unique ID is already pending: the push is skipped and returns
	// nil instead of an error.
	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("pending")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	// Another unique ID is a different job and is pushed.
	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("other")).OnConnection("database").Dispatch())
	s.Equal(int64(2), s.pendingJobs())
}

func (s *QueueUniqueTestSuite) TestUniqueJobWithoutUniqueIDIsUniqueBySignature() {
	s.T().Cleanup(func() { s.releaseLock("test_unique", "") })

	s.NoError(facades.Queue().Job(&jobs.TestUnique{}).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	s.NoError(facades.Queue().Job(&jobs.TestUnique{}).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())
}

func (s *QueueUniqueTestSuite) TestUniqueJobIsReleasedWhenItSucceeds() {
	s.T().Cleanup(func() { s.releaseLock("test_unique", "success") })

	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("success")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	stop := s.startWorker()
	defer stop()

	s.Require().Eventually(func() bool {
		return s.pendingJobs() == 0 && jobs.TestUniqueResultLen() == 1
	}, 5*time.Second, 25*time.Millisecond)

	// The job is done, so the lock is gone and the same unique ID is pushed
	// again instead of being skipped: the worker handles the second copy too.
	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("success")).OnConnection("database").Dispatch())

	s.Require().Eventually(func() bool {
		return s.pendingJobs() == 0 && jobs.TestUniqueResultLen() == 2
	}, 5*time.Second, 25*time.Millisecond)

	stop()
	s.Equal([]any{"success", "success"}, jobs.TestUniqueResult)
}

func (s *QueueUniqueTestSuite) TestUniqueJobLockIsHeldWhileItWaitsForRetry() {
	s.T().Cleanup(func() { s.releaseLock("test_unique_retryable", "retry") })

	// The job fails on its first attempt, is released for one retry 2s later,
	// and gives up on the second attempt.
	jobs.TestUniqueRetryableNeverSucceed = true
	jobs.TestUniqueRetryableFailUntil = 1
	jobs.TestUniqueRetryableDelay = 2 * time.Second

	s.NoError(facades.Queue().Job(&jobs.TestUniqueRetryable{}, uniqueArgs("retry")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	stop := s.startWorker()
	defer stop()

	// The first attempt failed and the job is back in the queue waiting for
	// its retry. The lock must still be held, so a duplicate dispatch is
	// skipped and the queue keeps a single row.
	s.Require().Eventually(func() bool {
		return jobs.TestUniqueRetryableResultLen() == 1
	}, 5*time.Second, 25*time.Millisecond)

	s.NoError(facades.Queue().Job(&jobs.TestUniqueRetryable{}, uniqueArgs("retry")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	// The retry ran and failed for good: the job lands in failed_jobs and the
	// lock is released.
	s.Require().Eventually(func() bool {
		failedJobs, err := facades.Queue().Failer().All()
		return err == nil && len(failedJobs) == 1
	}, 10*time.Second, 25*time.Millisecond)

	// The failed job is recorded before the queue row is deleted, so wait for
	// the row to disappear before asserting on the queue size.
	s.Require().Eventually(func() bool {
		return s.pendingJobs() == 0
	}, 5*time.Second, 25*time.Millisecond)

	stop()

	// The failure was final, so the lock is gone and the same unique ID is
	// pushed again.
	s.NoError(facades.Queue().Job(&jobs.TestUniqueRetryable{}, uniqueArgs("retry")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())
}

func (s *QueueUniqueTestSuite) TestUniqueJobLockExpiresAfterUniqueFor() {
	s.T().Cleanup(func() { s.releaseLock("test_unique", "ttl") })

	s.NoError(facades.Queue().Job(jobs.NewTestUnique(500*time.Millisecond), uniqueArgs("ttl")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	// Held while UniqueFor has not elapsed.
	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("ttl")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	time.Sleep(700 * time.Millisecond)

	// The lock expired, so the same unique ID is pushed again.
	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("ttl")).OnConnection("database").Dispatch())
	s.Equal(int64(2), s.pendingJobs())
}

func (s *QueueUniqueTestSuite) TestDispatchSyncIgnoresUniqueness() {
	s.T().Cleanup(func() { s.releaseLock("test_unique", "sync") })

	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("sync")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())

	// DispatchSync runs the job right away, even though the unique lock of the
	// pending copy is held, and leaves the lock untouched.
	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("sync")).DispatchSync())
	s.Equal([]any{"sync"}, jobs.TestUniqueResult)
	s.Equal(int64(1), s.pendingJobs())

	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("sync")).OnConnection("database").Dispatch())
	s.Equal(int64(1), s.pendingJobs())
}

func (s *QueueUniqueTestSuite) TestUniqueJobOnSyncConnectionRunsEveryTime() {
	if facades.Config().GetString("queue.default") != "sync" {
		s.T().Skip("skip test due to only for sync")
	}

	// The sync driver runs the job inside Dispatch and releases the lock right
	// away, so a second dispatch is not skipped.
	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("sync-connection")).Dispatch())
	s.NoError(facades.Queue().Job(&jobs.TestUnique{}, uniqueArgs("sync-connection")).Dispatch())

	s.Equal([]any{"sync-connection", "sync-connection"}, jobs.TestUniqueResult)
}
