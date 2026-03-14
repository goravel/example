package feature

import (
"testing"
"time"

"github.com/stretchr/testify/suite"

"goravel/app/facades"
)

type ConfigTestSuite struct {
suite.Suite
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
}
