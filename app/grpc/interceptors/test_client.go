package interceptors

import (
	"context"

	"google.golang.org/grpc"
)

func TestClient(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	return nil
}
