package telemetry

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPoll(t *testing.T) {
	defer func(interval, timeout time.Duration) {
		awaitInterval, awaitTimeout = interval, timeout
	}(awaitInterval, awaitTimeout)
	awaitInterval, awaitTimeout = 10*time.Millisecond, 300*time.Millisecond

	t.Run("returns once all substrings appear", func(t *testing.T) {
		var hits atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			if hits.Add(1) < 3 {
				fmt.Fprint(w, "warming up")
				return
			}
			fmt.Fprint(w, "trace-a trace-b")
		}))
		defer server.Close()

		body, ok := poll(func() string { return server.URL }, []string{"trace-a", "trace-b"})

		assert.True(t, ok)
		assert.Contains(t, body, "trace-a")
		assert.GreaterOrEqual(t, hits.Load(), int32(3))
	})

	t.Run("times out with last body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			fmt.Fprint(w, "empty result")
		}))
		defer server.Close()

		body, ok := poll(func() string { return server.URL }, []string{"never"})

		assert.False(t, ok)
		assert.Equal(t, "empty result", body)
	})
}
