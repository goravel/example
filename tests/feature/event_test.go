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
	frameworkerrors "github.com/goravel/framework/errors"
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
}

func TestEventTestSuite(t *testing.T) {
	suite.Run(t, &EventTestSuite{})
}

func (s *EventTestSuite) SetupTest() {
	listeners.TestResultOfSendShipmentNotification = nil

	// GetEvents returns a copy, so it can no longer be used to unregister. Events
	// are resolved by name now, so each scenario below declares its own event type
	// and Register overwrites only the listeners it registered for that name.
}

func (s *EventTestSuite) TestDispatchBootstrappedEvents() {
	s.NoError(facades.Event().Job(&events.OrderShipped{}, []event.Arg{
		{Type: "string", Value: "I'm OrderShipped"},
	}).Dispatch())

	s.NoError(facades.Event().Job(&events.OrderCanceled{}, []event.Arg{
		{Type: "string", Value: "I'm OrderCanceled"},
	}).Dispatch())

	s.True(waitUntil(3*time.Second, 20*time.Millisecond, func() bool {
		return len(listeners.TestResultOfSendShipmentNotification) == 2
	}))

	s.ElementsMatch([]string{
		"I'm OrderShipped",
		"I'm OrderCanceled",
	}, listeners.TestResultOfSendShipmentNotification)
}

func (s *EventTestSuite) TestDispatchUnregisteredEvent() {
	eventInstance := &integrationEvent{
		handle: func(args []event.Arg) ([]event.Arg, error) {
			return args, nil
		},
	}

	err := facades.Event().Job(eventInstance, nil).Dispatch()

	s.Equal(frameworkerrors.EventListenerNotBind.Args(eventInstance), err)
}

func (s *EventTestSuite) TestDispatchReturnsEventHandleError() {
	expectedErr := errors.New("event handle error")
	eventInstance := &handleErrorEvent{err: expectedErr}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature:   s.uniqueName("event_handle_error_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     capture,
	}
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			listenerInstance,
		},
	})

	err := facades.Event().Job(eventInstance, []event.Arg{
		{Type: "string", Value: "test"},
	}).Dispatch()

	s.Equal(expectedErr, err)
	s.Empty(capture.Handled())
}

func (s *EventTestSuite) TestDispatchSyncListenerWithTransformedArgs() {
	eventInstance := &transformingEvent{}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature:   s.uniqueName("sync_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     capture,
	}
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			listenerInstance,
		},
	})

	err := facades.Event().Job(eventInstance, []event.Arg{
		{Type: "string", Value: "goravel"},
	}).Dispatch()

	s.NoError(err)
	s.Equal([][]any{
		{"goravel_transformed", 2},
	}, capture.Handled())
	s.Equal(1, capture.QueueCallCount())
}

func (s *EventTestSuite) TestDispatchStopsAfterListenerError() {
	expectedErr := errors.New("listener handle error")
	eventInstance := &stopOnErrorEvent{}
	failedCapture := &listenerCapture{}
	skippedCapture := &listenerCapture{}
	failedListener := &integrationListener{
		signature:   s.uniqueName("failed_listener"),
		queueConfig: event.Queue{Enable: false},
		handleErr:   expectedErr,
		capture:     failedCapture,
	}
	skippedListener := &integrationListener{
		signature:   s.uniqueName("skipped_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     skippedCapture,
	}
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			failedListener,
			skippedListener,
		},
	})

	err := facades.Event().Job(eventInstance, []event.Arg{
		{Type: "string", Value: "should stop"},
	}).Dispatch()

	s.Equal(expectedErr, err)
	s.Len(failedCapture.Handled(), 1)
	s.Empty(skippedCapture.Handled())
}

func (s *EventTestSuite) TestDispatchQueuedListenerEventually() {
	eventInstance := &queuedEvent{}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature: s.uniqueName("queued_listener"),
		queueConfig: event.Queue{
			Enable: true,
		},
		capture: capture,
	}
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			listenerInstance,
		},
	})

	err := facades.Event().Job(eventInstance, []event.Arg{
		{Type: "string", Value: "queued"},
	}).Dispatch()

	s.NoError(err)
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
	s.True(file.Contains(listenerPath, "func (receiver *"+listenerName+") Handle(eventName string, args ...any) error {"))
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

func (s *EventTestSuite) TestListenStringEventAndDispatch() {
	eventName := s.uniqueName("user.created.")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(eventName, &integrationListener{
		signature: s.uniqueName("listen_string_listener"),
		capture:   capture,
	}))

	result := facades.Event().Dispatch(eventName, []event.Arg{
		{Type: "string", Value: "goravel"},
	})

	s.False(result.Failed())
	s.NoError(result.Error())
	s.Equal([]string{eventName}, capture.Names())
	s.Equal([][]any{{"goravel"}}, capture.Handled())
}

func (s *EventTestSuite) TestListenMultipleEventsAndListeners() {
	placed := s.uniqueName("order.placed.")
	paid := s.uniqueName("order.paid.")
	first := &listenerCapture{}
	second := &listenerCapture{}
	s.NoError(facades.Event().Listen([]string{placed, paid},
		&integrationListener{signature: s.uniqueName("listen_multi_first"), capture: first},
		&integrationListener{signature: s.uniqueName("listen_multi_second"), capture: second},
	))

	s.False(facades.Event().Dispatch(placed).Failed())
	s.False(facades.Event().Dispatch(paid).Failed())

	s.Equal([]string{placed, paid}, first.Names())
	s.Equal([]string{placed, paid}, second.Names())
}

func (s *EventTestSuite) TestListenEventValuesAndListener() {
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen([]event.Event{&listenEvent{}, &otherListenEvent{}},
		&integrationListener{signature: s.uniqueName("listen_event_values"), capture: capture},
	))

	s.False(facades.Event().Dispatch(&listenEvent{}).Failed())
	s.False(facades.Event().Dispatch(&otherListenEvent{}).Failed())

	// An object event is named by its import path and type, which is also what
	// travels as the first argument of a queued job.
	s.Equal([]string{
		"goravel/tests/feature.listenEvent",
		"goravel/tests/feature.otherListenEvent",
	}, capture.Names())
}

func (s *EventTestSuite) TestListenWildcardReceivesTheMatchedName() {
	prefix := s.uniqueName("invoice")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(prefix+".*", &integrationListener{
		signature: s.uniqueName("listen_wildcard_listener"),
		capture:   capture,
	}))

	s.False(facades.Event().Dispatch(prefix + ".issued").Failed())
	s.False(facades.Event().Dispatch(prefix + ".paid").Failed())
	s.False(facades.Event().Dispatch("shipment.sent").Failed())

	// The pattern is what it registered on, the matched name is what it receives.
	s.Equal([]string{prefix + ".issued", prefix + ".paid"}, capture.Names())
}

func (s *EventTestSuite) TestListenClosures() {
	var (
		mu       sync.Mutex
		untyped  []any
		typedFor *listenEvent
	)
	eventName := s.uniqueName("report.generated.")

	s.NoError(facades.Event().Listen(eventName, func(evt any, args ...any) error {
		mu.Lock()
		defer mu.Unlock()
		untyped = append(untyped, evt)

		return nil
	}))

	// No listener argument: the event comes from the closure's parameter type.
	s.NoError(facades.Event().Listen(func(evt *listenEvent) error {
		mu.Lock()
		defer mu.Unlock()
		typedFor = evt

		return nil
	}))

	s.False(facades.Event().Dispatch(eventName).Failed())
	dispatched := &listenEvent{Name: "typed"}
	s.False(facades.Event().Dispatch(dispatched).Failed())

	mu.Lock()
	defer mu.Unlock()
	s.Equal([]any{eventName}, untyped)
	s.Same(dispatched, typedFor)
}

func (s *EventTestSuite) TestDispatchCollectsEveryListenerError() {
	eventName := s.uniqueName("payment.failed.")
	first := errors.New("first listener failed")
	second := errors.New("second listener failed")
	survived := &listenerCapture{}

	s.NoError(facades.Event().Listen(eventName,
		&integrationListener{signature: s.uniqueName("collect_first"), handleErr: first},
		&integrationListener{signature: s.uniqueName("collect_survivor"), capture: survived},
		&integrationListener{signature: s.uniqueName("collect_second"), handleErr: second},
	))

	result := facades.Event().Dispatch(eventName)

	// Unlike the deprecated Job flow, Dispatch runs every listener.
	s.True(result.Failed())
	s.Len(result.Errors(), 2)
	s.ErrorContains(result.Error(), first.Error())
	s.ErrorContains(result.Error(), second.Error())
	s.Len(survived.Handled(), 1)
}

func (s *EventTestSuite) TestDispatchRecoversFromAPanickingListener() {
	eventName := s.uniqueName("panic.raised.")
	survived := &listenerCapture{}

	s.NoError(facades.Event().Listen(eventName,
		&integrationListener{signature: s.uniqueName("panicking_listener"), panics: true},
		&integrationListener{signature: s.uniqueName("panic_survivor"), capture: survived},
	))

	result := facades.Event().Dispatch(eventName)

	// The panic becomes an error on the result, the listener behind it still runs.
	s.True(result.Failed())
	s.Len(result.Errors(), 1)
	s.ErrorContains(result.Error(), "panicked")
	s.Len(survived.Handled(), 1)
}

func (s *EventTestSuite) TestDispatchQueuedListenerRegisteredThroughListen() {
	eventName := s.uniqueName("newsletter.sent.")
	signature := s.uniqueName("listen_queued_listener")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(eventName, &integrationListener{
		signature:   signature,
		queueConfig: event.Queue{Enable: true},
		capture:     capture,
	}))

	// Listen registers the job with the queue, so a worker can resolve it from
	// the signature alone.
	job, err := facades.Queue().GetJob(signature)
	s.NoError(err)
	s.Equal(signature, job.Signature())

	result := facades.Event().Dispatch(eventName, []event.Arg{
		{Type: "string", Value: "queued through listen"},
	})

	s.False(result.Failed())
	s.True(waitUntil(5*time.Second, 20*time.Millisecond, func() bool {
		return len(capture.Handled()) == 1
	}))
	// The event name leads the queued payload and is handed back to the listener
	// as its first parameter, not as part of the payload.
	s.Equal([]string{eventName}, capture.Names())
	s.Equal([][]any{{"queued through listen"}}, capture.Handled())
}

func (s *EventTestSuite) TestDispatchQueuedWildcardListenerCarriesTheMatchedName() {
	prefix := s.uniqueName("audit")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(prefix+".*", &integrationListener{
		signature:   s.uniqueName("listen_queued_wildcard"),
		queueConfig: event.Queue{Enable: true},
		capture:     capture,
	}))

	s.False(facades.Event().Dispatch(prefix + ".login").Failed())

	s.True(waitUntil(5*time.Second, 20*time.Millisecond, func() bool {
		return len(capture.Handled()) == 1
	}))
	s.Equal([]string{prefix + ".login"}, capture.Names())
}

func (s *EventTestSuite) TestDispatchBootstrappedEventThroughDispatch() {
	// The path a real application takes: an event registered at boot, fired with
	// the new Dispatch instead of the deprecated Job.
	result := facades.Event().Dispatch(&events.OrderShipped{}, []event.Arg{
		{Type: "string", Value: "dispatched not jobbed"},
	})

	s.False(result.Failed())
	s.True(waitUntil(3*time.Second, 20*time.Millisecond, func() bool {
		return len(listeners.TestResultOfSendShipmentNotification) == 1
	}))
	s.Equal([]string{"dispatched not jobbed"}, listeners.TestResultOfSendShipmentNotification)
}

func (s *EventTestSuite) TestJobReachesListenersRegisteredThroughListen() {
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(&jobSeesListenEvent{}, &integrationListener{
		signature: s.uniqueName("job_sees_listen"),
		capture:   capture,
	}))

	// The deprecated Job now resolves listeners by event name, so it reaches the
	// ones registered through Listen too.
	s.NoError(facades.Event().Job(&jobSeesListenEvent{}, []event.Arg{
		{Type: "string", Value: "seen by job"},
	}).Dispatch())

	s.Equal([]string{"goravel/tests/feature.jobSeesListenEvent"}, capture.Names())
	s.Equal([][]any{{"seen by job"}}, capture.Handled())
}

func (s *EventTestSuite) TestListenRejectsInvalidRegistrations() {
	eventName := s.uniqueName("rejected.")

	tests := []struct {
		name        string
		listen      func() error
		expectedErr error
	}{
		{
			name:        "NotAListener",
			listen:      func() error { return facades.Event().Listen(eventName, "not a listener") },
			expectedErr: frameworkerrors.EventInvalidListener.Args("string"),
		},
		{
			name:        "NilPointerListener",
			listen:      func() error { return facades.Event().Listen(eventName, (*integrationListener)(nil)) },
			expectedErr: frameworkerrors.EventListenerNotPointer.Args("goravel/tests/feature.integrationListener"),
		},
		{
			name:        "EmptySignature",
			listen:      func() error { return facades.Event().Listen(eventName, &integrationListener{}) },
			expectedErr: frameworkerrors.EventListenerEmptySignature.Args("goravel/tests/feature.integrationListener"),
		},
		{
			name:        "TypedClosureOnAnotherEvent",
			listen:      func() error { return facades.Event().Listen(eventName, func(evt *listenEvent) error { return nil }) },
			expectedErr: frameworkerrors.EventListenerEventMismatch.Args("goravel/tests/feature.listenEvent", eventName),
		},
		{
			name:        "EmptyEventName",
			listen:      func() error { return facades.Event().Listen("", &integrationListener{signature: "unused"}) },
			expectedErr: frameworkerrors.EventInvalidEvent.Args(""),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.Equal(test.expectedErr, test.listen())
		})
	}
}

func (s *EventTestSuite) TestDispatchRejectsMoreThanOnePayload() {
	eventName := s.uniqueName("too.many.payloads.")

	result := facades.Event().Dispatch(eventName,
		[]event.Arg{{Type: "string", Value: "first"}},
		[]event.Arg{{Type: "string", Value: "second"}},
	)

	s.True(result.Failed())
	s.Equal(frameworkerrors.EventTooManyPayloads.Args(eventName, 2), result.Error())
}

// uniqueName is process wide on purpose. Nothing ever unregisters an event or a
// queue signature, so a counter that restarted with the suite would collide with
// its own previous run under -count=2.
func (s *EventTestSuite) uniqueName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, atomic.AddUint64(&eventTestCounter, 1))
}

var eventTestCounter uint64

// listenEvent, otherListenEvent and jobSeesListenEvent belong to the Listen
// based tests. jobSeesListenEvent exists so the Job test does not also fire the
// listener bootstrapped on events.OrderShipped.
type listenEvent struct {
	Name string
}

func (receiver *listenEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type otherListenEvent struct{}

func (receiver *otherListenEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type jobSeesListenEvent struct{}

func (receiver *jobSeesListenEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

// Each scenario uses its own event type: listeners are resolved by event name,
// which is derived from the type, so scenarios sharing a type would share
// listeners.
type handleErrorEvent struct {
	err error
}

func (receiver *handleErrorEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return nil, receiver.err
}

type transformingEvent struct{}

func (receiver *transformingEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return []event.Arg{
		{Type: "string", Value: castString(args[0].Value) + "_transformed"},
		{Type: "int", Value: 2},
	}, nil
}

type stopOnErrorEvent struct{}

func (receiver *stopOnErrorEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

type queuedEvent struct{}

func (receiver *queuedEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
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

type integrationListener struct {
	signature   string
	queueConfig event.Queue
	handleErr   error
	panics      bool
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
	if receiver.panics {
		panic("listener panicked")
	}

	if receiver.capture != nil {
		receiver.capture.AddHandled(eventName, args)
	}

	return receiver.handleErr
}

type listenerCapture struct {
	mu        sync.Mutex
	names     []string
	handled   [][]any
	queueArgs [][]any
}

func (receiver *listenerCapture) AddHandled(eventName string, args []any) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.names = append(receiver.names, eventName)
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

// Names returns the event name each Handle call received.
func (receiver *listenerCapture) Names() []string {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	return append([]string(nil), receiver.names...)
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
