package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
	"goravel/tests"
)

type OrmTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestOrmTestSuite(t *testing.T) {
	suite.Run(t, &OrmTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *OrmTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *OrmTestSuite) TearDownTest() {
}

func (s *OrmTestSuite) TestCreate() {
	s.Run("create with map", func() {
		s.Require().NoError(facades.Orm().Query().Model(&models.User{}).Create(map[string]any{"name": "Goravel"}))

		var user models.User
		s.Require().NoError(facades.Orm().Query().First(&user))
		s.Equal("Goravel", user.Name)
	})
}

func (s *OrmTestSuite) TestRestore() {
	s.Require().NoError(facades.Orm().Query().Model(&models.User{}).Create(map[string]any{"name": "restore"}))

	var user models.User
	s.Require().NoError(facades.Orm().Query().Where("name", "restore").First(&user))
	s.Equal("restore", user.Name)

	res, err := facades.Orm().Query().Delete(&user)
	s.Require().NoError(err)
	s.Equal(int64(1), res.RowsAffected)

	var user1 models.User
	s.Require().NoError(facades.Orm().Query().Find(&user1, user.ID))
	s.Empty(user1.Name)

	res, err = facades.Orm().Query().WithTrashed().Restore(&user)
	s.Require().NoError(err)
	s.Equal(int64(1), res.RowsAffected)

	var user2 models.User
	s.Require().NoError(facades.Orm().Query().Where("name", "restore").First(&user2))
	s.Equal("restore", user2.Name)
}
