package main

import (
	"os"
	"testing"

	"goravel/app/facades"
	"goravel/bootstrap"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/suite"
)

func TestMain(m *testing.M) {
	_ = bootstrap.Boot()

	os.Exit(m.Run())
}

type MainTestSuite struct {
	suite.Suite
}

func TestMainTestSuite(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

func (s *MainTestSuite) SetupSuite() {}

func (s *MainTestSuite) TearDownTest() {
	res := facades.Process().Run("git", "checkout", ".")
	s.False(res.Failed())

	res = facades.Process().Run("git", "clean", "-fd")
	s.False(res.Failed())

	res = facades.Process().Run("go", "mod", "tidy")
	s.False(res.Failed())
}

func (s *MainTestSuite) TestMakeCommand() {
	s.NoError(facades.Artisan().Call("make:command TestCommand"))
	s.True(file.Contains(path.Bootstrap("commands.go"), "&commands.TestCommand{},"))
	s.True(file.Exists(path.App("console", "commands", "test_command.go")))
	s.True(file.Contains(path.Bootstrap("app.go"), "WithCommands(Commands)."))
}

func (s *MainTestSuite) TestMakeFilter() {
	s.NoError(facades.Artisan().Call("make:filter TestFilter"))
	s.True(file.Contains(path.Bootstrap("filters.go"), "&filters.TestFilter{},"))
	s.True(file.Exists(path.App("filters", "test_filter.go")))
	s.True(file.Contains(path.Bootstrap("app.go"), "WithFilters(Filters)."))
}

func (s *MainTestSuite) TestMakeJob() {
	s.NoError(facades.Artisan().Call("make:job TestJob"))
	s.True(file.Contains(path.Bootstrap("jobs.go"), "&jobs.TestJob{},"))
	s.True(file.Exists(path.App("jobs", "test_job.go")))
	s.True(file.Contains(path.Bootstrap("app.go"), "WithJobs(Jobs)."))
}

func (s *MainTestSuite) TestMakeMiddleware() {
	s.NoError(facades.Artisan().Call("make:middleware TestMiddleware"))
	s.True(file.Exists(path.App("http", "middleware", "test_middleware.go")))
}

func (s *MainTestSuite) TestMakeMigration() {
	carbon.SetTestNow(carbon.Parse("2026-01-02 12:34:56"))
	defer carbon.ClearTestNow()
	s.NoError(facades.Artisan().Call("make:migration create_users_table -m User"))
	s.True(file.Contains(path.Bootstrap("migrations.go"), "&migrations.M20260102123456CreateUsersTable{},"))
	s.True(file.Contains(path.Bootstrap("app.go"), "WithMigrations(Migrations)."))
	s.True(file.Exists(path.Migration("20260102123456_create_users_table.go")))
	s.True(file.Contains(path.Migration("20260102123456_create_users_table.go"), `table.BigIncrements("id")`))
}

func (s *MainTestSuite) TestMakeProvider() {
	s.NoError(facades.Artisan().Call("make:provider TestProvider"))
	s.True(file.Contains(path.Bootstrap("providers.go"), "&providers.TestProvider{},"))
	s.True(file.Exists(path.App("providers", "test_provider.go")))
	s.True(file.Contains(path.Bootstrap("app.go"), "WithProviders(Providers)."))
}

func (s *MainTestSuite) TestMakeRule() {
	s.NoError(facades.Artisan().Call("make:rule TestRule"))
	s.True(file.Contains(path.Bootstrap("rules.go"), "&rules.TestRule{},"))
	s.True(file.Exists(path.App("rules", "test_rule.go")))
	s.True(file.Contains(path.Bootstrap("app.go"), "WithRules(Rules)."))
}

func (s *MainTestSuite) TestMakeSeeder() {
	s.NoError(facades.Artisan().Call("make:seeder TestSeeder"))
	s.True(file.Contains(path.Bootstrap("seeders.go"), "&seeders.TestSeeder{},"))
	s.True(file.Exists(path.Database("seeders", "test_seeder.go")))
	s.True(file.Contains(path.Bootstrap("app.go"), "WithSeeders(Seeders)."))
}

func (s *MainTestSuite) TestMakeView() {
	s.NoError(facades.Artisan().Call("make:view TestView"))
	s.True(file.Exists(path.View("TestView.tmpl")))
}
