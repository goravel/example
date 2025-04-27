package feature

import (
	"fmt"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/suite"
	"goravel/tests"
	"testing"
	"time"
)

type TransactionsSuite struct {
	suite.Suite
	tests.TestCase
}

func TestTransactionsSuite(t *testing.T) {
	database, err := facades.Testing().Docker().Database("postgres")
	if err != nil {
		panic(err)
	}
	fmt.Println(database)

	if err := database.Build(); err != nil {
		panic(err)
	}

	if err := database.Migrate(); err != nil {
		panic(err)
	}
	fmt.Println(database.Config().Host)
	fmt.Println(database.Config().Username)
	fmt.Println(database.Config().Password)
	fmt.Println(database.Config().Database)
	fmt.Println(database.Config().Port)
	time.Sleep(120 * time.Second)
	suite.Run(t, new(TransactionsSuite))
	if err := file.Remove("storage"); err != nil {
		panic(err)
	}
	if err := database.Shutdown(); err != nil {
		panic(err)
	}
}

// SetupTest will run before each test in the suite.
func (s *TransactionsSuite) SetupTest() {
	fmt.Println("SetupTest")
}

// TearDownTest will run after each test in the suite.
func (s *TransactionsSuite) TearDownTest() {
	s.RefreshDatabase()
}

func (s *TransactionsSuite) Testdb() {
	fmt.Println("Test example")
}
