package telemetry

import (
	"fmt"
	"io"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/goravel/framework/errors"
)

var (
	awaitInterval = 2 * time.Second
	awaitTimeout  = 60 * time.Second
)

// AwaitTraces polls Jaeger until every want substring appears in the traces
// recorded for service, failing the test on timeout.
func AwaitTraces(t *testing.T, service string, want ...string) {
	t.Helper()

	target := "http://localhost:16686/api/traces?service=" + url.QueryEscape(service)
	await(t, func() string { return target }, want)
}

// AwaitMetric polls Prometheus until every want substring appears in the
// instant query result for promQL, failing the test on timeout.
func AwaitMetric(t *testing.T, promQL string, want ...string) {
	t.Helper()

	target := "http://localhost:9090/api/v1/query?query=" + url.QueryEscape(promQL)
	await(t, func() string { return target }, want)
}

// AwaitLogs polls Loki until every want substring appears in the last five
// minutes of logs matching logQL, failing the test on timeout.
func AwaitLogs(t *testing.T, logQL string, want ...string) {
	t.Helper()

	await(t, func() string {
		end := time.Now().UnixNano()
		start := time.Now().Add(-5 * time.Minute).UnixNano()
		return fmt.Sprintf("http://localhost:3100/loki/api/v1/query_range?query=%s&start=%d&end=%d",
			url.QueryEscape(logQL), start, end)
	}, want)
}

func await(t *testing.T, target func() string, want []string) {
	t.Helper()

	body, ok := poll(target, want)
	if !ok {
		t.Fatalf("telemetry not found at %s\nwant: %q\nlast response: %s", target(), want, body)
	}
}

func poll(target func() string, want []string) (string, bool) {
	deadline := time.Now().Add(awaitTimeout)
	for {
		body := fetch(target())
		if containsAll(body, want) {
			return body, true
		}
		if time.Now().After(deadline) {
			return body, false
		}
		time.Sleep(awaitInterval)
	}
}

func fetch(target string) string {
	resp, err := httpClient.Get(target)
	if err != nil {
		return err.Error()
	}
	defer errors.Ignore(resp.Body.Close)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err.Error()
	}

	return string(body)
}

func containsAll(body string, want []string) bool {
	for _, item := range want {
		if !strings.Contains(body, item) {
			return false
		}
	}
	return true
}
