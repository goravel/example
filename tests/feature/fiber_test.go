package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
)

func TestFiberDriver(t *testing.T) {
	facades.Config().Add("http.default", "fiber")
	if err := facades.App().Restart(); err != nil {
		panic(err)
	}

	suite.Run(t, new(HttpTestSuite))
	suite.Run(t, new(ViewTestSuite))

	facades.Config().Add("http.default", "gin")
	if err := facades.App().Restart(); err != nil {
		panic(err)
	}
}
