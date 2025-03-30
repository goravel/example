package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"
)

func TestPostgresDriver(t *testing.T) {
	database, err := facades.Testing().Docker().Database("postgres")
	if err != nil {
		panic(err)
	}

	if err := database.Build(); err != nil {
		panic(err)
	}

	facades.Config().Add("database.default", "postgres")
	facades.Config().Add("database.connections.postgres.port", database.Config().Port)

	facades.App().Refresh()
	facades.App().Boot()

	if err := database.Migrate(); err != nil {
		panic(err)
	}

	suite.Run(t, &DBTestSuite{})
	suite.Run(t, &OrmTestSuite{})

	facades.Config().Add("database.default", "sqlite")
	facades.App().Refresh()
	facades.App().Boot()
}

func TestMysqlDriver(t *testing.T) {
	facades.Config().Add("database.default", "mysql")
	facades.App().Refresh()
	facades.App().Boot()

	suite.Run(t, &DBTestSuite{})
	suite.Run(t, &OrmTestSuite{})

	facades.Config().Add("database.default", "sqlite")
	facades.App().Refresh()
	facades.App().Boot()
}

func TestSqliteDriver(t *testing.T) {
	facades.Config().Add("database.default", "sqlite")
	facades.App().Refresh()
	facades.App().Boot()

	suite.Run(t, &DBTestSuite{})
	suite.Run(t, &OrmTestSuite{})

	facades.Config().Add("database.default", "sqlite")
	facades.App().Refresh()
	facades.App().Boot()
}
