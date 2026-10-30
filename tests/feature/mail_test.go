package feature

import (
	"testing"

	"github.com/goravel/framework/contracts/mail"
	frameworkerrors "github.com/goravel/framework/errors"
	"github.com/goravel/framework/notification/channels"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/mails"
	"goravel/app/models"
	"goravel/app/notifications"
	"goravel/tests"
)

// MailTestSuite covers Mail.Fake (framework#1574): every mail a test drives
// through the facade is recorded instead of delivered, so no SMTP server is
// needed.
type MailTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestMailTestSuite(t *testing.T) {
	suite.Run(t, &MailTestSuite{})
}

// SetupTest empties the jobs table so "nothing was dispatched" assertions are
// meaningful, and leaves the mailer un-faked: the Mail binding hands the same
// *mail.Fake to every Application it resolves, so an enabled fake would leak
// into the next test and into other suites of the same process.
func (s *MailTestSuite) SetupTest() {
	s.RefreshDatabase()

	facades.Mail().Fake().Reset()
}

func (s *MailTestSuite) TearDownTest() {
	facades.Mail().Fake().Reset()
}

func (s *MailTestSuite) TestFakeRecordsSentMail() {
	fake := facades.Mail().Fake()
	s.True(fake.AssertNothingOutgoing())

	s.NoError(facades.Mail().
		From(mail.Address{Address: "team@example.com", Name: "Example Team"}).
		To([]string{"Ana <Ana@Example.com>"}).
		Cc([]string{"cc@example.com"}).
		Bcc([]string{"bcc@example.com"}).
		Attach([]string{"invoice.pdf"}).
		Subject("Welcome").
		Content(mail.Content{Html: "<h1>Hi</h1>", Text: "Hi"}).
		Headers(map[string]string{"X-Tag": "welcome"}).
		Send())

	s.Equal([]mail.Message{{
		Subject:     "Welcome",
		Html:        "<h1>Hi</h1>",
		Text:        "Hi",
		From:        mail.Address{Address: "team@example.com", Name: "Example Team"},
		To:          []string{"Ana <Ana@Example.com>"},
		Cc:          []string{"cc@example.com"},
		Bcc:         []string{"bcc@example.com"},
		Attachments: []string{"invoice.pdf"},
		Headers:     map[string]string{"X-Tag": "welcome"},
	}}, fake.Sent())

	// Faking again through a new facade call shares the same records.
	s.Same(fake, facades.Mail().Fake())

	s.True(fake.AssertSentCount(1))
	s.False(fake.AssertSentCount(2))
	s.True(fake.AssertNothingQueued())
	s.False(fake.AssertNothingOutgoing())
	s.True(fake.AssertSent(func(message mail.Message) bool {
		return message.HasTo("ana@example.com") &&
			message.HasCc("cc@example.com") &&
			message.HasBcc("bcc@example.com") &&
			message.HasAttachment("invoice.pdf")
	}))
	s.True(fake.AssertNotSent(func(message mail.Message) bool {
		return message.HasTo("someone-else@example.com")
	}))
}

func (s *MailTestSuite) TestFakeUsesDefaultSender() {
	fake := facades.Mail().Fake()

	s.NoError(facades.Mail().
		To([]string{"ana@example.com"}).
		Subject("Hi").
		Content(mail.Content{Html: "<p>Hi</p>"}).
		Send())

	s.Equal([]mail.Message{{
		Subject: "Hi",
		Html:    "<p>Hi</p>",
		From: mail.Address{
			Address: facades.Config().GetString("mail.from.address"),
			Name:    facades.Config().GetString("mail.from.name"),
		},
		To: []string{"ana@example.com"},
	}}, fake.Sent())
}

func (s *MailTestSuite) TestFakeRendersMailableView() {
	fake := facades.Mail().Fake()
	welcome := mails.NewWelcome("Ana", "ana@example.com")

	s.NoError(facades.Mail().Send(welcome))

	sent := fake.Sent()
	s.Require().Len(sent, 1)
	s.Same(welcome, sent[0].Mailable)
	s.Equal("Welcome to Goravel, Ana", sent[0].Subject)
	s.Equal([]string{"ana@example.com"}, sent[0].To)
	s.Equal(map[string]string{"X-Mailer": "goravel-example"}, sent[0].Headers)
	// The view keeps its surrounding newlines, so only the body is asserted.
	s.Contains(sent[0].Html, "<p>Hi Ana, welcome to Goravel.</p>")
	s.True(fake.AssertSent(func(message mail.Message) bool {
		return message.HasTo("ana@example.com") && message.Subject == "Welcome to Goravel, Ana"
	}))
}

func (s *MailTestSuite) TestFakeRecordsQueuedMail() {
	fake := facades.Mail().Fake()
	invoice := mails.NewInvoice("1001", "ana@example.com")

	s.NoError(facades.Mail().Queue(invoice))

	queued := fake.Queued()
	s.Require().Len(queued, 1)
	s.Same(invoice, queued[0].Mailable)
	s.Equal(&mail.Queue{Connection: "database", Queue: "mails"}, queued[0].Queue)
	s.Equal("<p>Invoice 1001 is ready.</p>", queued[0].Html)

	s.True(fake.AssertQueuedCount(1))
	s.False(fake.AssertQueuedCount(2))
	s.True(fake.AssertNothingSent())
	s.True(fake.AssertNotQueued(func(message mail.Message) bool { return message.Subject == "2002" }))

	// Faking records the mail instead of dispatching the send job.
	count, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), count)
}

func (s *MailTestSuite) TestFakeFilters() {
	fake := facades.Mail().Fake()

	for _, subject := range []string{"a", "b", "a"} {
		s.Require().NoError(facades.Mail().
			To([]string{"ana@example.com"}).
			Subject(subject).
			Content(mail.Content{Html: "<p>" + subject + "</p>"}).
			Send())
	}
	s.Require().NoError(facades.Mail().Queue(mails.NewInvoice("1001", "ana@example.com")))

	isA := func(message mail.Message) bool { return message.Subject == "a" }
	isB := func(message mail.Message) bool { return message.Subject == "b" }

	tests := []struct {
		name   string
		filter []func(message mail.Message) bool
		expect int
	}{
		{name: "no filter returns every sent mail", filter: nil, expect: 3},
		{name: "a nil filter keeps every mail", filter: []func(message mail.Message) bool{nil}, expect: 3},
		{name: "one filter narrows the records", filter: []func(message mail.Message) bool{isA}, expect: 2},
		{name: "every filter must match", filter: []func(message mail.Message) bool{isA, isB}, expect: 0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Len(fake.Sent(tt.filter...), tt.expect)
		})
	}

	s.Empty(fake.Queued(isA))
	s.Len(fake.Queued(func(message mail.Message) bool { return message.HasTo("ana@example.com") }), 1)
	s.True(fake.AssertSentCount(3))
	s.False(fake.AssertQueuedCount(0))
	s.True(fake.AssertNotSent(func(message mail.Message) bool { return message.Subject == "c" }))
}

func (s *MailTestSuite) TestFakeCapturesNotificationMail() {
	fake := facades.Mail().Fake()

	s.NoError(facades.Notification().Send(
		&models.User{Name: "Bowen", Mail: "bowen@example.com"},
		notifications.NewOrderShipped("42"),
	))

	sent := fake.Sent()
	s.Require().Len(sent, 1)
	s.Equal("Order 42 has shipped", sent[0].Subject)
	s.Equal("<p>Order 42 has shipped.</p>", sent[0].Html)
	s.Equal([]string{"Bowen <bowen@example.com>"}, sent[0].To)
	s.IsType(&channels.NotificationMailable{}, sent[0].Mailable)
	s.True(fake.AssertSent(func(message mail.Message) bool {
		// Display names are ignored, so the bare address matches.
		return message.HasTo("bowen@example.com")
	}))
}

func (s *MailTestSuite) TestFakeResetRestoresDelivery() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default":        "database",
		"app.disabled_runners": []string{"app:queue:database", "app:queue:test"},
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	// The restart above rebuilt the app, so the fake is obtained afterwards.
	fake := facades.Mail().Fake()
	s.NoError(facades.Mail().Queue(mails.NewInvoice("1001", "ana@example.com")))
	s.True(fake.AssertQueuedCount(1))

	fake.Reset()

	s.NoError(facades.Mail().Queue(mails.NewInvoice("1001", "ana@example.com")))
	s.True(fake.AssertNothingOutgoing())

	var jobs []jobRow // shared with notification_test.go, same package
	s.NoError(facades.DB().Table("jobs").Get(&jobs))
	s.Require().Len(jobs, 1)
	s.Equal("mails", jobs[0].Queue)
}

func (s *MailTestSuite) TestFakeTemplateErrorStillFails() {
	fake := facades.Mail().Fake()

	err := facades.Mail().Send(&missingViewMailable{})

	s.ErrorIs(err, frameworkerrors.MailTemplateParseFailed)
	s.True(fake.AssertNothingSent())
}

// missingViewMailable references a view that does not exist, so rendering
// fails before anything is recorded.
type missingViewMailable struct{}

func (r *missingViewMailable) Attachments() []string  { return nil }
func (r *missingViewMailable) Content() *mail.Content { return &mail.Content{HtmlView: "missing.tmpl"} }
func (r *missingViewMailable) Envelope() *mail.Envelope {
	return &mail.Envelope{To: []string{"ana@example.com"}, Subject: "Missing"}
}
func (r *missingViewMailable) Headers() map[string]string { return nil }
func (r *missingViewMailable) Queue() *mail.Queue         { return nil }
