package feature

import (
	"context"
	"fmt"
	"testing"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/redis"
	"github.com/stretchr/testify/suite"
)

func TestRedisDriver(t *testing.T) {
	facades.Config().Add("cache.default", "redis")
	facades.Config().Add("queue.default", "redis")
	facades.App().Refresh()

	initMachineryQueue()

	go func() {
		if err := facades.Queue().Worker().Run(); err != nil {
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

// initMachineryQueue will write some tasks to the queue, because the machinery worker will not start if the queue is empty
func initMachineryQueue() {
	ctx := context.Background()
	redisQueue, err := redis.NewQueue(ctx, facades.Config(), facades.Queue(), facades.App().GetJson(), "redis")
	if err != nil {
		panic(err)
	}

	appName := facades.Config().GetString("app.name")
	queueKey := fmt.Sprintf("%s_%s:%s", appName, "queues", "default")
	testQueueKey := fmt.Sprintf("%s_%s:%s", appName, "queues", "test")

	dispatch := "{\"UUID\":\"task_db1f04df-01a4-420d-984b-b5366bbf4954\",\"Name\":\"test\",\"RoutingKey\":\"Goravel_queues:default\",\"ETA\":null,\"GroupUUID\":\"\",\"GroupTaskCount\":0,\"Args\":[],\"Headers\":{},\"Priority\":0,\"Immutable\":false,\"RetryCount\":0,\"RetryTimeout\":0,\"OnSuccess\":null,\"OnError\":null,\"ChordCallback\":null,\"BrokerMessageGroupId\":\"\",\"SQSReceiptHandle\":\"\",\"StopTaskDeletionOnError\":false,\"IgnoreWhenTaskNotRegistered\":false}"

	err = redisQueue.Instance().RPush(ctx, queueKey, dispatch).Err()
	if err != nil {
		panic(err)
	}

	dispatchWithDelay := "{\"UUID\":\"task_3ee9dbfb-d578-4f11-b675-7561503f0b66\",\"Name\":\"test\",\"RoutingKey\":\"Goravel_queues:default\",\"ETA\":\"2025-04-16T19:30:06.018975+08:00\",\"GroupUUID\":\"\",\"GroupTaskCount\":0,\"Args\":[],\"Headers\":{},\"Priority\":0,\"Immutable\":false,\"RetryCount\":0,\"RetryTimeout\":0,\"OnSuccess\":null,\"OnError\":null,\"ChordCallback\":null,\"BrokerMessageGroupId\":\"\",\"SQSReceiptHandle\":\"\",\"StopTaskDeletionOnError\":false,\"IgnoreWhenTaskNotRegistered\":false}"

	err = redisQueue.Instance().RPush(ctx, queueKey, dispatchWithDelay).Err()
	if err != nil {
		panic(err)
	}

	dispatchWithConnectionAndQueue := "{\"UUID\":\"task_18868508-9928-4a22-a05a-fb79b38fa905\",\"Name\":\"test\",\"RoutingKey\":\"Goravel_queues:test\",\"ETA\":null,\"GroupUUID\":\"\",\"GroupTaskCount\":0,\"Args\":[],\"Headers\":{},\"Priority\":0,\"Immutable\":false,\"RetryCount\":0,\"RetryTimeout\":0,\"OnSuccess\":null,\"OnError\":null,\"ChordCallback\":null,\"BrokerMessageGroupId\":\"\",\"SQSReceiptHandle\":\"\",\"StopTaskDeletionOnError\":false,\"IgnoreWhenTaskNotRegistered\":false}"

	err = redisQueue.Instance().RPush(ctx, testQueueKey, dispatchWithConnectionAndQueue).Err()
	if err != nil {
		panic(err)
	}
}
