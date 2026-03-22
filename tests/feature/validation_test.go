package feature

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/goravel/framework/support/http"
	"github.com/goravel/framework/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/app/models"
	"goravel/tests"
)

type ValidationTestSuite struct {
	suite.Suite
	tests.TestCase
}

func TestValidationTestSuite(t *testing.T) {
	suite.Run(t, &ValidationTestSuite{})
}

func (s *ValidationTestSuite) SetupTest() {
	s.RefreshDatabase()
}

func (s *ValidationTestSuite) TestRules() {
	cases := []struct {
		name     string
		rule     string
		errRule  string
		passData map[string]any
		failData map[string]any
		message  string
	}{
		{name: "required", rule: "required", errRule: "required", passData: map[string]any{"f": "ok"}, failData: map[string]any{"f": ""}, message: "required failed"},
		{name: "required_if", rule: "required_if:cond,yes", errRule: "required_if", passData: map[string]any{"cond": "yes", "f": "ok"}, failData: map[string]any{"cond": "yes", "f": ""}, message: "required_if failed"},
		{name: "required_unless", rule: "required_unless:cond,no", errRule: "required_unless", passData: map[string]any{"cond": "yes", "f": "ok"}, failData: map[string]any{"cond": "yes", "f": ""}, message: "required_unless failed"},
		{name: "required_with", rule: "required_with:cond", errRule: "required_with", passData: map[string]any{"cond": "x", "f": "ok"}, failData: map[string]any{"cond": "x", "f": ""}, message: "required_with failed"},
		{name: "required_with_all", rule: "required_with_all:a,b", errRule: "required_with_all", passData: map[string]any{"a": "x", "b": "y", "f": "ok"}, failData: map[string]any{"a": "x", "b": "y", "f": ""}, message: "required_with_all failed"},
		{name: "required_without", rule: "required_without:cond", errRule: "required_without", passData: map[string]any{"f": "ok"}, failData: map[string]any{"f": ""}, message: "required_without failed"},
		{name: "required_without_all", rule: "required_without_all:a,b", errRule: "required_without_all", passData: map[string]any{"f": "ok"}, failData: map[string]any{"f": ""}, message: "required_without_all failed"},
		{name: "int", rule: "int", errRule: "int", passData: map[string]any{"f": 1}, failData: map[string]any{"f": "a"}, message: "int failed"},
		{name: "uint", rule: "uint", errRule: "uint", passData: map[string]any{"f": 1}, failData: map[string]any{"f": -1}, message: "uint failed"},
		{name: "bool", rule: "bool", errRule: "bool", passData: map[string]any{"f": true}, failData: map[string]any{"f": "not_bool"}, message: "bool failed"},
		{name: "string", rule: "string", errRule: "string", passData: map[string]any{"f": "ok"}, failData: map[string]any{"f": 1}, message: "string failed"},
		{name: "float", rule: "float", errRule: "float", passData: map[string]any{"f": 1.2}, failData: map[string]any{"f": "a"}, message: "float failed"},
		{name: "slice", rule: "slice", errRule: "slice", passData: map[string]any{"f": []int{1, 2}}, failData: map[string]any{"f": "a"}, message: "slice failed"},
		{name: "in", rule: "in:a,b", errRule: "in", passData: map[string]any{"f": "a"}, failData: map[string]any{"f": "c"}, message: "in failed"},
		{name: "not_in", rule: "not_in:a,b", errRule: "not_in", passData: map[string]any{"f": "c"}, failData: map[string]any{"f": "a"}, message: "not_in failed"},
		{name: "starts_with", rule: "starts_with:go", errRule: "starts_with", passData: map[string]any{"f": "goravel"}, failData: map[string]any{"f": "oravel"}, message: "starts_with failed"},
		{name: "ends_with", rule: "ends_with:vel", errRule: "ends_with", passData: map[string]any{"f": "goravel"}, failData: map[string]any{"f": "gorave"}, message: "ends_with failed"},
		{name: "between", rule: "between:1,10", errRule: "between", passData: map[string]any{"f": 5}, failData: map[string]any{"f": 11}, message: "between failed"},
		{name: "max", rule: "max:10", errRule: "max", passData: map[string]any{"f": 10}, failData: map[string]any{"f": 11}, message: "max failed"},
		{name: "min", rule: "min:2", errRule: "min", passData: map[string]any{"f": 2}, failData: map[string]any{"f": 1}, message: "min failed"},
		{name: "eq", rule: "eq:5", errRule: "eq", passData: map[string]any{"f": "5"}, failData: map[string]any{"f": "4"}, message: "eq failed"},
		{name: "ne", rule: "ne:5", errRule: "ne", passData: map[string]any{"f": "4"}, failData: map[string]any{"f": "5"}, message: "ne failed"},
		{name: "lt", rule: "lt:5", errRule: "lt", passData: map[string]any{"f": 4}, failData: map[string]any{"f": 6}, message: "lt failed"},
		{name: "gt", rule: "gt:5", errRule: "gt", passData: map[string]any{"f": 6}, failData: map[string]any{"f": 4}, message: "gt failed"},
		{name: "len", rule: "len:3", errRule: "len", passData: map[string]any{"f": "abc"}, failData: map[string]any{"f": "ab"}, message: "len failed"},
		{name: "min_len", rule: "min_len:3", errRule: "min_len", passData: map[string]any{"f": "abc"}, failData: map[string]any{"f": "ab"}, message: "min_len failed"},
		{name: "max_len", rule: "max_len:3", errRule: "max_len", passData: map[string]any{"f": "abc"}, failData: map[string]any{"f": "abcd"}, message: "max_len failed"},
		{name: "email", rule: "email", errRule: "email", passData: map[string]any{"f": "a@b.com"}, failData: map[string]any{"f": "a"}, message: "email failed"},
		{name: "array", rule: "array", errRule: "array", passData: map[string]any{"f": []any{"a"}}, failData: map[string]any{"f": "a"}, message: "array failed"},
		{name: "map", rule: "map", errRule: "map", passData: map[string]any{"f": map[string]any{"a": 1}}, failData: map[string]any{"f": []any{1}}, message: "map failed"},
		{name: "eq_field", rule: "eq_field:other", errRule: "eq_field", passData: map[string]any{"f": "a", "other": "a"}, failData: map[string]any{"f": "a", "other": "b"}, message: "eq_field failed"},
		{name: "ne_field", rule: "ne_field:other", errRule: "ne_field", passData: map[string]any{"f": "a", "other": "b"}, failData: map[string]any{"f": "a", "other": "a"}, message: "ne_field failed"},
		{name: "gt_field", rule: "gt_field:other", errRule: "gt_field", passData: map[string]any{"f": 3, "other": 2}, failData: map[string]any{"f": 2, "other": 3}, message: "gt_field failed"},
		{name: "gte_field", rule: "gte_field:other", errRule: "gte_field", passData: map[string]any{"f": 3, "other": 3}, failData: map[string]any{"f": 2, "other": 3}, message: "gte_field failed"},
		{name: "lt_field", rule: "lt_field:other", errRule: "lt_field", passData: map[string]any{"f": 2, "other": 3}, failData: map[string]any{"f": 3, "other": 2}, message: "lt_field failed"},
		{name: "lte_field", rule: "lte_field:other", errRule: "lte_field", passData: map[string]any{"f": 3, "other": 3}, failData: map[string]any{"f": 4, "other": 3}, message: "lte_field failed"},
		{name: "date", rule: "date", errRule: "date", passData: map[string]any{"f": "2024-01-02 03:04:05"}, failData: map[string]any{"f": "x"}, message: "date failed"},
		{name: "gt_date", rule: "gt_date:2024-01-01 00:00:00", errRule: "gt_date", passData: map[string]any{"f": "2024-01-02 00:00:00"}, failData: map[string]any{"f": "2023-12-31 00:00:00"}, message: "gt_date failed"},
		{name: "lt_date", rule: "lt_date:2024-01-01 00:00:00", errRule: "lt_date", passData: map[string]any{"f": "2023-12-31 00:00:00"}, failData: map[string]any{"f": "2024-01-02 00:00:00"}, message: "lt_date failed"},
		{name: "gte_date", rule: "gte_date:2024-01-01 00:00:00", errRule: "gte_date", passData: map[string]any{"f": "2024-01-01 00:00:00"}, failData: map[string]any{"f": "2023-12-31 00:00:00"}, message: "gte_date failed"},
		{name: "lte_date", rule: "lte_date:2024-01-01 00:00:00", errRule: "lte_date", passData: map[string]any{"f": "2024-01-01 00:00:00"}, failData: map[string]any{"f": "2024-01-02 00:00:00"}, message: "lte_date failed"},
		{name: "alpha", rule: "alpha", errRule: "alpha", passData: map[string]any{"f": "abc"}, failData: map[string]any{"f": "a1"}, message: "alpha failed"},
		{name: "alpha_num", rule: "alpha_num", errRule: "alpha_num", passData: map[string]any{"f": "a1"}, failData: map[string]any{"f": "a-1"}, message: "alpha_num failed"},
		{name: "alpha_dash", rule: "alpha_dash", errRule: "alpha_dash", passData: map[string]any{"f": "a-1_"}, failData: map[string]any{"f": "a@1"}, message: "alpha_dash failed"},
		{name: "json", rule: "json", errRule: "json", passData: map[string]any{"f": `{"a":1}`}, failData: map[string]any{"f": "x"}, message: "json failed"},
		{name: "number", rule: "number", errRule: "number", passData: map[string]any{"f": "123"}, failData: map[string]any{"f": "abc"}, message: "number failed"},
		{name: "full_url", rule: "full_url", errRule: "full_url", passData: map[string]any{"f": "https://goravel.dev/docs?a=1"}, failData: map[string]any{"f": "goravel.dev"}, message: "full_url failed"},
		{name: "ip", rule: "ip", errRule: "ip", passData: map[string]any{"f": "127.0.0.1"}, failData: map[string]any{"f": "999.999.999.999"}, message: "ip failed"},
		{name: "ipv4", rule: "ipv4", errRule: "ipv4", passData: map[string]any{"f": "127.0.0.1"}, failData: map[string]any{"f": "2001:db8::1"}, message: "ipv4 failed"},
		{name: "ipv6", rule: "ipv6", errRule: "ipv6", passData: map[string]any{"f": "2001:db8::1"}, failData: map[string]any{"f": "127.0.0.1"}, message: "ipv6 failed"},
		{name: "regex", rule: "regex:^[a-z0-9]+$", errRule: "regex", passData: map[string]any{"f": "abc123"}, failData: map[string]any{"f": "ABC"}, message: "regex failed"},
		{name: "uuid", rule: "uuid", errRule: "uuid", passData: map[string]any{"f": "550e8400-e29b-41d4-a716-446655440000"}, failData: map[string]any{"f": "x"}, message: "uuid failed"},
		{name: "uuid3", rule: "uuid3", errRule: "uuid3", passData: map[string]any{"f": "f47ac10b-58cc-3372-a567-0e02b2c3d479"}, failData: map[string]any{"f": "550e8400-e29b-41d4-a716-446655440000"}, message: "uuid3 failed"},
		{name: "uuid4", rule: "uuid4", errRule: "uuid4", passData: map[string]any{"f": "550e8400-e29b-41d4-a716-446655440000"}, failData: map[string]any{"f": "f47ac10b-58cc-3372-a567-0e02b2c3d479"}, message: "uuid4 failed"},
		{name: "uuid5", rule: "uuid5", errRule: "uuid5", passData: map[string]any{"f": "987fbc97-4bed-5078-9f07-9141ba07c9f3"}, failData: map[string]any{"f": "550e8400-e29b-41d4-a716-446655440000"}, message: "uuid5 failed"},
	}

	for _, tc := range cases {
		s.Run(tc.name, func() {
			validator, err := facades.Validation().Make(context.Background(), tc.passData, map[string]any{
				"f": tc.rule,
			})
			s.Require().NoError(err)
			s.Require().NotNil(validator)
			s.False(validator.Fails(), tc.name+" pass case")

			validator, err = facades.Validation().Make(context.Background(), tc.failData, map[string]any{
				"f": tc.rule,
			}, validation.Messages(map[string]string{
				"f." + tc.errRule: tc.message,
			}))
			s.Require().NoError(err)
			s.Require().NotNil(validator)
			s.True(validator.Fails(), tc.name+" fail case")
			s.Equal(map[string]map[string]string{
				"f": {
					tc.errRule: tc.message,
				},
			}, validator.Errors().All())
		})
	}
}

func (s *ValidationTestSuite) TestFileRule() {
	tempDir := s.T().TempDir()
	txtPath := filepath.Join(tempDir, "f.txt")

	s.Require().NoError(os.WriteFile(txtPath, []byte("plain text"), 0o644))

	passBody := http.NewBody().SetField("rule", "file").SetField("message", "file failed").SetFiles(map[string][]string{
		"f": {txtPath},
	})
	passPayload, err := passBody.Build()
	s.Require().NoError(err)
	resp, err := s.Http(s.T()).WithHeader("Content-Type", passPayload.ContentType()).Post("/validation/upload", passPayload.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	failPayload, err := http.NewBody().SetField("rule", "file").SetField("message", "file failed").SetField("f", "not-file").Build()
	s.Require().NoError(err)
	resp, err = s.Http(s.T()).WithHeader("Content-Type", failPayload.ContentType()).Post("/validation/upload", failPayload.Reader())
	s.Require().NoError(err)
	resp.AssertBadRequest()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Contains(content, "file")
	s.Contains(content, "file failed")
}

func (s *ValidationTestSuite) TestImageRule() {
	tempDir := s.T().TempDir()
	txtPath := filepath.Join(tempDir, "f.txt")
	pngPath := filepath.Join(tempDir, "f.png")

	s.Require().NoError(os.WriteFile(txtPath, []byte("plain text"), 0o644))
	s.Require().NoError(os.WriteFile(pngPath, []byte{
		0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a,
		0x00, 0x00, 0x00, 0x0d, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1f, 0x15, 0xc4,
		0x89, 0x00, 0x00, 0x00, 0x0a, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9c, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0d, 0x0a, 0x2d, 0xb4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4e, 0x44, 0xae,
		0x42, 0x60, 0x82,
	}, 0o644))

	passBody := http.NewBody().SetField("rule", "image").SetField("message", "image failed").SetFiles(map[string][]string{
		"f": {pngPath},
	})
	passPayload, err := passBody.Build()
	s.Require().NoError(err)
	resp, err := s.Http(s.T()).WithHeader("Content-Type", passPayload.ContentType()).Post("/validation/upload", passPayload.Reader())
	s.Require().NoError(err)
	resp.AssertSuccessful()

	failPayload, err := http.NewBody().SetField("rule", "image").SetField("message", "image failed").SetFiles(map[string][]string{
		"f": {txtPath},
	}).Build()
	s.Require().NoError(err)
	resp, err = s.Http(s.T()).WithHeader("Content-Type", failPayload.ContentType()).Post("/validation/upload", failPayload.Reader())
	s.Require().NoError(err)
	resp.AssertBadRequest()

	content, err := resp.Content()
	s.Require().NoError(err)
	s.Contains(content, "image")
	s.Contains(content, "image failed")
}

func (s *ValidationTestSuite) TestFilters() {
	type filterCase struct {
		name       string
		filter     string
		passData   map[string]any
		passRule   string
		failData   map[string]any
		failRule   string
		message    string
		passAssert func(*testing.T, any)
	}

	stringEqCase := func(alias, in, out, failIn string) filterCase {
		return filterCase{
			name:     alias,
			filter:   alias,
			passData: map[string]any{"f": in},
			passRule: "eq:" + out,
			failData: map[string]any{"f": failIn},
			failRule: "eq:" + out,
			message:  alias + " failed",
			passAssert: func(t *testing.T, actual any) {
				require.IsType(t, "", actual)
				require.Equal(t, out, actual)
			},
		}
	}

	tests := []filterCase{
		{name: "int", filter: "int", passData: map[string]any{"f": "12"}, passRule: "required", failData: map[string]any{"f": "13"}, failRule: "eq:12", message: "int failed", passAssert: func(t *testing.T, actual any) { require.IsType(t, int(0), actual); require.Equal(t, 12, actual) }},
		{name: "toInt", filter: "toInt", passData: map[string]any{"f": "12"}, passRule: "required", failData: map[string]any{"f": "13"}, failRule: "eq:12", message: "toInt failed", passAssert: func(t *testing.T, actual any) { require.IsType(t, int(0), actual); require.Equal(t, 12, actual) }},
		{name: "uint", filter: "uint", passData: map[string]any{"f": "12"}, passRule: "required", failData: map[string]any{"f": "13"}, failRule: "eq:12", message: "uint failed", passAssert: func(t *testing.T, actual any) { require.IsType(t, uint(0), actual); require.Equal(t, uint(12), actual) }},
		{name: "toUint", filter: "toUint", passData: map[string]any{"f": "12"}, passRule: "required", failData: map[string]any{"f": "13"}, failRule: "eq:12", message: "toUint failed", passAssert: func(t *testing.T, actual any) { require.IsType(t, uint(0), actual); require.Equal(t, uint(12), actual) }},
		{name: "int64", filter: "int64", passData: map[string]any{"f": "12"}, passRule: "required", failData: map[string]any{"f": "13"}, failRule: "eq:12", message: "int64 failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, int64(0), actual)
			require.Equal(t, int64(12), actual)
		}},
		{name: "toInt64", filter: "toInt64", passData: map[string]any{"f": "12"}, passRule: "required", failData: map[string]any{"f": "13"}, failRule: "eq:12", message: "toInt64 failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, int64(0), actual)
			require.Equal(t, int64(12), actual)
		}},
		{name: "float", filter: "float", passData: map[string]any{"f": "1.5"}, passRule: "required", failData: map[string]any{"f": "2.5"}, failRule: "eq:1.5", message: "float failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, float64(0), actual)
			require.InDelta(t, 1.5, actual, 0.00001)
		}},
		{name: "toFloat", filter: "toFloat", passData: map[string]any{"f": "1.5"}, passRule: "eq:1.5", failData: map[string]any{"f": "2.5"}, failRule: "eq:1.5", message: "toFloat failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, float64(0), actual)
			require.InDelta(t, 1.5, actual, 0.00001)
		}},
		{name: "bool", filter: "bool", passData: map[string]any{"f": "true"}, passRule: "required", failData: map[string]any{"f": ""}, failRule: "accepted", message: "bool failed", passAssert: func(t *testing.T, actual any) { require.IsType(t, false, actual); require.Equal(t, true, actual) }},
		{name: "toBool", filter: "toBool", passData: map[string]any{"f": "true"}, passRule: "required", failData: map[string]any{"f": ""}, failRule: "accepted", message: "toBool failed", passAssert: func(t *testing.T, actual any) { require.IsType(t, false, actual); require.Equal(t, true, actual) }},
		stringEqCase("trim", "  Goravel  ", "Goravel", "  Laravel  "),
		stringEqCase("trimSpace", "  Goravel  ", "Goravel", "  Laravel  "),
		stringEqCase("ltrim", "  Goravel", "Goravel", "  Laravel"),
		stringEqCase("trimLeft", "  Goravel", "Goravel", "  Laravel"),
		stringEqCase("rtrim", "Goravel  ", "Goravel", "Laravel  "),
		stringEqCase("trimRight", "Goravel  ", "Goravel", "Laravel  "),
		{name: "integer", filter: "integer", passData: map[string]any{"f": "12"}, passRule: "eq:12", failData: map[string]any{"f": "13"}, failRule: "eq:12", message: "integer failed", passAssert: func(t *testing.T, actual any) { require.IsType(t, int(0), actual); require.Equal(t, 12, actual) }},
		stringEqCase("lower", "GoRavel", "goravel", "LarAvel"),
		stringEqCase("lowercase", "GoRavel", "goravel", "LarAvel"),
		stringEqCase("upper", "GoRavel", "GORAVEL", "Laravel"),
		stringEqCase("uppercase", "GoRavel", "GORAVEL", "Laravel"),
		stringEqCase("lcFirst", "Goravel", "goravel", "Laravel"),
		stringEqCase("lowerFirst", "Goravel", "goravel", "Laravel"),
		stringEqCase("ucFirst", "goravel", "Goravel", "laravel"),
		stringEqCase("upperFirst", "goravel", "Goravel", "laravel"),
		stringEqCase("ucWord", "goravel framework", "Goravel Framework", "laravel framework"),
		stringEqCase("upperWord", "goravel framework", "Goravel Framework", "laravel framework"),
		stringEqCase("camel", "goravel_framework", "goravelFramework", "laravel_framework"),
		stringEqCase("camelCase", "goravel_framework", "goravelFramework", "laravel_framework"),
		stringEqCase("snake", "goravelFramework", "goravel_framework", "laravelFramework"),
		stringEqCase("snakeCase", "goravelFramework", "goravel_framework", "laravelFramework"),
		{name: "escapeJs", filter: "escapeJs", passData: map[string]any{"f": `"`}, passRule: "required", failData: map[string]any{"f": ""}, failRule: "required", message: "escapeJs failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, "", actual)
			escaped := actual.(string)
			require.NotEqual(t, `"`, escaped)
			require.Contains(t, escaped, `\`)
		}},
		{name: "escapeJS", filter: "escapeJS", passData: map[string]any{"f": `"`}, passRule: "required", failData: map[string]any{"f": ""}, failRule: "required", message: "escapeJS failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, "", actual)
			escaped := actual.(string)
			require.NotEqual(t, `"`, escaped)
			require.Contains(t, escaped, `\`)
		}},
		{name: "escapeHtml", filter: "escapeHtml", passData: map[string]any{"f": "<b>x</b>"}, passRule: "required", failData: map[string]any{"f": ""}, failRule: "required", message: "escapeHtml failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, "", actual)
			escaped := actual.(string)
			require.Equal(t, "&lt;b&gt;x&lt;/b&gt;", escaped)
		}},
		{name: "escapeHTML", filter: "escapeHTML", passData: map[string]any{"f": "<b>x</b>"}, passRule: "required", failData: map[string]any{"f": ""}, failRule: "required", message: "escapeHTML failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, "", actual)
			escaped := actual.(string)
			require.Equal(t, "&lt;b&gt;x&lt;/b&gt;", escaped)
		}},
		{name: "str2ints", filter: "str2ints", passData: map[string]any{"f": "1,2,3"}, passRule: "len:3", failData: map[string]any{"f": "1,2"}, failRule: "len:3", message: "str2ints failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, []int{}, actual)
			require.Equal(t, []int{1, 2, 3}, actual)
		}},
		{name: "strToInts", filter: "strToInts", passData: map[string]any{"f": "1,2,3"}, passRule: "len:3", failData: map[string]any{"f": "1,2"}, failRule: "len:3", message: "strToInts failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, []int{}, actual)
			require.Equal(t, []int{1, 2, 3}, actual)
		}},
		{name: "str2time", filter: "str2time", passData: map[string]any{"f": "2024-01-02 03:04:05"}, passRule: "required", failData: map[string]any{"f": ""}, failRule: "after:2020-01-01", message: "str2time failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, time.Time{}, actual)
			require.False(t, actual.(time.Time).IsZero())
		}},
		{name: "strToTime", filter: "strToTime", passData: map[string]any{"f": "2024-01-02 03:04:05"}, passRule: "required", failData: map[string]any{"f": ""}, failRule: "after:2020-01-01", message: "strToTime failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, time.Time{}, actual)
			require.False(t, actual.(time.Time).IsZero())
		}},
		{name: "str2arr", filter: "str2arr", passData: map[string]any{"f": "a,b"}, passRule: "len:2", failData: map[string]any{"f": "a"}, failRule: "len:2", message: "str2arr failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, []string{}, actual)
			require.Equal(t, []string{"a", "b"}, actual)
		}},
		{name: "str2array", filter: "str2array", passData: map[string]any{"f": "a,b"}, passRule: "len:2", failData: map[string]any{"f": "a"}, failRule: "len:2", message: "str2array failed", passAssert: func(t *testing.T, actual any) {
			require.IsType(t, []string{}, actual)
			require.Equal(t, []string{"a", "b"}, actual)
		}},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			validator, err := facades.Validation().Make(context.Background(), tt.passData, map[string]any{
				"f": tt.passRule,
			}, validation.Filters(map[string]any{
				"f": tt.filter,
			}))
			s.Require().NoError(err)
			s.Require().NotNil(validator)
			s.False(validator.Fails(), tt.name+" pass case")

			if tt.passAssert != nil {
				var data struct {
					F any `json:"f" form:"f"`
				}
				err = validator.Bind(&data)
				s.Require().NoError(err)
				tt.passAssert(s.T(), data.F)
			}

			validator, err = facades.Validation().Make(context.Background(), tt.failData, map[string]any{
				"f": tt.failRule,
			}, validation.Filters(map[string]any{
				"f": tt.filter,
			}), validation.Messages(map[string]string{
				"f." + ruleKey(tt.failRule): tt.message,
			}))
			s.Require().NoError(err)
			s.True(validator.Fails(), tt.name+" fail case")
			errors := validator.Errors().All()
			s.Contains(errors, "f")
			s.Equal(tt.message, errors["f"][ruleKey(tt.failRule)])
		})
	}
}

func (s *ValidationTestSuite) TestCustomRule() {
	s.Require().NoError(facades.Orm().Query().Model(&models.User{}).Create(map[string]any{
		"name": "Goravel",
		"mail": "team@goravel.dev",
	}))

	validator, err := facades.Validation().Make(context.Background(), map[string]any{
		"f": "Goravel",
	}, map[string]any{
		"f": "custom_exists:users,name",
	}, validation.Messages(map[string]string{
		"f.custom_exists": "custom_exists failed",
	}))
	s.Require().NoError(err)
	s.False(validator.Fails())

	validator, err = facades.Validation().Make(context.Background(), map[string]any{
		"f": "Unknown",
	}, map[string]any{
		"f": "custom_exists:users,name",
	}, validation.Messages(map[string]string{
		"f.custom_exists": "custom_exists failed",
	}))
	s.Require().NoError(err)
	s.True(validator.Fails())
	s.Equal(map[string]map[string]string{
		"f": {
			"custom_exists": "custom_exists failed",
		},
	}, validator.Errors().All())
}

func (s *ValidationTestSuite) TestCustomFilter() {
	validator, err := facades.Validation().Make(context.Background(), map[string]any{
		"f": "goravel",
	}, map[string]any{
		"f": "eq:goravel_suffix",
	}, validation.Filters(map[string]any{
		"f": "append_suffix:_suffix",
	}), validation.Messages(map[string]string{
		"f.eq": "append_suffix failed",
	}))
	s.Require().NoError(err)
	s.False(validator.Fails())

	validator, err = facades.Validation().Make(context.Background(), map[string]any{
		"f": "goravel",
	}, map[string]any{
		"f": "eq:goravel_fail",
	}, validation.Filters(map[string]any{
		"f": "append_suffix:_suffix",
	}), validation.Messages(map[string]string{
		"f.eq": "append_suffix failed",
	}))
	s.Require().NoError(err)
	s.True(validator.Fails())
	s.Equal(map[string]map[string]string{
		"f": {
			"eq": "append_suffix failed",
		},
	}, validator.Errors().All())
}

func (s *ValidationTestSuite) TestValidateJson() {
	s.Run("success", func() {
		resp, err := s.Http(s.T()).Post("/validation/json", strings.NewReader(`{"context":"ctx","name":"Goravel","date":"2024-07-08 18:33:32","age":1,"items":[{"name":"item1"},{"name":"item2"}],"meta":{"name":"api","source":"api","trace":"abc"}}`))
		s.Require().NoError(err)
		resp.AssertSuccessful()

		result, err := resp.Json()
		s.Require().NoError(err)
		s.Equal(map[string]any{
			"context": "ctx_context",
			"name":    "Goravel",
			"date":    "2024-07-08 18:33:32",
			"age":     float64(1),
			"items":   []any{map[string]any{"name": "item1"}, map[string]any{"name": "item2"}},
			"meta":    map[string]any{"name": "api", "source": "api", "trace": "abc"},
		}, result)
	})

	s.Run("fail", func() {
		resp, err := s.Http(s.T()).Post("/validation/json", strings.NewReader(`{"context":"ctx","name":"Goravel","date":"2024-07-08 18:33:32","age":1,"items":[{"name":"item1"},{"name":""}],"meta":{"name":"","source":"api","trace":"abc"}}`))
		s.Require().NoError(err)
		resp.AssertBadRequest()

		result, err := resp.Json()
		s.Require().NoError(err)
		s.Equal(map[string]any{
			"items.1.name": map[string]any{
				"required": "The items.1.name field is required.",
			},
			"meta.name": map[string]any{
				"required": "The meta.name field is required.",
			},
		}, result["message"])
	})
}

func (s *ValidationTestSuite) TestValidateForm() {
	s.Run("success", func() {
		payload, err := http.NewBody().SetField("context", "ctx").SetField("name", "Goravel").SetField("age", 1).Build()
		s.Require().NoError(err)
		resp, err := s.Http(s.T()).WithHeader("Content-Type", payload.ContentType()).Post("/validation/form", payload.Reader())
		s.Require().NoError(err)
		resp.AssertSuccessful()

		content, err := resp.Content()
		s.Require().NoError(err)
		s.Equal("{\"age\":1,\"context\":\"ctx_context\",\"name\":\"Goravel\"}", content)
	})

	s.Run("fail", func() {
		payload, err := http.NewBody().SetField("context", "ctx").SetField("name", "").SetField("age", 1).Build()
		s.Require().NoError(err)
		resp, err := s.Http(s.T()).WithHeader("Content-Type", payload.ContentType()).Post("/validation/form", payload.Reader())
		s.Require().NoError(err)
		resp.AssertBadRequest()

		content, err := resp.Content()
		s.Require().NoError(err)
		s.Equal("{\"message\":{\"name\":{\"required\":\"The name field is required.\"}}}", content)
	})
}

func (s *ValidationTestSuite) TestValidateRequest() {
	s.Run("success", func() {
		resp, err := s.Http(s.T()).Post("/validation/request", strings.NewReader(`{"name":" Goravel ","context":"ctx","date":"2024-07-08 18:33:32","tags":["tag1","tag2"],"scores":[1,2],"items":[{"name":"item1"},{"name":"item2"}],"meta":{"name":"api","source":"api","trace":"abc"},"code":123456,"age":1}`))
		s.Require().NoError(err)
		resp.AssertSuccessful()

		result, err := resp.Json()
		s.Require().NoError(err)
		s.Equal(map[string]any{
			"context": "ctx_context",
			"name":    "Goravel",
			"date":    "2024-07-08 18:33:32",
			"tags":    []any{"tag1", "tag2"},
			"scores":  []any{float64(1), float64(2)},
			"items":   []any{map[string]any{"name": "item1"}, map[string]any{"name": "item2"}},
			"meta":    map[string]any{"name": "api", "source": "api", "trace": "abc"},
			"code":    float64(123456),
			"age":     float64(1),
		}, result)
	})

	s.Run("fail", func() {
		resp, err := s.Http(s.T()).Post("/validation/request", strings.NewReader(`{"name":" Goravel ","context":"ctx","date":"2024-07-08 18:33:32","tags":["tag1","tag2"],"scores":[1,2],"items":[{"name":"item1"},{"name":""}],"meta":{"name":"","source":"api","trace":"abc"},"code":123456,"age":1}`))
		s.Require().NoError(err)
		resp.AssertBadRequest()

		result, err := resp.Json()
		s.Require().NoError(err)
		s.Equal(map[string]any{
			"items.1.name": map[string]any{
				"required": "The items.1.name field is required.",
			},
			"meta.name": map[string]any{
				"required": "The meta.name field is required.",
			},
		}, result["message"])
	})
}

func (s *ValidationTestSuite) TestValidateMake() {
	s.Run("success", func() {
		validator, err := facades.Validation().Make(context.Background(), map[string]any{
			"context": "ctx",
			"name":    " Goravel ",
			"date":    "2024-07-08 18:33:32",
			"tags":    []any{"tag1", "tag2"},
			"scores":  []any{1, 2},
			"items": []any{
				map[string]any{"name": "item1"},
				map[string]any{"name": "item2"},
			},
			"meta": map[string]any{
				"name":   "api",
				"source": "api",
				"trace":  "abc",
			},
			"code": 123456,
			"age":  1,
		}, map[string]any{
			"name":         "required",
			"context":      "required",
			"tags.*":       "required|string",
			"scores.*":     "required|int",
			"items.*.name": "sometimes|required|string",
			"meta":         "sometimes|map",
			"meta.name":    "sometimes|required|string",
			"date":         "required|date",
			"code":         `required|regex:^\d{4,6}$`,
		}, validation.Filters(map[string]any{
			"name": "trim",
		}))
		s.Require().NoError(err)
		s.False(validator.Fails())
		s.Equal(map[string]any{
			"context": "ctx",
			"name":    "Goravel",
			"date":    "2024-07-08 18:33:32",
			"tags":    []any{"tag1", "tag2"},
			"scores":  []any{1, 2},
			"items": []any{
				map[string]any{"name": "item1"},
				map[string]any{"name": "item2"},
			},
			"meta": map[string]any{
				"name":   "api",
				"source": "api",
				"trace":  "abc",
			},
			"code": 123456,
		}, validator.Validated())
	})

	s.Run("fail", func() {
		validator, err := facades.Validation().Make(context.Background(), map[string]any{
			"context": "ctx",
			"name":    " Goravel ",
			"tags":    []any{"tag1", "tag2"},
			"scores":  []any{1, 2},
			"items": []any{
				map[string]any{"name": "item1"},
				map[string]any{"name": ""},
			},
			"meta": map[string]any{
				"name":   "",
				"source": "api",
				"trace":  "abc",
			},
			"code": 123456,
			"age":  1,
		}, map[string]any{
			"name":         "required",
			"context":      "required",
			"tags.*":       "required|string",
			"scores.*":     "required|int",
			"items.*.name": "sometimes|required|string",
			"meta":         "sometimes|map",
			"meta.name":    "sometimes|required|string",
			"date":         "required|date",
			"code":         `required|regex:^\d{4,6}$`,
		}, validation.Filters(map[string]any{
			"name": "trim",
		}), validation.Messages(map[string]string{
			"date.required": "date is required",
		}))
		s.Require().NoError(err)
		s.True(validator.Fails())
		s.Equal(map[string]map[string]string{
			"date": {
				"required": "date is required",
			},
			"items.1.name": {
				"required": "The items.1.name field is required.",
			},
			"meta.name": {
				"required": "The meta.name field is required.",
			},
		}, validator.Errors().All())
	})
}

func ruleKey(rule string) string {
	if idx := strings.Index(rule, ":"); idx >= 0 {
		return rule[:idx]
	}

	return rule
}

func TestRuleKey(t *testing.T) {
	assert.Equal(t, "eq", ruleKey("eq:1"))
	assert.Equal(t, "required", ruleKey("required"))
}
