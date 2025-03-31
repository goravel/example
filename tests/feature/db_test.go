//go:debug x509negativeserial=1

package feature

import (
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
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
