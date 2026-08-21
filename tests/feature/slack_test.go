package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/notification"
	slackcontracts "github.com/goravel/slack/contracts"

	"goravel/app/facades"
	"goravel/tests"
)

type SlackTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestSlackTestSuite(t *testing.T) {
	suite.Run(t, &SlackTestSuite{})
}

// requireSlack skips the test unless a bot token and a target channel are
// configured, and returns the channel to post to. Mirrors the mail live-test
// pattern: skipped locally, exercised for real in CI when secrets are set.
func (s *SlackTestSuite) requireSlack() string {
	if facades.Config().GetString("slack.token") == "" {
		s.T().Skip("skipping Slack test: set SLACK_BOT_TOKEN to run against a real Slack workspace")
	}
	channel := facades.Config().GetString("slack.channel")
	if channel == "" {
		s.T().Skip("skipping Slack test: set SLACK_CHANNEL to the target channel or user ID")
	}
	return channel
}

func (s *SlackTestSuite) TestSlackChannelRegistered() {
	s.NotNil(facades.Notification().Channel(slackcontracts.ChannelName))
}

func (s *SlackTestSuite) TestSendSlackTextNotification() {
	channel := s.requireSlack()

	err := facades.Notification().
		Route(slackcontracts.ChannelName, channel).
		NotifyNow(&slackTextNotification{})
	s.NoErrorf(err, "sending to channel %q", channel)
}

func (s *SlackTestSuite) TestSendSlackRichNotification() {
	channel := s.requireSlack()

	n := &slackRichNotification{msg: slackcontracts.Message{
		Text: "Invoice paid",
		Attachments: []slackcontracts.Attachment{
			{
				Title: "Details", Text: "Invoice #123", Color: "good",
				Fields: []slackcontracts.Field{
					{Title: "Amount", Value: "$99.00", Short: true},
					{Title: "Customer", Value: "Acme Corp", Short: true},
				},
			},
		},
	}}

	err := facades.Notification().
		Route(slackcontracts.ChannelName, channel).
		NotifyNow(n)
	s.NoErrorf(err, "sending to channel %q", channel)
}

type slackTextNotification struct{}

func (r *slackTextNotification) Via(_ notification.Notifiable) []string {
	return []string{slackcontracts.ChannelName}
}

func (r *slackTextNotification) ToSlack(_ notification.Notifiable) slackcontracts.Message {
	return slackcontracts.Message{Text: "Goravel Slack integration test — text delivery"}
}

type slackRichNotification struct {
	msg slackcontracts.Message
}

func (r *slackRichNotification) Via(_ notification.Notifiable) []string {
	return []string{slackcontracts.ChannelName}
}

func (r *slackRichNotification) ToSlack(_ notification.Notifiable) slackcontracts.Message {
	return r.msg
}
