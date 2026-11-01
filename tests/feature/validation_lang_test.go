package feature

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

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
}

func (s *ValidationLangTestSuite) removeValidationLangFiles() {
	for _, locale := range []string{"en", "cn", "fr", "de", "es"} {
		dir := filepath.Join(s.langPath, locale)
		_ = os.Remove(filepath.Join(dir, "validation.json"))
		_ = os.Remove(dir)
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

	s.setValidationMessage("fr", "required", "Custom required message for :attribute.")

	scope, err := tests.OverrideConfig(map[string]any{
		"app.locale": "fr",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

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

	s.setValidationMessage("de", "required", "Custom required message for :attribute.")

	scope, err := tests.OverrideConfig(map[string]any{
		"app.locale": "de",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

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
	messages["required"] = "Spanish: The :attribute field is required."
	messages["email"] = "Spanish: The :attribute field must be a valid email address."

	esData, err := json.MarshalIndent(messages, "", "  ")
	s.Require().NoError(err)

	esDir := filepath.Join(s.langPath, "es")
	s.Require().NoError(os.MkdirAll(esDir, 0o755))
	s.Require().NoError(os.WriteFile(filepath.Join(esDir, "validation.json"), esData, 0o644))

	scope, err := tests.OverrideConfig(map[string]any{
		"app.locale": "es",
	})
	s.Require().NoError(err)
	defer func() { s.NoError(scope.Restore()) }()

	validator, err := facades.Validation().Make(context.Background(),
		map[string]any{"name": ""},
		map[string]any{"name": "required"},
	)
	s.Require().NoError(err)
	s.True(validator.Fails())
	s.Equal("Spanish: The name field is required.", validator.Errors().One("name"))
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

// setValidationMessage writes one rule into lang/<locale>/validation.json, creating it if absent.
// locale must be one no other suite reads: the translation cache is per-locale and never expires.
func (s *ValidationLangTestSuite) setValidationMessage(locale, rule, message string) {
	validationFile := filepath.Join(s.langPath, locale, "validation.json")

	messages := make(map[string]any)
	if content, err := os.ReadFile(validationFile); err == nil {
		s.Require().NoError(json.Unmarshal(content, &messages))
	}

	messages[rule] = message

	modified, err := json.MarshalIndent(messages, "", "  ")
	s.Require().NoError(err)
	s.Require().NoError(os.MkdirAll(filepath.Dir(validationFile), 0o755))
	s.Require().NoError(os.WriteFile(validationFile, modified, 0o644))
}
