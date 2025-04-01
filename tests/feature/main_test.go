package feature

import (
	"os"
	"testing"
	"time"

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

	if err := database.Migrate(); err != nil {
		panic(err)
	}

	go func() {
		if err := facades.Route().Run(); err != nil {
			facades.Log().Errorf("Route run error: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)

	exit := m.Run()

	if err := file.Remove("storage"); err != nil {
		panic(err)
	}
	if err := database.Shutdown(); err != nil {
		panic(err)
	}

	os.Exit(exit)
}
