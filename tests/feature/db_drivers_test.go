package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
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
		facades.Config().Add("database.connections."+connection+".port", database.Config().Port)

		facades.App().Refresh()

		if err := database.Migrate(); err != nil {
			panic(err)
		}

		suite.Run(t, &DBTestSuite{})
		suite.Run(t, &OrmTestSuite{})

		facades.Config().Add("database.default", "sqlite")
		facades.App().Refresh()
	}
}
