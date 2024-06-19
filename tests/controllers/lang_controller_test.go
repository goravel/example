package controllers

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"

	"goravel/tests"
)

/*
*****************************************
We need add the lang folder in the testing package for now, will optimize it in v1.15
*****************************************
*/
type LangControllerTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestLangControllerTestSuite(t *testing.T) {
	suite.Run(t, &LangControllerTestSuite{})
}

// SetupTest will run before each test in the suite.
func (s *LangControllerTestSuite) SetupTest() {
}

// TearDownTest will run after each test in the suite.
func (s *LangControllerTestSuite) TearDownTest() {
}

func (s *LangControllerTestSuite) TestIndex() {
	tests := []struct {
		name           string
		lang           string
		expectResponse string
	}{
		{
			name:           "use default lang",
			expectResponse: "{\"current_locale\":\"en\",\"fallback\":\"Goravel 是一个基于 Go 语言的 Web 开发框架\",\"name\":\"Goravel Framework\"}",
		},
		{
			name:           "lang is cn",
			lang:           "cn",
			expectResponse: "{\"current_locale\":\"cn\",\"fallback\":\"Goravel 是一个基于 Go 语言的 Web 开发框架\",\"name\":\"Goravel 框架\"}",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			resp, err := http.Get(route(fmt.Sprintf("/lang?lang=%s", test.lang)))
			s.Require().NoError(err)

			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)

			s.Require().NoError(err)
			s.Equal(http.StatusOK, resp.StatusCode)
			s.Equal(test.expectResponse, string(body))
		})
	}
}
