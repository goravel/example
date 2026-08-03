package controllers

import (
	"context"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/http"

	"goravel/app/events"
	"goravel/app/facades"
)

type BroadcastController struct{}

func NewBroadcastController() *BroadcastController {
	return &BroadcastController{}
}

type DispatchRequest struct {
	ChannelType  string         `form:"channel_type" json:"channel_type"`
	Channel      string         `form:"channel" json:"channel"`
	OrderData    map[string]any `form:"order_data" json:"order_data"`
	TeamData     map[string]any `form:"team_data" json:"team_data"`
	ShouldFire   bool           `form:"should_fire" json:"should_fire"`
	QueueConn    string         `form:"queue_conn" json:"queue_conn"`
	QueueName    string         `form:"queue_name" json:"queue_name"`
	Conns        []string       `form:"conns" json:"conns"`
	Delay        int64          `form:"delay" json:"delay"`
	Timeout      int64          `form:"timeout" json:"timeout"`
	Tries        int            `form:"tries" json:"tries"`
	Backoff      []float64      `form:"backoff" json:"backoff"` // seconds per attempt, fractional allowed, e.g. [0.5,1]
	BroadcastNow bool           `form:"broadcast_now" json:"broadcast_now"`
}

func (c *BroadcastController) Dispatch(ctx http.Context) http.Response {
	var req DispatchRequest
	if err := ctx.Request().Bind(&req); err != nil {
		return ctx.Response().Json(http.StatusBadRequest, http.Json{
			"error": "invalid request body: " + err.Error(),
		})
	}

	if req.ChannelType == "" {
		req.ChannelType = "private"
	}

	switch req.ChannelType {
	case "public":
		return c.dispatchPublic(ctx, &req)
	case "presence":
		return c.dispatchPresence(ctx, &req)
	default:
		return c.dispatchPrivate(ctx, &req)
	}
}

func stripPrivatePrefix(name string) string {
	return strings.TrimPrefix(name, "private-")
}

func stripPresencePrefix(name string) string {
	return strings.TrimPrefix(name, "presence-")
}

func (c *BroadcastController) dispatchPrivate(ctx http.Context, req *DispatchRequest) http.Response {
	if req.ShouldFire && req.OrderData == nil {
		req.OrderData = map[string]any{
			"item":  "Laptop",
			"price": 1200,
		}
	}
	var delayedAt time.Time
	if req.Delay > 0 {
		delayedAt = time.Now().UTC().Add(time.Duration(req.Delay) * time.Second)
	}

	var backoff []time.Duration
	for _, b := range req.Backoff {
		backoff = append(backoff, time.Duration(b*float64(time.Second)))
	}

	err := facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		OrderData:          req.OrderData,
		ShouldFire:         req.ShouldFire,
		QueueName:          req.QueueName,
		Conns:              req.Conns,
		QueueConn:          req.QueueConn,
		DelayedAt:          delayedAt,
		Timeout:            time.Duration(req.Timeout) * time.Second,
		Tries:              req.Tries,
		Backoff:            backoff,
		ShouldBroadcastNow: req.BroadcastNow,
		ChannelName:        stripPrivatePrefix(req.Channel),
	})
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"error": err.Error()})
	}
	return ctx.Response().Success().Json(http.Json{
		"message": "private event dispatched",
		"channel": req.Channel,
	})
}

func (c *BroadcastController) dispatchPublic(ctx http.Context, req *DispatchRequest) http.Response {
	if req.OrderData == nil {
		req.OrderData = map[string]any{
			"item":  "Laptop",
			"price": 1200,
		}
	}
	err := facades.Broadcast().Dispatch(context.Background(), &events.OrderShippedBroadcast{
		ChannelType:        "public",
		OrderData:          req.OrderData,
		ShouldFire:         true,
		ShouldBroadcastNow: req.BroadcastNow,
		ChannelName:        req.Channel,
	})
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"error": err.Error()})
	}
	return ctx.Response().Success().Json(http.Json{
		"message": "public event dispatched",
		"channel": req.Channel,
	})
}

func (c *BroadcastController) dispatchPresence(ctx http.Context, req *DispatchRequest) http.Response {
	if req.TeamData == nil {
		req.TeamData = map[string]any{
			"name": "goravel",
		}
	}
	err := facades.Broadcast().Dispatch(context.Background(), &events.TeamPresenceBroadcast{
		TeamData:           req.TeamData,
		ChannelName:        stripPresencePrefix(req.Channel),
		ShouldBroadcastNow: req.BroadcastNow,
	})
	if err != nil {
		return ctx.Response().Json(http.StatusInternalServerError, http.Json{"error": err.Error()})
	}
	return ctx.Response().Success().Json(http.Json{
		"message": "presence event dispatched",
		"channel": req.Channel,
	})
}
