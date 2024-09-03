package feature

import (
	"os"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
)

func TestMain(m *testing.M) {
	database, err := facades.Testing().Docker().Database()
	if err != nil {
		panic(err)
	}

	if err := database.Build(); err != nil {
		panic(err)
	}

	exit := m.Run()

	file.Remove("storage")
	if err := database.Clear(); err != nil {
		panic(err)
	}

	os.Exit(exit)
}
