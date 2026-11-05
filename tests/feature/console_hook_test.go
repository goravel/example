package feature

import (
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/console/commands"
	"goravel/app/console/hooks"
	"goravel/app/facades"
	"goravel/tests"
)

type ConsoleHookTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestConsoleHookTestSuite(t *testing.T) {
	suite.Run(t, new(ConsoleHookTestSuite))
}

func (s *ConsoleHookTestSuite) SetupTest() {
	hooks.Reset()
	commands.ResetConsoleCaptures()
}

func (s *ConsoleHookTestSuite) TearDownTest() {
	hooks.Reset()
}

func (s *ConsoleHookTestSuite) TestCommandHooksAreRegisteredOnTheApp() {
	s.Equal(
		[]string{"*", "test:console-hook", "test:console-hook*"},
		slices.Sorted(maps.Keys(facades.App().CommandHooks())),
	)
}

func (s *ConsoleHookTestSuite) TestHooksRunAroundHandle() {
	s.NoError(facades.Artisan().Call("test:console-hook --target=production --approved"))

	s.Equal(&hooks.RunCapture{
		Events: []string{
			"trace:before",
			"audit:before",
			"approval:before",
			"handle",
			"audit:after",
			"trace:after",
		},
		Target:   "production",
		Approved: true,
	}, hooks.GetLatest())
}

func (s *ConsoleHookTestSuite) TestBeforeFalseSkipsHandleAndAfter() {
	s.NoError(facades.Artisan().Call("test:console-hook --target=staging"))

	s.Equal(&hooks.RunCapture{
		Events: []string{"trace:before", "audit:before", "approval:before"},
		Target: "staging",
	}, hooks.GetLatest())
}

func (s *ConsoleHookTestSuite) TestBeforeErrorSkipsHandleAndAfter() {
	err := facades.Artisan().Call("test:console-hook --target=production")

	s.Equal(hooks.ErrApprovalRequired, err)
	s.Equal(&hooks.RunCapture{
		Events: []string{"trace:before", "audit:before", "approval:before"},
		Target: "production",
	}, hooks.GetLatest())
}

// TestAfterErrorReplacesHandleError proves the asymmetry with Before: a Before
// error skips Handle and every After (TestBeforeErrorSkipsHandleAndAfter), while
// an After error does not abort the run — it is passed to the remaining After
// hooks and finally returned to the caller.
func (s *ConsoleHookTestSuite) TestAfterErrorReplacesHandleError() {
	err := facades.Artisan().Call("test:console-hook --target=production --approved --fail")

	s.EqualError(err, "audit hook: publish failed")
	s.True(errors.Is(err, commands.ErrConsoleHookPublishFailed), "the After hook must wrap, not drop, the error")
	s.Equal(&hooks.RunCapture{
		Events: []string{
			"trace:before",
			"audit:before",
			"approval:before",
			"handle",
			"audit:after",
			"trace:after",
		},
		Target:    "production",
		Approved:  true,
		HandleErr: commands.ErrConsoleHookPublishFailed,
	}, hooks.GetLatest())
}

func (s *ConsoleHookTestSuite) TestCommandHelpExitsCleanly() {
	s.NoError(facades.Artisan().Call("test:console-hook --help"))
}

func (s *ConsoleHookTestSuite) TestGlobalGlobHookIsInertForOtherCommands() {
	s.NoError(facades.Artisan().Call("test:console-shutdownable"))

	s.Nil(hooks.GetLatest(), "the '*' hook must not record anything outside test:console-hook")
}
