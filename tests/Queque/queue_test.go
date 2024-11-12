package Queque

import (
	"testing"

	"github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/app/jobs"
	"goravel/tests"
)

type TestQueueSuite struct {
	suite.Suite
	tests.TestCase
}

func TestTestQueueSuite(t *testing.T) {
	suite.Run(t, new(TestQueueSuite))
}

// SetupTest will run before each test in the suite.
func (s *TestQueueSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *TestQueueSuite) TearDownTest() {
}

func (s *TestQueueSuite) TestIndex() {
	// 1、config/queue.go 文件中需要指定：QUEUE_CONNECTION=redis
	// 2、根目录下执行air 启动项目服务
	// 3、在tests/Queue目录下执行go test -v
	// 4、查看logs/目录下的日志信息
	_ = facades.Queue().Job(&jobs.TestJob{}, []queue.Arg{}).OnQueue("test_job").Dispatch()
}
