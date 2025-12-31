package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
)

func TestFiberDriver(t *testing.T) {
	facades.Config().Add("http.default", "fiber")
	facades.App().Refresh()

	suite.Run(t, new(HttpTestSuite))

	facades.Config().Add("http.default", "gin")
	facades.App().Refresh()
}
