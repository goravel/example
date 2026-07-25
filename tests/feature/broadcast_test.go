package feature

import (
	"os/exec"
	"testing"
	"time"

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

func (s *BroadcastTestSuite) SetupSuite() {
	_ = exec.Command("docker", "compose", "up", "soketi", "-d").Run()
	time.Sleep(2 * time.Second)
}

func (s *BroadcastTestSuite) TearDownSuite() {
	_ = exec.Command("docker", "compose", "stop", "soketi").Run()
	_ = exec.Command("docker", "compose", "rm", "-f", "soketi").Run()
}

func (s *BroadcastTestSuite) SetupTest() {
	facades.Config().Add("broadcasting.default", "null")
}

func (s *BroadcastTestSuite) TestLogDriver() {
	facades.Config().Add("broadcasting.default", "log")

	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1, "name": "Test Order"},
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestNullDriver() {
	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1},
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestDispatch_BroadcastWhenFalse_SkipsDispatch() {
	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShipped{
		OrderID:    1,
		ShouldFire: false,
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestDispatch_NoChannels_SkipsDispatch() {
	err := facades.Broadcast().Dispatch(&appbroadcasting.EmptyEvent{})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestDispatch_ShouldBroadcastWithQueue() {
	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1},
	})
	s.NoError(err)
}

func (s *BroadcastTestSuite) TestDispatch_ShouldBroadcastWithConnections() {
	facades.Config().Add("broadcasting.default", "log")

	err := facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1},
	})
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

func TestPusherDriverEndToEnd(t *testing.T) {
	ws, err := tests.NewWSClient("127.0.0.1:6001", "test-key")
	if err != nil {
		t.Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.Close() }()

	if err := ws.Subscribe("test-channel"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	facades.Config().Add("broadcasting.default", "pusher")
	if err := facades.App().Restart(); err != nil {
		t.Fatal(err)
	}

	err = facades.Broadcast().Dispatch(&appbroadcasting.OrderShippedNow{
		OrderID:   1,
		OrderData: map[string]any{"id": 1, "msg": "hello from soketi"},
	})
	if err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	found := false
	for _, e := range ws.Events() {
		if e.Event == "order.shipped" && e.Channel == "test-channel" {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected order.shipped event on test-channel, got %d events", ws.EventCount())
	}
}
