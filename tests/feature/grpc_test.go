package feature

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"testing"

	proto "github.com/goravel/example-proto"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/encoding/protojson"
	protobufproto "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/protoadapt"

	"goravel/app/facades"
	"goravel/tests"

	"github.com/goravel/framework/support/color"
)

type GrpcFeatureSuite struct {
	suite.Suite
	tests.TestCase
}

func TestGrpcFeatureSuite(t *testing.T) {
	suite.Run(t, &GrpcFeatureSuite{})
}

func (s *GrpcFeatureSuite) TestHttpBridgeTableDriven() {
	cases := []struct {
		name         string
		token        string
		expectedBody string
	}{
		{
			name:         "numeric token",
			token:        "1",
			expectedBody: `{"id":1,"name":"Goravel","token":"1"}`,
		},
		{
			name:         "text token",
			token:        "goravel",
			expectedBody: `{"id":1,"name":"Goravel","token":"goravel"}`,
		},
		{
			name:         "mixed token",
			token:        "abc123",
			expectedBody: `{"id":1,"name":"Goravel","token":"abc123"}`,
		},
	}

	for _, c := range cases {
		s.Run(c.name, func() {
			resp, err := facades.Http().Get(fmt.Sprintf("/grpc/user?token=%s", url.QueryEscape(c.token)))
			s.Require().NoError(err)
			s.Equal(http.StatusOK, resp.Status())

			body, err := resp.Body()
			s.Require().NoError(err)
			s.Equal(c.expectedBody, body)
		})
	}
}

func (s *GrpcFeatureSuite) TestHttpBridgeConcurrentRequests() {
	const requestCount = 10

	type requestResult struct {
		status int
		body   string
		err    error
	}

	var wg sync.WaitGroup
	results := make(chan requestResult, requestCount)

	for i := 0; i < requestCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := facades.Http().Get("/grpc/user?token=1")
			if err != nil {
				results <- requestResult{
					err: err,
				}
				return
			}

			body, err := resp.Body()
			results <- requestResult{
				status: resp.Status(),
				body:   body,
				err:    err,
			}
		}()
	}

	wg.Wait()
	close(results)

	resultSlice := make([]requestResult, 0, requestCount)
	for result := range results {
		resultSlice = append(resultSlice, result)
	}

	s.Len(resultSlice, requestCount)
	for _, result := range resultSlice {
		s.NoError(result.err)
		s.Equal(http.StatusOK, result.status)
		s.Equal("{\"id\":1,\"name\":\"Goravel\",\"token\":\"1\"}", result.body)
	}
}

func (s *GrpcFeatureSuite) TestDirectGrpcClientConnection() {
	connection, err := facades.Grpc().Connect("user")
	s.Require().NoError(err)
	s.NotNil(connection)

	client := proto.NewUserServiceClient(connection)
	response, err := client.GetUser(context.Background(), &proto.UserRequest{Token: "direct"})
	s.Require().NoError(err)
	s.NotNil(response)

	// example-proto generates v1 proto.Message types. Bridge through protoadapt
	// so the non-deprecated protojson can marshal the wire form against a
	// single literal — no mirror map, no XXX_* cache fields to compare.
	actual, err := protojson.Marshal(protoadapt.MessageV2Of(response))
	s.Require().NoError(err)
	s.Equal(`{"code":200, "data":{"id":"1", "name":"Goravel", "token":"direct"}}`, string(actual))
}

// TestUserClientConnectionIsCached verifies that the gRPC application caches
// the *grpc.ClientConn for the "user" server so repeated facades.Grpc().Connect
// calls do not open new TCP connections. This is the user-side counterpart of
// the framework's `Application.servers` map, which `Shutdown` clears with the
// comment "allow Garbage Collection": if the cache were bypassed, every
// Connect would leak a new connection that Shutdown could not reach, and the
// HTTP bridge below would race against the controller's separately-held conn.
func (s *GrpcFeatureSuite) TestUserClientConnectionIsCached() {
	first, err := facades.Grpc().Connect("user")
	s.Require().NoError(err)
	s.NotNil(first)

	second, err := facades.Grpc().Connect("user")
	s.Require().NoError(err)
	s.Same(first, second, "Connect should return the cached *grpc.ClientConn, not open a new one")

	// Exercise the cached connection through the HTTP bridge so a real RPC
	// travels over it. If the controller held a different conn, this would
	// still pass at the HTTP layer, so the assertion above is the load-bearing
	// check; this one only confirms the cache is not stale.
	resp, err := facades.Http().Get("/grpc/user?token=cached")
	s.Require().NoError(err)
	s.Equal(http.StatusOK, resp.Status())

	body, err := resp.Body()
	s.Require().NoError(err)
	s.Equal(`{"id":1,"name":"Goravel","token":"cached"}`, body)
}

// TestGrpcCredentials exercises the client side of the gRPC transport
// credentials surface documented at goravel/docs#192 — specifically the
// `grpc.servers.<name>.credentials` config key and the fallback semantics
// that `WithGrpcClientCredentials` enables. The server-side
// `WithGrpcServerCredentials` path is not covered here because the running
// app's gRPC server is frozen by the time the suite runs (routes/grpc.go
// calls facades.Grpc().Server() to bind the UserService, which seeds the
// server singleton, and the framework has no path to reset the frozen
// server); a separate goravel/framework change would be needed to make
// that case testable from the example repo. All cases use the running
// insecure server on 127.0.0.1:3002, so they exercise the real
// resolveClientCredentials resolver rather than a stub.
func (s *GrpcFeatureSuite) TestGrpcCredentials() {
	s.Run("falls back to insecure when no credentials key set", func() {
		// Server entry without a `credentials` field.
		// resolveClientCredentials must return insecure.NewCredentials()
		// to preserve the v1.17 contract for apps that have not opted
		// into TLS. Pointing at the already-running insecure server on
		// 3002 keeps the test focused on the credentials resolver rather
		// than the TLS handshake.
		facades.Config().Add("grpc.servers.user-insecure-fallback", map[string]any{
			"host":           "127.0.0.1",
			"port":           "3002",
			"interceptors":   []string{},
			"stats_handlers": []string{},
		})

		conn, err := facades.Grpc().Connect("user-insecure-fallback")
		s.Require().NoError(err)
		s.NotNil(conn)
		defer func() { _ = conn.Close() }()

		client := proto.NewUserServiceClient(conn)
		resp, err := client.GetUser(context.Background(), &proto.UserRequest{Token: "insecure-fallback"})
		s.Require().NoError(err)
		s.NotNil(resp)

		s.True(protobufproto.Equal(
			protoadapt.MessageV2Of(&proto.UserResponse{
				Code: int32(http.StatusOK),
				Data: &proto.User{Id: 1, Name: "Goravel", Token: "insecure-fallback"},
			}),
			protoadapt.MessageV2Of(resp),
		), "insecure-fallback response did not match expected UserResponse")
	})

	s.Run("warns and falls back when credentials group is unknown", func() {
		// Register a group, but configure a different name so the lookup
		// misses. The framework should log the warning and fall back to
		// insecure. ServerName: "noop" makes the registered creds useless
		// on purpose — we never want the client to actually use them here.
		facades.Grpc().ClientCredentials(map[string]credentials.TransportCredentials{
			"tls-unknown": credentials.NewTLS(&tls.Config{ServerName: "noop"}),
		})

		facades.Config().Add("grpc.servers.user-tls-unknown", map[string]any{
			"host":           "127.0.0.1",
			"port":           "3002",
			"credentials":    "does-not-exist",
			"interceptors":   []string{},
			"stats_handlers": []string{},
		})

		var (
			conn *grpc.ClientConn
			cErr error
		)
		got := color.CaptureOutput(func(io.Writer) {
			conn, cErr = facades.Grpc().Connect("user-tls-unknown")
		})
		s.Require().NoError(cErr)
		s.NotNil(conn)
		defer func() { _ = conn.Close() }()

		s.Contains(got, `client credentials group "does-not-exist" is not registered for server "user-tls-unknown"`)
		s.Contains(got, "falling back to insecure credentials")
	})

	s.Run("returns insecure silently when groups registered but config key empty", func() {
		// Groups are registered for "tls-empty" but the server entry omits
		// the credentials key entirely. resolveClientCredentials must take
		// the silent-insecure path (no warning), distinct from the
		// "unknown group" path above.
		facades.Grpc().ClientCredentials(map[string]credentials.TransportCredentials{
			"tls-empty": credentials.NewTLS(&tls.Config{ServerName: "noop"}),
		})

		facades.Config().Add("grpc.servers.user-tls-empty", map[string]any{
			"host":           "127.0.0.1",
			"port":           "3002",
			"interceptors":   []string{},
			"stats_handlers": []string{},
		})

		var (
			conn *grpc.ClientConn
			cErr error
		)
		got := color.CaptureOutput(func(io.Writer) {
			conn, cErr = facades.Grpc().Connect("user-tls-empty")
		})
		s.Require().NoError(cErr)
		s.NotNil(conn)
		defer func() { _ = conn.Close() }()

		// Empty key is the documented "no creds wanted" path; the
		// framework only warns when a key was provided but unmatched.
		s.NotContains(got, "falling back to insecure credentials")
		s.NotContains(got, "is not registered for server")
	})
}
