package feature

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/notification"
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

// TestFacadeResolves covers the Manager contract plus Manager.Channel for
// both registered and unknown channel names.
func (s *NotificationTestSuite) TestFacadeResolves() {
	s.NotNil(facades.Notification())
	s.NotNil(facades.Notification().Channel("database"))
	s.NotNil(facades.Notification().Channel("mail"))
	s.Nil(facades.Notification().Channel("slack"))
}

// TestSendDatabaseNotification covers Via, Notifiable.RouteNotificationFor,
// DatabaseNotification.ToDatabase and NotificationWithID by asserting the
// persisted row's identity and payload.
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

// TestSendNowDatabaseNotification covers Manager.SendNow delivering
// synchronously.
func (s *NotificationTestSuite) TestSendNowDatabaseNotification() {
	user := &models.User{Name: "Now"}
	s.Require().NoError(facades.Orm().Query().Create(user))

	s.NoError(facades.Notification().SendNow(user, notifications.NewWelcome("Now")))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal(cast.ToString(user.ID), rows[0].NotifiableID)
	s.Contains(rows[0].Data, "Welcome Now")
}

// TestSendQueuedDatabaseNotification covers the ShouldQueue contract
// (default queue/connection) via the queued-dispatch round-trip. The
// example's default queue driver is sync, so Delivery executes inline and
// the row is visible once Send returns.
func (s *NotificationTestSuite) TestSendQueuedDatabaseNotification() {
	user := &models.User{Name: "Queued"}
	s.Require().NoError(facades.Orm().Query().Create(user))

	s.NoError(facades.Notification().Send(user, notifications.NewOrderProcessed("42")))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Contains(rows[0].Type, "OrderProcessed")
	s.Equal(cast.ToString(user.ID), rows[0].NotifiableID)
	s.Contains(rows[0].Data, "42")
}

// TestSendQueuedDatabaseNotificationWithQueueAndConnection covers
// ShouldQueue.OnQueue/OnConnection returning non-empty values. The sync
// driver accepts an explicit connection/queue and still delivers inline.
func (s *NotificationTestSuite) TestSendQueuedDatabaseNotificationWithQueueAndConnection() {
	user := &models.User{Name: "Routed"}
	s.Require().NoError(facades.Orm().Query().Create(user))

	s.NoError(facades.Notification().Send(user, &routedQueuedNotification{}))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal(cast.ToString(user.ID), rows[0].NotifiableID)
	s.Contains(rows[0].Data, "routed")
}

// TestOnDemandNotification covers Manager.Route + OnDemandNotifiable.Notify.
func (s *NotificationTestSuite) TestOnDemandNotification() {
	s.NoError(facades.Notification().Route("database", "123").Notify(notifications.NewWelcome("OnDemand")))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal("123", rows[0].NotifiableID)
	s.Contains(rows[0].Data, "Welcome OnDemand")
}

// TestOnDemandChainedRouteNotifyNow covers OnDemandNotifiable.Route chaining
// and NotifyNow. The notification routes to both the database and a custom
// channel, so every chained route is actually exercised.
func (s *NotificationTestSuite) TestOnDemandChainedRouteNotifyNow() {
	channelName := s.uniqueName("chained_capture")
	manager := facades.Notification()
	manager.Extend(&captureChannel{name: channelName})

	cn := &chainedViaNotification{channel: channelName}
	s.NoError(manager.Route("database", "456").Route(channelName, "route").NotifyNow(cn))

	var rows []notificationRow
	s.NoError(facades.DB().Table("notifications").Get(&rows))
	s.Require().Len(rows, 1)
	s.Equal("456", rows[0].NotifiableID)
	s.Contains(rows[0].Data, "chained")
	s.Equal([]any{cn}, capturedSends)
}

// TestShouldSendSkipsChannel covers NotificationWithShouldSend: returning
// false must skip delivery entirely (no row), true must deliver.
func (s *NotificationTestSuite) TestShouldSendSkipsChannel() {
	s.NoError(facades.Notification().Route("database", "789").NotifyNow(&shouldSendNotification{shouldSend: false}))

	count, err := facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(0), count)

	s.NoError(facades.Notification().Route("database", "789").NotifyNow(&shouldSendNotification{shouldSend: true}))

	count, err = facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(1), count)
}

// TestAfterSendingHook covers NotificationWithAfterSending: the hook must
// run after a successful channel delivery.
func (s *NotificationTestSuite) TestAfterSendingHook() {
	s.NoError(facades.Notification().Route("database", "1").NotifyNow(&afterSendingNotification{}))
	s.True(afterSendingCalled)

	count, err := facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(1), count)
}

// TestDatabaseRoutableConnection covers DatabaseRoutable.DatabaseConnection:
// an empty connection name routes to the default connection and the row is
// persisted there.
func (s *NotificationTestSuite) TestDatabaseRoutableConnection() {
	s.NoError(facades.Notification().Route("database", "101").NotifyNow(&databaseRoutableNotification{}))

	count, err := facades.DB().Table("notifications").Count()
	s.NoError(err)
	s.Equal(int64(1), count)
}

// TestCustomChannel covers Manager.Extend + Channel.Name/Channel.Send with a
// user channel whose Send produces an observable side effect.
func (s *NotificationTestSuite) TestCustomChannel() {
	channelName := s.uniqueName("capture_channel")
	manager := facades.Notification()
	manager.Extend(&captureChannel{name: channelName})
	s.NotNil(manager.Channel(channelName))

	cn := &customViaNotification{channel: channelName}
	s.NoError(manager.Route(channelName, "route").NotifyNow(cn))
	s.Equal([]any{cn}, capturedSends)
}

// TestSendToUnknownChannel covers the error path for an unregistered
// channel name.
func (s *NotificationTestSuite) TestSendToUnknownChannel() {
	err := facades.Notification().Route("slack", "route").NotifyNow(&unknownChannelNotification{})
	s.Error(err)
	s.ErrorIs(err, frameworkerrors.NotificationChannelNotFound.Args("slack"))
}

// TestSendMailNotification covers MailableNotification.ToMail plus
// Notifiable.RouteNotificationFor("mail") against a real SMTP server.
// Skipped when mail.host is unset (resolved from config, which loads .env).
func (s *NotificationTestSuite) TestSendMailNotification() {
	if facades.Config().GetString("mail.host") == "" {
		s.T().Skip("skipping mail notification test: set MAIL_HOST to run against a real SMTP server")
	}
	mailTo := os.Getenv("MAIL_TO") // MAIL_TO is intentionally env-only (not a mail config key)
	if mailTo == "" {
		mailTo = facades.Config().GetString("mail.from.address")
	}
	s.Require().NotEmpty(mailTo)

	s.NoError(facades.Notification().Send(&models.User{Mail: mailTo}, notifications.NewOrderShipped("42")))
}

// TestSendMailNotificationMailRoutable covers MailRoutable.RouteNotificationForMail
// against a real SMTP server. Skipped when mail.host is unset (resolved from
// config, which loads .env).
func (s *NotificationTestSuite) TestSendMailNotificationMailRoutable() {
	if facades.Config().GetString("mail.host") == "" {
		s.T().Skip("skipping mail notification test: set MAIL_HOST to run against a real SMTP server")
	}
	mailTo := os.Getenv("MAIL_TO") // MAIL_TO is intentionally env-only (not a mail config key)
	if mailTo == "" {
		mailTo = facades.Config().GetString("mail.from.address")
	}
	s.Require().NotEmpty(mailTo)

	s.NoError(facades.Notification().Send(&mailRoutableNotifiable{to: mailTo}, notifications.NewOrderShipped("42")))
}

// TestCommandMakeNotification covers the make:notification command: stub
// generation, already-exists behavior, nested packages and the --database
// variant.
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
	s.False(file.Contains(databasePath, "ToMail"))
	s.False(file.Contains(databasePath, `"github.com/goravel/framework/notification/mail"`))

	s.NoError(facades.Artisan().Call("--no-ansi make:notification " + nestedPackage + "/" + nestedNotificationName))
	s.True(file.Exists(nestedPath))
	s.True(file.Contains(nestedPath, "package "+nestedPackage))
	s.True(file.Contains(nestedPath, "type "+nestedNotificationName+" struct {"))
}

// TestCommandNotificationsTable covers the notifications:table command: it
// generates a migration file, auto-registers it in bootstrap/migrations.go
// and warns when the migration already exists.
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

func (s *NotificationTestSuite) uniqueName(prefix string) string {
	return fmt.Sprintf("%s%d", prefix, atomic.AddUint64(&s.counter, 1))
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

// ---- Test-local notifications / channels / notifiables ----

// routedQueuedNotification implements ShouldQueue with non-empty
// OnQueue/OnConnection values.
type routedQueuedNotification struct{}

func (r *routedQueuedNotification) Via(notifiable notification.Notifiable) []string {
	return []string{"database"}
}

func (r *routedQueuedNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"order_id": "routed"}
}

func (r *routedQueuedNotification) OnQueue() string      { return "notifications" }
func (r *routedQueuedNotification) OnConnection() string { return "sync" }

// shouldSendNotification implements NotificationWithShouldSend.
type shouldSendNotification struct {
	shouldSend bool
}

func (r *shouldSendNotification) Via(notifiable notification.Notifiable) []string {
	return []string{"database"}
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
	return []string{"database"}
}

func (r *afterSendingNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"message": "after"}
}

func (r *afterSendingNotification) AfterSending(notifiable notification.Notifiable, channel string) error {
	afterSendingCalled = true
	return nil
}

// databaseRoutableNotification implements DatabaseRoutable using the default
// connection (empty name).
type databaseRoutableNotification struct{}

func (r *databaseRoutableNotification) Via(notifiable notification.Notifiable) []string {
	return []string{"database"}
}

func (r *databaseRoutableNotification) ToDatabase(notifiable notification.Notifiable) map[string]any {
	return map[string]any{"message": "routable"}
}

func (r *databaseRoutableNotification) DatabaseConnection() string { return "" }

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
	return []string{"database", r.channel}
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
