package feature

import (
	"errors"
	"testing"

	ormcontract "github.com/goravel/framework/contracts/database/orm"
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

func (s *OrmTestSuite) TestFirstOrNew() {
	s.Require().NoError(facades.Orm().Query().Model(&models.User{}).Create(map[string]any{
		"name":   "first-or-new",
		"avatar": "exists.png",
	}))

	var existing models.User
	s.Require().NoError(facades.Orm().Query().FirstOrNew(&existing, map[string]any{
		"name": "first-or-new",
	}, map[string]any{
		"avatar": "new.png",
	}))
	s.NotZero(existing.ID)
	s.Equal("exists.png", existing.Avatar)

	var missing models.User
	s.Require().NoError(facades.Orm().Query().FirstOrNew(&missing, map[string]any{
		"name": "first-or-new-missing",
	}, map[string]any{
		"avatar": "missing.png",
	}))
	s.Zero(missing.ID)
	s.Equal("first-or-new-missing", missing.Name)
	s.Equal("missing.png", missing.Avatar)

	count, err := facades.Orm().Query().Model(&models.User{}).Where("name", "first-or-new-missing").Count()
	s.Require().NoError(err)
	s.Equal(int64(0), count)
}

func (s *OrmTestSuite) TestTransaction() {
	err := facades.Orm().Transaction(func(tx ormcontract.Query) error {
		if err := tx.Model(&models.User{}).Create(map[string]any{"name": "orm-tx-rollback"}); err != nil {
			return err
		}

		return errors.New("rollback")
	})
	s.Require().Error(err)

	count, err := facades.Orm().Query().Model(&models.User{}).Where("name", "orm-tx-rollback").Count()
	s.Require().NoError(err)
	s.Equal(int64(0), count)

	err = facades.Orm().Transaction(func(tx ormcontract.Query) error {
		return tx.Model(&models.User{}).Create(map[string]any{"name": "orm-tx-commit"})
	})
	s.Require().NoError(err)

	count, err = facades.Orm().Query().Model(&models.User{}).Where("name", "orm-tx-commit").Count()
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
