package feature

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/str"
	"github.com/stretchr/testify/suite"

	"goravel/app/events"
	"goravel/app/facades"
	"goravel/app/listeners"
	"goravel/tests"
)

type EventTestSuite struct {
	suite.Suite
	tests.TestCase
	counter uint64
}

func TestEventTestSuite(t *testing.T) {
	suite.Run(t, &EventTestSuite{})
}

func (s *EventTestSuite) SetupTest() {
	listeners.TestResultOfSendShipmentNotification = nil
}

func (s *EventTestSuite) TestDispatchBootstrappedEvents() {
	s.NoError(facades.Event().Dispatch(&events.OrderShipped{}, []event.Arg{
		{Type: "string", Value: "I'm OrderShipped"},
	}).Error())

	s.NoError(facades.Event().Dispatch(&events.OrderCanceled{}, []event.Arg{
		{Type: "string", Value: "I'm OrderCanceled"},
	}).Error())

	s.True(waitUntil(3*time.Second, 20*time.Millisecond, func() bool {
		return len(listeners.TestResultOfSendShipmentNotification) == 2
	}))

	s.ElementsMatch([]string{
		"I'm OrderShipped",
		"I'm OrderCanceled",
	}, listeners.TestResultOfSendShipmentNotification)
}

func (s *EventTestSuite) TestDispatchUnregisteredEvent() {
	eventInstance := &unregisteredIntegrationEvent{}

	// Dispatching an event nobody listens to is a silent success: the
	// deprecated Task's EventListenerNotBind error does not exist for Dispatch.
	result := facades.Event().Dispatch(eventInstance)

	s.False(result.Failed())
	s.NoError(result.Error())
}

func (s *EventTestSuite) TestDispatchReturnsEventHandleError() {
	expectedErr := errors.New("event handle error")
	eventInstance := &dispatchHandleErrorEvent{
		integrationEvent: integrationEvent{
			handle: func(args []event.Arg) ([]event.Arg, error) {
				return nil, expectedErr
			},
		},
	}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature:   s.uniqueName("event_handle_error_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     capture,
	}
	s.NoError(facades.Event().Listen(eventInstance, listenerInstance))

	result := facades.Event().Dispatch(eventInstance, []event.Arg{
		{Type: "string", Value: "test"},
	})

	// The event's Handle error still short-circuits the dispatch, so the
	// registered listener must not run.
	s.True(result.Failed())
	s.ErrorIs(result.Error(), expectedErr)
	s.Empty(capture.Handled())
}

func (s *EventTestSuite) TestDispatchSyncListenerWithTransformedArgs() {
	eventInstance := &dispatchTransformedArgsEvent{
		integrationEvent: integrationEvent{
			handle: func(args []event.Arg) ([]event.Arg, error) {
				return []event.Arg{
					{Type: "string", Value: castString(args[0].Value) + "_transformed"},
					{Type: "int", Value: 2},
				}, nil
			},
		},
	}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature:   s.uniqueName("sync_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     capture,
	}
	s.NoError(facades.Event().Listen(eventInstance, listenerInstance))

	result := facades.Event().Dispatch(eventInstance, []event.Arg{
		{Type: "string", Value: "goravel"},
	})

	s.False(result.Failed())
	s.NoError(result.Error())
	s.Equal([][]any{
		{"goravel_transformed", 2},
	}, capture.Handled())
	s.Equal(1, capture.QueueCallCount())
}

func (s *EventTestSuite) TestDispatchRunsAllListenersAndCollectsErrors() {
	expectedErr := errors.New("listener handle error")
	eventInstance := &dispatchMultipleListenersEvent{}
	failedCapture := &listenerCapture{}
	otherCapture := &listenerCapture{}
	failedListener := &integrationListener{
		signature:   s.uniqueName("failed_listener"),
		queueConfig: event.Queue{Enable: false},
		handleErr:   expectedErr,
		capture:     failedCapture,
	}
	otherListener := &integrationListener{
		signature:   s.uniqueName("other_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     otherCapture,
	}
	s.NoError(facades.Event().Listen(eventInstance, failedListener, otherListener))

	// Dispatch runs every listener and aggregates the errors instead of
	// stopping after the first failure, so both listeners handle the payload.
	result := facades.Event().Dispatch(eventInstance, []event.Arg{
		{Type: "string", Value: "should not stop"},
	})

	s.True(result.Failed())
	s.ErrorIs(result.Error(), expectedErr)
	s.Len(failedCapture.Handled(), 1)
	s.Len(otherCapture.Handled(), 1)
}

func (s *EventTestSuite) TestDispatchQueuedListenerEventually() {
	eventInstance := &dispatchQueuedListenerEvent{}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature: s.uniqueName("queued_listener"),
		queueConfig: event.Queue{
			Enable: true,
		},
		capture: capture,
	}
	s.NoError(facades.Event().Listen(eventInstance, listenerInstance))

	result := facades.Event().Dispatch(eventInstance, []event.Arg{
		{Type: "string", Value: "queued"},
	})

	s.False(result.Failed())
	s.NoError(result.Error())
	s.True(waitUntil(5*time.Second, 20*time.Millisecond, func() bool {
		return len(capture.Handled()) == 1
	}))
	s.Equal([][]any{
		{"queued"},
	}, capture.Handled())
	s.Equal(1, capture.QueueCallCount())
}

func (s *EventTestSuite) TestCommandMakeEvent() {
	eventName := s.uniqueName("EventFeature")
	nestedPackage := s.uniqueName("EventFeatureNested")
	nestedEventName := s.uniqueName("GeneratedEvent")
	eventPath := path.App("events", str.Of(eventName).Snake().String()+".go")
	nestedDir := path.App("events", nestedPackage)
	nestedPath := path.App("events", nestedPackage, str.Of(nestedEventName).Snake().String()+".go")

	s.NoError(os.RemoveAll(eventPath))
	s.NoError(os.RemoveAll(nestedDir))
	s.T().Cleanup(func() {
		s.NoError(os.RemoveAll(eventPath))
		s.NoError(os.RemoveAll(nestedDir))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:event " + eventName))
	s.True(file.Exists(eventPath))
	s.True(file.Contains(eventPath, "type "+eventName+" struct {"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") Handle(args []event.Arg) ([]event.Arg, error)"))

	originalContent, err := os.ReadFile(eventPath)
	s.NoError(err)

	output, err := s.CaptureArtisanOutput("--no-ansi make:event " + eventName)
	s.NoError(err)
	s.Contains(output, "already exists")

	currentContent, err := os.ReadFile(eventPath)
	s.NoError(err)
	s.Equal(string(originalContent), string(currentContent))

	s.NoError(facades.Artisan().Call("--no-ansi make:event " + nestedPackage + "/" + nestedEventName))
	s.True(file.Exists(nestedPath))
	s.True(file.Contains(nestedPath, "package "+nestedPackage))
	s.True(file.Contains(nestedPath, "type "+nestedEventName+" struct {"))
}

func (s *EventTestSuite) TestCommandMakeEventBroadcast() {
	eventName := s.uniqueName("EventBroadcast")
	eventPath := path.App("events", str.Of(eventName).Snake().String()+".go")

	s.NoError(os.RemoveAll(eventPath))
	s.T().Cleanup(func() {
		s.NoError(os.RemoveAll(eventPath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:event --broadcast " + eventName))
	s.True(file.Exists(eventPath))
	s.True(file.Contains(eventPath, `"github.com/goravel/framework/contracts/broadcasting"`))
	s.True(file.Contains(eventPath, "var _ broadcasting.ShouldBroadcast = (*"+eventName+")(nil)"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastOn() []string {"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastAs() string {"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastWith() map[string]any {"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastWhen() bool {"))
	s.False(file.Contains(eventPath, `"github.com/goravel/framework/contracts/event"`))
	s.False(file.Contains(eventPath, "Handle("))
	s.False(file.Contains(eventPath, "BroadcastNow()"))
}

func (s *EventTestSuite) TestCommandMakeEventBroadcastNow() {
	eventName := s.uniqueName("EventBroadcastNow")
	eventPath := path.App("events", str.Of(eventName).Snake().String()+".go")

	s.NoError(os.RemoveAll(eventPath))
	s.T().Cleanup(func() {
		s.NoError(os.RemoveAll(eventPath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:event --broadcast --now " + eventName))
	s.True(file.Exists(eventPath))
	s.True(file.Contains(eventPath, `"github.com/goravel/framework/contracts/broadcasting"`))
	s.True(file.Contains(eventPath, "var _ broadcasting.ShouldBroadcast = (*"+eventName+")(nil)"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastOn() []string {"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastAs() string {"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastWith() map[string]any {"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastWhen() bool {"))
	s.True(file.Contains(eventPath, "func (receiver *"+eventName+") BroadcastNow() bool {\n\treturn true\n}"))
	s.False(file.Contains(eventPath, `"github.com/goravel/framework/contracts/event"`))
	s.False(file.Contains(eventPath, "Handle("))
}

func (s *EventTestSuite) TestCommandMakeListener() {
	listenerName := s.uniqueName("ListenerFeature")
	nestedPackage := s.uniqueName("ListenerFeatureNested")
	nestedListenerName := s.uniqueName("GeneratedListener")
	listenerPath := path.App("listeners", str.Of(listenerName).Snake().String()+".go")
	nestedDir := path.App("listeners", nestedPackage)
	nestedPath := path.App("listeners", nestedPackage, str.Of(nestedListenerName).Snake().String()+".go")

	s.NoError(os.RemoveAll(listenerPath))
	s.NoError(os.RemoveAll(nestedDir))
	s.T().Cleanup(func() {
		s.NoError(os.RemoveAll(listenerPath))
		s.NoError(os.RemoveAll(nestedDir))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:listener " + listenerName))
	s.True(file.Exists(listenerPath))
	s.True(file.Contains(listenerPath, "type "+listenerName+" struct {"))
	s.True(file.Contains(listenerPath, "func (receiver *"+listenerName+") Signature() string {"))
	s.True(file.Contains(listenerPath, `return "`+str.Of(listenerName).Snake().String()+`"`))

	originalContent, err := os.ReadFile(listenerPath)
	s.NoError(err)

	output, err := s.CaptureArtisanOutput("--no-ansi make:listener " + listenerName)
	s.NoError(err)
	s.Contains(output, "already exists")

	currentContent, err := os.ReadFile(listenerPath)
	s.NoError(err)
	s.Equal(string(originalContent), string(currentContent))

	s.NoError(facades.Artisan().Call("--no-ansi make:listener " + nestedPackage + "/" + nestedListenerName))
	s.True(file.Exists(nestedPath))
	s.True(file.Contains(nestedPath, "package "+nestedPackage))
	s.True(file.Contains(nestedPath, "type "+nestedListenerName+" struct {"))
	s.True(file.Contains(nestedPath, `return "`+str.Of(nestedListenerName).Snake().String()+`"`))
}

func (s *EventTestSuite) uniqueName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, atomic.AddUint64(&s.counter, 1))
}

type integrationEvent struct {
	handle func(args []event.Arg) ([]event.Arg, error)
}

func (receiver *integrationEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	if receiver.handle == nil {
		return args, nil
	}

	return receiver.handle(args)
}

// unregisteredIntegrationEvent is the payload of TestDispatchUnregisteredEvent.
// It must not be integrationEvent itself: Listen appends registrations by the
// event's type name, so giving this payload its own never-registered type name
// guarantees the dispatch can't hit a foreign registration. Embedding gives
// this type the same Handle behaviour under its own name.
type unregisteredIntegrationEvent struct {
	integrationEvent
}

// The event types below give each dispatch test its own event name. Listen
// appends registrations instead of overwriting them, so two tests sharing the
// plain integrationEvent name would accumulate each other's listeners and make
// every dispatch run foreign captures.
type dispatchHandleErrorEvent struct {
	integrationEvent
}

type dispatchTransformedArgsEvent struct {
	integrationEvent
}

type dispatchMultipleListenersEvent struct {
	integrationEvent
}

type dispatchQueuedListenerEvent struct {
	integrationEvent
}

type integrationListener struct {
	signature   string
	queueConfig event.Queue
	handleErr   error
	capture     *listenerCapture
}

func (receiver *integrationListener) Signature() string {
	return receiver.signature
}

func (receiver *integrationListener) Queue(args ...any) event.Queue {
	if receiver.capture != nil {
		receiver.capture.AddQueueArgs(args)
	}

	return receiver.queueConfig
}

func (receiver *integrationListener) Handle(eventName string, args ...any) error {
	if receiver.capture != nil {
		receiver.capture.AddHandled(args)
	}

	return receiver.handleErr
}

type listenerCapture struct {
	mu        sync.Mutex
	handled   [][]any
	queueArgs [][]any
}

func (receiver *listenerCapture) AddHandled(args []any) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.handled = append(receiver.handled, copyAnySlice(args))
}

func (receiver *listenerCapture) AddQueueArgs(args []any) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.queueArgs = append(receiver.queueArgs, copyAnySlice(args))
}

func (receiver *listenerCapture) Handled() [][]any {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	result := make([][]any, len(receiver.handled))
	for i, args := range receiver.handled {
		result[i] = copyAnySlice(args)
	}

	return result
}

func (receiver *listenerCapture) QueueCallCount() int {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	return len(receiver.queueArgs)
}

func copyAnySlice(args []any) []any {
	copyArgs := make([]any, len(args))
	copy(copyArgs, args)

	return copyArgs
}

func castString(value any) string {
	result, ok := value.(string)
	if ok {
		return result
	}

	return fmt.Sprintf("%v", value)
}

// waitUntil polls until condition returns true or timeout occurs.
func waitUntil(timeout, interval time.Duration, condition func() bool) bool {
	deadline := time.Now().Add(timeout)
	for {
		if condition() {
			return true
		}
		if time.Now().After(deadline) {
			return false
		}

		time.Sleep(interval)
	}
}
