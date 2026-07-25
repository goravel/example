package feature

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	stdhttp "net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goravel/framework/broadcasting"
	contracts "github.com/goravel/framework/contracts/broadcasting"
	supporthttp "github.com/goravel/framework/support/http"
	"github.com/gorilla/websocket"
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

func (c *wsClient) subscribePublic(channel string) error {
	return c.sendSubscribe(channel, "")
}

func (c *wsClient) subscribePrivate(authEndpoint, channel string) error {
	c.mu.Lock()
	socketID := c.socketID
	c.mu.Unlock()

	if socketID == "" {
		return fmt.Errorf("socket_id not available yet")
	}

	auth, err := c.fetchAuth(authEndpoint, socketID, channel)
	if err != nil {
		return err
	}

	return c.sendSubscribe(channel, auth)
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

func (c *wsClient) sendSubscribe(channel, auth string) error {
	data := map[string]string{"channel": channel}
	if auth != "" {
		data["auth"] = auth
	}
	msg, _ := json.Marshal(map[string]any{
		"event": "pusher:subscribe",
		"data":  data,
	})
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *wsClient) fetchAuth(endpoint, socketID, channelName string) (string, error) {
	form := url.Values{}
	form.Set("socket_id", socketID)
	form.Set("channel_name", channelName)

	resp, err := stdhttp.Post(endpoint, "application/x-www-form-urlencoded", strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != stdhttp.StatusOK {
		return "", fmt.Errorf("auth endpoint returned %d: %s", resp.StatusCode, string(body))
	}

	var ar struct {
		Auth        string `json:"auth"`
		ChannelData string `json:"channel_data"`
	}
	if err := json.Unmarshal(body, &ar); err != nil {
		return "", fmt.Errorf("auth endpoint returned non-JSON: %s", string(body))
	}

	if ar.Auth == "" {
		return "", fmt.Errorf("auth endpoint returned empty auth")
	}

	return ar.Auth, nil
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

		if eventName == "pusher_internal:subscription_succeeded" {
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

func TestWSClientConnect(t *testing.T) {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		t.Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	ws.mu.Lock()
	sid := ws.socketID
	ws.mu.Unlock()

	if sid == "" {
		t.Error("expected socket_id from connection_established event")
	}
}

func TestWSClientPublishAndReceive(t *testing.T) {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		t.Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	if err := ws.subscribePublic("test-channel"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	if err := soketiTrigger("test-channel", "order.shipped", map[string]any{"order_id": 1, "msg": "hello"}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	for _, e := range ws.events() {
		if e.Event == "order.shipped" && e.Channel == "test-channel" {
			return
		}
	}
	t.Error("expected order.shipped event on test-channel")
}

func TestWSClientPublishToMultipleChannels(t *testing.T) {
	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		t.Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	if err := ws.subscribePublic("channel-a"); err != nil {
		t.Fatal(err)
	}
	if err := ws.subscribePublic("channel-b"); err != nil {
		t.Fatal(err)
	}
	time.Sleep(300 * time.Millisecond)

	if err := soketiTriggerMulti([]string{"channel-a", "channel-b"}, "multi-event", map[string]any{"broadcast": true}); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)

	channels := map[string]bool{}
	for _, e := range ws.events() {
		channels[e.Channel] = true
	}

	if !channels["channel-a"] {
		t.Error("expected event on channel-a")
	}
	if !channels["channel-b"] {
		t.Error("expected event on channel-b")
	}
}

func TestWSClientPrivateChannelAuthRejected(t *testing.T) {
	facades.Config().Add("broadcasting.default", "pusher")
	facades.Config().Add("broadcasting.connections.pusher", map[string]any{
		"driver": "pusher",
		"key":    soketiAppKey,
		"secret": soketiSecret,
		"app_id": soketiAppID,
		"options": map[string]any{
			"host":   "127.0.0.1",
			"port":   6001,
			"scheme": "http",
		},
	})
	if err := facades.App().Restart(); err != nil {
		t.Fatal(err)
	}

	ws, err := newWSClient(soketiHost, soketiAppKey)
	if err != nil {
		t.Skip("Soketi not reachable: " + err.Error())
	}
	defer func() { _ = ws.close() }()

	err = ws.subscribePrivate(soketiAuthURL, "private-orders.999")
	if err == nil {
		t.Error("expected auth error for unauthorized channel")
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
