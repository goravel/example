package interceptors

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func OpentracingClient(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	tracer, closer := NewJaegerTracer()
	defer closer.Close()

	span, _ := opentracing.StartSpanFromContext(
		ctx,
		"call gRPC",
		opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
		ext.SpanKindRPCClient,
	)
	defer span.Finish()

	md, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		md = metadata.New(nil)
	} else {
		md = md.Copy()
	}

	if err := tracer.Inject(span.Context(), opentracing.TextMap, MDReaderWriter{md}); err != nil {
		span.LogFields(log.Error(errors.WithMessage(err, "inject error")))

		return err
	}

	if err := invoker(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc, opts...); err != nil {
		span.LogFields(log.Error(errors.WithMessage(err, "call error")))

		return err
	}

	return nil
}
