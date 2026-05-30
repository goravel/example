package feature

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	frameworkgrpc "github.com/goravel/framework/grpc"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/stats"

	"goravel/app/facades"
	"goravel/tests"
)

const (
	grpcFrameworkClientGroup = "grpc_framework_integration"
	grpcFrameworkClientID    = "goravel-client"
	grpcFrameworkServerID    = "goravel-server"
)

var grpcFrameworkServerCounter atomic.Uint64

type grpcFrameworkContextKey string

const (
	grpcFrameworkContextClient grpcFrameworkContextKey = "client"
	grpcFrameworkContextServer grpcFrameworkContextKey = "server"
)

type GrpcFrameworkIntegrationSuite struct {
	suite.Suite
	tests.TestCase
}

func TestGrpcFrameworkIntegrationSuite(t *testing.T) {
	suite.Run(t, &GrpcFrameworkIntegrationSuite{})
}

func (s *GrpcFrameworkIntegrationSuite) TestRunWithHostAndConnect() {
	host := acquireFreeAddress(s.T())
	serverName := s.uniqueServerName("run_with_host")
	s.configureClientServer(serverName, host)

	app, statsProbe := s.newFrameworkApplication(true)
	runErrCh := runAsync(func() error {
		return app.Run(host)
	})
	defer shutdownApplication(s.T(), app, runErrCh)

	conn, err := app.Connect(serverName)
	s.NoError(err)

	client := frameworkgrpc.NewTestServiceClient(conn)
	resp := callGetWithRetry(s.T(), client, "run-host")

	s.assertTestResponse(resp, "run-host")
	s.GreaterOrEqual(statsProbe.rpcCount.Load(), int64(1))
	s.GreaterOrEqual(statsProbe.connCount.Load(), int64(1))
}

func (s *GrpcFrameworkIntegrationSuite) TestRunWithConfigHostAndPort() {
	address := acquireFreeAddress(s.T())
	host, port, err := net.SplitHostPort(address)
	s.Require().NoError(err)

	s.setConfig("grpc.host", host)
	s.setConfig("grpc.port", port)

	serverName := s.uniqueServerName("run_with_config")
	s.configureClientServer(serverName, address)

	app, statsProbe := s.newFrameworkApplication(true)
	runErrCh := runAsync(func() error {
		return app.Run()
	})
	defer shutdownApplication(s.T(), app, runErrCh)

	conn, err := app.Connect(serverName)
	s.NoError(err)

	client := frameworkgrpc.NewTestServiceClient(conn)
	resp := callGetWithRetry(s.T(), client, "run-config")

	s.assertTestResponse(resp, "run-config")
	s.GreaterOrEqual(statsProbe.rpcCount.Load(), int64(1))
	s.GreaterOrEqual(statsProbe.connCount.Load(), int64(1))
}

func (s *GrpcFrameworkIntegrationSuite) TestListenAndConnect() {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	s.Require().NoError(err)

	serverName := s.uniqueServerName("listen")
	s.configureClientServer(serverName, listener.Addr().String())

	app, statsProbe := s.newFrameworkApplication(true)
	runErrCh := runAsync(func() error {
		return app.Listen(listener)
	})
	defer shutdownApplication(s.T(), app, runErrCh)

	conn, err := app.Connect(serverName)
	s.NoError(err)

	client := frameworkgrpc.NewTestServiceClient(conn)
	resp := callGetWithRetry(s.T(), client, "listen")

	s.assertTestResponse(resp, "listen")
	s.GreaterOrEqual(statsProbe.rpcCount.Load(), int64(1))
	s.GreaterOrEqual(statsProbe.connCount.Load(), int64(1))
}

func (s *GrpcFrameworkIntegrationSuite) TestConnectCachingSerial() {
	host := acquireFreeAddress(s.T())
	serverName := s.uniqueServerName("cache_serial")
	s.configureClientServer(serverName, host)

	app, statsProbe := s.newFrameworkApplication(true)
	runErrCh := runAsync(func() error {
		return app.Run(host)
	})
	defer shutdownApplication(s.T(), app, runErrCh)

	conn1, err := app.Connect(serverName)
	s.NoError(err)

	conn2, err := app.Connect(serverName)
	s.NoError(err)

	s.Same(conn1, conn2)

	client := frameworkgrpc.NewTestServiceClient(conn2)
	resp := callGetWithRetry(s.T(), client, "cache-serial")

	s.assertTestResponse(resp, "cache-serial")
	s.GreaterOrEqual(statsProbe.rpcCount.Load(), int64(1))
}

func (s *GrpcFrameworkIntegrationSuite) TestConnectCachingConcurrent() {
	host := acquireFreeAddress(s.T())
	serverName := s.uniqueServerName("cache_concurrent")
	s.configureClientServer(serverName, host)

	app, statsProbe := s.newFrameworkApplication(true)
	runErrCh := runAsync(func() error {
		return app.Run(host)
	})
	defer shutdownApplication(s.T(), app, runErrCh)

	const concurrency = 20
	type connectResult struct {
		conn *grpc.ClientConn
		err  error
	}

	var wg sync.WaitGroup
	results := make(chan connectResult, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := app.Connect(serverName)
			results <- connectResult{
				conn: conn,
				err:  err,
			}
		}()
	}

	wg.Wait()
	close(results)

	connections := make([]*grpc.ClientConn, 0, concurrency)
	for result := range results {
		s.NoError(result.err)
		s.NotNil(result.conn)
		connections = append(connections, result.conn)
	}

	s.Len(connections, concurrency)
	first := connections[0]
	for _, conn := range connections[1:] {
		s.Same(first, conn)
	}

	client := frameworkgrpc.NewTestServiceClient(first)
	resp := callGetWithRetry(s.T(), client, "cache-concurrent")

	s.assertTestResponse(resp, "cache-concurrent")
	s.GreaterOrEqual(statsProbe.rpcCount.Load(), int64(1))
}

func (s *GrpcFrameworkIntegrationSuite) TestShutdownReconnectAndCall() {
	host := acquireFreeAddress(s.T())
	serverName := s.uniqueServerName("shutdown_reconnect")
	s.configureClientServer(serverName, host)

	app, statsProbe := s.newFrameworkApplication(true)
	runErrCh := runAsync(func() error {
		return app.Run(host)
	})

	stopped := false
	defer func() {
		if !stopped {
			shutdownApplication(s.T(), app, runErrCh)
		}
	}()

	conn1, err := app.Connect(serverName)
	s.NoError(err)

	client1 := frameworkgrpc.NewTestServiceClient(conn1)
	resp1 := callGetWithRetry(s.T(), client1, "before-shutdown")
	s.assertTestResponse(resp1, "before-shutdown")

	shutdownApplication(s.T(), app, runErrCh)
	stopped = true

	s.True(waitForCondition(3*time.Second, 20*time.Millisecond, func() bool {
		return conn1.GetState() == connectivity.Shutdown
	}))

	serverApp, _ := s.newFrameworkApplication(true)
	serverRunErrCh := runAsync(func() error {
		return serverApp.Run(host)
	})
	defer shutdownApplication(s.T(), serverApp, serverRunErrCh)

	conn2, err := app.Connect(serverName)
	s.NoError(err)
	s.NotSame(conn1, conn2)

	client2 := frameworkgrpc.NewTestServiceClient(conn2)
	resp2 := callGetWithRetry(s.T(), client2, "after-shutdown")
	s.assertTestResponse(resp2, "after-shutdown")
	s.GreaterOrEqual(statsProbe.rpcCount.Load(), int64(1))
}

func (s *GrpcFrameworkIntegrationSuite) newFrameworkApplication(registerService bool) (*frameworkgrpc.Application, *grpcFrameworkStatsProbe) {
	app := frameworkgrpc.NewApplication(facades.Config())
	statsProbe := &grpcFrameworkStatsProbe{}

	app.UnaryServerInterceptors([]grpc.UnaryServerInterceptor{
		grpcFrameworkServerInterceptor,
	})
	app.ServerStatsHandlers([]stats.Handler{
		statsProbe,
	})
	app.UnaryClientInterceptorGroups(map[string][]grpc.UnaryClientInterceptor{
		grpcFrameworkClientGroup: {
			grpcFrameworkClientInterceptor,
		},
	})
	app.ClientStatsHandlerGroups(map[string][]stats.Handler{
		grpcFrameworkClientGroup: {
			statsProbe,
		},
	})

	if registerService {
		frameworkgrpc.RegisterTestServiceServer(app.Server(), &grpcFrameworkTestService{})
	}

	return app, statsProbe
}

func (s *GrpcFrameworkIntegrationSuite) configureClientServer(serverName, host string) {
	s.setConfig(fmt.Sprintf("grpc.servers.%s.host", serverName), host)
	s.setConfig(fmt.Sprintf("grpc.servers.%s.interceptors", serverName), []string{grpcFrameworkClientGroup})
	s.setConfig(fmt.Sprintf("grpc.servers.%s.stats_handlers", serverName), []string{grpcFrameworkClientGroup})
}

func (s *GrpcFrameworkIntegrationSuite) setConfig(key string, value any) {
	original := facades.Config().Get(key)
	facades.Config().Add(key, value)
	s.T().Cleanup(func() {
		facades.Config().Add(key, original)
	})
}

func (s *GrpcFrameworkIntegrationSuite) uniqueServerName(prefix string) string {
	return fmt.Sprintf("%s_%d", prefix, grpcFrameworkServerCounter.Add(1))
}

func (s *GrpcFrameworkIntegrationSuite) assertTestResponse(resp *frameworkgrpc.TestResponse, name string) {
	s.Equal(int32(http.StatusOK), resp.GetCode())
	s.Equal(fmt.Sprintf("server=%s,client=%s,name=%s", grpcFrameworkServerID, grpcFrameworkClientID, name), resp.GetMessage())
}

func runAsync(callback func() error) <-chan error {
	runErrCh := make(chan error, 1)
	go func() {
		runErrCh <- callback()
	}()

	return runErrCh
}

func shutdownApplication(t *testing.T, app *frameworkgrpc.Application, runErrCh <-chan error) {
	t.Helper()

	require.NoError(t, app.Shutdown(true))

	select {
	case runErr := <-runErrCh:
		if runErr == nil {
			return
		}
		if errors.Is(runErr, grpc.ErrServerStopped) {
			return
		}

		require.NoError(t, runErr)
	case <-time.After(3 * time.Second):
		require.FailNow(t, "timed out waiting for gRPC server shutdown")
	}
}

func callGetWithRetry(t *testing.T, client frameworkgrpc.TestServiceClient, name string) *frameworkgrpc.TestResponse {
	t.Helper()

	deadline := time.Now().Add(6 * time.Second)
	var lastErr error

	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
		resp, err := client.Get(ctx, &frameworkgrpc.TestRequest{Name: name})
		cancel()
		if err == nil {
			return resp
		}

		lastErr = err
		time.Sleep(60 * time.Millisecond)
	}

	require.NoError(t, lastErr)

	return nil
}

func acquireFreeAddress(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	address := listener.Addr().String()
	require.NoError(t, listener.Close())

	return address
}

func waitForCondition(timeout, interval time.Duration, condition func() bool) bool {
	expire := time.Now().Add(timeout)
	for {
		if condition() {
			return true
		}
		if time.Now().After(expire) {
			return false
		}
		time.Sleep(interval)
	}
}

func grpcFrameworkClientInterceptor(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	outgoingMD, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		outgoingMD = metadata.New(nil)
	} else {
		outgoingMD = outgoingMD.Copy()
	}

	outgoingMD.Set("client", grpcFrameworkClientID)

	return invoker(metadata.NewOutgoingContext(ctx, outgoingMD), method, req, reply, cc, opts...)
}

func grpcFrameworkServerInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	incomingMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		incomingMD = metadata.New(nil)
	}

	client := ""
	values := incomingMD.Get("client")
	if len(values) > 0 {
		client = values[0]
	}

	ctx = context.WithValue(ctx, grpcFrameworkContextClient, client)
	ctx = context.WithValue(ctx, grpcFrameworkContextServer, grpcFrameworkServerID)

	return handler(ctx, req)
}

type grpcFrameworkStatsProbe struct {
	rpcCount  atomic.Int64
	connCount atomic.Int64
}

func (g *grpcFrameworkStatsProbe) TagRPC(ctx context.Context, _ *stats.RPCTagInfo) context.Context {
	return ctx
}

func (g *grpcFrameworkStatsProbe) HandleRPC(context.Context, stats.RPCStats) {
	g.rpcCount.Add(1)
}

func (g *grpcFrameworkStatsProbe) TagConn(ctx context.Context, _ *stats.ConnTagInfo) context.Context {
	return ctx
}

func (g *grpcFrameworkStatsProbe) HandleConn(context.Context, stats.ConnStats) {
	g.connCount.Add(1)
}

type grpcFrameworkTestService struct {
	frameworkgrpc.UnimplementedTestServiceServer
}

func (g *grpcFrameworkTestService) Get(ctx context.Context, req *frameworkgrpc.TestRequest) (*frameworkgrpc.TestResponse, error) {
	server := ctx.Value(grpcFrameworkContextServer)
	client := ctx.Value(grpcFrameworkContextClient)

	return &frameworkgrpc.TestResponse{
		Code:    int32(http.StatusOK),
		Message: fmt.Sprintf("server=%v,client=%v,name=%s", server, client, req.GetName()),
	}, nil
}
