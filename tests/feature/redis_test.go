package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"
)

func TestRedisDriver(t *testing.T) {
	facades.Config().Add("cache.default", "redis")
	facades.App().Refresh()

	suite.Run(t, &HttpTestSuite{})

	facades.Config().Add("cache.default", "memory")
	facades.App().Refresh()
}
