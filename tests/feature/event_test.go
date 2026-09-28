package feature

import (
	"errors"
	"fmt"
	"os"
	"reflect"
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

// eventNameCounter backs uniqueName at package scope: the event registry keeps
// registrations for the whole process, so names must stay unique across suite
// re-instantiations (`go test -count=2`), not just within one.
var eventNameCounter atomic.Uint64

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

func (s *EventTestSuite) TestGetEventsReturnsACopy() {
	eventInstance := &getEventsEvent{}
	listener := &integrationListener{
		signature:   s.uniqueName("get_events_listener"),
		queueConfig: event.Queue{},
		capture:     nil,
	}

	// Register (deprecated) is the only entry that records the event under its
	// value, which is what GetEvents returns a defensive copy of.
	//nolint:staticcheck
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			listener,
		},
	})

	// GetEvents returns a copy, so lookups are keyed on the event value.
	//nolint:staticcheck
	events := facades.Event().GetEvents()

	s.Equal([]event.Listener{listener}, events[eventInstance])

	events[eventInstance] = nil

	//nolint:staticcheck
	s.Equal([]event.Listener{listener}, facades.Event().GetEvents()[eventInstance])
}

func (s *EventTestSuite) TestJobDispatchUnregisteredEvent() {
	eventInstance := &unregisteredIntegrationEvent{}

	// The deprecated Task requires a bound listener; unlike Dispatch, an
	// unbound event fails with EventListenerNotBind.
	//nolint:staticcheck
	err := facades.Event().Job(eventInstance, nil).Dispatch()

	s.Equal(frameworkerrors.EventListenerNotBind.Args(eventInstance), err)
}

func (s *EventTestSuite) TestJobDispatchReturnsEventHandleError() {
	expectedErr := errors.New("event handle error")
	eventInstance := &integrationEvent{
		handle: func(args []event.Arg) ([]event.Arg, error) {
			return nil, expectedErr
		},
	}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature:   s.uniqueName("event_handle_error_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     capture,
	}

	// Register (deprecated) replaces the legacy listeners it bound to the shared
	// integrationEvent name, keeping the Job tests isolated from each other.
	//nolint:staticcheck
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			listenerInstance,
		},
	})

	// The event's own Handle error short-circuits the Task and its error is
	// returned directly, so the registered listener must not run.
	//nolint:staticcheck
	err := facades.Event().Job(eventInstance, []event.Arg{
		{Type: "string", Value: "test"},
	}).Dispatch()

	s.Equal(expectedErr, err)
	s.Empty(capture.Handled())
}

func (s *EventTestSuite) TestJobDispatchSyncListenerWithTransformedArgs() {
	eventInstance := &integrationEvent{
		handle: func(args []event.Arg) ([]event.Arg, error) {
			return []event.Arg{
				{Type: "string", Value: castString(args[0].Value) + "_transformed"},
				{Type: "int", Value: 2},
			}, nil
		},
	}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature:   s.uniqueName("sync_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     capture,
	}

	//nolint:staticcheck
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			listenerInstance,
		},
	})

	// The Task delivers the Handle-transformed payload to the sync listener and
	// asks for its queue options once.
	//nolint:staticcheck
	err := facades.Event().Job(eventInstance, []event.Arg{
		{Type: "string", Value: "goravel"},
	}).Dispatch()

	s.NoError(err)
	s.Equal([][]any{
		{"goravel_transformed", 2},
	}, capture.Handled())
	s.Equal(1, capture.QueueCallCount())
}

func (s *EventTestSuite) TestJobDispatchStopsAtFirstListenerError() {
	expectedErr := errors.New("listener handle error")
	eventInstance := &integrationEvent{
		handle: func(args []event.Arg) ([]event.Arg, error) {
			return args, nil
		},
	}
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

	//nolint:staticcheck
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			failedListener,
			skippedListener,
		},
	})

	// The Task is fail-fast: the first failing listener returns its error
	// directly and the listener behind it is never invoked.
	//nolint:staticcheck
	err := facades.Event().Job(eventInstance, []event.Arg{
		{Type: "string", Value: "should stop"},
	}).Dispatch()

	s.Equal(expectedErr, err)
	s.Len(failedCapture.Handled(), 1)
	s.Empty(skippedCapture.Handled())
}

func (s *EventTestSuite) TestJobDispatchQueuedListenerEventually() {
	eventInstance := &integrationEvent{
		handle: func(args []event.Arg) ([]event.Arg, error) {
			return args, nil
		},
	}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature: s.uniqueName("queued_listener"),
		queueConfig: event.Queue{
			Enable: true,
		},
		capture: capture,
	}

	//nolint:staticcheck
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {
			listenerInstance,
		},
	})

	// A queueable listener is pushed onto the queue, the Task reports no error
	// and the listener handles the payload once through the worker.
	//nolint:staticcheck
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
	s.True(file.Contains(listenerPath, `return "`+str.Of(listenerName).Snake().String()+`"`))
	s.True(file.Contains(listenerPath, "func (receiver *"+listenerName+") Queue(args ...any) event.Queue {"))
	s.True(file.Contains(listenerPath, "func (receiver *"+listenerName+") Handle(eventName string, args ...any) error {"))

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

// TestListenStringEventDeliversNameAndArgs proves a string event reaches a
// Listener with the canonical event name and the payload values.
func (s *EventTestSuite) TestListenStringEventDeliversNameAndArgs() {
	eventName := s.uniqueName("order.confirmed")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(eventName, &integrationListener{
		signature:   s.uniqueName("order_confirmed_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     capture,
	}))

	result := facades.Event().Dispatch(eventName, []event.Arg{
		{Type: "string", Value: "A-1"},
	})

	s.False(result.Failed())
	s.NoError(result.Error())
	s.Equal([]string{eventName}, capture.EventNames())
	s.Equal([][]any{{"A-1"}}, capture.Handled())
}

// TestListenClosureOnStringEvent covers the func(event any, args ...any) error
// form, whose first argument is the dispatched value (the name, for a string).
func (s *EventTestSuite) TestListenClosureOnStringEvent() {
	eventName := s.uniqueName("user.created")
	var received []any
	s.NoError(facades.Event().Listen(eventName, func(evt any, args ...any) error {
		received = append(received, evt, args)

		return nil
	}))

	s.False(facades.Event().Dispatch(eventName, []event.Arg{
		{Type: "int", Value: 7},
	}).Failed())

	s.Equal([]any{eventName, []any{7}}, received)
}

// TestListenWildcardMatchesPrefix proves a pattern matches every event sharing
// its prefix, delivers the matched name, and ignores unrelated names.
func (s *EventTestSuite) TestListenWildcardMatchesPrefix() {
	prefix := s.uniqueName("shipment")
	var matched []string
	s.NoError(facades.Event().Listen(prefix+".*", func(evt any, args ...any) error {
		matched = append(matched, castString(evt))

		return nil
	}))

	s.False(facades.Event().Dispatch(prefix + ".created").Failed())
	s.False(facades.Event().Dispatch(prefix + ".updated").Failed())
	s.False(facades.Event().Dispatch(s.uniqueName("invoice")).Failed())

	s.ElementsMatch([]string{prefix + ".created", prefix + ".updated"}, matched)
}

// TestListenWildcardOrderIsRegistrationOrder pins the deterministic invocation
// order of overlapping patterns.
func (s *EventTestSuite) TestListenWildcardOrderIsRegistrationOrder() {
	prefix := s.uniqueName("audit")
	var order []string
	register := func(pattern, label string) {
		s.NoError(facades.Event().Listen(pattern, func(evt any, args ...any) error {
			order = append(order, label)

			return nil
		}))
	}

	suffix := s.uniqueName("created")

	register(prefix+".*", "A")
	register("*."+suffix, "B")
	register(prefix+".*", "C")

	s.False(facades.Event().Dispatch(prefix + "." + suffix).Failed())

	s.Equal([]string{"A", "B", "C"}, order)
}

// TestListenTypedClosureResolvesTheEventFromItsParameter covers the implicit
// closure-only form.
func (s *EventTestSuite) TestListenTypedClosureResolvesTheEventFromItsParameter() {
	var received []*listenClosureEvent
	s.NoError(facades.Event().Listen(func(evt *listenClosureEvent) error {
		received = append(received, evt)

		return nil
	}))

	evt := &listenClosureEvent{}
	s.False(facades.Event().Dispatch(evt).Failed())

	s.Equal([]*listenClosureEvent{evt}, received)
}

// TestListenTypedClosureRejectsAMismatchedEvent verifies Listen reports the
// mismatch instead of registering something that could never fire.
func (s *EventTestSuite) TestListenTypedClosureRejectsAMismatchedEvent() {
	err := facades.Event().Listen(s.uniqueName("audit"), func(evt *listenClosureEvent) error {
		return nil
	})

	s.Error(err)
	s.ErrorIs(err, frameworkerrors.EventListenerEventMismatch)
}

// TestListenSliceOfStringEvents registers one listener for many string events.
func (s *EventTestSuite) TestListenSliceOfStringEvents() {
	created := s.uniqueName("account.created")
	updated := s.uniqueName("account.updated")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen([]string{created, updated}, &integrationListener{
		signature:   s.uniqueName("account_audit_listener"),
		queueConfig: event.Queue{Enable: false},
		capture:     capture,
	}))

	s.False(facades.Event().Dispatch(created).Failed())
	s.False(facades.Event().Dispatch(updated).Failed())

	s.ElementsMatch([]string{created, updated}, capture.EventNames())
}

// TestListenSliceOfEventValues covers []event.Event registrations.
func (s *EventTestSuite) TestListenSliceOfEventValues() {
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(
		[]event.Event{&listenSliceEventA{}, &listenSliceEventB{}},
		&integrationListener{
			signature:   s.uniqueName("slice_event_listener"),
			queueConfig: event.Queue{Enable: false},
			capture:     capture,
		},
	))

	s.False(facades.Event().Dispatch(&listenSliceEventA{}).Failed())
	s.False(facades.Event().Dispatch(&listenSliceEventB{}).Failed())

	nameOf := func(v any) string {
		t := reflect.TypeOf(v)

		return t.PkgPath() + "." + t.Name()
	}
	s.ElementsMatch([]string{nameOf(listenSliceEventA{}), nameOf(listenSliceEventB{})}, capture.EventNames())
}

// TestListenRejectsInvalidRegistrations checks every rejected form returns its
// declared error through the facade.
func (s *EventTestSuite) TestListenRejectsInvalidRegistrations() {
	tests := []struct {
		name   string
		events any
		listen any
		expect error
	}{
		{
			name:   "nil listener",
			events: s.uniqueName("nil_listener"),
			listen: nil,
			expect: frameworkerrors.EventInvalidListener,
		},
		{
			name:   "value listener",
			events: s.uniqueName("value_listener"),
			listen: valueListener{},
			expect: frameworkerrors.EventListenerNotPointer,
		},
		{
			name:   "empty signature",
			events: s.uniqueName("empty_signature"),
			listen: &integrationListener{},
			expect: frameworkerrors.EventListenerEmptySignature,
		},
		{
			name:   "invalid event",
			events: 123,
			listen: func(evt any, args ...any) error { return nil },
			expect: frameworkerrors.EventInvalidEvent,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.ErrorIs(facades.Event().Listen(tt.events, tt.listen), tt.expect)
		})
	}
}

// TestDispatchCollectsEveryListenerError covers Result.Failed/Errors/Error.
func (s *EventTestSuite) TestDispatchCollectsEveryListenerError() {
	eventName := s.uniqueName("payment")
	errA := errors.New("payment a failed")
	errB := errors.New("payment b failed")
	s.NoError(facades.Event().Listen(eventName,
		&integrationListener{signature: s.uniqueName("payment_a"), handleErr: errA},
		&integrationListener{signature: s.uniqueName("payment_b"), handleErr: errB},
	))

	result := facades.Event().Dispatch(eventName)

	s.True(result.Failed())
	s.ErrorIs(result.Error(), errA)
	s.ErrorIs(result.Error(), errB)
	s.Len(result.Errors(), 2)
}

// TestDispatchRejectsMoreThanOnePayload covers the payload-validation path.
func (s *EventTestSuite) TestDispatchRejectsMoreThanOnePayload() {
	eventName := s.uniqueName("invoice")
	payload := []event.Arg{{Type: "string", Value: "x"}}

	result := facades.Event().Dispatch(eventName, payload, payload)

	s.Require().True(result.Failed())
	// Compare the formatted errors: Result.Error() joins the collected errors
	// and Goravel's errorString.Is matches on the format text alone, so ErrorIs
	// would not pin the event name or the payload count.
	s.Equal(
		frameworkerrors.EventTooManyPayloads.Args(eventName, 2).Error(),
		result.Error().Error(),
	)
}

// TestDispatchWithoutListenersSkipsEventHandle proves an unbound event does not
// run its own Handle, unlike the deprecated Task which errors.
func (s *EventTestSuite) TestDispatchWithoutListenersSkipsEventHandle() {
	evt := &skippedHandleEvent{integrationEvent: integrationEvent{
		handle: func(args []event.Arg) ([]event.Arg, error) {
			panic("Handle must not run without listeners")
		},
	}}

	result := facades.Event().Dispatch(evt)

	s.False(result.Failed())
	s.Empty(result.Errors())
	s.NoError(result.Error())
}

// TestDispatchRecoversFromAPanickingListener proves one panic fails itself
// while the listeners behind it still run.
func (s *EventTestSuite) TestDispatchRecoversFromAPanickingListener() {
	eventName := s.uniqueName("panic.listener")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(eventName,
		&panickingListener{signature: s.uniqueName("panicking_listener")},
		&integrationListener{signature: s.uniqueName("surviving_listener"), capture: capture},
	))

	result := facades.Event().Dispatch(eventName)

	s.True(result.Failed())
	s.ErrorIs(result.Error(), frameworkerrors.EventListenerPanic)
	s.Len(capture.Handled(), 1)
}

// TestDispatchRecoversFromAPanickingEvent proves the event's own Handle panic
// is contained and short-circuits the listeners.
func (s *EventTestSuite) TestDispatchRecoversFromAPanickingEvent() {
	capture := &listenerCapture{}
	evt := &panickingEvent{integrationEvent: integrationEvent{
		handle: func(args []event.Arg) ([]event.Arg, error) { panic("event boom") },
	}}
	s.NoError(facades.Event().Listen(evt, &integrationListener{
		signature: s.uniqueName("panicking_event_listener"),
		capture:   capture,
	}))

	result := facades.Event().Dispatch(evt)

	s.True(result.Failed())
	s.ErrorIs(result.Error(), frameworkerrors.EventHandlePanic)
	s.Empty(capture.Handled())
}

// TestDispatchQueuedListenerReceivesEventNameFirst routes the listener through
// the database queue connection, where the app's app:queue:database boot worker
// processes it: the event name must lead the payload so the worker can resolve
// it.
func (s *EventTestSuite) TestDispatchQueuedListenerReceivesEventNameFirst() {
	eventName := s.uniqueName("queued.event")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(eventName, &integrationListener{
		signature:   s.uniqueName("queued_listener"),
		queueConfig: event.Queue{Enable: true, Connection: "database"},
		capture:     capture,
	}))

	result := facades.Event().Dispatch(eventName, []event.Arg{
		{Type: "string", Value: "payload"},
	})

	s.False(result.Failed())
	s.NoError(result.Error())
	s.True(waitUntil(5*time.Second, 20*time.Millisecond, func() bool {
		return len(capture.Handled()) == 1
	}))
	s.Equal([]string{eventName}, capture.EventNames())
	s.Equal([][]any{{"payload"}}, capture.Handled())
}

// TestConcurrentDispatch runs many dispatches at once to exercise the facade's
// locking under the race detector.
func (s *EventTestSuite) TestConcurrentDispatch() {
	eventName := s.uniqueName("concurrent")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(eventName, &integrationListener{
		signature: s.uniqueName("concurrent_listener"),
		capture:   capture,
	}))

	const dispatches = 20
	eventInstance := facades.Event()
	var wg sync.WaitGroup
	for i := 0; i < dispatches; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			eventInstance.Dispatch(eventName)
		}()
	}
	wg.Wait()

	s.Len(capture.EventNames(), dispatches)
}

// TestDispatchReachesListenersRegisteredThroughRegister keeps the deprecated
// flow working: a Register listener is reachable by Dispatch.
func (s *EventTestSuite) TestDispatchReachesListenersRegisteredThroughRegister() {
	eventInstance := &registerReachedEvent{}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature: s.uniqueName("register_reached_listener"),
		capture:   capture,
	}

	//nolint:staticcheck
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {listenerInstance},
	})

	result := facades.Event().Dispatch(eventInstance, []event.Arg{
		{Type: "string", Value: "goravel"},
	})

	s.False(result.Failed())
	s.Equal([][]any{{"goravel"}}, capture.Handled())
}

// TestListenAndRegisterCoexistOnTheSameEvent proves Register overwrites only
// its own listeners and never the ones added through Listen.
func (s *EventTestSuite) TestListenAndRegisterCoexistOnTheSameEvent() {
	eventInstance := &coalesceEvent{}
	listenCapture := &listenerCapture{}
	registerCapture := &listenerCapture{}

	s.NoError(facades.Event().Listen(eventInstance, &integrationListener{
		signature: s.uniqueName("listen_listener"),
		capture:   listenCapture,
	}))

	legacy := &integrationListener{
		signature: s.uniqueName("register_listener"),
		capture:   registerCapture,
	}

	// Register is called twice to mirror the framework guarantee that repeated
	// Register calls only drop the legacy listeners.
	//nolint:staticcheck
	for i := 0; i < 2; i++ {
		facades.Event().Register(map[event.Event][]event.Listener{
			eventInstance: {legacy},
		})
	}

	result := facades.Event().Dispatch(eventInstance)

	s.False(result.Failed())
	s.Len(listenCapture.Handled(), 1)
	s.Len(registerCapture.Handled(), 1)
}

// TestJobReachesListenersRegisteredThroughListen covers the deprecated Task
// finding listeners added by the new Listen API.
func (s *EventTestSuite) TestJobReachesListenersRegisteredThroughListen() {
	eventInstance := &jobListenReachedEvent{}
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(eventInstance, &integrationListener{
		signature: s.uniqueName("job_listen_listener"),
		capture:   capture,
	}))

	//nolint:staticcheck
	s.NoError(facades.Event().Job(eventInstance, []event.Arg{
		{Type: "string", Value: "goravel"},
	}).Dispatch())

	s.Equal([][]any{{"goravel"}}, capture.Handled())
}

// TestJobReachesWildcardListeners covers the Task path also matching wildcards.
// The leading * absorbs the package path, leaving the type name as a suffix,
// so the test stays independent of the package import path.
func (s *EventTestSuite) TestJobReachesWildcardListeners() {
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen("*jobWildcardEvent", &integrationListener{
		signature: s.uniqueName("job_wildcard_listener"),
		capture:   capture,
	}))

	//nolint:staticcheck
	s.NoError(facades.Event().Job(&jobWildcardEvent{}, nil).Dispatch())

	s.Len(capture.Handled(), 1)
}

// TestJobReachesAFreshInstanceOfARegisteredEvent covers the identity change:
// Job resolves listeners by event name, not by the value it is called with.
func (s *EventTestSuite) TestJobReachesAFreshInstanceOfARegisteredEvent() {
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature: s.uniqueName("fresh_instance_listener"),
		capture:   capture,
	}

	//nolint:staticcheck
	facades.Event().Register(map[event.Event][]event.Listener{
		&jobIdentifiedEvent{id: 1}: {listenerInstance},
	})

	//nolint:staticcheck
	s.NoError(facades.Event().Job(&jobIdentifiedEvent{id: 2}, nil).Dispatch())

	s.Len(capture.Handled(), 1)
}

// TestDispatchQueuesAListenerRegisteredThroughRegister proves a queued legacy
// listener really runs through the database queue connection, processed by the
// app's app:queue:database boot worker, with the event name leading.
func (s *EventTestSuite) TestDispatchQueuesAListenerRegisteredThroughRegister() {
	eventInstance := &queuedRegisterEvent{}
	capture := &listenerCapture{}
	listenerInstance := &integrationListener{
		signature:   s.uniqueName("queued_register_listener"),
		queueConfig: event.Queue{Enable: true, Connection: "database"},
		capture:     capture,
	}

	//nolint:staticcheck
	facades.Event().Register(map[event.Event][]event.Listener{
		eventInstance: {listenerInstance},
	})

	result := facades.Event().Dispatch(eventInstance, []event.Arg{
		{Type: "string", Value: "goravel"},
	})

	s.False(result.Failed())
	s.True(waitUntil(5*time.Second, 20*time.Millisecond, func() bool {
		return len(capture.Handled()) == 1
	}))
	s.Equal([][]any{{"goravel"}}, capture.Handled())
}

// TestDispatchQueuesAWildcardListenerWithTheMatchedName proves the queue
// carries the matched event name, not the pattern, when a queued wildcard runs
// through the database queue connection's app:queue:database boot worker.
func (s *EventTestSuite) TestDispatchQueuesAWildcardListenerWithTheMatchedName() {
	prefix := s.uniqueName("queued.wildcard")
	capture := &listenerCapture{}
	s.NoError(facades.Event().Listen(prefix+".*", &integrationListener{
		signature:   s.uniqueName("queued_wildcard_listener"),
		queueConfig: event.Queue{Enable: true, Connection: "database"},
		capture:     capture,
	}))

	matched := prefix + ".created"
	s.False(facades.Event().Dispatch(matched).Failed())

	s.True(waitUntil(5*time.Second, 20*time.Millisecond, func() bool {
		return len(capture.Handled()) == 1
	}))
	s.Equal([]string{matched}, capture.EventNames())
}

func (s *EventTestSuite) uniqueName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, eventNameCounter.Add(1))
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

// getEventsEvent is the payload of TestGetEventsReturnsACopy. Register keys the
// deprecated events registry by the event value, so this dedicated type keeps
// the GetEvents copy assertions clear of the bootstrapped OrderShipped and
// OrderCanceled entries as well as the integrationEvent registrations of the
// other deprecated tests.
type getEventsEvent struct {
	integrationEvent
}

// panickingListener panics from Handle to exercise the dispatcher's recovery.
type panickingListener struct {
	signature string
}

func (receiver *panickingListener) Signature() string { return receiver.signature }

func (receiver *panickingListener) Queue(args ...any) event.Queue { return event.Queue{} }

func (receiver *panickingListener) Handle(eventName string, args ...any) error {
	panic("listener panic")
}

// valueListener uses value receivers on purpose so a value satisfies
// event.Listener and Listen rejects it with EventListenerNotPointer.
type valueListener struct{}

func (valueListener) Signature() string { return "value_listener" }

func (valueListener) Queue(args ...any) event.Queue { return event.Queue{} }

func (valueListener) Handle(eventName string, args ...any) error { return nil }

// Dedicated event types give each test its own event name, since Listen appends.
type listenClosureEvent struct{ integrationEvent }

type listenSliceEventA struct{ integrationEvent }

type listenSliceEventB struct{ integrationEvent }

type panickingEvent struct{ integrationEvent }

type skippedHandleEvent struct{ integrationEvent }

type registerReachedEvent struct{ integrationEvent }

type coalesceEvent struct{ integrationEvent }

type jobListenReachedEvent struct{ integrationEvent }

type jobWildcardEvent struct{ integrationEvent }

type queuedRegisterEvent struct{ integrationEvent }

// jobIdentifiedEvent has a field so the value identity and the event name differ.
type jobIdentifiedEvent struct {
	id int
}

func (receiver *jobIdentifiedEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
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
		// Record the name first: a waiter polling Handled() must not observe the
		// payload before the name that belongs with it.
		receiver.capture.AddEventName(eventName)
		receiver.capture.AddHandled(args)
	}

	return receiver.handleErr
}

type listenerCapture struct {
	mu         sync.Mutex
	handled    [][]any
	queueArgs  [][]any
	eventNames []string
}

func (receiver *listenerCapture) AddHandled(args []any) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.handled = append(receiver.handled, copyAnySlice(args))
}

func (receiver *listenerCapture) AddEventName(eventName string) {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	receiver.eventNames = append(receiver.eventNames, eventName)
}

func (receiver *listenerCapture) EventNames() []string {
	receiver.mu.Lock()
	defer receiver.mu.Unlock()

	names := make([]string, len(receiver.eventNames))
	copy(names, receiver.eventNames)

	return names
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
