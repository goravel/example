package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestDBDrivers(t *testing.T) {
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
		facades.App().Refresh()

		suite.Run(t, &OrmTestSuite{})
		suite.Run(t, &MigrationTestSuite{})

		facades.Config().Add("database.default", "sqlite")
		facades.App().Refresh()

		assert.NoError(t, database.Shutdown())
	}
}
