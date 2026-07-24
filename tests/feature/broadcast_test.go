package feature

import (
	stdhttp "net/http"
	"testing"

	"github.com/goravel/framework/broadcasting"
	contracts "github.com/goravel/framework/contracts/broadcasting"
	supporthttp "github.com/goravel/framework/support/http"
	"github.com/stretchr/testify/suite"

	appbroadcasting "goravel/app/broadcasting"
	"goravel/app/facades"
	"goravel/tests"
)

type BroadcastTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestBroadcastTestSuite(t *testing.T) {
	suite.Run(t, &BroadcastTestSuite{})
}

func (s *BroadcastTestSuite) SetupTest() {
	config := facades.Config()
	config.Add("broadcasting.connections.log", map[string]any{
		"driver": "log",
	})
	config.Add("broadcasting.connections.null", map[string]any{
		"driver": "null",
	})
	config.Add("broadcasting.default", "null")
}

func (s *BroadcastTestSuite) TestBroadcastDispatch_ShouldBroadcastNow_NullDriver() {
	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1},
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestBroadcastDispatch_ShouldBroadcastNow_LogDriver() {
	facades.Config().Add("broadcasting.default", "log")

	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1, "name": "Test Order"},
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestBroadcastDispatch_ShouldBroadcastWithConnections() {
	facades.Config().Add("broadcasting.default", "log")

	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1},
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestBroadcastDispatch_ShouldBroadcastWithQueue() {
	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1},
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestBroadcastDispatch_BroadcastWhenFalse() {
	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShipped{
		OrderID:    1,
		ShouldFire: false,
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestBroadcastDispatch_NoChannels() {
	err := facades.Broadcast().Dispatch(&appbroadcasting.EmptyEvent{})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestChannelAuth_PublicChannel() {
	body, err := supporthttp.NewBody().
		SetField("socket_id", "1234.5678").
		SetField("channel_name", "public-channel").
		Build()
	s.NoError(err)

	resp, err := s.Http(s.T()).Post("/broadcasting/auth", body.Reader())
	s.NoError(err)
	resp.AssertOk()
}

func (s *BroadcastTestSuite) TestChannelAuth_MissingParams() {
	body, err := supporthttp.NewBody().
		SetField("socket_id", "1234.5678").
		Build()
	s.NoError(err)

	resp, err := s.Http(s.T()).Post("/broadcasting/auth", body.Reader())
	s.NoError(err)
	resp.AssertBadRequest()
}

func (s *BroadcastTestSuite) TestChannelAuth_MissingSocketID() {
	body, err := supporthttp.NewBody().
		SetField("channel_name", "private-orders.1").
		Build()
	s.NoError(err)

	resp, err := s.Http(s.T()).Post("/broadcasting/auth", body.Reader())
	s.NoError(err)
	resp.AssertBadRequest()
}

func (s *BroadcastTestSuite) TestChannelHelpers() {
	public := broadcasting.PublicChannel("my-channel")
	s.Equal("my-channel", public.Name)
	s.False(broadcasting.IsPrivateChannel(public))
	s.False(broadcasting.IsPresenceChannel(public))
	s.Equal("my-channel", broadcasting.ChannelBaseName(public))

	private := broadcasting.PrivateChannel("orders.123")
	s.Equal("private-orders.123", private.Name)
	s.True(broadcasting.IsPrivateChannel(private))
	s.False(broadcasting.IsPresenceChannel(private))
	s.Equal("orders.123", broadcasting.ChannelBaseName(private))

	presence := broadcasting.PresenceChannel("chat")
	s.Equal("presence-chat", presence.Name)
	s.False(broadcasting.IsPrivateChannel(presence))
	s.True(broadcasting.IsPresenceChannel(presence))
	s.Equal("chat", broadcasting.ChannelBaseName(presence))
}

func (s *BroadcastTestSuite) TestConstants() {
	s.Equal("private-", contracts.ChannelPrefixPrivate)
	s.Equal("presence-", contracts.ChannelPrefixPresence)
}

func (s *BroadcastTestSuite) TestAuthResponseType() {
	resp := contracts.AuthResponse{
		Auth:        "test-key:signature",
		ChannelData: `{"user_id":"1"}`,
	}
	s.Equal("test-key:signature", resp.Auth)
	s.Equal(`{"user_id":"1"}`, resp.ChannelData)
}

func (s *BroadcastTestSuite) TestPusherDriverDispatch_RoundTrip() {
	if !s.isRelayCloudRunning() {
		s.T().Skip("RelayCloud not running")
	}

	facades.Config().Add("broadcasting.default", "pusher")
	facades.Config().Add("broadcasting.connections.pusher", map[string]any{
		"driver":  "pusher",
		"key":     "test-key",
		"secret":  "test-secret",
		"app_id":  "test-app",
		"options": map[string]any{
			"host":   "127.0.0.1",
			"port":   6001,
			"scheme": "http",
		},
	})

	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"test": "relaycloud"},
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) isRelayCloudRunning() bool {
	resp, err := stdhttp.Get("http://127.0.0.1:6001")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return true
}
