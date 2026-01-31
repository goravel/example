package feature

import (
	"goravel/app/facades"
	"testing"

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

	dailyLogPath := path.Storage("logs", "goravel-"+now.ToDateString()+".log")

	facades.Log().Info("This is an info log")
	assert.True(t, file.Contains(path.Storage("logs", "goravel.log"), `{"environment":"local","level":"info","message":"This is an info log","time":"`+dateTimeMilli+`"}`))
	assert.True(t, file.Contains(dailyLogPath, "["+dateTimeMilli+"] local.info: This is an info log"))
}
