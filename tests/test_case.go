package tests

import (
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/testing"

	"goravel/bootstrap"
)

func init() {
	bootstrap.Boot()

	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	}()
}

type TestCase struct {
	testing.TestCase
}
