package services

import (
	"testing"

	"goravel/app/facades"

	"github.com/goravel/framework/support/file"
)

func TestMain(m *testing.M) {
	if err := facades.Artisan().Call("migrate"); err != nil {
		panic(err)
	}

	m.Run()

	_ = file.Remove("goravel")
}
