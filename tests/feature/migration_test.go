package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
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
}

// TearDownTest will run after each test in the suite.
func (s *MigrationTestSuite) TearDownTest() {
}

func (s *MigrationTestSuite) TestMigrate() {
	s.True(facades.Schema().HasTable("users"))
}
