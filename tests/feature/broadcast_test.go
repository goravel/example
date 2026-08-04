package feature

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	frameworkmodels "github.com/goravel/framework/queue/models"
	supporthttp "github.com/goravel/framework/support/http"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"

	"goravel/app/events"
	"goravel/app/facades"
	"goravel/app/models"
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
	if err := exec.Command("docker", "compose", "up", "soketi", "-d", "--wait").Run(); err != nil {
		s.T().Fatalf("failed to start soketi: %v", err)
	}
}

func (s *BroadcastTestSuite) SetupTest() {
	s.RefreshDatabase()
}

func (s *BroadcastTestSuite) TearDownSuite() {
	if err := exec.Command("docker", "compose", "stop", "soketi").Run(); err != nil {
		s.T().Logf("failed to stop soketi: %v", err)
	}
	if err := exec.Command("docker", "compose", "rm", "-f", "soketi").Run(); err != nil {
		s.T().Logf("failed to remove soketi container: %v", err)
	}
}

func (s *BroadcastTestSuite) TestDispatchWithPusher() {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	s.NoError(ws.subscribePublic("orders"))
	time.Sleep(300 * time.Millisecond)

	err = facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelType:        "public",
		ChannelName:        "orders",
		ShouldFire:         true,
		ShouldBroadcastNow: true,
		OrderData:          map[string]any{"id": 1, "name": "Test Order"},
	})
	s.NoError(err)
	time.Sleep(1 * time.Second)

	found := false
	expectedData := `{"order":{"id":1,"name":"Test Order"}}`
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "orders" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on orders channel")
}

func (s *BroadcastTestSuite) TestDispatch_BroadcastWhenFalse_SkipsDispatch() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default": "database",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	jwtToken := s.jwtLogin("broadcast-skip-when")

	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	auth, _ := s.authChannel(ws.socketID, "private-orders.1", jwtToken)
	s.NoError(ws.Subscribe("private-orders.1", auth, ""))
	time.Sleep(300 * time.Millisecond)

	s.NoError(facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelName: "orders.1",
		ShouldFire:  false,
		Conns:       []string{"pusher"},
	}))
	time.Sleep(1 * time.Second)

	count, err := facades.DB().Table("jobs").Where("queue", "default").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "no job should be queued when BroadcastWhen is false")

	found := false
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "private-orders.1" {
			found = true
		}
	}
	s.False(found, "no broadcast should fire when BroadcastWhen is false")
}

func (s *BroadcastTestSuite) TestDispatch_NoChannels_SkipsDispatch() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default": "database",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	s.NoError(facades.Broadcast().Dispatch(context.Background(), &events.EmptyBroadcastEvent{}))
	time.Sleep(500 * time.Millisecond)

	count, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "no job should be queued when BroadcastOn returns no channels")
}

func (s *BroadcastTestSuite) TestChannelAuth_PublicChannel() {
	jwtToken := s.jwtLogin("broadcast-public-channel")

	body, err := supporthttp.NewBody().
		SetField("socket_id", "1234.5678").
		SetField("channel_name", "public-channel").
		Build()
	s.NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", jwtToken).
		Post("/broadcasting/auth", body.Reader())
	s.NoError(err)
	resp.AssertOk()
}

func (s *BroadcastTestSuite) TestChannelAuth_MissingParams() {
	jwtToken := s.jwtLogin("broadcast-missing-params")

	body, err := supporthttp.NewBody().
		SetField("socket_id", "1234.5678").
		Build()
	s.NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", jwtToken).
		Post("/broadcasting/auth", body.Reader())
	s.NoError(err)
	resp.AssertBadRequest()
}

func (s *BroadcastTestSuite) TestChannelAuth_MissingSocketID() {
	jwtToken := s.jwtLogin("broadcast-missing-sid")

	body, err := supporthttp.NewBody().
		SetField("channel_name", "private-orders.1").
		Build()
	s.NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", jwtToken).
		Post("/broadcasting/auth", body.Reader())
	s.NoError(err)
	resp.AssertBadRequest()
}

func (s *BroadcastTestSuite) TestChannelAuth_UsersChannel_Authorized() {
	jwtToken := s.jwtLogin("broadcast-users-test")

	var user models.User
	s.Require().NoError(facades.Orm().Query().Where("name", "broadcast-users-test").First(&user))

	socketID := "1234.5678"
	channelName := fmt.Sprintf("private-users.%d", user.ID)

	body, err := supporthttp.NewBody().
		SetField("socket_id", socketID).
		SetField("channel_name", channelName).
		Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", jwtToken).
		Post("/broadcasting/auth", body.Reader())
	s.Require().NoError(err)
	resp.AssertOk()

	var ar struct {
		Auth string `json:"auth"`
	}
	s.NoError(resp.Bind(&ar))
	s.NotEmpty(ar.Auth)
}

func (s *BroadcastTestSuite) TestChannelAuth_UsersChannel_Unauthorized() {
	jwtToken := s.jwtLogin("broadcast-users-unauthorized")

	var user models.User
	s.Require().NoError(facades.Orm().Query().Where("name", "broadcast-users-unauthorized").First(&user))

	socketID := "1234.5678"
	channelName := fmt.Sprintf("private-users.%d", user.ID+1)

	body, err := supporthttp.NewBody().
		SetField("socket_id", socketID).
		SetField("channel_name", channelName).
		Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", jwtToken).
		Post("/broadcasting/auth", body.Reader())
	s.Require().NoError(err)
	resp.AssertForbidden()
}

func (s *BroadcastTestSuite) TestChannelAuth_UsersChannel_ReturnsUser() {
	jwtToken := s.jwtLogin("broadcast-users-presence")

	var user models.User
	s.Require().NoError(facades.Orm().Query().Where("name", "broadcast-users-presence").First(&user))

	socketID := "1234.5678"
	channelName := fmt.Sprintf("presence-users.%d", user.ID)

	body, err := supporthttp.NewBody().
		SetField("socket_id", socketID).
		SetField("channel_name", channelName).
		Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", jwtToken).
		Post("/broadcasting/auth", body.Reader())
	s.Require().NoError(err)
	resp.AssertOk()

	var ar struct {
		Auth        string `json:"auth"`
		ChannelData string `json:"channel_data"`
	}
	s.NoError(resp.Bind(&ar))
	s.NotEmpty(ar.Auth)
	s.NotEmpty(ar.ChannelData)

	var cd struct {
		UserID   string         `json:"user_id"`
		UserInfo map[string]any `json:"user_info"`
	}
	s.NoError(json.Unmarshal([]byte(ar.ChannelData), &cd))
	s.Equal(user.Name, cd.UserInfo["Name"])
	s.NotZero(cd.UserInfo["id"])
}

type wsClient struct {
	conn      *websocket.Conn
	socketID  string
	collected []wsEvent
	mu        sync.Mutex
}

type wsEvent struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Data    string `json:"data"`
}

func newWSClient(host, appKey string) (*wsClient, error) {
	urlStr := fmt.Sprintf("ws://%s/app/%s?protocol=7&client=js&version=7.0.0", host, appKey)
	conn, _, err := websocket.DefaultDialer.Dial(urlStr, nil)
	if err != nil {
		return nil, err
	}

	c := &wsClient{conn: conn}
	go c.readLoop()
	time.Sleep(300 * time.Millisecond)
	return c, nil
}

func (c *wsClient) Subscribe(channel, auth, channelData string) error {
	return c.sendSubscribe(channel, auth, channelData)
}

func (c *wsClient) subscribePublic(channel string) error {
	return c.sendSubscribe(channel, "", "")
}

func (c *wsClient) subscribePrivate(authEndpoint, channel string, headers map[string]string) error {
	c.mu.Lock()
	socketID := c.socketID
	c.mu.Unlock()

	if socketID == "" {
		return fmt.Errorf("socket_id not available yet")
	}

	auth, channelData, err := c.fetchAuth(authEndpoint, socketID, channel, headers)
	if err != nil {
		return err
	}

	return c.sendSubscribe(channel, auth, channelData)
}

func (c *wsClient) events() []wsEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]wsEvent, len(c.collected))
	copy(result, c.collected)
	return result
}

func (c *wsClient) close() error {
	return c.conn.Close()
}

func (c *wsClient) sendSubscribe(channel, auth, channelData string) error {
	data := map[string]string{"channel": channel}
	if auth != "" {
		data["auth"] = auth
	}
	if channelData != "" {
		data["channel_data"] = channelData
	}
	msg, _ := json.Marshal(map[string]any{
		"event": "pusher:subscribe",
		"data":  data,
	})
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *wsClient) fetchAuth(endpoint, socketID, channelName string, headers map[string]string) (string, string, error) {
	form := url.Values{}
	form.Set("socket_id", socketID)
	form.Set("channel_name", channelName)

	req, err := stdhttp.NewRequest("POST", endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := stdhttp.DefaultClient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	if resp.StatusCode != stdhttp.StatusOK {
		return "", "", fmt.Errorf("auth endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var ar struct {
		Auth        string `json:"auth"`
		ChannelData string `json:"channel_data"`
	}
	if err := json.Unmarshal(body, &ar); err != nil {
		return "", "", fmt.Errorf("auth endpoint returned non-JSON: %s", string(body))
	}

	if ar.Auth == "" {
		return "", "", fmt.Errorf("auth endpoint returned empty auth")
	}

	return ar.Auth, ar.ChannelData, nil
}

func (c *wsClient) readLoop() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var raw map[string]any
		if err := json.Unmarshal(msg, &raw); err != nil {
			continue
		}

		eventName, _ := raw["event"].(string)

		if eventName == "pusher:connection_established" {
			sid := extractSocketID(raw)
			if sid != "" {
				c.mu.Lock()
				c.socketID = sid
				c.mu.Unlock()
			}
			continue
		}

		channel, _ := raw["channel"].(string)
		data, _ := raw["data"].(string)

		c.mu.Lock()
		c.collected = append(c.collected, wsEvent{
			Event:   eventName,
			Channel: channel,
			Data:    data,
		})
		c.mu.Unlock()
	}
}

func extractSocketID(raw map[string]any) string {
	if data, ok := raw["data"].(map[string]any); ok {
		sid, _ := data["socket_id"].(string)
		return sid
	}
	if str, ok := raw["data"].(string); ok {
		var parsed map[string]any
		if json.Unmarshal([]byte(str), &parsed) == nil {
			sid, _ := parsed["socket_id"].(string)
			return sid
		}
	}
	return ""
}

const (
	soketiHost    = "127.0.0.1:6001"
	soketiAppID   = "test-app"
	soketiAppKey  = "test-key"
	soketiSecret  = "test-secret"
	soketiAuthURL = "http://127.0.0.1:8080/broadcasting/auth"
)

func (s *BroadcastTestSuite) TestWSClientConnect() {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	ws.mu.Lock()
	sid := ws.socketID
	ws.mu.Unlock()

	s.NotEmpty(sid, "expected socket_id from connection_established event")
}

func (s *BroadcastTestSuite) TestWSClientPublishAndReceive() {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	s.NoError(ws.subscribePublic("test-channel"))
	time.Sleep(300 * time.Millisecond)

	s.NoError(soketiTrigger("test-channel", "order.shipped", map[string]any{"order_id": 1, "msg": "hello"}))
	time.Sleep(1 * time.Second)

	found := false
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "test-channel" {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on test-channel")
}

func (s *BroadcastTestSuite) TestWSClientPublishToMultipleChannels() {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	s.NoError(ws.subscribePublic("channel-a"))
	s.NoError(ws.subscribePublic("channel-b"))
	time.Sleep(300 * time.Millisecond)

	s.NoError(soketiTriggerMulti([]string{"channel-a", "channel-b"}, "multi-event", map[string]any{"broadcast": true}))
	time.Sleep(1 * time.Second)

	channels := map[string]bool{}
	for _, e := range ws.events() {
		channels[e.Channel] = true
	}

	s.True(channels["channel-a"], "expected event on channel-a")
	s.True(channels["channel-b"], "expected event on channel-b")
}

func (s *BroadcastTestSuite) TestWSClientPrivateChannelAuthRejected() {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	err = ws.subscribePrivate(soketiAuthURL, "private-orders.999", nil)
	s.Error(err, "expected auth error for unauthorized channel")
}

func (s *BroadcastTestSuite) jwtLogin(name string) string {
	body, err := supporthttp.NewBody().SetField("name", name).Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).Post("/jwt/login", body.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	authHeader := resp.Headers().Get("Authorization")
	s.Require().NotEmpty(authHeader)
	return authHeader
}

func (s *BroadcastTestSuite) authChannel(socketID, channel, jwtToken string) (string, string) {
	body, err := supporthttp.NewBody().
		SetField("socket_id", socketID).
		SetField("channel_name", channel).
		Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Authorization", jwtToken).
		Post("/broadcasting/auth", body.Reader())
	s.Require().NoError(err)
	resp.AssertOk()

	var ar struct {
		Auth        string `json:"auth"`
		ChannelData string `json:"channel_data"`
	}
	s.NoError(resp.Bind(&ar))
	s.NotEmpty(ar.Auth)
	return ar.Auth, ar.ChannelData
}

func (s *BroadcastTestSuite) TestPublicChannelFullFlow() {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	s.NoError(ws.subscribePublic("orders"))
	time.Sleep(300 * time.Millisecond)

	err = facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelType:        "public",
		ChannelName:        "orders",
		ShouldFire:         true,
		ShouldBroadcastNow: true,
		OrderData: map[string]any{
			"id":   1,
			"name": "Test Order",
		},
	})
	s.NoError(err)
	time.Sleep(1 * time.Second)

	found := false
	expectedData := `{"order":{"id":1,"name":"Test Order"}}`
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "orders" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on orders channel")
}

func (s *BroadcastTestSuite) TestPublicChannelFullFlow_CustomChannelName() {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	s.NoError(ws.subscribePublic("public-updates"))
	time.Sleep(300 * time.Millisecond)

	err = facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelType:        "public",
		ChannelName:        "public-updates",
		ShouldFire:         true,
		ShouldBroadcastNow: true,
		OrderData: map[string]any{
			"item":  "Laptop",
			"price": 1,
		},
	})
	s.NoError(err)
	time.Sleep(1 * time.Second)

	found := false
	expectedData := `{"order":{"item":"Laptop","price":1}}`
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "public-updates" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on public-updates channel")
}

func (s *BroadcastTestSuite) TestPrivateChannelFullFlow() {
	jwtToken := s.jwtLogin("broadcast-private-test")

	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	auth, _ := s.authChannel(ws.socketID, "private-orders.1", jwtToken)
	s.NoError(ws.Subscribe("private-orders.1", auth, ""))
	time.Sleep(300 * time.Millisecond)

	err = facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelName: "orders.1",
		ShouldFire:  true,
		OrderData: map[string]any{
			"id":   1,
			"name": "Private Order",
		},
	})
	s.NoError(err)
	time.Sleep(1 * time.Second)

	found := false
	expectedData := `{"order":{"id":1,"name":"Private Order"}}`
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "private-orders.1" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on private-orders.1 channel")
}

func (s *BroadcastTestSuite) TestPresenceChannelFullFlow() {
	jwtTokenA := s.jwtLogin("broadcast-presence-a")
	jwtTokenB := s.jwtLogin("broadcast-presence-b")

	wsA, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = wsA.close() }()

	auth, channelData := s.authChannel(wsA.socketID, "presence-team.1", jwtTokenA)
	s.NoError(wsA.Subscribe("presence-team.1", auth, channelData))
	time.Sleep(300 * time.Millisecond)

	wsB, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}

	authB, channelDataB := s.authChannel(wsB.socketID, "presence-team.1", jwtTokenB)
	s.NoError(wsB.Subscribe("presence-team.1", authB, channelDataB))
	time.Sleep(500 * time.Millisecond)

	s.NoError(wsB.close())
	time.Sleep(500 * time.Millisecond)

	subscribed := false
	memberAdded := false
	memberRemoved := false
	for _, e := range wsA.events() {
		switch {
		case e.Event == "pusher_internal:subscription_succeeded" && e.Channel == "presence-team.1":
			subscribed = true
		case e.Event == "pusher_internal:member_added" && e.Channel == "presence-team.1":
			memberAdded = true
		case e.Event == "pusher_internal:member_removed" && e.Channel == "presence-team.1":
			memberRemoved = true
		}
	}
	s.True(subscribed, "expected subscription_succeeded on presence-team.1 channel")
	s.True(memberAdded, "expected member_added on presence-team.1 channel")
	s.True(memberRemoved, "expected member_removed on presence-team.1 channel")

	err = facades.Broadcast().Dispatch(context.Background(), &events.TeamPresenceBroadcast{
		ChannelName: "team.1",
		TeamData: map[string]any{
			"id":   1,
			"name": "Test Team",
		},
	})
	s.NoError(err)
	time.Sleep(1 * time.Second)

	found := false
	expectedData := `{"team":{"id":1,"name":"Test Team"}}`
	for _, e := range wsA.events() {
		if e.Event == "team.presence" && e.Channel == "presence-team.1" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected team.presence event on presence-team.1 channel")
}

func (s *BroadcastTestSuite) TestDispatch_WithQueue() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default": "database",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	jwtToken := s.jwtLogin("broadcast-queue-test")

	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	auth, _ := s.authChannel(ws.socketID, "private-orders.1", jwtToken)
	s.NoError(ws.Subscribe("private-orders.1", auth, ""))
	time.Sleep(300 * time.Millisecond)

	const expectedData = `{"order":{"id":1,"name":"Queued Order"}}`

	s.NoError(facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelName: "orders.1",
		ShouldFire:  true,
		QueueName:   "custom-broadcast-queue",
		Conns:       []string{"pusher"},
		OrderData: map[string]any{
			"id":   1,
			"name": "Queued Order",
		},
	}))

	var jobs []frameworkmodels.Job
	s.NoError(facades.DB().Table("jobs").Where("queue", "custom-broadcast-queue").Get(&jobs))
	s.Require().Len(jobs, 1)
	s.Equal("custom-broadcast-queue", jobs[0].Queue)

	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "custom-broadcast-queue",
		Concurrent: 1,
		Tries:      1,
	})
	go func() { _ = worker.Run() }()

	time.Sleep(2 * time.Second)
	_ = worker.Shutdown()

	count, err := facades.DB().Table("jobs").Where("queue", "custom-broadcast-queue").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "queued broadcast job should be consumed")

	found := false
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "private-orders.1" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on private-orders.1 after queued job consumed")
}

func (s *BroadcastTestSuite) TestDispatch_WithQueueConnection() {
	jwtToken := s.jwtLogin("broadcast-queue-conn-test")

	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	auth, _ := s.authChannel(ws.socketID, "private-orders.1", jwtToken)
	s.NoError(ws.Subscribe("private-orders.1", auth, ""))
	time.Sleep(300 * time.Millisecond)

	const expectedData = `{"order":{"id":1,"name":"Queued Conn Order"}}`

	s.NoError(facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelName: "orders.1",
		ShouldFire:  true,
		QueueConn:   "database",
		Conns:       []string{"pusher"},
		OrderData: map[string]any{
			"id":   1,
			"name": "Queued Conn Order",
		},
	}))

	var jobs []frameworkmodels.Job
	s.NoError(facades.DB().Table("jobs").Where("queue", "default").Get(&jobs))
	s.Require().Len(jobs, 1)

	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "default",
		Concurrent: 1,
		Tries:      1,
	})
	go func() { _ = worker.Run() }()

	time.Sleep(2 * time.Second)
	_ = worker.Shutdown()

	count, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "queued broadcast job should be consumed")

	found := false
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "private-orders.1" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on private-orders.1 after queued job consumed")
}

func (s *BroadcastTestSuite) TestDispatch_WithDelay() {
	scope, err := tests.OverrideConfig(map[string]any{
		"queue.default": "database",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	jwtToken := s.jwtLogin("broadcast-delay-test")

	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	auth, _ := s.authChannel(ws.socketID, "private-orders.1", jwtToken)
	s.NoError(ws.Subscribe("private-orders.1", auth, ""))
	time.Sleep(300 * time.Millisecond)

	const (
		delay        = 3 * time.Second
		expectedData = `{"order":{"id":1,"name":"Delayed Order"}}`
	)

	s.NoError(facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelName: "orders.1",
		ShouldFire:  true,
		DelayedAt:   time.Now().UTC().Add(delay),
		QueueName:   "custom-delay-queue",
		Conns:       []string{"pusher"},
		OrderData: map[string]any{
			"id":   1,
			"name": "Delayed Order",
		},
	}))

	var jobs []frameworkmodels.Job
	s.NoError(facades.DB().Table("jobs").Where("queue", "custom-delay-queue").Get(&jobs))
	s.Require().Len(jobs, 1)
	s.Require().NotNil(jobs[0].AvailableAt)
	s.Greater(jobs[0].AvailableAt.StdTime(), time.Now())

	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "custom-delay-queue",
		Concurrent: 1,
		Tries:      1,
	})
	go func() { _ = worker.Run() }()

	// Before the delay elapses, the worker must not consume the job nor fire the broadcast.
	time.Sleep(1 * time.Second)

	count, err := facades.DB().Table("jobs").Where("queue", "custom-delay-queue").Count()
	s.NoError(err)
	s.Equal(int64(1), count, "delayed job should not be consumed before delay expires")

	foundEarly := false
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "private-orders.1" && e.Data == expectedData {
			foundEarly = true
		}
	}
	s.False(foundEarly, "delayed broadcast should not fire before delay expires")

	// After the delay elapses, the worker consumes the job and the broadcast is delivered.
	time.Sleep(5 * time.Second)
	_ = worker.Shutdown()

	count, err = facades.DB().Table("jobs").Where("queue", "custom-delay-queue").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "delayed job should be consumed after delay expires")

	found := false
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "private-orders.1" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on private-orders.1 after delayed job consumed")
}

func (s *BroadcastTestSuite) TestDispatch_WithConnections() {
	jwtToken := s.jwtLogin("broadcast-connections-test")

	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		s.T().Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	auth, _ := s.authChannel(ws.socketID, "private-orders.1", jwtToken)
	s.NoError(ws.Subscribe("private-orders.1", auth, ""))
	time.Sleep(300 * time.Millisecond)

	err = facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelName: "orders.1",
		ShouldFire:  true,
		Conns:       []string{"pusher", "null", "log"},
		OrderData: map[string]any{
			"id":   1,
			"name": "Connections Order",
		},
	})
	s.NoError(err)
	time.Sleep(1 * time.Second)

	found := false
	expectedData := `{"order":{"id":1,"name":"Connections Order"}}`
	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "private-orders.1" && e.Data == expectedData {
			found = true
		}
	}
	s.True(found, "expected order.shipped event on private-orders.1")

	matches, err := filepath.Glob("../../storage/logs/*.log")
	s.Require().NoError(err)
	s.Require().NotEmpty(matches, "expected at least one log file under storage/logs/")

	var logContent []byte
	for _, m := range matches {
		if b, rErr := os.ReadFile(m); rErr == nil {
			logContent = append(logContent, b...)
		}
	}
	logText := string(logContent)
	s.Contains(logText, "Broadcasting event", "expected log driver to emit 'Broadcasting event' message")
	s.Contains(logText, "order.shipped", "expected log driver to record the event name")
	s.Contains(logText, "private-orders.1", "expected log driver to record the channel name")
	s.Contains(logText, "Connections Order", "expected log driver to record the broadcast payload")
}

func (s *BroadcastTestSuite) TestDispatch_WithTriesAndBackoff() {
	scope, err := tests.OverrideConfig(map[string]any{"queue.default": "database"})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	s.NoError(facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelType: "public",
		ChannelName: "retry-orders",
		ShouldFire:  true,
		// "broken" is intentionally NOT in config/broadcasting.go: broadcastToConns
		// iterates conns in order, so the "log" driver emits its per-attempt line
		// first, then cfg.Connection("broken") returns connection-not-found and the
		// attempt fails. "log" must stay first; adding a "broken" connection to the
		// config would silently change this test's behavior.
		Conns:     []string{"log", "broken"},
		Tries:     3,
		Backoff:   []time.Duration{1 * time.Second, 2 * time.Second},
		OrderData: map[string]any{"id": 1, "name": "Retryable Broadcast"},
	}))

	// Dispatch-time capture: queued payload carries tries/backoff (ms).
	var jobs []frameworkmodels.Job
	s.NoError(facades.DB().Table("jobs").Where("queue", "default").Get(&jobs))
	s.Require().Len(jobs, 1)
	var payload struct {
		Args []struct {
			Value string `json:"value"`
		} `json:"args"`
	}
	s.NoError(json.Unmarshal([]byte(jobs[0].Payload), &payload))
	s.Require().Len(payload.Args, 1)
	var item struct {
		Tries   int     `json:"tries"`
		Backoff []int64 `json:"backoff"`
	}
	s.NoError(json.Unmarshal([]byte(payload.Args[0].Value), &item))
	s.Equal(3, item.Tries)
	s.Equal([]int64{1000, 2000}, item.Backoff)

	before := s.countBroadcastEvents("Retryable Broadcast")
	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "default",
		Concurrent: 1,
		Tries:      1, // event policy must override the worker's tries
	})
	workerErr := make(chan error, 1)
	start := time.Now()
	go func() { workerErr <- worker.Run() }()

	// The queued broadcast always FAILS (the "broken" connection), so the
	// completion signal is the failed job landing in the failer. Poll with a
	// bounded timeout instead of a fixed sleep so a slow worker can't strand
	// the job or mask a startup failure.
	elapsed, ok := s.waitForFailedBroadcast(start, 10*time.Second)
	_ = worker.Shutdown()
	s.Require().True(ok, "expected goravel_broadcast job to fail within 10s")
	s.NoError(<-workerErr, "worker should start and stop cleanly")

	// The worker releases the job back to the queue between retries
	// (Release(delay)), so jobs==0 alone is not a timing signal — keep it as a
	// separate check.
	count, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "job should be consumed")

	// Without the 1s+2s release delays this would be ~0 and FAIL; the upper
	// bound catches pathological slowness.
	s.GreaterOrEqual(elapsed, 3*time.Second, "1s + 2s backoff should have elapsed before the final attempt")
	s.Less(elapsed, 8*time.Second, "retries should complete within 8s")
	s.Equal(before+3, s.countBroadcastEvents("Retryable Broadcast"),
		"event BroadcastTries=3 should override worker Tries=1")
}

func (s *BroadcastTestSuite) TestDispatch_WithoutTries_SingleShot() {
	scope, err := tests.OverrideConfig(map[string]any{"queue.default": "database"})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	s.NoError(facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelType: "public",
		ChannelName: "single-shot-orders",
		ShouldFire:  true,
		// Same phantom-connection mechanism as TestDispatch_WithTriesAndBackoff:
		// "log" emits its line first, then the unconfigured "broken" connection
		// fails the attempt (broadcastToConns short-circuits on the first error).
		Conns:     []string{"log", "broken"},
		OrderData: map[string]any{"id": 2, "name": "Single Shot Broadcast"},
	}))

	var jobs []frameworkmodels.Job
	s.NoError(facades.DB().Table("jobs").Where("queue", "default").Get(&jobs))
	s.Require().Len(jobs, 1)
	s.NotContains(jobs[0].Payload, `"tries"`) // omitempty: no retry policy serialized
	s.NotContains(jobs[0].Payload, `"backoff"`)

	before := s.countBroadcastEvents("Single Shot Broadcast")
	worker := facades.Queue().Worker(contractsqueue.Args{
		Connection: "database",
		Queue:      "default",
		Concurrent: 1,
		Tries:      3, // worker would retry, but the event stays single-shot
	})
	workerErr := make(chan error, 1)
	start := time.Now()
	go func() { workerErr <- worker.Run() }()

	// Same completion signal as the retry test: the failed job landing in the
	// failer, polled with a bounded timeout instead of a fixed sleep.
	_, ok := s.waitForFailedBroadcast(start, 5*time.Second)
	_ = worker.Shutdown()
	s.Require().True(ok, "expected goravel_broadcast job to fail within 5s")
	s.NoError(<-workerErr, "worker should start and stop cleanly")

	count, err := facades.DB().Table("jobs").Count()
	s.NoError(err)
	s.Equal(int64(0), count)
	s.Equal(before+1, s.countBroadcastEvents("Single Shot Broadcast"),
		"broadcast without BroadcastTries must fail after a single attempt despite worker Tries=3")
}

func (s *BroadcastTestSuite) TestDispatch_BackoffWithoutTries_Suppressed() {
	scope, err := tests.OverrideConfig(map[string]any{"queue.default": "database"})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	s.NoError(facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelType: "public",
		ChannelName: "backoff-only-orders",
		ShouldFire:  true,
		Conns:       []string{"null"},
		Backoff:     []time.Duration{time.Second}, // no BroadcastTries → ignored
		OrderData:   map[string]any{"id": 3},
	}))

	var jobs []frameworkmodels.Job
	s.NoError(facades.DB().Table("jobs").Where("queue", "default").Get(&jobs))
	s.Require().Len(jobs, 1)
	s.NotContains(jobs[0].Payload, `"backoff"`, "backoff must not serialize without tries")
}

func (s *BroadcastTestSuite) TestDispatchViaHTTP_Public_WithQueueOptions() {
	scope, err := tests.OverrideConfig(map[string]any{"queue.default": "database"})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	body, err := supporthttp.NewBody().
		SetField("channel_type", "public").
		SetField("channel", "public-http-orders").
		SetField("order_data", map[string]any{"id": 1, "name": "Public HTTP Broadcast"}).
		SetField("should_fire", true).
		SetField("queue_name", "public-http-queue").
		SetField("delay", 1).
		SetField("tries", 3).
		SetField("backoff", []float64{1, 2}).
		Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Content-Type", body.ContentType()).
		Post("/broadcasting/dispatch", body.Reader())
	s.NoError(err)
	resp.AssertOk()

	// dispatchPublic must forward queue options: the job lands in the named
	// queue, the payload serializes tries/backoff (ms), and the delay is
	// converted seconds → available_at (future).
	var jobs []frameworkmodels.Job
	s.NoError(facades.DB().Table("jobs").Where("queue", "public-http-queue").Get(&jobs))
	s.Require().Len(jobs, 1)
	s.Equal("public-http-queue", jobs[0].Queue)
	s.Require().NotNil(jobs[0].AvailableAt)
	s.Greater(jobs[0].AvailableAt.StdTime(), time.Now())
	var payload struct {
		Args []struct {
			Value string `json:"value"`
		} `json:"args"`
	}
	s.NoError(json.Unmarshal([]byte(jobs[0].Payload), &payload))
	s.Require().Len(payload.Args, 1)
	var item struct {
		Tries   int     `json:"tries"`
		Backoff []int64 `json:"backoff"`
	}
	s.NoError(json.Unmarshal([]byte(payload.Args[0].Value), &item))
	s.Equal(3, item.Tries)
	s.Equal([]int64{1000, 2000}, item.Backoff) // backoff 1s/2s → ms
}

func (s *BroadcastTestSuite) TestDispatchViaHTTP_Public_ShouldFireFalse_SkipsDispatch() {
	scope, err := tests.OverrideConfig(map[string]any{"queue.default": "database"})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	body, err := supporthttp.NewBody().
		SetField("channel_type", "public").
		SetField("channel", "public-http-skip").
		SetField("order_data", map[string]any{"id": 2, "name": "Public HTTP Skip"}).
		SetField("should_fire", false).
		SetField("queue_name", "public-http-skip-queue").
		SetField("tries", 3).
		Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Content-Type", body.ContentType()).
		Post("/broadcasting/dispatch", body.Reader())
	s.NoError(err)
	resp.AssertOk()

	// dispatchPublic must honor should_fire (not hardcode true): with
	// should_fire=false nothing is queued, even though queue options are set.
	count, err := facades.DB().Table("jobs").Where("queue", "public-http-skip-queue").Count()
	s.NoError(err)
	s.Equal(int64(0), count, "no job should be queued when should_fire is false")
}

func (s *BroadcastTestSuite) TestDispatchViaHTTP_Presence_WithQueueOptions() {
	scope, err := tests.OverrideConfig(map[string]any{"queue.default": "database"})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	body, err := supporthttp.NewBody().
		SetField("channel_type", "presence").
		SetField("channel", "presence-team.http").
		SetField("team_data", map[string]any{"id": 1, "name": "Presence HTTP Broadcast"}).
		SetField("queue_name", "presence-http-queue").
		SetField("tries", 3).
		Build()
	s.Require().NoError(err)

	resp, err := s.Http(s.T()).
		WithHeader("Content-Type", body.ContentType()).
		Post("/broadcasting/dispatch", body.Reader())
	s.NoError(err)
	resp.AssertOk()

	// dispatchPresence must forward queue options: the job lands in the named
	// queue and the payload serializes the retry policy.
	var jobs []frameworkmodels.Job
	s.NoError(facades.DB().Table("jobs").Where("queue", "presence-http-queue").Get(&jobs))
	s.Require().Len(jobs, 1)
	s.Equal("presence-http-queue", jobs[0].Queue)
	var payload struct {
		Args []struct {
			Value string `json:"value"`
		} `json:"args"`
	}
	s.NoError(json.Unmarshal([]byte(jobs[0].Payload), &payload))
	s.Require().Len(payload.Args, 1)
	var item struct {
		Tries int `json:"tries"`
	}
	s.NoError(json.Unmarshal([]byte(payload.Args[0].Value), &item))
	s.Equal(3, item.Tries)
}

// countBroadcastEvents returns the number of "Broadcasting event" log entries
// that carry the given OrderData marker. Each entry corresponds to exactly one
// broadcast attempt by the log driver. Counting only entries that contain the
// marker keeps the count immune to (a) SQL log lines that embed the same
// marker via the queued/failed payload, and (b) accumulation across test runs
// when combined with a before/after delta.
//
// Formatter coupling: the matcher relies on the "single" channel's JSON
// formatter putting `"message":"Broadcasting event"` and the payload on one
// line. The "daily" text channel splits the payload onto its own `[With]`
// line, so those entries are naturally excluded. If config/logging.go ever
// switches "single" to a non-JSON formatter (or "daily" to JSON), the count
// will silently drift — the first symptom is `before+3`/`before+1` failing.
func (s *BroadcastTestSuite) countBroadcastEvents(marker string) int {
	matches, err := filepath.Glob("../../storage/logs/*.log")
	s.Require().NoError(err)
	count := 0
	for _, m := range matches {
		if b, rErr := os.ReadFile(m); rErr == nil {
			for _, line := range strings.Split(string(b), "\n") {
				// Tighten to the JSON formatter's line shape so a formatter
				// change fails loudly instead of double-counting (daily→json).
				if strings.Contains(line, `"message":"Broadcasting event"`) && strings.Contains(line, marker) {
					count++
				}
			}
		}
	}
	return count
}

// waitForFailedBroadcast polls the queue failer until a failed goravel_broadcast
// job appears (the queued broadcast always fails via the phantom "broken"
// connection), returning the elapsed time since start. The failure is recorded
// by the worker's failed-job processor after all retries (in-process or
// release-based) are exhausted, so its appearance in the failer is the
// reliable completion signal —
// unlike the jobs table, which the worker empties before retries run. Returns
// ok=false if the timeout elapses.
//
// Invariants this relies on: (a) SetupTest's RefreshDatabase (migrate:refresh)
// drops and recreates failed_jobs before every test, so the table starts empty;
// and (b) each new test dispatches exactly one queued broadcast, so the first
// failed goravel_broadcast row this helper sees is that job's. If a future test
// stops calling RefreshDatabase or dispatches a second queued broadcast, the
// poll would latch onto a stale/foreign row instead of timing out.
func (s *BroadcastTestSuite) waitForFailedBroadcast(start time.Time, timeout time.Duration) (time.Duration, bool) {
	deadline := time.Now().Add(timeout)
	for {
		failedJobs, err := facades.Queue().Failer().All()
		if err != nil {
			s.T().Logf("failed to list failed jobs: %v", err)
		} else {
			for _, fj := range failedJobs {
				if fj.Signature() == "goravel_broadcast" {
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

func soketiTrigger(channel, event string, data map[string]any) error {
	return soketiTriggerMulti([]string{channel}, event, data)
}

func soketiTriggerMulti(channels []string, event string, data map[string]any) error {
	dataJSON, _ := json.Marshal(data)
	body := map[string]any{
		"name":     event,
		"channels": channels,
		"data":     string(dataJSON),
	}
	bodyBytes, _ := json.Marshal(body)

	bodyMD5 := fmt.Sprintf("%x", md5.Sum(bodyBytes))
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	signing := fmt.Sprintf("POST\n/apps/%s/events\nauth_key=%s&auth_timestamp=%s&auth_version=1.0&body_md5=%s",
		soketiAppID, soketiAppKey, timestamp, bodyMD5)

	mac := hmac.New(sha256.New, []byte(soketiSecret))
	_, _ = mac.Write([]byte(signing))
	sig := hex.EncodeToString(mac.Sum(nil))

	reqURL := fmt.Sprintf("http://%s/apps/%s/events?auth_key=%s&auth_timestamp=%s&auth_version=1.0&body_md5=%s&auth_signature=%s",
		soketiHost, soketiAppID, soketiAppKey, timestamp, bodyMD5, sig)

	resp, err := stdhttp.Post(reqURL, "application/json", bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("soketi responded %d: %s", resp.StatusCode, string(body))
	}

	return nil
}
