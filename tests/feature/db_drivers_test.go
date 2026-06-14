package feature

import (
	"testing"

	"github.com/goravel/framework/support/env"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
)

func TestDBDrivers(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping database driver tests on Windows due to potential issues with Docker.")
	}

	connections := []string{"postgres", "mysql", "sqlserver"}

	for _, connection := range connections {
		database, err := facades.Testing().Docker().Database(connection)
		if err != nil {
			panic(err)
		}

		if err := database.Build(); err != nil {
			panic(err)
		}

		if err := database.Ready(); err != nil {
			panic(err)
		}

		facades.Config().Add("database.default", connection)
		facades.Config().Add("database.connections."+connection+".port", database.Config().Port)

		if err := facades.App().Restart(); err != nil {
			panic(err)
		}

		suite.Run(t, &DBTestSuite{})
		suite.Run(t, &OrmTestSuite{})
		suite.Run(t, &MigrationTestSuite{})

		facades.Config().Add("database.default", "sqlite")
		if err := facades.App().Restart(); err != nil {
			panic(err)
		}

		assert.NoError(t, database.Shutdown())
	}
}
