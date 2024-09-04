package controllers

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
)

func TestMain(m *testing.M) {
	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	}()

	m.Run()

	file.Remove("storage")
}
