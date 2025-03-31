package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/sqlite"
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

func (s *MigrationTestSuite) TestMigrate() {
	s.True(facades.Schema().HasTable("users"))
}

func (s *MigrationTestSuite) TestChange() {
	if facades.Schema().Orm().Config().Driver == sqlite.Name {
		s.T().Skip("sqlite does not support change column")
	}

	columns, err := facades.Schema().GetColumns("users")

	s.Require().NoError(err)

	for _, column := range columns {
		if column.Name == "alias" {
			s.Contains(column.Default, "test")
		}
	}
}
