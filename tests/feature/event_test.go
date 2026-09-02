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
	counter uint64
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

// listenEvent is only ever used by the Listen based tests below.
type listenEvent struct {
	Name string
}

func (receiver *listenEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

func (s *EventTestSuite) TestListenStringEventAndDispatch() {
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen("user.created", &integrationListener{
		signature: s.uniqueName("listen_string_listener"),
		capture:   capture,
	}))

	result := facades.Event().Dispatch("user.created", []event.Arg{
		{Type: "string", Value: "goravel"},
	})

	s.False(result.Failed())
	s.NoError(result.Error())
	s.Equal([]string{"user.created"}, capture.Names())
	s.Equal([][]any{{"goravel"}}, capture.Handled())
}

func (s *EventTestSuite) TestListenMultipleEventsAndListeners() {
	first := &listenerCapture{}
	second := &listenerCapture{}
	s.NoError(facades.Event().Listen([]string{"order.placed", "order.paid"},
		&integrationListener{signature: s.uniqueName("listen_multi_first"), capture: first},
		&integrationListener{signature: s.uniqueName("listen_multi_second"), capture: second},
	))

	s.False(facades.Event().Dispatch("order.placed").Failed())
	s.False(facades.Event().Dispatch("order.paid").Failed())

	s.Equal([]string{"order.placed", "order.paid"}, first.Names())
	s.Equal([]string{"order.placed", "order.paid"}, second.Names())
}

func (s *EventTestSuite) TestListenWildcardReceivesTheMatchedName() {
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen("invoice.*", &integrationListener{
		signature: s.uniqueName("listen_wildcard_listener"),
		capture:   capture,
	}))

	s.False(facades.Event().Dispatch("invoice.issued").Failed())
	s.False(facades.Event().Dispatch("invoice.paid").Failed())
	s.False(facades.Event().Dispatch("shipment.sent").Failed())

	// The pattern is what it registered on, the matched name is what it receives.
	s.Equal([]string{"invoice.issued", "invoice.paid"}, capture.Names())
}

func (s *EventTestSuite) TestListenClosures() {
	var (
		mu       sync.Mutex
		untyped  []any
		typedFor *listenEvent
	)

	s.NoError(facades.Event().Listen("report.generated", func(evt any, args ...any) error {
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

	s.False(facades.Event().Dispatch("report.generated").Failed())
	dispatched := &listenEvent{Name: "typed"}
	s.False(facades.Event().Dispatch(dispatched).Failed())

	mu.Lock()
	defer mu.Unlock()
	s.Equal([]any{"report.generated"}, untyped)
	s.Same(dispatched, typedFor)
}

func (s *EventTestSuite) TestDispatchCollectsEveryListenerError() {
	first := errors.New("first listener failed")
	second := errors.New("second listener failed")
	survived := &listenerCapture{}

	s.NoError(facades.Event().Listen("payment.failed",
		&integrationListener{signature: s.uniqueName("collect_first"), handleErr: first},
		&integrationListener{signature: s.uniqueName("collect_survivor"), capture: survived},
		&integrationListener{signature: s.uniqueName("collect_second"), handleErr: second},
	))

	result := facades.Event().Dispatch("payment.failed")

	// Unlike the deprecated Job flow, Dispatch runs every listener.
	s.True(result.Failed())
	s.Len(result.Errors(), 2)
	s.ErrorContains(result.Error(), first.Error())
	s.ErrorContains(result.Error(), second.Error())
	s.Len(survived.Handled(), 1)
}

func (s *EventTestSuite) TestDispatchQueuedListenerRegisteredThroughListen() {
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen("newsletter.sent", &integrationListener{
		signature:   s.uniqueName("listen_queued_listener"),
		queueConfig: event.Queue{Enable: true},
		capture:     capture,
	}))

	result := facades.Event().Dispatch("newsletter.sent", []event.Arg{
		{Type: "string", Value: "queued through listen"},
	})

	s.False(result.Failed())
	s.True(waitUntil(5*time.Second, 20*time.Millisecond, func() bool {
		return len(capture.Handled()) == 1
	}))
	// The event name leads the queued payload, so the worker can pass it on.
	s.Equal([]string{"newsletter.sent"}, capture.Names())
	s.Equal([][]any{{"queued through listen"}}, capture.Handled())
}

func (s *EventTestSuite) TestJobReachesListenersRegisteredThroughListen() {
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(&events.OrderShipped{}, &integrationListener{
		signature: s.uniqueName("job_sees_listen"),
		capture:   capture,
	}))

	// The deprecated Job now resolves listeners by event name, so it reaches the
	// ones registered through Listen too.
	s.NoError(facades.Event().Job(&events.OrderShipped{}, []event.Arg{
		{Type: "string", Value: "seen by job"},
	}).Dispatch())

	s.Equal([][]any{{"seen by job"}}, capture.Handled())
}

func (s *EventTestSuite) TestListenRejectsInvalidRegistrations() {
	// Not a listener and not a closure.
	s.Error(facades.Event().Listen("user.created", "not a listener"))
	// A nil pointer cannot be told apart from another of its type.
	s.Error(facades.Event().Listen("user.created", (*integrationListener)(nil)))
	// An empty signature would claim the empty key in the queue registry.
	s.Error(facades.Event().Listen("user.created", &integrationListener{}))
	// A typed closure cannot be registered on an event it does not name.
	s.Error(facades.Event().Listen("user.created", func(evt *listenEvent) error { return nil }))
	// An event that is neither a non-empty string nor a named struct.
	s.Error(facades.Event().Listen("", &integrationListener{signature: s.uniqueName("unused")}))
}

func (s *EventTestSuite) TestDispatchRejectsMoreThanOnePayload() {
	result := facades.Event().Dispatch("user.created",
		[]event.Arg{{Type: "string", Value: "first"}},
		[]event.Arg{{Type: "string", Value: "second"}},
	)

	s.True(result.Failed())
}
