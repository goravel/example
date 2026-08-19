package feature

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	sqlitefacades "github.com/goravel/sqlite/facades"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/suite"

	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/contracts/notification"
	contractsqueue "github.com/goravel/framework/contracts/queue"
	frameworkerrors "github.com/goravel/framework/errors"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/str"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/app/notifications"
	"goravel/tests"
)

var (
	afterSendingCalled bool
	capturedSends      []any
)

type NotificationTestSuite struct {
	suite.Suite
	tests.TestCase
	counter uint64
}

func TestNotificationTestSuite(t *testing.T) {
	suite.Run(t, &NotificationTestSuite{})
}

func (s *NotificationTestSuite) SetupTest() {
	s.RefreshDatabase()

	afterSendingCalled = false
	capturedSends = nil
}

func (s *NotificationTestSuite) TestFacadeResolves() {
	s.NotNil(facades.Notification())
	s.NotNil(facades.Notification().Channel(notification.ChannelDatabase))
	s.NotNil(facades.Notification().Channel(notification.ChannelMail))
	s.Nil(facades.Notification().Channel("slack"))
}

func (s *NotificationTestSuite) TestSendDatabaseNotification() {
	user := &models.User{Name: "Bowen", Mail: "bowen@example.com"}
	s.Require().NoError(facades.Orm().Query().Create(user))

	s.NoError(facades.Notification().Send(user, notifications.NewWelcome("Bowen")))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal("welcome-notification", rows[0].ID)
	s.Contains(rows[0].Type, "Welcome")
	s.Equal(cast.ToString(user.ID), rows[0].NotifiableID)
	s.Contains(rows[0].Data, "Welcome Bowen")
}

func (s *NotificationTestSuite) TestSendNowDatabaseNotification() {
	user := &models.User{Name: "Now"}
	s.Require().NoError(facades.Orm().Query().Create(user))

	// Use the async database queue driver and disable the booted queue
	// workers so a queued notification could never be delivered by a
	// background runner.
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default":        "database",
		"app.disabled_runners": []string{"app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	s.NoError(facades.Notification().SendNow(user, notifications.NewOrderProcessed("Now")))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal(cast.ToString(user.ID), rows[0].NotifiableID)
	s.Contains(rows[0].Data, "Now")

	// SendNow must not have queued anything.
	jobsCount, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), jobsCount)
}

func (s *NotificationTestSuite) TestSendQueuedDatabaseNotification() {
	user := &models.User{Name: "Queued"}
	s.Require().NoError(facades.Orm().Query().Create(user))

	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default":        "database",
		"app.disabled_runners": []string{"app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	s.NoError(facades.Notification().Send(user, notifications.NewOrderProcessed("42")))

	// The notification is queued, not delivered inline.
	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 0)

	// Empty OnQueue/OnConnection default to the "default" queue on the
	// database connection.
	var jobRows []jobRow
	s.NoError(facades.DB().Table("jobs").Get(&jobRows))
	s.Require().Len(jobRows, 1)
	s.Equal("default", jobRows[0].Queue)

	worker := facades.Queue().Worker(contractsqueue.Args{Queue: "default", Concurrent: 1})
	workerErr := make(chan error, 1)
	go func() { workerErr <- worker.Run() }()
	var shutdownOnce sync.Once
	defer func() { shutdownOnce.Do(func() { _ = worker.Shutdown() }) }()

	s.Require().Eventually(func() bool {
		rows = nil
		if err := facades.DB().Table("notifications").Get(&rows); err != nil {
			return false
		}
		return len(rows) == 1
	}, 5*time.Second, 25*time.Millisecond)
	s.Contains(rows[0].Type, "OrderProcessed")
	s.Equal(cast.ToString(user.ID), rows[0].NotifiableID)
	s.Contains(rows[0].Data, "42")

	shutdownOnce.Do(func() { _ = worker.Shutdown() })
	s.NoError(<-workerErr, "worker should start and stop cleanly")

	// The job was consumed: the queue must be drained.
	jobsCount, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), jobsCount)
}

func (s *NotificationTestSuite) TestSendQueuedDatabaseNotificationWithQueueAndConnection() {
	user := &models.User{Name: "Routed"}
	s.Require().NoError(facades.Orm().Query().Create(user))

	s.NoError(facades.Notification().Send(user, &routedQueuedNotification{}))

	// Not delivered inline: the notification is queued.
	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 0)

	// OnConnection("database") + OnQueue("notifications") are both observable
	// in the jobs table.
	var jobRows []jobRow
	s.NoError(facades.DB().Table("jobs").Get(&jobRows))
	s.Require().Len(jobRows, 1)
	s.Equal("notifications", jobRows[0].Queue)

	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "notifications",
		Concurrent: 1,
	})
	workerErr := make(chan error, 1)
	go func() { workerErr <- worker.Run() }()
	var shutdownOnce sync.Once
	defer func() { shutdownOnce.Do(func() { _ = worker.Shutdown() }) }()

	s.Require().Eventually(func() bool {
		rows = nil
		if err := facades.DB().Table("notifications").Get(&rows); err != nil {
			return false
		}
		return len(rows) == 1
	}, 5*time.Second, 25*time.Millisecond)
	s.Equal(cast.ToString(user.ID), rows[0].NotifiableID)
	s.Contains(rows[0].Data, "routed")

	shutdownOnce.Do(func() { _ = worker.Shutdown() })
	s.NoError(<-workerErr, "worker should start and stop cleanly")

	// The job was consumed: the queue must be drained.
	jobsCount, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), jobsCount)
}

func (s *NotificationTestSuite) TestOnDemandNotification() {
	s.NoError(facades.Notification().Route(notification.ChannelDatabase, "123").Notify(notifications.NewWelcome("OnDemand")))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal("123", rows[0].NotifiableID)
	s.Contains(rows[0].Data, "Welcome OnDemand")
}

func (s *NotificationTestSuite) TestOnDemandChainedRouteNotifyNow() {
	channelName := s.uniqueName("chained_capture")
	manager := facades.Notification()
	manager.Extend(&captureChannel{name: channelName})

	cn := &chainedViaNotification{channel: channelName}
	s.NoError(manager.Route(notification.ChannelDatabase, "456").Route(channelName, "route").NotifyNow(cn))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal("456", rows[0].NotifiableID)
	s.Contains(rows[0].Data, "chained")
	s.Equal([]any{cn}, capturedSends)
}

func (s *NotificationTestSuite) TestShouldSendSkipsChannel() {
	s.NoError(facades.Notification().Route(notification.ChannelDatabase, "789").NotifyNow(&shouldSendNotification{shouldSend: false}))

	count, err := facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(0), count)

	s.NoError(facades.Notification().Route(notification.ChannelDatabase, "789").NotifyNow(&shouldSendNotification{shouldSend: true}))

	count, err = facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(1), count)
}

func (s *NotificationTestSuite) TestAfterSendingHook() {
	s.NoError(facades.Notification().Route(notification.ChannelDatabase, "1").NotifyNow(&afterSendingNotification{}))
	s.True(afterSendingCalled)

	count, err := facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(1), count)
}

func (s *NotificationTestSuite) TestNotificationWithDatabaseConnectionDefault() {
	s.NoError(facades.Notification().Route(notification.ChannelDatabase, "101").NotifyNow(&defaultConnectionNotification{}))

	count, err := facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(1), count)
}

func (s *NotificationTestSuite) TestNotificationWithDatabaseConnectionCustomConnection() {
	scope, err := tests.OverrideConfig(map[string]any{
		"database.connections.reporting": map[string]any{
			"database": filepath.Join(s.T().TempDir(), "reporting.db"),
			"prefix":   "",
			"singular": false,
			"via": func() (contractsdriver.Driver, error) {
				return sqlitefacades.Sqlite("reporting")
			},
		},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	// Mirror database/migrations/20260805000001_create_notifications_table.go
	// on the reporting connection; keep this blueprint in sync with that
	// migration if it changes.
	s.Require().NoError(facades.Schema().Connection("reporting").Create("notifications", func(table contractsschema.Blueprint) {
		table.String("id", 36)
		table.Primary("id")
		table.String("type")
		table.String("notifiable_type")
		table.String("notifiable_id")
		table.Text("data")
		table.Timestamp("read_at").Nullable()
		table.Timestamps()
		table.Index("notifiable_type", "notifiable_id")
	}))

	s.NoError(facades.Notification().Route(notification.ChannelDatabase, "101").NotifyNow(&reportingConnectionNotification{}))

	count, err := facades.DB().Connection("reporting").Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(1), count)

	count, err = facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(0), count)
}

func (s *NotificationTestSuite) TestCustomChannel() {
	channelName := s.uniqueName("capture_channel")
	manager := facades.Notification()
	manager.Extend(&captureChannel{name: channelName})
	s.NotNil(manager.Channel(channelName))

	cn := &customViaNotification{channel: channelName}
	s.NoError(manager.Route(channelName, "route").NotifyNow(cn))
	s.Equal([]any{cn}, capturedSends)
}

func (s *NotificationTestSuite) TestSendToUnknownChannel() {
	err := facades.Notification().Route("slack", "route").NotifyNow(&unknownChannelNotification{})
	s.Error(err)
	s.ErrorIs(err, frameworkerrors.NotificationChannelNotFound.Args("slack"))
}

func (s *NotificationTestSuite) TestSendMailNotification() {
	mailTo := s.mailRecipient()

	s.NoError(facades.Notification().Send(&models.User{Mail: mailTo}, notifications.NewOrderShipped("notification-mail")))
}

func (s *NotificationTestSuite) TestSendMailNotificationMailRoutable() {
	mailTo := s.mailRecipient()

	s.NoError(facades.Notification().Send(&mailRoutableNotifiable{to: mailTo}, notifications.NewOrderShipped("notification-mail-routable")))
}

func (s *NotificationTestSuite) TestDatabaseRoutableTypedRoute() {
	s.NoError(facades.Notification().Send(
		&databaseRoutableNotifiable{typed: "42", fallback: "wrong-route"},
		notifications.NewWelcome("typed"),
	))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal("42", rows[0].NotifiableID)
}

func (s *NotificationTestSuite) TestDatabaseRoutableFallbackToGenericRoute() {
	s.NoError(facades.Notification().Send(
		&databaseRoutableNotifiable{typed: "", fallback: "43"},
		notifications.NewWelcome("fallback"),
	))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal("43", rows[0].NotifiableID)
}

func (s *NotificationTestSuite) TestDatabaseRoutableEmptyRouteError() {
	err := facades.Notification().Send(
		&databaseRoutableNotifiable{},
		notifications.NewWelcome("empty"),
	)
	s.ErrorIs(err, frameworkerrors.NotificationDatabaseEmptyRoute)
}

func (s *NotificationTestSuite) TestSendMailNotificationEmptyRoute() {
	err := facades.Notification().Send(
		&mailEmptyRouteNotifiable{},
		notifications.NewOrderShipped("empty-mail"),
	)
	s.ErrorIs(err, frameworkerrors.NotificationMailEmptyRoute)
}

func (s *NotificationTestSuite) TestCommandMakeNotification() {
	notificationName := s.uniqueName("NotificationFeature")
	nestedPackage := s.uniqueName("NotificationFeatureNested")
	nestedNotificationName := s.uniqueName("GeneratedNotification")
	notificationPath := path.App("notifications", str.Of(notificationName).Snake().String()+".go")
	nestedDir := path.App("notifications", nestedPackage)
	nestedPath := path.App("notifications", nestedPackage, str.Of(nestedNotificationName).Snake().String()+".go")

	s.NoError(os.RemoveAll(notificationPath))
	s.NoError(os.RemoveAll(nestedDir))
	s.T().Cleanup(func() {
		s.NoError(os.RemoveAll(notificationPath))
		s.NoError(os.RemoveAll(nestedDir))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:notification " + notificationName))
	s.True(file.Exists(notificationPath))
	s.True(file.Contains(notificationPath, "type "+notificationName+" struct {"))
	s.True(file.Contains(notificationPath, "func (r *"+notificationName+") Via(notifiable notification.Notifiable) []string {"))
	s.True(file.Contains(notificationPath, "func (r *"+notificationName+") ToMail(notifiable notification.Notifiable) notification.MailMessage {"))
	s.True(file.Contains(notificationPath, `"github.com/goravel/framework/notification/mail"`))
	s.True(file.Contains(notificationPath, "return []string{notification.ChannelMail}"))

	originalContent, err := os.ReadFile(notificationPath)
	s.NoError(err)

	output, err := s.CaptureArtisanOutput("--no-ansi make:notification " + notificationName)
	s.NoError(err)
	s.Contains(output, "already exists")

	currentContent, err := os.ReadFile(notificationPath)
	s.NoError(err)
	s.Equal(string(originalContent), string(currentContent))

	databaseName := s.uniqueName("NotificationDatabase")
	databasePath := path.App("notifications", str.Of(databaseName).Snake().String()+".go")
	s.NoError(os.RemoveAll(databasePath))
	s.T().Cleanup(func() {
		s.NoError(os.RemoveAll(databasePath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:notification --database " + databaseName))
	s.True(file.Exists(databasePath))
	s.True(file.Contains(databasePath, "func (r *"+databaseName+") ToDatabase(notifiable notification.Notifiable) map[string]any {"))
	s.True(file.Contains(databasePath, "return []string{notification.ChannelDatabase}"))
	s.False(file.Contains(databasePath, "ToMail"))
	s.False(file.Contains(databasePath, `"github.com/goravel/framework/notification/mail"`))

	s.NoError(facades.Artisan().Call("--no-ansi make:notification " + nestedPackage + "/" + nestedNotificationName))
	s.True(file.Exists(nestedPath))
	s.True(file.Contains(nestedPath, "package "+nestedPackage))
	s.True(file.Contains(nestedPath, "type "+nestedNotificationName+" struct {"))
}

func (s *NotificationTestSuite) TestCommandNotificationsTable() {
	migrationsPath := path.Bootstrap("migrations.go")
	originalContent, err := os.ReadFile(migrationsPath)
	s.Require().NoError(err)
	s.T().Cleanup(func() {
		s.Require().NoError(os.WriteFile(migrationsPath, originalContent, 0o644))
	})

	beforeFiles := s.listRelativeMigrationFiles()

	s.NoError(facades.Artisan().Call("--no-ansi notifications:table"))

	afterFiles := s.listRelativeMigrationFiles()
	var createdFiles []string
	for item := range afterFiles {
		if _, ok := beforeFiles[item]; !ok && strings.Contains(item, "_create_notifications_table.go") {
			createdFiles = append(createdFiles, item)
		}
	}
	s.Require().Len(createdFiles, 1)
	s.Contains(createdFiles[0], "_create_notifications_table.go")

	createdPath := filepath.Join("database", "migrations", createdFiles[0])
	s.Require().FileExists(createdPath)

	// Re-run: if it lands in the same second as the first run, the command
	// detects the existing migration and warns; if a second boundary is
	// crossed, it generates a new migration with a fresh timestamp instead.
	// Both are valid behaviors — assert whichever occurred.
	output, err := s.CaptureArtisanOutput("--no-ansi notifications:table")
	s.NoError(err)

	rerunFiles := s.listRelativeMigrationFiles()
	var rerunCreated []string
	for item := range rerunFiles {
		if _, ok := beforeFiles[item]; !ok && strings.Contains(item, "_create_notifications_table.go") {
			rerunCreated = append(rerunCreated, item)
		}
	}
	if strings.Contains(output, "already exists") {
		s.Equal(createdFiles, rerunCreated)
	} else {
		// A second boundary was crossed: the re-run created one more file.
		s.Len(rerunCreated, len(createdFiles)+1)
	}

	// Clean up every migration file the command generated during this test
	// (from either the create or the re-run call).
	s.T().Cleanup(func() {
		for _, item := range rerunCreated {
			s.NoError(file.Remove(filepath.Join("database", "migrations", item)))
		}
		_ = os.Remove(filepath.Join("database", "migrations"))
		_ = os.Remove("database")
	})

	updatedBootstrap, err := os.ReadFile(migrationsPath)
	s.Require().NoError(err)

	// The command names the generated file "<timestamp>_create_notifications_table.go"
	// and registers "&migrations.M<timestamp>CreateNotificationsTable{}", so
	// assert the exact entry for the file created in this test rather than a
	// generic prefix that the permanent notifications migration would match.
	ts := strings.TrimSuffix(createdFiles[0], "_create_notifications_table.go")
	s.Contains(string(updatedBootstrap), "&migrations.M"+ts+"CreateNotificationsTable{}")
}

// mailRecipient skips the test when mail is not configured and resolves the
// recipient address: MAIL_TO env if set, otherwise the global
// mail.from.address config.
func (s *NotificationTestSuite) mailRecipient() string {
	if facades.Config().GetString("mail.host") == "" {
		s.T().Skip("skipping mail notification test: set MAIL_HOST to run against a real SMTP server")
	}
	mailTo := os.Getenv("MAIL_TO") // MAIL_TO is intentionally env-only (not a mail config key)
	if mailTo == "" {
		mailTo = facades.Config().GetString("mail.from.address")
	}
	s.Require().NotEmpty(mailTo)
	return mailTo
}

func (s *NotificationTestSuite) uniqueName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, atomic.AddUint64(&s.counter, 1))
}

// extendFlakyChannel registers a flaky custom channel on the notification
// Manager that the queued DispatchJob will use. The Manager is bound as a
// singleton (framework#1537), so facades.Notification() returns the same
// instance that the boot-time DispatchJob registered by registerJobs holds:
// runtime Extend calls are visible to the worker's delivery without any
// re-registration.
//
// Caution: this mutates global queue state (the shared JobStorer), so the
// tests that call it MUST NOT be run with t.Parallel(). Cleanup relies on
// each test's OverrideConfig defer scope.Restore() → facades.App().Restart(),
// which re-runs registerJobs and re-registers the default DispatchJob.
func (s *NotificationTestSuite) extendFlakyChannel(name string, failUntil int, alwaysFail bool) *flakyChannel {
	flaky := &flakyChannel{name: name, failUntil: failUntil, alwaysFail: alwaysFail}
	facades.Notification().Extend(flaky)

	return flaky
}

func (s *NotificationTestSuite) TestSendQueuedNotificationWithTriesAndBackoff() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default":        "database",
		"app.disabled_runners": []string{"app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	channelName := s.uniqueName("flaky")
	flaky := s.extendFlakyChannel(channelName, 0, true)

	s.NoError(facades.Notification().Route(channelName, "route").Notify(
		&retryableNotification{channel: channelName, tries: 3, backoff: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}},
	))

	// Dispatch-time capture: the retry policy is serialized to the payload.
	item := s.readQueuedNotificationItem("default")
	s.Equal(3, item.Tries)
	s.Equal([]int64{100, 200}, item.Backoff) // backoff in ms

	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "default",
		Concurrent: 1,
		Tries:      1, // notification policy must override the worker's tries
	})
	workerErr := make(chan error, 1)
	start := time.Now()
	go func() { workerErr <- worker.Run() }()

	elapsed, ok := s.waitForFailedNotification(start, 5*time.Second)
	_ = worker.Shutdown()
	s.Require().True(ok, "expected goravel_notifications:dispatch job to fail within 5s")
	s.NoError(<-workerErr, "worker should start and stop cleanly")

	count, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "job should be consumed")

	s.Equal(3, flaky.attempts(), "notification Tries=3 should override worker Tries=1")
	// Coarse sanity check only: the database worker's pop loop sleeps a fixed
	// ~1s interval between retries, so elapsed is dominated by that poll
	// interval rather than by the 100ms/200ms backoff. The lower bound catches
	// a missing retry path (e.g. no retries at all), not backoff regressions.
	s.GreaterOrEqual(elapsed, 300*time.Millisecond, "retries should not complete before the release delays")
	s.Less(elapsed, 5*time.Second, "retries should complete within 5s")
}

func (s *NotificationTestSuite) TestSendQueuedNotificationWithoutTriesFallsBackToWorkerTries() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default": "database",
		// Disable the framework's boot-time queue runner (Tries=1) so it can't
		// steal the re-released job between attempts.
		"app.disabled_runners": []string{"goravel:queue", "app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	channelName := s.uniqueName("flaky")
	flaky := s.extendFlakyChannel(channelName, 0, true)

	s.NoError(facades.Notification().Route(channelName, "route").Notify(
		&retryableNotification{channel: channelName, tries: 0},
	))

	item := s.readQueuedNotificationItem("default")
	s.Zero(item.Tries) // omitempty: no retry policy serialized
	s.Empty(item.Backoff)

	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "default",
		Concurrent: 1,
		Tries:      3, // no declared Tries → the worker's Tries now governs
	})
	workerErr := make(chan error, 1)
	start := time.Now()
	go func() { workerErr <- worker.Run() }()

	_, ok := s.waitForFailedNotification(start, 10*time.Second)
	_ = worker.Shutdown()
	s.Require().True(ok, "expected goravel_notifications:dispatch job to fail within 10s")
	s.NoError(<-workerErr, "worker should start and stop cleanly")

	count, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), count)
	s.Equal(3, flaky.attempts(), "notification without Tries must fall back to the worker's Tries=3")
}

func (s *NotificationTestSuite) TestSendQueuedNotificationBackoffWithoutTriesSerialized() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default":        "database",
		"app.disabled_runners": []string{"app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	channelName := s.uniqueName("flaky")
	s.extendFlakyChannel(channelName, 0, true)

	s.NoError(facades.Notification().Route(channelName, "route").Notify(
		&retryableNotification{channel: channelName, tries: 0, backoff: []time.Duration{time.Second}},
	))

	// Dispatch only (no worker): backoff must serialize even without tries,
	// so worker-driven retries honor the declared delay.
	item := s.readQueuedNotificationItem("default")
	s.Zero(item.Tries)
	s.Equal([]int64{1000}, item.Backoff) // 1s → ms
}

func (s *NotificationTestSuite) TestSendQueuedNotificationBackoffWithoutTriesFallsBackToWorkerTries() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default": "database",
		// Disable the framework's boot-time queue runner (Tries=1) so it can't
		// steal the job during the backoff release window.
		"app.disabled_runners": []string{"goravel:queue", "app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	channelName := s.uniqueName("flaky")
	flaky := s.extendFlakyChannel(channelName, 0, true)

	s.NoError(facades.Notification().Route(channelName, "route").Notify(
		&retryableNotification{channel: channelName, tries: 0, backoff: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}},
	))

	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "default",
		Concurrent: 1,
		Tries:      3, // no declared Tries → worker's Tries governs, backoff still applies
	})
	workerErr := make(chan error, 1)
	start := time.Now()
	go func() { workerErr <- worker.Run() }()

	elapsed, ok := s.waitForFailedNotification(start, 10*time.Second)
	_ = worker.Shutdown()
	s.Require().True(ok, "expected goravel_notifications:dispatch job to fail within 10s")
	s.NoError(<-workerErr, "worker should start and stop cleanly")

	s.Equal(3, flaky.attempts(), "worker Tries=3 should govern a notification without its own Tries")
	// Coarse sanity check only: the database worker's pop loop sleeps a fixed
	// ~1s interval between retries, so elapsed is dominated by that poll
	// interval rather than by the 100ms/200ms backoff. This only confirms the
	// retry path exists (3 attempts); it does not verify backoff is honored.
	s.GreaterOrEqual(elapsed, 300*time.Millisecond, "retries should not complete before the release delays")
}

func (s *NotificationTestSuite) TestSendQueuedOrderFailedCapturesRetryPolicy() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default":        "database",
		"app.disabled_runners": []string{"app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	user := &models.User{Name: "Bowen", Mail: "bowen@example.com"}
	s.Require().NoError(facades.Orm().Query().Create(user))

	s.NoError(facades.Notification().Send(user, notifications.NewOrderFailed("42")))

	item := s.readQueuedNotificationItem("default")
	s.Equal(3, item.Tries)
	s.Equal([]int64{1000, 2000}, item.Backoff) // demo [1s, 2s] → ms
}

func (s *NotificationTestSuite) TestSendQueuedNotificationRetriesTransientFailure() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default":        "database",
		"app.disabled_runners": []string{"app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	channelName := s.uniqueName("flaky")
	flaky := s.extendFlakyChannel(channelName, 2, false)

	s.NoError(facades.Notification().Route(channelName, "route").Notify(
		&retryableNotification{channel: channelName, tries: 3, backoff: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond}},
	))

	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "default",
		Concurrent: 1,
		Tries:      1,
	})
	workerErr := make(chan error, 1)
	go func() { workerErr <- worker.Run() }()

	s.Require().Eventually(func() bool { return flaky.attempts() >= 3 }, 5*time.Second, 100*time.Millisecond)
	_ = worker.Shutdown()
	s.NoError(<-workerErr, "worker should start and stop cleanly")

	s.Equal(3, flaky.attempts(), "delivery should have succeeded on the third attempt")
	count, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "job should be consumed after recovery")

	// No failed job: the retry recovered instead of failing.
	failedJobs, err := facades.Queue().Failer().All()
	s.NoError(err)
	s.Empty(failedJobs)
}

// waitForFailedNotification polls the queue failer until a failed
// goravel_notifications:dispatch job appears, returning elapsed time since
// start. Mirrors waitForFailedBroadcast; relies on RefreshDatabase resetting
// failed_jobs in SetupTest and each test dispatching exactly one queued
// notification.
func (s *NotificationTestSuite) waitForFailedNotification(start time.Time, timeout time.Duration) (time.Duration, bool) {
	deadline := time.Now().Add(timeout)
	for {
		failedJobs, err := facades.Queue().Failer().All()
		if err != nil {
			s.T().Logf("failed to list failed jobs: %v", err)
		} else {
			for _, fj := range failedJobs {
				if fj.Signature() == "goravel_notifications:dispatch" {
					return time.Since(start), true
				}
			}
		}
		if time.Now().After(deadline) {
			return time.Since(start), false
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// readQueuedNotificationItem decodes the dispatch-time Tries/Backoff carried
// in the jobs table payload for the given queue.
func (s *NotificationTestSuite) readQueuedNotificationItem(queue string) queuedNotificationItem {
	var jobs []jobRow
	s.Require().NoError(facades.DB().Table("jobs").Where("queue", queue).Get(&jobs))
	s.Require().Len(jobs, 1)

	var payload struct {
		Args []struct {
			Value string `json:"value"`
		} `json:"args"`
	}
	s.Require().NoError(json.Unmarshal([]byte(jobs[0].Payload), &payload))
	s.Require().Len(payload.Args, 1)

	var item queuedNotificationItem
	s.Require().NoError(json.Unmarshal([]byte(payload.Args[0].Value), &item))
	return item
}

// listRelativeMigrationFiles lists *.go files under the cwd-relative
// database/migrations directory (the path the notifications:table command
// writes to).
func (s *NotificationTestSuite) listRelativeMigrationFiles() map[string]struct{} {
	entries, err := os.ReadDir("database/migrations")
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]struct{}{}
		}
		s.NoError(err)
	}

	files := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".go") {
			files[entry.Name()] = struct{}{}
		}
	}
	return files
}

// notificationRow mirrors the notifications table columns asserted in tests.
type notificationRow struct {
	ID             string `db:"id"`
	Type           string `db:"type"`
	NotifiableType string `db:"notifiable_type"`
	NotifiableID   string `db:"notifiable_id"`
	Data           string `db:"data"`
}

// jobRow mirrors the jobs table column asserted in the queue tests.
type jobRow struct {
	Queue   string `db:"queue"`
	Payload string `db:"payload"`
}

// flakyChannel implements notification.ResolvableChannel. Deliver fails the
// first failUntil invocations (or always, when alwaysFail is set), then
// succeeds, recording every invocation so tests can assert the exact attempt
// count without racing the worker goroutine.
type flakyChannel struct {
	name       string
	failUntil  int
	alwaysFail bool

	mu           sync.Mutex
	deliverCount int
}

func (f *flakyChannel) Name() string { return f.name }

// Send satisfies Channel (embedded in ResolvableChannel); the queued path
// calls Resolve+Deliver directly, so this mirrors DatabaseChannel.Send.
func (f *flakyChannel) Send(notifiable notification.Notifiable, n notification.Notification) error {
	route, payload, err := f.Resolve(notifiable, n)
	if err != nil {
		return err
	}
	return f.Deliver(route, payload)
}

func (f *flakyChannel) Resolve(notifiable notification.Notifiable, n notification.Notification) (string, []byte, error) {
	return "route", []byte(`{}`), nil
}

func (f *flakyChannel) Deliver(route string, payload []byte) error {
	f.mu.Lock()
	f.deliverCount++
	count := f.deliverCount
	f.mu.Unlock()

	if f.alwaysFail || count <= f.failUntil {
		return errors.New("transient delivery failure")
	}
	return nil
}

func (f *flakyChannel) attempts() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.deliverCount
}

// retryableNotification implements ShouldQueue + NotificationWithTries +
// NotificationWithBackoff, routing to a flaky custom channel so the queued
// DispatchJob's Tries/Backoff policy is exercised end-to-end.
type retryableNotification struct {
	channel string
	tries   int
	backoff []time.Duration
}

func (r *retryableNotification) Via(notifiable notification.Notifiable) []string {
	return []string{r.channel}
}

func (r *retryableNotification) OnQueue() string      { return "" }
func (r *retryableNotification) OnConnection() string { return "" }

func (r *retryableNotification) Tries(channel string) int {
	return r.tries
}

func (r *retryableNotification) Backoff(channel string) []time.Duration {
	return r.backoff
}

// queuedNotificationItem decodes the dispatch-time Tries/Backoff carried in
// the jobs table payload. The dispatch arg is a JSON string whose fields
// mirror broadcasting's broadcastItem wire format (backoff in ms).
type queuedNotificationItem struct {
	Tries   int     `json:"tries"`
	Backoff []int64 `json:"backoff"`
}

// ---- Test-local notifications / channels / notifiables ----

// routedQueuedNotification implements ShouldQueue with non-empty
// OnQueue/OnConnection values.
type routedQueuedNotification struct{}

func (r *routedQueuedNotification) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelDatabase}
}

func (r *routedQueuedNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"order_id": "routed"}
}

func (r *routedQueuedNotification) OnQueue() string      { return "notifications" }
func (r *routedQueuedNotification) OnConnection() string { return "database" }

// shouldSendNotification implements NotificationWithShouldSend.
type shouldSendNotification struct {
	shouldSend bool
}

func (r *shouldSendNotification) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelDatabase}
}

func (r *shouldSendNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"message": "hello"}
}

func (r *shouldSendNotification) ShouldSend(notifiable notification.Notifiable, channel string) bool {
	return r.shouldSend
}

// afterSendingNotification implements NotificationWithAfterSending.
type afterSendingNotification struct{}

func (r *afterSendingNotification) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelDatabase}
}

func (r *afterSendingNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"message": "after"}
}

func (r *afterSendingNotification) AfterSending(notifiable notification.Notifiable, channel string) error {
	afterSendingCalled = true
	return nil
}

// defaultConnectionNotification implements
// NotificationWithDatabaseConnection using the default connection (empty
// name).
type defaultConnectionNotification struct{}

func (r *defaultConnectionNotification) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelDatabase}
}

func (r *defaultConnectionNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"message": "routable"}
}

func (r *defaultConnectionNotification) DatabaseConnection() string { return "" }

// reportingConnectionNotification implements
// NotificationWithDatabaseConnection using a custom "reporting" connection.
type reportingConnectionNotification struct{}

func (r *reportingConnectionNotification) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelDatabase}
}

func (r *reportingConnectionNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"message": "reporting"}
}

func (r *reportingConnectionNotification) DatabaseConnection() string { return "reporting" }

// captureChannel implements notification.Channel with an observable Send
// side effect.
type captureChannel struct {
	name string
}

func (r *captureChannel) Name() string { return r.name }

func (r *captureChannel) Send(notifiable notification.Notifiable, n notification.Notification) error {
	capturedSends = append(capturedSends, n)
	return nil
}

// customViaNotification routes to the channel name it was configured with.
type customViaNotification struct {
	channel string
}

func (r *customViaNotification) Via(notifiable notification.Notifiable) []string {
	return []string{r.channel}
}

// chainedViaNotification routes to both the database channel and a custom
// channel, so chained routes are exercised on every leg.
type chainedViaNotification struct {
	channel string
}

func (r *chainedViaNotification) Via(notifiable notification.Notifiable) []string {
	return []string{notification.ChannelDatabase, r.channel}
}

func (r *chainedViaNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"message": "chained"}
}

// unknownChannelNotification routes to a channel that is never registered.
type unknownChannelNotification struct{}

func (r *unknownChannelNotification) Via(notifiable notification.Notifiable) []string {
	return []string{"slack"}
}

// mailRoutableNotifiable implements MailRoutable with a fixed recipient.
type mailRoutableNotifiable struct {
	to string
}

func (r *mailRoutableNotifiable) RouteNotificationFor(channel string) any {
	return nil
}

func (r *mailRoutableNotifiable) RouteNotificationForMail(notification notification.Notification) map[string]string {
	return map[string]string{r.to: "Test User"}
}

// databaseRoutableNotifiable implements contracts/notification.DatabaseRoutable
// (typed route) in addition to Notifiable. typed is what
// RouteNotificationForDatabase returns and fallback is what
// RouteNotificationFor(ChannelDatabase) returns, so tests can pin which
// route the database channel prefers and what happens when either is empty.
type databaseRoutableNotifiable struct {
	typed    string
	fallback any
}

func (r *databaseRoutableNotifiable) RouteNotificationFor(channel string) any {
	if channel == notification.ChannelDatabase {
		return r.fallback
	}
	return nil
}

func (r *databaseRoutableNotifiable) RouteNotificationForDatabase() string {
	return r.typed
}

// mailEmptyRouteNotifiable provides no mail route: RouteNotificationFor
// returns nil for every channel, so the mail channel reports an empty
// route without touching SMTP.
type mailEmptyRouteNotifiable struct{}

func (r *mailEmptyRouteNotifiable) RouteNotificationFor(channel string) any {
	return nil
}
