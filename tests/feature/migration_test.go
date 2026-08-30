package feature

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cast"

	contractsdriver "github.com/goravel/framework/contracts/database/driver"
	contractsschema "github.com/goravel/framework/contracts/database/schema"
	"github.com/goravel/framework/database/migration"
	databaseschema "github.com/goravel/framework/database/schema"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/mysql"
	"github.com/goravel/sqlite"
	sqlitefacades "github.com/goravel/sqlite/facades"
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

func (s *MigrationTestSuite) TestMigrator_NonDefaultConnectionLedgerRows() {
	const (
		defaultName   = "migration_e2e_default"
		reportingName = "migration_e2e_reporting"
	)

	dir := s.T().TempDir()

	scope, err := tests.OverrideConfig(map[string]any{
		"database.default":                      defaultName,
		"database.connections." + defaultName:   sqliteConnectionConfig(defaultName, filepath.Join(dir, defaultName+".db")),
		"database.connections." + reportingName: sqliteConnectionConfig(reportingName, filepath.Join(dir, reportingName+".db")),
	})
	s.Require().NoError(err)
	defer func() { s.Require().NoError(scope.Restore()) }()

	// Build a constrained schema holding only the runtime migrations, so the
	// migrator doesn't re-run the app's full bootstrap migration set on the
	// throwaway default DB.
	v := facades.Config().Get("database.connections." + defaultName + ".via")
	via, ok := v.(func() (contractsdriver.Driver, error))
	s.Require().Truef(ok, "via closure for %s has unexpected type %T", defaultName, v)
	defaultDriver, err := via()
	s.Require().NoError(err)

	schema, err := databaseschema.NewSchema(facades.Config(), facades.Log(), facades.Orm(), defaultDriver, nil)
	s.Require().NoError(err)

	schema.Register([]contractsschema.Migration{
		newRuntimeConnectionMigration(schema, reportingName, "20260826160940_create_migration_e2e_users", "migration_e2e_users"),
		newRuntimeConnectionMigration(schema, reportingName, "20260826161140_create_migration_e2e_user_tokens", "migration_e2e_user_tokens"),
	})

	migrationTable := facades.Config().GetString("database.migrations.table")
	migrator := migration.NewMigrator(nil, schema, migrationTable)
	s.Require().NoError(migrator.Run())

	// Both tables created on the non-default (reporting) connection.
	s.True(schema.Connection(reportingName).HasTable("migration_e2e_users"))
	s.True(schema.Connection(reportingName).HasTable("migration_e2e_user_tokens"))

	// The non-default connection never received the migrations ledger.
	s.False(schema.Connection(reportingName).HasTable(migrationTable))

	// Both ledger rows were written to the default connection.
	for _, signature := range []string{
		"20260826160940_create_migration_e2e_users",
		"20260826161140_create_migration_e2e_user_tokens",
	} {
		count, err := facades.DB().Table(migrationTable).Where("migration", signature).Count()
		s.Require().NoError(err)
		s.Equal(int64(1), count)
	}
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
	ranOutput, err := s.CaptureArtisanOutput("--no-ansi migrate:status")
	s.NoError(err)
	s.Contains(ranOutput, "Migration name")
	s.Contains(ranOutput, "Batch / Status")
	s.Contains(ranOutput, "20210101000001_create_users_table")
	s.Contains(ranOutput, "20210101000002_create_jobs_table")
	s.Contains(ranOutput, "20250330911908_add_columns_to_users_table")
	s.Contains(ranOutput, "20250331093125_alert_columns_of_users_table")
	s.Contains(ranOutput, "Ran")

	s.NoError(facades.Artisan().Call("--no-ansi migrate:reset"))

	pendingOutput, err := s.CaptureArtisanOutput("--no-ansi migrate:status")
	s.NoError(err)
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

// sqliteConnectionConfig returns the config for a non-default sqlite connection
// backed by the given database file. The connection is resolved lazily via the
// sqlite facade's via closure, mirroring the framework's
// database.connections.<name> config shape.
func sqliteConnectionConfig(name, database string) map[string]any {
	return map[string]any{
		"database": database,
		"prefix":   "",
		"singular": false,
		"via": func() (contractsdriver.Driver, error) {
			return sqlitefacades.Sqlite(name)
		},
	}
}

// runtimeConnectionMigration is a schema.Migration whose Up/Down run on a
// non-default connection via the migrator's SetConnection flow.
type runtimeConnectionMigration struct {
	schema     contractsschema.Schema
	connection string
	signature  string
	table      string
}

func newRuntimeConnectionMigration(schema contractsschema.Schema, connection, signature, table string) *runtimeConnectionMigration {
	return &runtimeConnectionMigration{schema: schema, connection: connection, signature: signature, table: table}
}

func (r *runtimeConnectionMigration) Signature() string { return r.signature }

func (r *runtimeConnectionMigration) Connection() string { return r.connection }

func (r *runtimeConnectionMigration) Up() error {
	return r.schema.Create(r.table, func(table contractsschema.Blueprint) {
		table.String("name")
	})
}

func (r *runtimeConnectionMigration) Down() error {
	return r.schema.DropIfExists(r.table)
}
