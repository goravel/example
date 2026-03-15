//go:debug x509negativeserial=1

package feature

import (
	"errors"
	"testing"

	"github.com/goravel/framework/support/carbon"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
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
	result, err := facades.DB().Table("users").Insert(map[string]any{"name": "Goravel"})
	s.Require().NoError(err)
	s.Equal(int64(1), result.RowsAffected)

	var user User
	s.Require().NoError(facades.DB().Table("users").Where("name", "Goravel").First(&user))
	s.True(user.ID > 0)
	s.Equal("Goravel", user.Name)

	result, err = facades.DB().Table("users").Where("id", user.ID).Update(map[string]any{"name": "Goravel2"})
	s.Require().NoError(err)
	s.Equal(int64(1), result.RowsAffected)

	s.Require().NoError(facades.DB().Table("users").Where("name", "Goravel2").First(&user))
	s.Equal("Goravel2", user.Name)

	result, err = facades.DB().Table("users").Where("id", user.ID).Delete()
	s.Require().NoError(err)
	s.Equal(int64(1), result.RowsAffected)
}

func (s *DBTestSuite) TestTransaction() {
	tx, err := facades.DB().BeginTransaction()
	s.Require().NoError(err)

	result, err := tx.Table("users").Insert(map[string]any{"name": "rollback"})
	s.Require().NoError(err)
	s.Equal(int64(1), result.RowsAffected)
	s.Require().NoError(tx.Rollback())

	exists, err := facades.DB().Table("users").Where("name", "rollback").Exists()
	s.Require().NoError(err)
	s.False(exists)

	tx, err = facades.DB().BeginTransaction()
	s.Require().NoError(err)

	result, err = tx.Table("users").Insert(map[string]any{"name": "commit"})
	s.Require().NoError(err)
	s.Equal(int64(1), result.RowsAffected)
	s.Require().NoError(tx.Commit())

	exists, err = facades.DB().Table("users").Where("name", "commit").Exists()
	s.Require().NoError(err)
	s.True(exists)
}

func (s *DBTestSuite) TestQueryHelpers() {
	result, err := facades.DB().Table("users").Insert([]map[string]any{
		{"name": "alpha"},
		{"name": "beta", "mail": "beta@example.com"},
		{"name": "gamma"},
	})
	s.Require().NoError(err)
	s.Equal(int64(3), result.RowsAffected)

	count, err := facades.DB().Table("users").Count()
	s.Require().NoError(err)
	s.Equal(int64(3), count)

	exists, err := facades.DB().Table("users").Where("name", "beta").Exists()
	s.Require().NoError(err)
	s.True(exists)

	notExists, err := facades.DB().Table("users").Where("name", "delta").DoesntExist()
	s.Require().NoError(err)
	s.True(notExists)

	withMailCount, err := facades.DB().Table("users").WhereNotNull("mail").Count()
	s.Require().NoError(err)
	s.Equal(int64(1), withMailCount)
}

func (s *DBTestSuite) TestFirstOrAndFindOrFail() {
	_, err := facades.DB().Table("users").Insert(map[string]any{"name": "exists"})
	s.Require().NoError(err)

	var existing User
	s.Require().NoError(facades.DB().Table("users").Where("name", "exists").FirstOr(&existing, func() error {
		return errors.New("should not execute callback for existing record")
	}))
	s.Equal("exists", existing.Name)

	var missing User
	err = facades.DB().Table("users").Where("name", "missing").FirstOr(&missing, func() error {
		return errors.New("fallback")
	})
	s.Require().Error(err)
	s.Equal("fallback", err.Error())

	err = facades.DB().Table("users").Where("name", "missing").FindOrFail(&missing)
	s.Require().Error(err)
}

func (s *DBTestSuite) TestUpdateOrInsertAndPluck() {
	_, err := facades.DB().Table("users").Insert([]map[string]any{
		{"name": "alpha"},
		{"name": "beta"},
	})
	s.Require().NoError(err)

	_, err = facades.DB().Table("users").UpdateOrInsert(
		map[string]any{"name": "alpha"},
		map[string]any{"avatar": "alpha.png"},
	)
	s.Require().NoError(err)

	_, err = facades.DB().Table("users").UpdateOrInsert(
		map[string]any{"name": "gamma"},
		map[string]any{"avatar": "gamma.png"},
	)
	s.Require().NoError(err)

	var names []string
	s.Require().NoError(facades.DB().Table("users").OrderBy("id").Pluck("name", &names))
	s.Equal([]string{"alpha", "beta", "gamma"}, names)

	var avatar string
	s.Require().NoError(facades.DB().Table("users").Where("name", "alpha").Value("avatar", &avatar))
	s.Equal("alpha.png", avatar)

	count, err := facades.DB().Table("users").Count()
	s.Require().NoError(err)
	s.Equal(int64(3), count)
}

// TODO use orm.BaseModel when https://github.com/goravel/framework/pull/976 is merged
type User struct {
	ID        uint             `db:"id"`
	Name      string           `db:"name"`
	Avatar    string           `db:"avatar"`
	Alias     string           `db:"alias"`
	Mail      *string          `db:"mail"`
	CreatedAt *carbon.DateTime `db:"created_at"`
	UpdatedAt *carbon.DateTime `db:"updated_at"`
	DeletedAt *carbon.DateTime `db:"deleted_at"`
}
