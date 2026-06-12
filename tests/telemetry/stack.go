package telemetry

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/facades"
	"github.com/stretchr/testify/require"
)

// Service names mirror the services defined in docker-compose.yml.
const (
	ServiceCollector    = "otel-collector"
	ServiceCollectorTLS = "otel-collector-tls"
	ServiceJaeger       = "jaeger"
	ServicePrometheus   = "prometheus"
	ServiceLoki         = "loki"
)

var (
	readyInterval = 2 * time.Second
	readyTimeout  = 90 * time.Second

	httpClient = &http.Client{Timeout: 5 * time.Second}

	readiness = map[string]func() error{
		ServiceJaeger:       httpReady("http://localhost:16686/"),
		ServicePrometheus:   httpReady("http://localhost:9090/-/ready"),
		ServiceLoki:         httpReady("http://localhost:3100/ready"),
		ServiceCollector:    tcpReady("localhost:4318"),
		ServiceCollectorTLS: tcpReady("localhost:4319"),
	}
)

// EnsureStack starts the requested compose services if they are not already
// running and blocks until each reports ready. Suites must not stop the
// services themselves; TeardownStack owns shutdown via TestMain.
func EnsureStack(t *testing.T, services ...string) {
	t.Helper()

	require.NotEmpty(t, services)

	for _, service := range services {
		_, ok := readiness[service]
		require.True(t, ok, "no readiness probe for service "+service)
	}

	result := facades.Process().Path(composeDir()).Run("docker compose up -d " + strings.Join(services, " "))
	require.False(t, result.Failed(), result.ErrorOutput())

	for _, service := range services {
		require.NoError(t, waitFor(readiness[service]), service+" not ready within "+readyTimeout.String())
	}
}

// TeardownStack stops all compose services, including profile-gated ones, and
// removes generated TLS material. Call it once from TestMain after m.Run.
func TeardownStack() {
	facades.Process().Path(composeDir()).Run("docker compose --profile tls down")
	if certsDir != "" {
		_ = os.RemoveAll(certsDir)
	}
}

func composeDir() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..")
}

func waitFor(probe func() error) error {
	deadline := time.Now().Add(readyTimeout)
	err := probe()
	for err != nil && time.Now().Before(deadline) {
		time.Sleep(readyInterval)
		err = probe()
	}
	return err
}

func httpReady(url string) func() error {
	return func() error {
		resp, err := httpClient.Get(url)
		if err != nil {
			return err
		}
		defer errors.Ignore(resp.Body.Close)
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s returned %d", url, resp.StatusCode)
		}
		return nil
	}
}

func tcpReady(addr string) func() error {
	return func() error {
		conn, err := net.DialTimeout("tcp", addr, time.Second)
		if err != nil {
			return err
		}
		return conn.Close()
	}
}
