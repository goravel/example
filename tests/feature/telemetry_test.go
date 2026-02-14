package feature

import (
	"testing"
	"time"

	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/assert"
)

func TestTelemetry(t *testing.T) {
	appName := facades.Config().GetString("app.name")
	_, err := facades.Http().Get("/user/1")
	assert.NoError(t, err)

	time.Sleep(7 * time.Second)

	t.Run("Check Jaeger for Traces", func(t *testing.T) {
		resp, err := facades.Http().Get("http://localhost:16686/api/traces?service=" + appName)
		assert.NoError(t, err)

		var result struct {
			Data []any `json:"data"`
		}
		assert.NoError(t, resp.Bind(&result))
		assert.NotEmpty(t, result.Data, "Telemetry failed to reach Jaeger")
	})
}
