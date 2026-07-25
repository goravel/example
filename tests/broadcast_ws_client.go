package tests

import (
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
)

type WSClient struct {
	conn   *websocket.Conn
	events []WSEvent
	mu     sync.Mutex
}

type WSEvent struct {
	Event   string `json:"event"`
	Channel string `json:"channel"`
	Data    string `json:"data"`
}

func NewWSClient(host, appKey string) (*WSClient, error) {
	url := fmt.Sprintf("ws://%s/app/%s?protocol=7&client=js&version=7.0.0", host, appKey)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	c := &WSClient{conn: conn}
	go c.readLoop()
	return c, nil
}

func (c *WSClient) Subscribe(channel string) error {
	msg, _ := json.Marshal(map[string]any{
		"event": "pusher:subscribe",
		"data": map[string]string{
			"channel": channel,
		},
	})
	return c.conn.WriteMessage(websocket.TextMessage, msg)
}

func (c *WSClient) Events() []WSEvent {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]WSEvent, len(c.events))
	copy(result, c.events)
	return result
}

func (c *WSClient) EventCount() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.events)
}

func (c *WSClient) Close() error {
	return c.conn.Close()
}

func (c *WSClient) readLoop() {
	for {
		_, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		var event WSEvent
		if err := json.Unmarshal(msg, &event); err != nil {
			continue
		}

		c.mu.Lock()
		c.events = append(c.events, event)
		c.mu.Unlock()
	}
}
