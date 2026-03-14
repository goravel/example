package feature

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
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
	s.RefreshDatabase()
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

func (s *OrmTestSuite) TestFirstOrCreate() {
	var created models.User
	s.Require().NoError(facades.Orm().Query().FirstOrCreate(&created, models.User{
		Name: "first-or-create",
	}))
	s.NotZero(created.ID)
	s.Equal("first-or-create", created.Name)

	var existing models.User
	s.Require().NoError(facades.Orm().Query().FirstOrCreate(&existing, models.User{
		Name: "first-or-create",
	}, models.User{
		Avatar: "new-avatar.png",
	}))
	s.Equal(created.ID, existing.ID)
	s.Equal("", existing.Avatar)

	count, err := facades.Orm().Query().Model(&models.User{}).Where("name", "first-or-create").Count()
	s.Require().NoError(err)
	s.Equal(int64(1), count)
}

func (s *OrmTestSuite) TestUpdateOrCreate() {
	var created models.User
	s.Require().NoError(facades.Orm().Query().UpdateOrCreate(&created, map[string]any{
		"name": "update-or-create",
	}, map[string]any{
		"avatar": "old-avatar.png",
	}))
	s.NotZero(created.ID)

	var updated models.User
	s.Require().NoError(facades.Orm().Query().UpdateOrCreate(&updated, map[string]any{
		"name": "update-or-create",
	}, map[string]any{
		"avatar": "new-avatar.png",
	}))
	s.Equal(created.ID, updated.ID)
	s.Equal("new-avatar.png", updated.Avatar)

	count, err := facades.Orm().Query().Model(&models.User{}).Where("name", "update-or-create").Count()
	s.Require().NoError(err)
	s.Equal(int64(1), count)
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
