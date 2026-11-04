package feature

import (
	"errors"
	"maps"
	"slices"
	"testing"

	frameworkconsole "github.com/goravel/framework/console"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/console/command"
	"github.com/stretchr/testify/suite"

	"goravel/app/console/commands"
	"goravel/app/console/hooks"
	"goravel/app/facades"
	"goravel/bootstrap"
	"goravel/tests"
)

type ConsoleHookTestSuite struct {
	suite.Suite
	tests.TestCase
}

// conflictingFlagHook adds --target, which test:console-hook already defines,
// so building an Artisan application with it must fail.
type conflictingFlagHook struct {
	contractsconsole.BaseCommandHook
}

func (conflictingFlagHook) Flags() []command.Flag {
	return []command.Flag{
		&command.StringFlag{Name: "target"},
	}
}

func TestConsoleHookTestSuite(t *testing.T) {
	suite.Run(t, new(ConsoleHookTestSuite))
}

func (s *ConsoleHookTestSuite) SetupTest() {
	hooks.Reset()
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

func (s *ConsoleHookTestSuite) TestAfterReplacesHandleError() {
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

func (s *ConsoleHookTestSuite) TestCommandHelpSucceedsWithHookFlag() {
	s.NoError(facades.Artisan().Call("test:console-hook --help"))
}

func (s *ConsoleHookTestSuite) TestHookFlagConflictFailsRegistration() {
	artisan := frameworkconsole.NewApplication(
		"artisan", "Goravel Framework", "artisan command [options] [arguments...]", "1.0.0", true,
		map[string][]contractsconsole.CommandHook{
			"test:console-hook": {conflictingFlagHook{}},
		},
	)
	artisan.Register(bootstrap.Commands())

	s.EqualError(
		artisan.Call("test:console-hook"),
		"failed to register command 'test:console-hook': command hook flag 'target' conflicts with an existing flag",
	)
}

func (s *ConsoleHookTestSuite) TestGlobalGlobHookIsInertForOtherCommands() {
	s.NoError(facades.Artisan().Call("test:console-shutdownable"))

	s.Nil(hooks.GetLatest(), "the '*' hook must not record anything outside test:console-hook")
}
