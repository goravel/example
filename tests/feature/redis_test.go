package feature

import (
	"fmt"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

type RedisTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestRedisTestSuite(t *testing.T) {
	suite.Run(t, &RedisTestSuite{})
}

func (s *RedisTestSuite) SetupSuite() {
	facades.Config().Add("cache.default", "redis")
	facades.App().Refresh()
	facades.App().Boot()
	fmt.Println("RedisTestSuite", facades.Config().GetString("cache.default"))
}

// SetupTest will run before each test in the suite.
func (s *RedisTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *RedisTestSuite) TearDownTest() {
}

// TearDownSuite will run after each test in the suite.
func (s *RedisTestSuite) TearDownSuite() {
	facades.Config().Add("cache.default", "memory")
	facades.App().Refresh()
	facades.App().Boot()
	fmt.Println("RedisTestSuite", facades.Config().GetString("cache.default"))
}

func (s *RedisTestSuite) TestThrottle() {
	routeTestSuite := RouteTestSuite{
		Suite:    s.Suite,
		TestCase: s.TestCase,
	}

	routeTestSuite.TestThrottle()

	facades.Cache().Forever("test", "test")
}
