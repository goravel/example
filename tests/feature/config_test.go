package feature

import (
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
)

type ConfigTestSuite struct {
	suite.Suite
	tests.TestCase
}

// retentionPolicy stands in for a user-defined config value: All copies maps
// and slices but shares everything else, so a pointer must come back identical.
type retentionPolicy struct {
	Minutes int
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, &ConfigTestSuite{})
}

func (s *ConfigTestSuite) TestEnv() {
	s.T().Setenv("TESTS_CONFIG_ENV", "goravel")

	s.Equal("goravel", facades.Config().Env("TESTS_CONFIG_ENV"))
	s.Equal("default", facades.Config().Env("TESTS_CONFIG_ENV_MISSING", "default"))
}

func (s *ConfigTestSuite) TestEnvString() {
	s.T().Setenv("TESTS_CONFIG_ENV_STRING", "goravel")

	s.Equal("goravel", facades.Config().EnvString("TESTS_CONFIG_ENV_STRING"))
	s.Equal("default", facades.Config().EnvString("TESTS_CONFIG_ENV_STRING_MISSING", "default"))
}

func (s *ConfigTestSuite) TestEnvBool() {
	s.T().Setenv("TESTS_CONFIG_ENV_BOOL", "true")

	s.True(facades.Config().EnvBool("TESTS_CONFIG_ENV_BOOL"))
	s.True(facades.Config().EnvBool("TESTS_CONFIG_ENV_BOOL_MISSING", true))
}

func (s *ConfigTestSuite) TestAddAndGet() {
	facades.Config().Add("tests.config.any", "goravel")

	s.Equal("goravel", facades.Config().Get("tests.config.any"))
	s.Equal("default", facades.Config().Get("tests.config.any.missing", "default"))
}

func (s *ConfigTestSuite) TestGetString() {
	facades.Config().Add("tests.config.string", "goravel")

	s.Equal("goravel", facades.Config().GetString("tests.config.string"))
	s.Equal("default", facades.Config().GetString("tests.config.string.missing", "default"))
}

func (s *ConfigTestSuite) TestGetInt() {
	facades.Config().Add("tests.config.int", 1)

	s.Equal(1, facades.Config().GetInt("tests.config.int"))
	s.Equal(2, facades.Config().GetInt("tests.config.int.missing", 2))
}

func (s *ConfigTestSuite) TestGetBool() {
	facades.Config().Add("tests.config.bool", true)

	s.True(facades.Config().GetBool("tests.config.bool"))
	s.True(facades.Config().GetBool("tests.config.bool.missing", true))
}

func (s *ConfigTestSuite) TestGetDuration() {
	facades.Config().Add("tests.config.duration", 3*time.Second)

	s.Equal(3*time.Second, facades.Config().GetDuration("tests.config.duration"))
	s.Equal(2*time.Second, facades.Config().GetDuration("tests.config.duration.missing", 2*time.Second))
}

func (s *ConfigTestSuite) TestUnmarshalKey() {
	facades.Config().Add("tests.config.unmarshal", map[string]any{
		"name":    "goravel",
		"enabled": true,
		"retries": 3,
	})

	type testConfig struct {
		Name    string `mapstructure:"name"`
		Enabled bool   `mapstructure:"enabled"`
		Retries int    `mapstructure:"retries"`
	}

	var config testConfig
	err := facades.Config().UnmarshalKey("tests.config.unmarshal", &config)
	s.Require().NoError(err)
	s.Equal(testConfig{Name: "goravel", Enabled: true, Retries: 3}, config)

	err = facades.Config().UnmarshalKey("tests.config.unmarshal", config)
	s.Error(err)
	s.ErrorContains(err, "pointer")
}

// TestAppConfig pins the typed getters against the config the app boots with,
// so they are proven on config/*.go and not only on values a test just added.
func (s *ConfigTestSuite) TestAppConfig() {
	s.Equal("en", facades.Config().GetString("app.locale"))
	s.Equal([]string{}, facades.Config().GetStringSlice("app.disabled_runners"))
}

func (s *ConfigTestSuite) TestGetInt64() {
	s.Equal(int64(3), facades.Config().GetInt64("http.request_timeout"))
	s.Equal(int64(10), facades.Config().GetInt64("database.pool.max_idle_conns"))

	facades.Config().Add("tests.config.int64", map[string]any{"string": "9000000000"})

	// Env-style string values are cast, as they are for GetInt.
	s.Equal(int64(9000000000), facades.Config().GetInt64("tests.config.int64.string"))
	s.Equal(int64(0), facades.Config().GetInt64("tests.config.int64.missing"))
	s.Equal(int64(9), facades.Config().GetInt64("tests.config.int64.missing", 9))
}

func (s *ConfigTestSuite) TestGetFloat64() {
	// http.request_timeout is a plain 3 in config/http.go, so the cast does not
	// depend on the environment the way telemetry.traces.sampler.ratio does.
	s.Equal(3.0, facades.Config().GetFloat64("http.request_timeout"))

	facades.Config().Add("tests.config.float64", map[string]any{
		"typed":  0.25,
		"string": "1.5",
	})

	s.Equal(0.25, facades.Config().GetFloat64("tests.config.float64.typed"))
	s.Equal(1.5, facades.Config().GetFloat64("tests.config.float64.string"))
	s.Equal(0.0, facades.Config().GetFloat64("tests.config.float64.missing"))
	s.Equal(0.25, facades.Config().GetFloat64("tests.config.float64.missing", 0.25))
}

func (s *ConfigTestSuite) TestGetTime() {
	releasedAt := time.Date(2026, 10, 30, 8, 5, 12, 0, time.UTC)
	facades.Config().Add("tests.config.time", map[string]any{
		"typed":    releasedAt,
		"rfc3339":  "2026-10-30T08:05:12Z",
		"datetime": "2026-10-30 08:05:12",
		"date":     "2026-10-30",
		"invalid":  "not-a-time",
		"number":   20261030,
	})

	tests := []struct {
		name string
		path string
		want time.Time
	}{
		{name: "should return a stored time.Time as is", path: "tests.config.time.typed", want: releasedAt},
		{name: "should parse an RFC3339 string", path: "tests.config.time.rfc3339", want: releasedAt},
		{name: "should parse a date time string", path: "tests.config.time.datetime", want: releasedAt},
		{name: "should parse a date string", path: "tests.config.time.date", want: time.Date(2026, 10, 30, 0, 0, 0, 0, time.UTC)},
		{name: "should return the zero time when the value can not be parsed", path: "tests.config.time.invalid", want: time.Time{}},
		{name: "should return the zero time for a value that is not time or string", path: "tests.config.time.number", want: time.Time{}},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.want, facades.Config().GetTime(tt.path))
		})
	}

	fallback := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Equal(time.Time{}, facades.Config().GetTime("tests.config.time.missing"))
	s.Equal(fallback, facades.Config().GetTime("tests.config.time.missing", fallback))
	// Set but unparseable does not fall back to the default.
	s.Equal(time.Time{}, facades.Config().GetTime("tests.config.time.invalid", fallback))
}

func (s *ConfigTestSuite) TestGetIntSlice() {
	s.Equal([]int{2, 100}, facades.Config().GetIntSlice("session.lottery"))

	facades.Config().Add("tests.config.int_slice", map[string]any{
		"values": []int{1, 2, 3},
		"empty":  []int{},
	})

	s.Equal([]int{1, 2, 3}, facades.Config().GetIntSlice("tests.config.int_slice.values"))
	// A set, empty slice is returned as is instead of the default.
	s.Equal([]int{}, facades.Config().GetIntSlice("tests.config.int_slice.empty", []int{9}))

	s.Nil(facades.Config().GetIntSlice("tests.config.int_slice.missing"))
	s.Equal([]int{9}, facades.Config().GetIntSlice("tests.config.int_slice.missing", []int{9}))
}

func (s *ConfigTestSuite) TestGetStringMap() {
	s.Equal(map[string]any{"driver": "memory"}, facades.Config().GetStringMap("cache.stores.memory"))

	s.Nil(facades.Config().GetStringMap("tests.config.string_map.missing"))
	s.Equal(map[string]any{"fallback": true}, facades.Config().GetStringMap("tests.config.string_map.missing", map[string]any{"fallback": true}))
}

func (s *ConfigTestSuite) TestGetStringMapString() {
	s.Equal(map[string]string{"driver": "memory"}, facades.Config().GetStringMapString("cache.stores.memory"))

	s.Nil(facades.Config().GetStringMapString("tests.config.string_map_string.missing"))
	s.Equal(map[string]string{"X-Retry": "3"}, facades.Config().GetStringMapString("tests.config.string_map_string.missing", map[string]string{"X-Retry": "3"}))
}

func (s *ConfigTestSuite) TestHas() {
	s.True(facades.Config().Has("app.name"))
	s.True(facades.Config().Has("session.lottery"))
	// Paths are matched without case, the same way Get and Forget treat them.
	s.True(facades.Config().Has("APP.NAME"))

	s.False(facades.Config().Has("app.missing"))
	s.False(facades.Config().Has("tests.config.not_set"))
}

func (s *ConfigTestSuite) TestForget() {
	facades.Config().Add("tests.config.forget", map[string]any{
		"driver": "file",
		"store":  map[string]any{"name": "local"},
	})
	// Forget lasts for the whole process: put the key back for whatever runs next.
	s.T().Cleanup(func() {
		facades.Config().Add("tests.config.forget", map[string]any{
			"driver": "file",
			"store":  map[string]any{"name": "local"},
		})
	})

	facades.Config().Forget("tests.config.forget")

	s.False(facades.Config().Has("tests.config.forget"))
	// Everything nested under a forgotten path reads as missing too.
	s.False(facades.Config().Has("tests.config.forget.driver"))
	s.False(facades.Config().Has("tests.config.forget.store.name"))
	s.Nil(facades.Config().Get("tests.config.forget.driver"))
	s.Equal("fallback", facades.Config().GetString("tests.config.forget.driver", "fallback"))

	// UnmarshalKey reports no error and leaves the target at its zero value.
	var forgotten struct {
		Driver string `json:"driver"`
	}
	s.Require().NoError(facades.Config().UnmarshalKey("tests.config.forget", &forgotten))
	s.Equal("", forgotten.Driver)

	// The forgotten key is dropped from All, its parents stay.
	testsConfig, ok := facades.Config().All()["tests"].(map[string]any)
	s.Require().True(ok)
	configSection, ok := testsConfig["config"].(map[string]any)
	s.Require().True(ok)
	s.NotContains(configSection, "forget")

	// Add on the forgotten path makes it visible again.
	facades.Config().Add("tests.config.forget", map[string]any{"driver": "file"})
	s.True(facades.Config().Has("tests.config.forget.driver"))
}

func (s *ConfigTestSuite) TestForget_AppConfigPath() {
	original := facades.Config().Get("app.maintenance")
	s.Require().NotNil(original)
	s.T().Cleanup(func() { facades.Config().Add("app.maintenance", original) })

	facades.Config().Forget("app.maintenance")

	s.False(facades.Config().Has("app.maintenance"))
	s.False(facades.Config().Has("app.maintenance.driver"))
	s.Equal("cache", facades.Config().GetString("app.maintenance.driver", "cache"))

	// The rest of the app config is untouched and app itself stays in All.
	s.True(facades.Config().Has("app.name"))
	app, ok := facades.Config().All()["app"].(map[string]any)
	s.Require().True(ok)
	s.NotContains(app, "maintenance")

	facades.Config().Add("app.maintenance", original)
	s.Equal(original, facades.Config().Get("app.maintenance"))
}

func (s *ConfigTestSuite) TestAll() {
	name := facades.Config().GetString("app.name")
	lottery := facades.Config().GetIntSlice("session.lottery")

	all := facades.Config().All()
	app, ok := all["app"].(map[string]any)
	s.Require().True(ok)
	s.Equal(name, app["name"])

	// Nested maps and slices are copies, so writing through All leaves the
	// configuration untouched.
	app["name"] = "mutated"
	session, ok := all["session"].(map[string]any)
	s.Require().True(ok)
	allLottery, ok := session["lottery"].([]int)
	s.Require().True(ok)
	allLottery[0] = 99
	s.Equal(name, facades.Config().GetString("app.name"))
	s.Equal(lottery, facades.Config().GetIntSlice("session.lottery"))

	// Structs and pointers are shared rather than copied.
	policy := &retentionPolicy{Minutes: 30}
	facades.Config().Add("tests.all_policy", policy)

	testsConfig, ok := facades.Config().All()["tests"].(map[string]any)
	s.Require().True(ok)
	s.Same(policy, testsConfig["all_policy"])
}

func (s *ConfigTestSuite) TestAll_ExcludesProcessEnv() {
	s.T().Setenv("TESTS_CONFIG_ALL_ENV_ONLY", "goravel")

	s.Equal("goravel", facades.Config().Env("TESTS_CONFIG_ALL_ENV_ONLY"))
	// AutomaticEnv makes the key readable, but All lists config values rather
	// than every process env variable.
	s.NotContains(facades.Config().All(), "tests_config_all_env_only")
}
