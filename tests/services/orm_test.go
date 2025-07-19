//go:debug x509negativeserial=1

package services

import (
	"github.com/goravel/framework/contracts/database/orm"
	"testing"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/suite"

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
}

// TearDownTest will run after each test in the suite.
func (s *OrmTestSuite) TearDownTest() {
}

func (s *OrmTestSuite) TestOrm() {

	personEducationBuilder := facades.Orm().Query().Model(&models.Person{})

	personEducationBuilder = personEducationBuilder.Where("type", 1)
	personEducationBuilder = personEducationBuilder.Where("street_code", "c110101017")
	personEducationBuilder = personEducationBuilder.Where("education", 2)

	type educationQueryItem struct {
		Query orm.Query
		Slug  string
	}

	var queryBuilder = make([]educationQueryItem, 0)

	queryBuilder = append(queryBuilder, educationQueryItem{
		Query: personEducationBuilder,
		Slug:  "total",
	})
	queryBuilder = append(queryBuilder, educationQueryItem{
		Query: personEducationBuilder.Where("sex", 3),
		Slug:  "female",
	})

	for _, queryBuilderItem := range queryBuilder {
		var person models.Person
		queryBuilderItem.Query.Where("graduate", 4).First(&person)
	}
}
