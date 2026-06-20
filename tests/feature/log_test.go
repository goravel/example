package feature

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"goravel/app/facades"
	"goravel/tests"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/stretchr/testify/assert"
)

func TestLog(t *testing.T) {
	now := carbon.Now()
	dateTimeMilli := now.ToDateTimeMilliString()
	carbon.SetTestNow(carbon.Parse(dateTimeMilli))
	defer carbon.ClearTestNow()

	facades.Log().Info("This is an info log")

	assert.True(t, file.Contains(path.Storage("logs", "goravel.log"), `{"environment":"local","level":"info","message":"This is an info log","time":"`+dateTimeMilli+`"}`))

	dailyLogPath := path.Storage("logs", "goravel-"+now.ToDateString()+".log")
	assert.True(t, file.Contains(dailyLogPath, "["+dateTimeMilli+"] local.info: This is an info log"))

	// The next day's log file can be created correctly
	now = carbon.Now().AddDay()
	dateTimeMilli = now.ToDateTimeMilliString()
	carbon.SetTestNow(carbon.Parse(dateTimeMilli))

	facades.Log().Error("This is an error log")

	assert.True(t, file.Contains(path.Storage("logs", "goravel.log"), `{"environment":"local","level":"error","message":"This is an error log","time":"`+dateTimeMilli+`","trace":{`))

	newDailyLogPath := path.Storage("logs", "goravel-"+now.ToDateString()+".log")
	assert.True(t, file.Contains(newDailyLogPath, "["+dateTimeMilli+"] local.error: This is an error log"))
}

func TestLogWithContext(t *testing.T) {
	testCase := &tests.TestCase{}

	resp, err := testCase.Http(t).Get("/log/with-context")
	assert.NoError(t, err)
	resp.AssertSuccessful()

	body, err := os.ReadFile(path.Storage("logs", "goravel.log"))
	assert.NoError(t, err)
	assert.NotEmpty(t, body)

	line := findLogLine(string(body), `"message":"log with context example"`)
	assert.NotEmpty(t, line, "log line with marker message not found")

	var entry map[string]any
	assert.NoError(t, json.Unmarshal([]byte(line), &entry))

	logCtx, ok := entry["context"].(map[string]any)
	assert.True(t, ok, "log entry must include a context object")
	if !ok {
		return
	}

	assert.NotContains(t, logCtx, "GoravelAuthJwt")
	assert.NotContains(t, logCtx, "access_token")
	assert.NotContains(t, logCtx, "secret_key")

	assert.Equal(t, "req-abc-123", logCtx["request_id"])
	assert.Equal(t, "trace-xyz-987", logCtx["trace_id"])
	assert.Equal(t, "user-42", logCtx["controllers.logSentinel"])
}

func findLogLine(body, marker string) string {
	var found string
	for _, line := range strings.Split(body, "\n") {
		if strings.Contains(line, marker) {
			found = line
		}
	}
	return found
}
