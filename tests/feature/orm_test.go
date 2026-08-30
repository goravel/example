package feature

import (
	"path/filepath"
	"sync"
	"testing"

	contractsorm "github.com/goravel/framework/contracts/database/orm"
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

func (s *OrmTestSuite) TestConnection_ReflectsRequestedConnection() {
	const name = "orm_connection_reflect"
	database, cleanup := s.registerConnection(name)
	defer cleanup()

	cold := facades.Orm().Connection(name) // lazily builds the connection
	s.assertConnection(cold, name, database)

	warm := facades.Orm().Connection(name) // cached path — before #1540 this reported the default
	s.assertConnection(warm, name, database)
}

func (s *OrmTestSuite) TestConnection_ConcurrentSafe() {
	const name = "orm_connection_concurrent"
	database, cleanup := s.registerConnection(name)
	defer cleanup()

	const n = 20
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		results = make([]contractsorm.Orm, 0, n)
		start   = make(chan struct{})
	)

	// Capture the shared facade once so all goroutines exercise the same
	// singleton's cached Connection path concurrently.
	orm := facades.Orm()

	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			<-start
			connection := orm.Connection(name)
			mu.Lock()
			results = append(results, connection)
			mu.Unlock()
		}()
	}

	close(start)
	wg.Wait()

	s.Require().Len(results, n)
	for _, orm := range results {
		s.assertConnection(orm, name, database)
	}
}

// registerConnection registers a non-default sqlite connection backed by a temp
// file and returns its expected database name plus a restore func.
func (s *OrmTestSuite) registerConnection(name string) (database string, cleanup func()) {
	database = filepath.Join(s.T().TempDir(), name+".db")

	scope, err := tests.OverrideConfig(map[string]any{
		"database.connections." + name: sqliteConnectionConfig(name, database),
	})
	s.Require().NoError(err)

	return database, func() { s.Require().NoError(scope.Restore()) }
}

func (s *OrmTestSuite) assertConnection(orm contractsorm.Orm, name, database string) {
	s.Require().NotNil(orm)
	s.Equal(name, orm.Name())
	s.Equal(database, orm.DatabaseName())
	s.Equal(name, orm.Config().Connection)
	s.Equal(database, orm.Config().Database)
}
