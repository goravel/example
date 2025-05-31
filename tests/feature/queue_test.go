package feature

import (
	"testing"
	"time"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/queue"
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
	s.NoError(facades.Queue().Job(&jobs.Test{}, testQueueArgs).Dispatch())

	time.Sleep(1 * time.Second)

	s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchWithDelay() {
	s.NoError(facades.Queue().Job(&jobs.Test{}, testQueueArgs).Delay(time.Now().Add(1 * time.Second)).Dispatch())

	time.Sleep(2 * time.Second)

	s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchChain() {
	s.NoError(facades.Queue().Chain([]contractsqueue.ChainJob{
		{
			Job:  &jobs.Test{},
			Args: testQueueArgs,
		},
		{
			Job:  &jobs.Test{},
			Args: testQueueArgs,
		},
	}).Dispatch())

	time.Sleep(1 * time.Second)

	var args []any
	for i := 0; i < 2; i++ {
		args = append(args, queue.ConvertArgs(testQueueArgs)...)
	}

	s.Equal(args, jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchWithQueue() {
	s.NoError(facades.Queue().Job(&jobs.Test{}, testQueueArgs).OnQueue("test").Dispatch())

	time.Sleep(1 * time.Second)

	s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
}

func (s *QueueTestSuite) TestDispatchWithConnectionAndQueue() {
	if facades.Config().GetString("queue.default") == "sync" {
		s.T().Skip("skip test due to only for redis")
	}

	s.NoError(facades.Queue().Job(&jobs.Test{}, testQueueArgs).OnConnection("redis1").OnQueue("test").Dispatch())

	time.Sleep(1 * time.Second)

	s.Equal(queue.ConvertArgs(testQueueArgs), jobs.TestResult)
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
