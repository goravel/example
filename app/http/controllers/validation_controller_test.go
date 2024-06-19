package controllers

import (
	"testing"

	"github.com/goravel/framework/contracts/http"
	testingmock "github.com/goravel/framework/testing/mock"
	"github.com/goravel/gin"
	"github.com/stretchr/testify/suite"

	"goravel/app/models"
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
	mockContext.EXPECT().Request().Return(mockRequest).Once()
	mockRequest.EXPECT().Validate(map[string]string{
		"name": "required",
	}).Return(mockValidator, nil).Once()
	mockValidator.EXPECT().Fails().Return(false).Once()
	var user models.User
	mockValidator.EXPECT().Bind(&user).Run(func(user any) {
		user.(*models.User).Name = "Goravel"
	}).Return(nil).Once()
	mockContext.EXPECT().Response().Return(mockResponse).Once()
	mockResponseStatus := mockFactory.ResponseStatus()
	mockResponse.EXPECT().Success().Return(mockResponseStatus).Once()

	resp := &gin.JsonResponse{}
	mockResponseStatus.EXPECT().Json(http.Json{
		"name": "Goravel",
	}).Return(resp).Once()

	s.Equal(resp, NewValidationController().Json(mockContext))

	mockContext.AssertExpectations(s.T())
	mockRequest.AssertExpectations(s.T())
	mockResponse.AssertExpectations(s.T())
	mockValidator.AssertExpectations(s.T())
	mockResponseStatus.AssertExpectations(s.T())
}
