package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"
)

func TestFiberDriver(t *testing.T) {
	facades.Config().Add("http.default", "fiber")
	facades.App().Refresh()

	suite.Run(t, new(HttpTestSuite))

	facades.Config().Add("http.default", "gin")
	facades.App().Refresh()
}
