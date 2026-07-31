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
