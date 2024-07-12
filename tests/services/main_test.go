package services

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
)

func TestMain(m *testing.M) {
	facades.Artisan().Call("migrate")

	m.Run()

	file.Remove("goravel")
}
