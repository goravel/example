package feature

import (
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"goravel/app/facades"
)

func TestGrpc(t *testing.T) {
	var wg sync.WaitGroup

	// Make several requests to test the gRPC client functionality
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := facades.Http().Get("/grpc/user?token=1")

			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, resp.Status())

			body, err := resp.Body()
			assert.NoError(t, err)
			assert.Equal(t, "{\"id\":1,\"name\":\"Goravel\",\"token\":\"1\"}", body)
		}()
	}

	wg.Wait()
}
