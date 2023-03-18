package interceptors

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func OpentracingServer(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	tracer, closer := NewJaegerTracer()
	defer closer.Close()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	}

	spanContext, err := tracer.Extract(opentracing.TextMap, MDReaderWriter{md})
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		return nil, fmt.Errorf("extract from metadata error: %v", err)
	}

	span := tracer.StartSpan(
		info.FullMethod,
		ext.RPCServerOption(spanContext),
		opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
		ext.SpanKindRPCServer,
	)
	defer span.Finish()

	return handler(opentracing.ContextWithSpan(ctx, span), req)
}
