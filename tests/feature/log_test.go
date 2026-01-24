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
	carbon.SetTestNow(carbon.Parse("2026-01-02 12:34:56.123"))
	defer carbon.ClearTestNow()

	facades.Log().Info("This is an info log")
	assert.True(t, file.Contains(path.Storage("logs", "goravel.log"), `{"environment":"local","level":"info","message":"This is an info log","time":"2026-01-02 12:34:56.123"}`))
}
