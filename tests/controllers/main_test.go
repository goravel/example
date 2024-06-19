package controllers

import (
	"fmt"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/str"
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

func route(path string) string {
	return fmt.Sprintf("http://%s:%s/%s",
		facades.Config().GetString("APP_HOST"),
		facades.Config().GetString("APP_PORT"),
		str.Of(path).LTrim("/").String())
}
