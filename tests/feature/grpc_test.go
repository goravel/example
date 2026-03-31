package feature

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"

	proto "github.com/goravel/example-proto"
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
	tests := []struct {
		name         string
		token        string
		expectedBody string
	}{
		{
			name:         "token is 1",
			token:        "1",
			expectedBody: "{\"id\":1,\"name\":\"Goravel\",\"token\":\"1\"}",
		},
		{
			name:         "token is text",
			token:        "goravel",
			expectedBody: "{\"id\":1,\"name\":\"Goravel\",\"token\":\"goravel\"}",
		},
		{
			name:         "token is alphanumeric",
			token:        "abc123",
			expectedBody: "{\"id\":1,\"name\":\"Goravel\",\"token\":\"abc123\"}",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			resp, err := facades.Http().Get(fmt.Sprintf("/grpc/user?token=%s", url.QueryEscape(test.token)))

			s.NoError(err)
			s.Equal(http.StatusOK, resp.Status())

			body, err := resp.Body()
			s.NoError(err)
			s.Equal(test.expectedBody, body)
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
	s.NoError(err)

	client := proto.NewUserServiceClient(connection)
	response, err := client.GetUser(context.Background(), &proto.UserRequest{
		Token: "direct",
	})
	s.NoError(err)

	s.Equal(map[string]any{
		"code":    int32(http.StatusOK),
		"message": "",
		"data": map[string]any{
			"id":    uint64(1),
			"name":  "Goravel",
			"token": "direct",
		},
	}, map[string]any{
		"code":    response.GetCode(),
		"message": response.GetMessage(),
		"data": map[string]any{
			"id":    response.GetData().GetId(),
			"name":  response.GetData().GetName(),
			"token": response.GetData().GetToken(),
		},
	})
}
