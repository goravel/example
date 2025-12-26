package services

import (
	"goravel/app/models"
	"goravel/tests"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/debug"
	"github.com/stretchr/testify/suite"
)

type ExampleTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestExampleTestSuite(t *testing.T) {
	suite.Run(t, new(ExampleTestSuite))
}

// SetupTest will run before each test in the suite.
func (s *ExampleTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *ExampleTestSuite) TearDownTest() {
}

func (s *ExampleTestSuite) TestIndex() {
	type UserListItem struct {
		No   int    `json:"no"`
		Name string `json:"name"`
	}

	selectColumns := []string{
		"users.id as no",
		//"users.alias as name", // 注释这行代码会报错
	}

	var total int64

	var userInfos []UserListItem
	facades.Orm().Query().Model(&models.User{}).
		Select(selectColumns...).Paginate(1, 10, &userInfos, &total)

	debug.Dump(userInfos)

	s.True(true)
}
