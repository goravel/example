package services

import (
	"testing"

	"github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
)

type UserTestSuite struct {
	suite.Suite
	user User
}

func TestUserTestSuite(t *testing.T) {
	suite.Run(t, &UserTestSuite{
		user: NewUserImpl(),
	})
}

func (s *UserTestSuite) SetupTest() {

}

func (s *UserTestSuite) TestCreate() {
	mockOrm, mockDb, _, _ := mock.Orm()
	mockOrm.On("Query").Return(mockDb).Once()
	mockDb.On("Create", &models.User{
		Name:   "name",
		Avatar: "avatar",
	}).Return(nil).Once()
	user, err := s.user.Create()
	s.Nil(err)
	s.Equal("name", user.Name)
	mockOrm.AssertExpectations(s.T())
	mockDb.AssertExpectations(s.T())
}
