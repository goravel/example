package feature

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	_ "unsafe"

	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

//go:linkname translationLoaded github.com/goravel/framework/translation.loaded
var translationLoaded map[string]map[string]map[string]any

type ValidationLangTestSuite struct {
	suite.Suite
	tests.TestCase

	langPath string
}

func TestValidationLangTestSuite(t *testing.T) {
	suite.Run(t, &ValidationLangTestSuite{})
}

func (s *ValidationLangTestSuite) SetupTest() {
	s.langPath = facades.App().LangPath()
	s.removeValidationLangFiles()
}

func (s *ValidationLangTestSuite) TearDownTest() {
	s.removeValidationLangFiles()
	translationLoaded = make(map[string]map[string]map[string]any)
}

func (s *ValidationLangTestSuite) removeValidationLangFiles() {
	for _, locale := range []string{"en", "cn"} {
		_ = os.Remove(filepath.Join(s.langPath, locale, "validation.json"))
	}
}

func (s *ValidationLangTestSuite) TestDefaultMessagesWithoutPublishing() {
	validator, err := facades.Validation().Make(context.Background(),
		map[string]any{"name": ""},
		map[string]any{"name": "required"},
	)
	s.Require().NoError(err)
	s.True(validator.Fails())
	s.Equal("The name field is required.", validator.Errors().One("name"))
}

func (s *ValidationLangTestSuite) TestLangPublishCommand() {
	output, err := s.CaptureArtisanOutput("lang:publish")
	s.Require().NoError(err)
	s.Contains(output, "Publishing complete")

	targetFile := filepath.Join(s.langPath, "en", "validation.json")
	s.True(file.Exists(targetFile))

	content, err := os.ReadFile(targetFile)
	s.Require().NoError(err)
	s.Contains(string(content), `"required"`)
	s.Contains(string(content), `"email"`)
}

func (s *ValidationLangTestSuite) TestLangPublishForce() {
	s.Require().NoError(facades.Artisan().Call("lang:publish"))

	targetFile := filepath.Join(s.langPath, "en", "validation.json")
	s.Require().True(file.Exists(targetFile))

	output, err := s.CaptureArtisanOutput("lang:publish")
	s.Require().NoError(err)
	s.Contains(output, "Publishing complete")

	s.Require().NoError(facades.Artisan().Call("lang:publish --force"))
}

func (s *ValidationLangTestSuite) TestOverridePublishedMessages() {
	s.Require().NoError(facades.Artisan().Call("lang:publish"))

	s.setValidationMessage("en", "required", "Custom required message for :attribute.")

	validator, err := facades.Validation().Make(context.Background(),
		map[string]any{"name": ""},
		map[string]any{"name": "required"},
	)
	s.Require().NoError(err)
	s.True(validator.Fails())
	s.Equal("Custom required message for name.", validator.Errors().One("name"))
}

func (s *ValidationLangTestSuite) TestNonOverriddenMessagesUseDefaults() {
	s.Require().NoError(facades.Artisan().Call("lang:publish"))

	s.setValidationMessage("en", "required", "Custom required message for :attribute.")

	validator, err := facades.Validation().Make(context.Background(),
		map[string]any{"email": "not-an-email"},
		map[string]any{"email": "email"},
	)
	s.Require().NoError(err)
	s.True(validator.Fails())
	s.Equal("The email field must be a valid email address.", validator.Errors().One("email"))
}

func (s *ValidationLangTestSuite) TestLanguageSwitch() {
	s.Require().NoError(facades.Artisan().Call("lang:publish"))

	validationFile := filepath.Join(s.langPath, "en", "validation.json")
	content, err := os.ReadFile(validationFile)
	s.Require().NoError(err)

	var messages map[string]any
	s.Require().NoError(json.Unmarshal(content, &messages))
	messages["required"] = "Chinese: The :attribute field is required."
	messages["email"] = "Chinese: The :attribute field must be a valid email address."

	cnData, err := json.MarshalIndent(messages, "", "  ")
	s.Require().NoError(err)

	cnDir := filepath.Join(s.langPath, "cn")
	s.Require().NoError(os.MkdirAll(cnDir, 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(cnDir, "validation.json"), cnData, 0o644))

	scope, err := tests.OverrideConfig(map[string]any{
		"app.locale": "cn",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	validator, err := facades.Validation().Make(context.Background(),
		map[string]any{"name": ""},
		map[string]any{"name": "required"},
	)
	s.Require().NoError(err)
	s.True(validator.Fails())
	s.Equal("Chinese: The name field is required.", validator.Errors().One("name"))
}

func (s *ValidationLangTestSuite) TestFallbackLocale() {
	s.Require().NoError(facades.Artisan().Call("lang:publish"))

	scope, err := tests.OverrideConfig(map[string]any{
		"app.locale": "jp",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	validator, err := facades.Validation().Make(context.Background(),
		map[string]any{"name": ""},
		map[string]any{"name": "required"},
	)
	s.Require().NoError(err)
	s.True(validator.Fails())
	s.Equal("The name field is required.", validator.Errors().One("name"))
}

func (s *ValidationLangTestSuite) setValidationMessage(locale, rule, message string) {
	validationFile := filepath.Join(s.langPath, locale, "validation.json")
	content, err := os.ReadFile(validationFile)
	s.Require().NoError(err)

	var messages map[string]any
	s.Require().NoError(json.Unmarshal(content, &messages))
	messages[rule] = message

	modified, err := json.MarshalIndent(messages, "", "  ")
	s.Require().NoError(err)
	s.Require().NoError(os.WriteFile(validationFile, modified, 0o644))
}
