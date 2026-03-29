package feature

import (
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cast"

	"github.com/goravel/framework/support/color"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/mysql"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

type MigrationTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestMigrationTestSuite(t *testing.T) {
	suite.Run(t, &MigrationTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *MigrationTestSuite) SetupTest() {
	s.RefreshDatabase()
}

// TearDownTest will run after each test in the suite.
func (s *MigrationTestSuite) TearDownTest() {
}

func (s *MigrationTestSuite) TestChange() {
	if facades.Schema().Orm().Config().Driver == sqlite.Name {
		s.T().Skip("sqlite does not support change column")
	}

	columns, err := facades.Schema().GetColumns("users")

	s.Require().NoError(err)

	var mailExists bool
	for _, column := range columns {
		if column.Name == "alias" {
			s.Contains(column.Default, "test")
		}

		if column.Name == "mail" {
			mailExists = true
		}
	}

	s.True(mailExists)
}

func (s *MigrationTestSuite) TestFirst_After() {
	if facades.Schema().Orm().Config().Driver != mysql.Name {
		s.T().Skip("only mysql supports first and after")
	}

	columns, err := facades.Schema().GetColumns("users")
	s.Require().NoError(err)

	s.Equal("mail", columns[0].Name)
	s.Equal("alias", columns[3].Name)
}

func (s *MigrationTestSuite) TestMigrate() {
	s.True(facades.Schema().HasTable("users"))
}

func (s *MigrationTestSuite) TestCommandMigrate() {
	total, err := s.migrationCount()
	s.Require().NoError(err)

	s.NoError(facades.Artisan().Call("--no-ansi migrate:reset"))

	count, err := s.migrationCount()
	s.NoError(err)
	s.Zero(count)
	s.False(facades.Schema().HasTable("users"))

	s.NoError(facades.Artisan().Call("--no-ansi migrate"))

	count, err = s.migrationCount()
	s.NoError(err)
	s.Equal(total, count)

	s.True(facades.Schema().HasTable("users"))
	s.True(facades.Schema().HasTable("jobs"))
	s.True(facades.Schema().HasTable("failed_jobs"))
	s.True(s.columnExists("users", "mail"))
}

func (s *MigrationTestSuite) TestCommandMigrateReset() {
	s.True(facades.Schema().HasTable("users"))

	s.NoError(facades.Artisan().Call("--no-ansi migrate:reset"))

	count, err := s.migrationCount()
	s.NoError(err)
	s.Zero(count)

	s.False(facades.Schema().HasTable("users"))
	s.False(facades.Schema().HasTable("jobs"))
	s.False(facades.Schema().HasTable("failed_jobs"))
}

func (s *MigrationTestSuite) TestCommandMigrateRefresh() {
	total, err := s.migrationCount()
	s.Require().NoError(err)

	s.NoError(facades.Artisan().Call("--no-ansi migrate:refresh"))
	afterRefresh, err := s.migrationCount()
	s.NoError(err)
	s.Equal(total, afterRefresh)
	s.True(facades.Schema().HasTable("users"))
	s.True(s.columnExists("users", "mail"))

	s.NoError(facades.Artisan().Call("--no-ansi migrate:refresh --step 1"))

	afterStepRefresh, err := s.migrationCount()
	s.NoError(err)
	s.Equal(total, afterStepRefresh)

	lastBatch, err := s.latestMigrationBatch()
	s.NoError(err)
	s.Equal(lastBatch, 2)

	s.True(facades.Schema().HasTable("users"))
	s.True(s.columnExists("users", "mail"))
}

func (s *MigrationTestSuite) TestCommandMigrateFresh() {
	total, err := s.migrationCount()
	s.Require().NoError(err)

	s.NoError(facades.Artisan().Call("--no-ansi migrate:fresh --seed --seeder DatabaseSeeder"))

	count, err := s.migrationCount()
	s.NoError(err)
	s.Equal(total, count)

	s.True(facades.Schema().HasTable("users"))
	s.True(facades.Schema().HasTable("jobs"))
	s.True(facades.Schema().HasTable("failed_jobs"))
	s.True(s.columnExists("users", "mail"))

	var user models.User
	s.NoError(facades.Orm().Query().Where("mail", "migration@goravel.dev").FirstOrFail(&user))
	s.Equal("migration", user.Name)
}

func (s *MigrationTestSuite) TestCommandMigrateRollback() {
	total, err := s.migrationCount()
	s.Require().NoError(err)

	s.NoError(facades.Artisan().Call("--no-ansi migrate:rollback"))
	afterDefaultRollback, err := s.migrationCount()
	s.NoError(err)
	s.Zero(afterDefaultRollback)

	s.RefreshDatabase()

	s.NoError(facades.Artisan().Call("--no-ansi migrate:rollback --step 1"))
	afterStepRollback, err := s.migrationCount()
	s.NoError(err)
	s.Equal(total-1, afterStepRollback)

	s.RefreshDatabase()

	s.NoError(facades.Artisan().Call("--no-ansi migrate:rollback --step 1"))
	s.NoError(facades.Artisan().Call("--no-ansi migrate"))

	latestBatch, err := s.latestMigrationBatch()
	s.NoError(err)
	s.Equal(2, latestBatch)

	s.NoError(facades.Artisan().Call("--no-ansi migrate:rollback --batch " + cast.ToString(latestBatch)))
	afterBatchRollback, err := s.migrationCount()
	s.NoError(err)
	s.Equal(total-1, afterBatchRollback)
}

func (s *MigrationTestSuite) TestCommandMigrateStatus() {
	ranOutput := s.captureArtisanOutput("--no-ansi migrate:status")
	s.Contains(ranOutput, "Migration name")
	s.Contains(ranOutput, "Batch / Status")
	s.Contains(ranOutput, "20210101000001_create_users_table")
	s.Contains(ranOutput, "20210101000002_create_jobs_table")
	s.Contains(ranOutput, "20250330911908_add_columns_to_users_table")
	s.Contains(ranOutput, "20250331093125_alert_columns_of_users_table")
	s.Contains(ranOutput, "Ran")

	s.NoError(facades.Artisan().Call("--no-ansi migrate:reset"))

	pendingOutput := s.captureArtisanOutput("--no-ansi migrate:status")
	s.Contains(pendingOutput, "Migration name")
	s.Contains(pendingOutput, "Batch / Status")
	s.Contains(pendingOutput, "20210101000001_create_users_table")
	s.Contains(pendingOutput, "20210101000002_create_jobs_table")
	s.Contains(pendingOutput, "20250330911908_add_columns_to_users_table")
	s.Contains(pendingOutput, "20250331093125_alert_columns_of_users_table")
	s.Contains(pendingOutput, "Pending")
}

func (s *MigrationTestSuite) TestCommandMakeMigration() {
	migrationsPath := path.Bootstrap("migrations.go")
	originalContent, err := os.ReadFile(migrationsPath)
	if err != nil {
		s.T().Fatalf("read %s failed: %v", migrationsPath, err)
	}

	s.T().Cleanup(func() {
		if err := os.WriteFile(migrationsPath, originalContent, 0o644); err != nil {
			s.T().Fatalf("restore %s failed: %v", migrationsPath, err)
		}
	})

	beforeFiles := s.listMigrationFiles()

	driver := facades.Orm().Config().Driver
	migrationName := "test_" + driver + "_" + cast.ToString(time.Now().UnixNano())
	s.NoError(facades.Artisan().Call("--no-ansi make:migration " + migrationName))

	afterFiles := s.listMigrationFiles()
	var createdFiles []string
	for item := range afterFiles {
		if _, ok := beforeFiles[item]; !ok {
			createdFiles = append(createdFiles, item)
		}
	}

	s.Require().NotEmpty(createdFiles)

	migrationPath := path.Migration(createdFiles[0])
	s.Require().FileExists(migrationPath)

	s.T().Cleanup(func() {
		if migrationPath != "" {
			s.NoError(file.Remove(migrationPath))
		}
	})

	content, err := os.ReadFile(migrationPath)
	s.Require().NoError(err)

	re := regexp.MustCompile(`type\s+(M[^\s]+)\s+struct`)
	matches := re.FindStringSubmatch(string(content))
	s.Require().Len(matches, 2)

	structName := matches[1]
	updatedBootstrap, err := os.ReadFile(migrationsPath)
	s.Require().NoError(err)
	s.Contains(string(updatedBootstrap), "&migrations."+structName+"{}")
}

func (s *MigrationTestSuite) TestTableComment() {
	if facades.Schema().Orm().Config().Driver == sqlite.Name || facades.Schema().Orm().Config().Driver == sqlserver.Name {
		s.T().Skip("sqlite and sqlserver does not support table comment")
	}

	tables, err := facades.Schema().GetTables()
	s.Require().NoError(err)

	for _, table := range tables {
		if table.Name == "users" {
			s.Equal("user table", table.Comment)
		}
	}
}

func (s *MigrationTestSuite) migrationCount() (int64, error) {
	table := facades.Config().GetString("database.migrations.table")
	return facades.DB().Table(table).Count()
}

func (s *MigrationTestSuite) latestMigrationBatch() (int, error) {
	table := facades.Config().GetString("database.migrations.table")

	var batch int
	err := facades.DB().Table(table).OrderByDesc("batch").Limit(1).Value("batch", &batch)
	if err != nil {
		return 0, err
	}

	return batch, nil
}

func (s *MigrationTestSuite) columnExists(table, column string) bool {
	return facades.Schema().HasColumn(table, column)
}

func (s *MigrationTestSuite) captureArtisanOutput(command string) string {
	return color.CaptureOutput(func(_ io.Writer) {
		s.NoError(facades.Artisan().Call(command))
	})
}

func (s *MigrationTestSuite) listMigrationFiles() map[string]struct{} {
	migrationDir := path.Migration()
	entries, err := os.ReadDir(migrationDir)
	s.NoError(err)

	files := make(map[string]struct{})
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".go") {
			files[entry.Name()] = struct{}{}
		}
	}

	return files
}
