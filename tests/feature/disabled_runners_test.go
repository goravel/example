package feature

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	contractsfoundation "github.com/goravel/framework/contracts/foundation"
	"github.com/stretchr/testify/suite"

	"goravel/bootstrap"
	"goravel/tests"
)

type trackerRunner struct {
	sig     string
	started chan struct{}
	done    chan struct{}
	once    sync.Once
	didRun  atomic.Bool
}

func newTrackerRunner(sig string) *trackerRunner {
	return &trackerRunner{
		sig:     sig,
		started: make(chan struct{}),
		done:    make(chan struct{}),
	}
}

func (r *trackerRunner) Signature() string { return r.sig }
func (r *trackerRunner) ShouldRun() bool   { return true }

func (r *trackerRunner) Run() error {
	r.didRun.Store(true)
	close(r.started)
	<-r.done
	return nil
}

func (r *trackerRunner) Shutdown() error {
	r.once.Do(func() { close(r.done) })
	return nil
}

func (r *trackerRunner) Started() bool { return r.didRun.Load() }

type DisabledRunnersTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestDisabledRunnersTestSuite(t *testing.T) {
	suite.Run(t, new(DisabledRunnersTestSuite))
}

func (s *DisabledRunnersTestSuite) SetupTest() {
	bootstrap.BootstrapRunnerList = nil
}

func (s *DisabledRunnersTestSuite) TearDownTest() {
	bootstrap.BootstrapRunnerList = nil
}

func (s *DisabledRunnersTestSuite) withRunners(runners []contractsfoundation.Runner, disabledPatterns []string) func() {
	bootstrap.BootstrapRunnerList = runners
	scope, err := tests.OverrideConfig(map[string]any{
		"app.disabled_runners": disabledPatterns,
	})
	s.Require().NoError(err)

	return func() {
		bootstrap.BootstrapRunnerList = nil
		s.Require().NoError(scope.Restore())
	}
}

func (s *DisabledRunnersTestSuite) waitForStart(runner *trackerRunner) {
	s.Require().Eventually(func() bool {
		return runner.Started()
	}, 3*time.Second, 50*time.Millisecond)
}

func (s *DisabledRunnersTestSuite) waitForNotStarted(runner *trackerRunner) {
	s.Require().Never(func() bool {
		return runner.Started()
	}, 2*time.Second, 50*time.Millisecond)
}

func (s *DisabledRunnersTestSuite) TestDisabledRunners_RunnerStartsWhenNotDisabled() {
	runner := newTrackerRunner("test:demo")
	cleanup := s.withRunners([]contractsfoundation.Runner{runner}, []string{})
	defer cleanup()

	s.waitForStart(runner)
}

func (s *DisabledRunnersTestSuite) TestDisabledRunners_RunnerSkippedWhenDisabled() {
	runner := newTrackerRunner("test:demo")
	cleanup := s.withRunners([]contractsfoundation.Runner{runner}, []string{"test:demo"})
	defer cleanup()

	s.waitForNotStarted(runner)
}

func (s *DisabledRunnersTestSuite) TestDisabledRunners_WildcardSkipsAll() {
	runner := newTrackerRunner("test:demo")
	cleanup := s.withRunners([]contractsfoundation.Runner{runner}, []string{"*"})
	defer cleanup()

	s.waitForNotStarted(runner)
}

func (s *DisabledRunnersTestSuite) TestDisabledRunners_NamespaceWildcard() {
	runner := newTrackerRunner("test:demo")
	cleanup := s.withRunners([]contractsfoundation.Runner{runner}, []string{"test:*"})
	defer cleanup()

	s.waitForNotStarted(runner)
}

func (s *DisabledRunnersTestSuite) TestDisabledRunners_SelectiveDisable() {
	alpha := newTrackerRunner("test:alpha")
	beta := newTrackerRunner("test:beta")
	cleanup := s.withRunners([]contractsfoundation.Runner{alpha, beta}, []string{"test:alpha"})
	defer cleanup()

	s.waitForNotStarted(alpha)
	s.waitForStart(beta)
}

func (s *DisabledRunnersTestSuite) TestDisabledRunners_RestoreEnablesRunner() {
	runner := newTrackerRunner("test:demo")
	cleanup := s.withRunners([]contractsfoundation.Runner{runner}, []string{"test:demo"})
	s.waitForNotStarted(runner)
	cleanup()

	runner2 := newTrackerRunner("test:demo")
	cleanup = s.withRunners([]contractsfoundation.Runner{runner2}, []string{})
	defer cleanup()
	s.waitForStart(runner2)
}
