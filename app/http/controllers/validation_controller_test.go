package controllers

import (
	"testing"

	"github.com/goravel/framework/contracts/http"
	"github.com/goravel/framework/support/carbon"
	testingmock "github.com/goravel/framework/testing/mock"
	"github.com/goravel/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"goravel/app/http/requests"
)

type ValidationControllerTestSuite struct {
	suite.Suite
}

func TestValidationControllerTestSuite(t *testing.T) {
	suite.Run(t, &ValidationControllerTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *ValidationControllerTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *ValidationControllerTestSuite) TearDownTest() {
}

func (s *ValidationControllerTestSuite) TestJson() {
	mockFactory := testingmock.Factory()
	mockContext := mockFactory.Context()
	mockRequest := mockFactory.ContextRequest()
	mockResponse := mockFactory.ContextResponse()
	mockValidator := mockFactory.ValidationValidator()
	mockContext.EXPECT().WithValue("ctx", "context").Once()
	mockContext.EXPECT().Request().Return(mockRequest).Once()
	mockRequest.EXPECT().Validate(map[string]any{
		"context":      "required",
		"name":         "required",
		"date":         "required|date",
		"items.*.name": "sometimes|required|string",
		"meta":         "sometimes|map",
		"meta.name":    "sometimes|required|string",
	}, mock.AnythingOfType("validation.Option")).Return(mockValidator, nil).Once()
	mockValidator.EXPECT().Fails().Return(false).Once()
	var user User
	mockValidator.EXPECT().Bind(&user).Run(func(user any) {
		user.(*User).Context = "ctx_context"
		user.(*User).Name = "Goravel"
		user.(*User).Date = carbon.NewDateTime(carbon.Parse("2024-07-08 22:34:31"))
		user.(*User).Age = 1
		user.(*User).Items = []requests.ValidationItem{{Name: "item1"}}
		user.(*User).Meta = map[string]any{"source": "api"}
	}).Return(nil).Once()
	mockContext.EXPECT().Response().Return(mockResponse).Once()
	mockResponseStatus := mockFactory.ResponseStatus()
	mockResponse.EXPECT().Success().Return(mockResponseStatus).Once()

	resp := &gin.JsonResponse{}
	mockResponseStatus.EXPECT().Json(http.Json{
		"context": "ctx_context",
		"name":    "Goravel",
		"date":    "2024-07-08 22:34:31",
		"age":     1,
		"items":   []requests.ValidationItem{{Name: "item1"}},
		"meta":    map[string]any{"source": "api"},
	}).Return(resp).Once()

	s.Equal(resp, NewValidationController().Json(mockContext))

	mockContext.AssertExpectations(s.T())
	mockRequest.AssertExpectations(s.T())
	mockResponse.AssertExpectations(s.T())
	mockValidator.AssertExpectations(s.T())
	mockResponseStatus.AssertExpectations(s.T())
}
