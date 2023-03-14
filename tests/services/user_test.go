package services

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/app/services"
	"goravel/bootstrap"
)

/*********************************
1. Init .env
2. Init DB
	users: id, name, avatar, created_at, updated_at
 ********************************/

type UserTestSuite struct {
	suite.Suite
	user services.User
}

func TestUserTestSuite(t *testing.T) {
	// If you run test through Mock, don't need this line.
	bootstrap.Boot()

	suite.Run(t, &UserTestSuite{
		user: services.NewUserImpl(),
	})
}

func (s *UserTestSuite) SetupTest() {

}

// Run test through a real environment(DB, redis, etc.), `bootstrap.Boot()` is required.
// This way is not perfect now(need init .env in the same directory, etc.)
// We will optimize it in the future and make a full document for it.(#73: https://github.com/goravel/goravel/issues/73)
func (s *UserTestSuite) TestCreateByConfig() {
	user, err := s.user.Create()
	s.Nil(err)
	s.True(user.ID > 0)
	s.Equal("name", user.Name)
}
