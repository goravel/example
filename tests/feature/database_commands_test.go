package feature

import (
	"io"
	"os"
	"testing"

	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

type DatabaseCommandsTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestDatabaseCommandsTestSuite(t *testing.T) {
	suite.Run(t, &DatabaseCommandsTestSuite{})
}

func (s *DatabaseCommandsTestSuite) SetupTest() {
	s.RefreshDatabase()
}

func (s *DatabaseCommandsTestSuite) TestCommandDBSeed() {
	s.NoError(facades.Artisan().Call("--no-ansi db:seed"))

	var user models.User
	s.NoError(facades.Orm().Query().Where("mail", "migration@goravel.dev").FirstOrFail(&user))
	s.Equal("migration", user.Name)
}

func (s *DatabaseCommandsTestSuite) TestCommandDBShow() {
	output := s.captureArtisanOutput("--no-ansi db:show")
	s.Contains(output, "Database")
	s.Contains(output, "Tables")
	s.Contains(output, "users")
}

func (s *DatabaseCommandsTestSuite) TestCommandDBTable() {
	output := s.captureArtisanOutput("--no-ansi db:table users")
	s.Contains(output, "users")
	s.Contains(output, "Columns")
	s.Contains(output, "id")
}

func (s *DatabaseCommandsTestSuite) TestCommandDBWipe() {
	s.NoError(facades.Artisan().Call("--no-ansi db:wipe"))

	s.False(facades.Schema().HasTable("users"))
	s.False(facades.Schema().HasTable("jobs"))
	s.False(facades.Schema().HasTable("failed_jobs"))
}

func (s *DatabaseCommandsTestSuite) TestCommandMakeModel() {
	modelPath := path.Model("test_db_command_model.go")
	s.NoError(file.Remove(modelPath))
	s.T().Cleanup(func() {
		s.NoError(file.Remove(modelPath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:model TestDbCommandModel"))
	s.True(file.Exists(modelPath))
	s.True(file.Contains(modelPath, "type TestDbCommandModel struct"))
}

func (s *DatabaseCommandsTestSuite) TestCommandMakeObserver() {
	observerPath := path.App("observers", "test_db_command_observer.go")
	s.NoError(file.Remove(observerPath))
	s.T().Cleanup(func() {
		s.NoError(file.Remove(observerPath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:observer TestDbCommandObserver"))
	s.True(file.Exists(observerPath))
	s.True(file.Contains(observerPath, "type TestDbCommandObserver struct"))
}

func (s *DatabaseCommandsTestSuite) TestCommandMakeFactory() {
	factoryPath := path.Database("factories", "test_db_command_factory.go")
	s.NoError(file.Remove(factoryPath))
	s.T().Cleanup(func() {
		s.NoError(file.Remove(factoryPath))
	})

	s.NoError(facades.Artisan().Call("--no-ansi make:factory TestDbCommandFactory"))
	s.True(file.Exists(factoryPath))
	s.True(file.Contains(factoryPath, "type TestDbCommandFactory struct"))
}

func (s *DatabaseCommandsTestSuite) TestCommandMakeSeeder() {
	seederPath := path.Database("seeders", "test_db_command_seeder.go")
	seedersBootstrapPath := path.Bootstrap("seeders.go")

	seedersBootstrapContent, err := os.ReadFile(seedersBootstrapPath)
	if err != nil {
		s.T().Fatalf("read %s failed: %v", seedersBootstrapPath, err)
	}

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

func (s *DatabaseCommandsTestSuite) captureArtisanOutput(command string) string {
	return color.CaptureOutput(func(_ io.Writer) {
		s.NoError(facades.Artisan().Call(command))
	})
}
