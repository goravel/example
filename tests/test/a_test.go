package test

import (
	"fmt"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

type DBTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, &DBTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *DBTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *DBTestSuite) TearDownTest() {
}

func (s *DBTestSuite) TestCRUD() {
	fmt.Println(facades.Config().GetString("DB_DATABASE"))
	s.True(false)
}
