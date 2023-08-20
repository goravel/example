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

func OpentracingClient(tracer opentracing.Tracer, parentCtx context.Context) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		span, _ := opentracing.StartSpanFromContext(parentCtx,
			"call gRPC",
			opentracing.Tag{Key: string(ext.Component), Value: "gRPC"},
			ext.SpanKindRPCClient)

		defer span.Finish()

		md, ok := metadata.FromOutgoingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		} else {
			md = md.Copy()
		}

		err := tracer.Inject(span.Context(), opentracing.TextMap, MDReaderWriter{md})
		if err != nil {
			span.LogFields(log.Error(errors.WithMessage(err, "inject-error")))
		}

		newCtx := metadata.NewOutgoingContext(ctx, md)
		err = invoker(newCtx, method, req, reply, cc, opts...)
		if err != nil {
			span.LogFields(log.Error(errors.WithMessage(err, "call-error")))
		}
		return err
	}
}
