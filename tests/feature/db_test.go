//go:debug x509negativeserial=1

package feature

import (
	"os"
	"testing"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/models"
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

func (s *DBTestSuite) TestCommandDBSeed() {
	s.NoError(facades.Artisan().Call("--no-ansi db:seed"))

	var user models.User
	s.NoError(facades.Orm().Query().Where("mail", "migration@goravel.dev").FirstOrFail(&user))
	s.Equal("migration", user.Name)
}

func (s *DBTestSuite) TestCommandDBShow() {
	output, err := s.CaptureArtisanOutput("--no-ansi db:show")
	s.NoError(err)
	s.Contains(output, "Database")
	s.Contains(output, "Tables")
	s.Contains(output, "users")
}

func (s *DBTestSuite) TestCommandDBTable() {
	output, err := s.CaptureArtisanOutput("--no-ansi db:table users")
	s.NoError(err)
	s.Contains(output, "users")
	s.Contains(output, "Columns")
	s.Contains(output, "id")
}

func (s *DBTestSuite) TestCommandDBWipe() {
	s.NoError(facades.Artisan().Call("--no-ansi db:wipe"))

	s.False(facades.Schema().HasTable("users"))
	s.False(facades.Schema().HasTable("jobs"))
	s.False(facades.Schema().HasTable("failed_jobs"))
}

func (s *DBTestSuite) TestCommandMakeModel() {
	modelPath := path.Model("test_db_command_model.go")
	s.NoError(file.Remove(modelPath))
	s.T().Cleanup(func() {
		s.NoError(file.Remove(modelPath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:model TestDbCommandModel"))
	s.True(file.Exists(modelPath))
	s.True(file.Contains(modelPath, "type TestDbCommandModel struct"))
}

func (s *DBTestSuite) TestCommandMakeObserver() {
	observerPath := path.App("observers", "test_db_command_observer.go")
	s.NoError(file.Remove(observerPath))
	s.T().Cleanup(func() {
		s.NoError(file.Remove(observerPath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:observer TestDbCommandObserver"))
	s.True(file.Exists(observerPath))
	s.True(file.Contains(observerPath, "type TestDbCommandObserver struct"))
}

func (s *DBTestSuite) TestCommandMakeFactory() {
	factoryPath := path.Database("factories", "test_db_command_factory.go")
	s.NoError(file.Remove(factoryPath))
	s.T().Cleanup(func() {
		s.NoError(file.Remove(factoryPath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:factory TestDbCommandFactory"))
	s.True(file.Exists(factoryPath))
	s.True(file.Contains(factoryPath, "type TestDbCommandFactory struct"))
}

func (s *DBTestSuite) TestCommandMakeSeeder() {
	seederPath := path.Database("seeders", "test_db_command_seeder.go")
	seedersBootstrapPath := path.Bootstrap("seeders.go")

	seedersBootstrapContent, err := os.ReadFile(seedersBootstrapPath)
	if err != nil {
		s.T().Fatalf("read %s failed: %v", seedersBootstrapPath, err)
	}

	s.NotContains(string(seedersBootstrapContent), "&seeders.TestDbCommandSeeder{}")

	s.NoError(file.Remove(seederPath))

	s.T().Cleanup(func() {
		s.NoError(file.Remove(seederPath))
		if err := os.WriteFile(seedersBootstrapPath, seedersBootstrapContent, 0o644); err != nil {
			s.T().Fatalf("restore %s failed: %v", seedersBootstrapPath, err)
		}
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:seeder TestDbCommandSeeder"))
	s.True(file.Exists(seederPath))
	s.True(file.Contains(seederPath, "type TestDbCommandSeeder struct"))

	updatedSeedersBootstrap, err := os.ReadFile(seedersBootstrapPath)
	s.Require().NoError(err)

	s.Contains(string(updatedSeedersBootstrap), "&seeders.TestDbCommandSeeder{}")
}

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
