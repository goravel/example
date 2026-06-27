package services

import (
	"context"
	"testing"

	"github.com/goravel/framework/contracts/mail"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/testing/mock"
	testifymock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

type MockExamplesTestSuite struct {
	suite.Suite
}

func TestMockExamplesTestSuite(t *testing.T) {
	suite.Run(t, &MockExamplesTestSuite{})
}

func (s *MockExamplesTestSuite) SetupTest() {}

func (s *MockExamplesTestSuite) TestApp() {
	mockFactory := mock.Factory()
	mockApp := mockFactory.App()
	mockApp.EXPECT().CurrentLocale(context.Background()).Return("en").Once()

	s.Equal("en", AppCurrentLocale())
}

func (s *MockExamplesTestSuite) TestArtisan() {
	mockFactory := mock.Factory()
	mockArtisan := mockFactory.Artisan()
	mockArtisan.EXPECT().Call("list").Return(nil).Once()

	s.Nil(ArtisanCall())
}

func (s *MockExamplesTestSuite) TestAuth() {
	mockFactory := mock.Factory()
	mockCtx := mockFactory.Context()
	mockAuth := mockFactory.Auth(mockCtx)
	mockAuth.EXPECT().Logout().Return(nil).Once()

	s.Nil(Auth(mockCtx))
}

func (s *MockExamplesTestSuite) TestCache() {
	mockFactory := mock.Factory()
	mockCache := mockFactory.Cache()
	mockCache.EXPECT().Put("name", "goravel", testifymock.Anything).Return(nil).Once()
	mockCache.EXPECT().Get("name", "test").Return("Goravel").Once()

	res := Cache()

	s.Equal("Goravel", res)
}

func (s *MockExamplesTestSuite) TestConfig() {
	mockFactory := mock.Factory()
	mockConfig := mockFactory.Config()
	mockConfig.EXPECT().GetString("app.name", "test").Return("Goravel").Once()

	res := Config()

	s.Equal("Goravel", res)
}

func (s *MockExamplesTestSuite) TestCrypt() {
	mockFactory := mock.Factory()
	mockCrypt := mockFactory.Crypt()
	mockCrypt.EXPECT().EncryptString("Goravel").Return("test", nil).Once()
	mockCrypt.EXPECT().DecryptString("test").Return("Goravel", nil).Once()

	res, err := Crypt("Goravel")

	s.Equal("Goravel", res)
	s.Nil(err)
}

func (s *MockExamplesTestSuite) TestEvent() {
	mockFactory := mock.Factory()
	mockEvent := mockFactory.Event()
	mockTask := mockFactory.EventTask()
	mockEvent.EXPECT().Job(testifymock.Anything, testifymock.Anything).Return(mockTask).Once()
	mockTask.EXPECT().Dispatch().Return(nil).Once()

	s.Nil(Event())
}

func (s *MockExamplesTestSuite) TestGate() {
	mockFactory := mock.Factory()
	mockGate := mockFactory.Gate()
	mockGate.EXPECT().Allows("update-post", map[string]any{
		"post": "test",
	}).Return(true).Once()

	s.True(Gate())
}

func (s *MockExamplesTestSuite) TestGrpc() {
	mockFactory := mock.Factory()
	mockGrpc := mockFactory.Grpc()
	mockGrpc.EXPECT().Client(context.Background(), "user").Return(nil, nil).Once()

	s.Nil(Grpc())
}

func (s *MockExamplesTestSuite) TestHash() {
	mockFactory := mock.Factory()
	mockHash := mockFactory.Hash()
	mockHash.EXPECT().Make("Goravel").Return("test", nil).Once()

	res, err := Hash()

	s.Equal("test", res)
	s.Nil(err)
}

func (s *MockExamplesTestSuite) TestLang() {
	mockFactory := mock.Factory()
	mockLang := mockFactory.Lang(context.Background())
	mockLang.EXPECT().Get("name").Return("Goravel").Once()

	s.Equal("Goravel", Lang(context.Background()))
}

func (s *MockExamplesTestSuite) TestLog() {
	mockFactory := mock.Factory()
	mockFactory.Log()

	s.NotPanics(func() {
		Log()
	})
}

func (s *MockExamplesTestSuite) TestMail() {
	mockFactory := mock.Factory()
	mockMail := mockFactory.Mail()
	mockMail.EXPECT().From(mail.Address{Address: "example@example.com", Name: "example"}).Return(mockMail).Once()
	mockMail.EXPECT().To([]string{"example@example.com"}).Return(mockMail).Once()
	mockMail.EXPECT().Subject("Subject").Return(mockMail).Once()
	mockMail.EXPECT().Content(mail.Content{Html: "<h1>Hello Goravel</h1>"}).Return(mockMail).Once()
	mockMail.EXPECT().Send().Return(nil).Once()

	s.Nil(Mail())
}

func (s *MockExamplesTestSuite) TestOrm() {
	mockFactory := mock.Factory()
	mockOrm := mockFactory.Orm()
	mockOrmQuery := mockFactory.OrmQuery()
	mockOrm.EXPECT().Query().Return(mockOrmQuery).Times(2)
	mockOrmQuery.EXPECT().Create(testifymock.Anything).Return(nil).Once()
	mockOrmQuery.EXPECT().Where("id = ?", 1).Return(mockOrmQuery).Once()
	mockOrmQuery.EXPECT().Find(testifymock.Anything).Return(nil).Once()

	s.Nil(Orm())
}

func (s *MockExamplesTestSuite) TestOrmTransaction() {
	mockFactory := mock.Factory()
	mockOrm := mockFactory.Orm()
	mockOrm.EXPECT().Transaction(testifymock.Anything).Return(nil).Once()

	s.Nil(OrmTransaction())
}

func (s *MockExamplesTestSuite) TestQueue() {
	mockFactory := mock.Factory()
	mockQueue := mockFactory.Queue()
	mockTask := mockFactory.QueueTask()
	mockQueue.EXPECT().Job(testifymock.Anything, testifymock.Anything).Return(mockTask).Once()
	mockTask.EXPECT().Dispatch().Return(nil).Once()

	s.Nil(Queue())
}

func (s *MockExamplesTestSuite) TestStorage() {
	mockFactory := mock.Factory()
	mockStorage := mockFactory.Storage()
	mockDriver := mockFactory.StorageDriver()
	mockStorage.EXPECT().WithContext(context.Background()).Return(mockDriver).Once()
	file, _ := filesystem.NewFile("1.txt")
	mockDriver.EXPECT().PutFile("file", file).Return("", nil).Once()

	path, err := Storage()

	s.Equal("", path)
	s.Nil(err)
}

func (s *MockExamplesTestSuite) TestValidation() {
	mockFactory := mock.Factory()
	mockValidation := mockFactory.Validation()
	mockValidator := mockFactory.ValidationValidator()
	mockErrors := mockFactory.ValidationErrors()
	mockValidation.EXPECT().Make(
		context.Background(),
		map[string]any{"a": "b"},
		map[string]any{"a": "required"},
	).Return(mockValidator, nil).Once()
	mockValidator.EXPECT().Errors().Return(mockErrors).Once()
	mockErrors.EXPECT().One("a").Return("error").Once()

	result := Validation()

	s.Equal("error", result)
}

func (s *MockExamplesTestSuite) TestView() {
	mockFactory := mock.Factory()
	mockView := mockFactory.View()
	mockView.EXPECT().Exists("welcome.tmpl").Return(true).Once()

	s.True(View())
}
