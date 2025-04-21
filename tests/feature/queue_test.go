package feature

import (
	"context"
	"fmt"
	"testing"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue"
	"github.com/goravel/redis"
	"github.com/stretchr/testify/suite"

	"goravel/app/jobs"
	"goravel/tests"
)

type QueueTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestQueueTestSuite(t *testing.T) {
	suite.Run(t, &QueueTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *QueueTestSuite) SetupTest() {
	jobs.TestResult = nil
}

// TearDownTest will run after each test in the suite.
func (s *QueueTestSuite) TearDownTest() {
}

func (s *QueueTestSuite) TestDispatch() {
	facades.Queue().Job(&jobs.Test{}, testQueueArgs).Dispatch()

	time.Sleep(1 * time.Second)

	s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchWithDelay() {
	facades.Queue().Job(&jobs.Test{}, testQueueArgs).Delay(time.Now().Add(1 * time.Second)).Dispatch()

	time.Sleep(2 * time.Second)

	s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchChain() {
	facades.Queue().Chain([]contractsqueue.Jobs{
		{
			Job:  &jobs.Test{},
			Args: testQueueArgs,
		},
		{
			Job:  &jobs.Test{},
			Args: testQueueArgs,
		},
	}).Dispatch()

	time.Sleep(1 * time.Second)

	var args []any
	for i := 0; i < 2; i++ {
		args = append(args, queue.ConvertArgs(testQueueArgs)...)
	}

	s.Equal(args, jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchWithConnectionAndQueue() {
	if facades.Config().GetString("queue.default") == "sync" {
		s.T().Skip("skip test due to only for redis")
	}

	facades.Queue().Job(&jobs.Test{}, testQueueArgs).OnConnection("redis1").OnQueue("test").Dispatch()

	time.Sleep(1 * time.Second)

	s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestMachinery() {
	if facades.Config().GetString("queue.default") == "sync" {
		s.T().Skip("skip test due to only for redis")
	}

	ctx := context.Background()
	redisQueue, err := redis.NewQueue(ctx, facades.Config(), facades.Queue(), facades.App().GetJson(), "redis")
	s.Nil(err)

	appName := facades.Config().GetString("app.name")

	s.Run("dispatch", func() {
		jobs.TestResult = nil

		dispatch := "{\"UUID\":\"task_db1f04df-01a4-420d-984b-b5366bbf4954\",\"Name\":\"test\",\"RoutingKey\":\"Goravel_queues:default\",\"ETA\":null,\"GroupUUID\":\"\",\"GroupTaskCount\":0,\"Args\":[{\"Name\":\"\",\"Type\":\"bool\",\"Value\":true},{\"Name\":\"\",\"Type\":\"int\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"float32\",\"Value\":1.1},{\"Name\":\"\",\"Type\":\"float64\",\"Value\":1.2},{\"Name\":\"\",\"Type\":\"string\",\"Value\":\"test\"},{\"Name\":\"\",\"Type\":\"[]bool\",\"Value\":[true,false]},{\"Name\":\"\",\"Type\":\"[]int\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int8\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint8\",\"Value\":\"AQID\"},{\"Name\":\"\",\"Type\":\"[]uint16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]float32\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]float64\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]string\",\"Value\":[\"test\",\"test2\",\"test3\"]}],\"Headers\":{},\"Priority\":0,\"Immutable\":false,\"RetryCount\":0,\"RetryTimeout\":0,\"OnSuccess\":null,\"OnError\":null,\"ChordCallback\":null,\"BrokerMessageGroupId\":\"\",\"SQSReceiptHandle\":\"\",\"StopTaskDeletionOnError\":false,\"IgnoreWhenTaskNotRegistered\":false}"

		s.Require().NoError(redisQueue.Instance().RPush(ctx, fmt.Sprintf("%s_%s:%s", appName, "queues", "default"), dispatch).Err())

		time.Sleep(1 * time.Second)

		s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
	})

	s.Run("dispatch chain", func() {
		jobs.TestResult = nil

		dispatchChain := "{\"UUID\":\"task_5d25e1a3-4be2-47d9-8dee-fa640135ab67\",\"Name\":\"test\",\"RoutingKey\":\"Goravel_queues:default\",\"ETA\":null,\"GroupUUID\":\"\",\"GroupTaskCount\":0,\"Args\":[{\"Name\":\"\",\"Type\":\"bool\",\"Value\":true},{\"Name\":\"\",\"Type\":\"int\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"float32\",\"Value\":1.1},{\"Name\":\"\",\"Type\":\"float64\",\"Value\":1.2},{\"Name\":\"\",\"Type\":\"string\",\"Value\":\"test\"},{\"Name\":\"\",\"Type\":\"[]bool\",\"Value\":[true,false]},{\"Name\":\"\",\"Type\":\"[]int\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int8\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint8\",\"Value\":\"AQID\"},{\"Name\":\"\",\"Type\":\"[]uint16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]float32\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]float64\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]string\",\"Value\":[\"test\",\"test2\",\"test3\"]}],\"Headers\":{},\"Priority\":0,\"Immutable\":false,\"RetryCount\":0,\"RetryTimeout\":0,\"OnSuccess\":[{\"UUID\":\"task_afaca469-4037-4df6-8c40-d3c51eb7b1b2\",\"Name\":\"test\",\"RoutingKey\":\"\",\"ETA\":null,\"GroupUUID\":\"\",\"GroupTaskCount\":0,\"Args\":[{\"Name\":\"\",\"Type\":\"bool\",\"Value\":true},{\"Name\":\"\",\"Type\":\"int\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"float32\",\"Value\":1.1},{\"Name\":\"\",\"Type\":\"float64\",\"Value\":1.2},{\"Name\":\"\",\"Type\":\"string\",\"Value\":\"test\"},{\"Name\":\"\",\"Type\":\"[]bool\",\"Value\":[true,false]},{\"Name\":\"\",\"Type\":\"[]int\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int8\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint8\",\"Value\":\"AQID\"},{\"Name\":\"\",\"Type\":\"[]uint16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]float32\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]float64\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]string\",\"Value\":[\"test\",\"test2\",\"test3\"]}],\"Headers\":null,\"Priority\":0,\"Immutable\":false,\"RetryCount\":0,\"RetryTimeout\":0,\"OnSuccess\":null,\"OnError\":null,\"ChordCallback\":null,\"BrokerMessageGroupId\":\"\",\"SQSReceiptHandle\":\"\",\"StopTaskDeletionOnError\":false,\"IgnoreWhenTaskNotRegistered\":false}],\"OnError\":null,\"ChordCallback\":null,\"BrokerMessageGroupId\":\"\",\"SQSReceiptHandle\":\"\",\"StopTaskDeletionOnError\":false,\"IgnoreWhenTaskNotRegistered\":false}"

		s.Require().NoError(redisQueue.Instance().RPush(ctx, fmt.Sprintf("%s_%s:%s", appName, "queues", "default"), dispatchChain).Err())

		time.Sleep(1 * time.Second)

		var args []any
		for i := 0; i < 2; i++ {
			args = append(args, queue.ConvertArgs(testQueueArgs)...)
		}

		s.Equal(args, jobs.TestResult)
	})

	s.Run("dispatch with delay", func() {
		jobs.TestResult = nil

		dispatchWithDelay := "{\"UUID\":\"task_3ee9dbfb-d578-4f11-b675-7561503f0b66\",\"Name\":\"test\",\"RoutingKey\":\"Goravel_queues:default\",\"ETA\":\"2025-04-16T19:30:06.018975+08:00\",\"GroupUUID\":\"\",\"GroupTaskCount\":0,\"Args\":[{\"Name\":\"\",\"Type\":\"bool\",\"Value\":true},{\"Name\":\"\",\"Type\":\"int\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"float32\",\"Value\":1.1},{\"Name\":\"\",\"Type\":\"float64\",\"Value\":1.2},{\"Name\":\"\",\"Type\":\"string\",\"Value\":\"test\"},{\"Name\":\"\",\"Type\":\"[]bool\",\"Value\":[true,false]},{\"Name\":\"\",\"Type\":\"[]int\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int8\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint8\",\"Value\":\"AQID\"},{\"Name\":\"\",\"Type\":\"[]uint16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]float32\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]float64\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]string\",\"Value\":[\"test\",\"test2\",\"test3\"]}],\"Headers\":{},\"Priority\":0,\"Immutable\":false,\"RetryCount\":0,\"RetryTimeout\":0,\"OnSuccess\":null,\"OnError\":null,\"ChordCallback\":null,\"BrokerMessageGroupId\":\"\",\"SQSReceiptHandle\":\"\",\"StopTaskDeletionOnError\":false,\"IgnoreWhenTaskNotRegistered\":false}"

		s.Require().NoError(redisQueue.Instance().RPush(ctx, fmt.Sprintf("%s_%s:%s", appName, "queues", "default"), dispatchWithDelay).Err())

		time.Sleep(1 * time.Second)

		s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
	})

	s.Run("dispatch with connection and queue", func() {
		jobs.TestResult = nil

		dispatchWithConnectionAndQueue := "{\"UUID\":\"task_18868508-9928-4a22-a05a-fb79b38fa905\",\"Name\":\"test\",\"RoutingKey\":\"Goravel_queues:test\",\"ETA\":null,\"GroupUUID\":\"\",\"GroupTaskCount\":0,\"Args\":[{\"Name\":\"\",\"Type\":\"bool\",\"Value\":true},{\"Name\":\"\",\"Type\":\"int\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"int64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint8\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint16\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint32\",\"Value\":1},{\"Name\":\"\",\"Type\":\"uint64\",\"Value\":1},{\"Name\":\"\",\"Type\":\"float32\",\"Value\":1.1},{\"Name\":\"\",\"Type\":\"float64\",\"Value\":1.2},{\"Name\":\"\",\"Type\":\"string\",\"Value\":\"test\"},{\"Name\":\"\",\"Type\":\"[]bool\",\"Value\":[true,false]},{\"Name\":\"\",\"Type\":\"[]int\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int8\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]int64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint8\",\"Value\":\"AQID\"},{\"Name\":\"\",\"Type\":\"[]uint16\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint32\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]uint64\",\"Value\":[1,2,3]},{\"Name\":\"\",\"Type\":\"[]float32\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]float64\",\"Value\":[1.1,1.2,1.3]},{\"Name\":\"\",\"Type\":\"[]string\",\"Value\":[\"test\",\"test2\",\"test3\"]}],\"Headers\":{},\"Priority\":0,\"Immutable\":false,\"RetryCount\":0,\"RetryTimeout\":0,\"OnSuccess\":null,\"OnError\":null,\"ChordCallback\":null,\"BrokerMessageGroupId\":\"\",\"SQSReceiptHandle\":\"\",\"StopTaskDeletionOnError\":false,\"IgnoreWhenTaskNotRegistered\":false}"

		s.Require().NoError(redisQueue.Instance().RPush(ctx, fmt.Sprintf("%s_%s:%s", appName, "queues", "test"), dispatchWithConnectionAndQueue).Err())

		time.Sleep(1 * time.Second)

		s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
	})
}

var (
	testQueueArgs = []contractsqueue.Arg{
		{
			Type:  "bool",
			Value: true,
		},
		{
			Type:  "int",
			Value: 1,
		},
		{
			Type:  "int8",
			Value: int8(1),
		},
		{
			Type:  "int16",
			Value: int16(1),
		},
		{
			Type:  "int32",
			Value: int32(1),
		},
		{
			Type:  "int64",
			Value: int64(1),
		},
		{
			Type:  "uint",
			Value: uint(1),
		},
		{
			Type:  "uint8",
			Value: uint8(1),
		},
		{
			Type:  "uint16",
			Value: uint16(1),
		},
		{
			Type:  "uint32",
			Value: uint32(1),
		},
		{
			Type:  "uint64",
			Value: uint64(1),
		},
		{
			Type:  "float32",
			Value: float32(1.1),
		},
		{
			Type:  "float64",
			Value: float64(1.2),
		},
		{
			Type:  "string",
			Value: "test",
		},
		{
			Type:  "[]bool",
			Value: []bool{true, false},
		},
		{
			Type:  "[]int",
			Value: []int{1, 2, 3},
		},
		{
			Type:  "[]int8",
			Value: []int8{1, 2, 3},
		},
		{
			Type:  "[]int16",
			Value: []int16{1, 2, 3},
		},
		{
			Type:  "[]int32",
			Value: []int32{1, 2, 3},
		},
		{
			Type:  "[]int64",
			Value: []int64{1, 2, 3},
		},
		{
			Type:  "[]uint",
			Value: []uint{1, 2, 3},
		},
		{
			Type:  "[]uint8",
			Value: []uint8{1, 2, 3},
		},
		{
			Type:  "[]uint16",
			Value: []uint16{1, 2, 3},
		},
		{
			Type:  "[]uint32",
			Value: []uint32{1, 2, 3},
		},
		{
			Type:  "[]uint64",
			Value: []uint64{1, 2, 3},
		},
		{
			Type:  "[]float32",
			Value: []float32{1.1, 1.2, 1.3},
		},
		{
			Type:  "[]float64",
			Value: []float64{1.1, 1.2, 1.3},
		},
		{
			Type:  "[]string",
			Value: []string{"test", "test2", "test3"},
		},
	}
)
