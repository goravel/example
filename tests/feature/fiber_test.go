package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"
)

func TestFiberDriver(t *testing.T) {
	facades.Config().Add("http.default", "fiber")
	facades.App().Refresh()
	facades.App().Boot()

	suite.Run(t, new(RouteTestSuite))

	facades.Config().Add("http.default", "gin")
	facades.App().Refresh()
	facades.App().Boot()
}
