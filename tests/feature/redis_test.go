package feature

import (
	"testing"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"
)

func TestRedisDriver(t *testing.T) {
	facades.Config().Add("cache.default", "redis")
	facades.Config().Add("queue.default", "redis")
	facades.App().Refresh()

	go func() {
		if err := facades.Queue().Worker().Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	go func() {
		if err := facades.Queue().Worker(queue.Args{
			Queue: "test",
		}).Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	go func() {
		if err := facades.Queue().Worker(queue.Args{
			Connection: "redis1",
			Queue:      "test",
		}).Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	suite.Run(t, &HttpTestSuite{})
	suite.Run(t, &QueueTestSuite{})

	facades.Config().Add("cache.default", "memory")
	facades.Config().Add("queue.default", "sync")
	facades.App().Refresh()
}

func TestMachineryDriver(t *testing.T) {
	facades.Config().Add("cache.default", "redis")
	facades.Config().Add("queue.default", "machinery")
	facades.App().Refresh()

	go func() {
		if err := facades.Queue().Worker().Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	go func() {
		if err := facades.Queue().Worker(queue.Args{
			Queue:      "test",
			Concurrent: 2,
		}).Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	go func() {
		if err := facades.Queue().Worker(queue.Args{
			Connection: "redis1",
			Queue:      "test",
		}).Run(); err != nil {
			facades.Log().Errorf("Queue run error: %v", err)
		}
	}()

	suite.Run(t, &QueueTestSuite{})

	facades.Config().Add("cache.default", "memory")
	facades.Config().Add("queue.default", "sync")
	facades.App().Refresh()
}
