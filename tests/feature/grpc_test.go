package feature

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"

	proto "github.com/goravel/example-proto"
	"github.com/golang/protobuf/jsonpb"
	"github.com/stretchr/testify/suite"

	"goravel/app/facades"
	"goravel/tests"
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

	// jsonpb.MarshalToString strips the unexported XXX_* cache fields so the
	// wire representation can be asserted as a single, stable literal.
	actual, err := (&jsonpb.Marshaler{}).MarshalToString(response)
	s.Require().NoError(err)
	s.Equal(`{"code":200,"data":{"id":"1","name":"Goravel","token":"direct"}}`, actual)
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
