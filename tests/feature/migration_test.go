package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/mysql"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

type MigrationTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, &MigrationTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *MigrationTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *MigrationTestSuite) TearDownTest() {
}

func (s *MigrationTestSuite) TestChange() {
	if facades.Schema().Orm().Config().Driver == sqlite.Name {
		s.T().Skip("sqlite does not support change column")
	}

	columns, err := facades.Schema().GetColumns("users")

	s.Require().NoError(err)

	var mailExists bool
	for _, column := range columns {
		if column.Name == "alias" {
			s.Contains(column.Default, "test")
		}

		if column.Name == "mail" {
			mailExists = true
		}
	}

	s.True(mailExists)
}

func (s *MigrationTestSuite) TestFirst_After() {
	if facades.Schema().Orm().Config().Driver != mysql.Name {
		s.T().Skip("only mysql supports first and after")
	}

	columns, err := facades.Schema().GetColumns("users")
	s.Require().NoError(err)

	s.Equal("mail", columns[0].Name)
	s.Equal("alias", columns[3].Name)
}
func (s *MigrationTestSuite) TestMigrate() {
	s.True(facades.Schema().HasTable("users"))
}

func (s *MigrationTestSuite) TestTableComment() {
	if facades.Schema().Orm().Config().Driver == sqlite.Name || facades.Schema().Orm().Config().Driver == sqlserver.Name {
		s.T().Skip("sqlite and sqlserver does not support table comment")
	}

	tables, err := facades.Schema().GetTables()
	s.Require().NoError(err)

	for _, table := range tables {
		if table.Name == "users" {
			s.Equal("user table", table.Comment)
		}
	}
}
