package interceptors

import (
	"io"
	"strings"

	"github.com/goravel/framework/facades"
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
	"google.golang.org/grpc/metadata"
)

type MDReaderWriter struct {
	metadata.MD
}

func (c MDReaderWriter) ForeachKey(handler func(key, val string) error) error {
	for k, vs := range c.MD {
		for _, v := range vs {
			if err := handler(k, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c MDReaderWriter) Set(key, val string) {
	key = strings.ToLower(key)
	c.MD[key] = append(c.MD[key], val)
}

func NewJaegerTracer() (opentracing.Tracer, io.Closer) {
	sender := transport.NewHTTPTransport(facades.Config().GetString("grpc.trace.endpoint"))
	tracer, closer := jaeger.NewTracer(
		facades.Config().GetString("app.env")+"."+facades.Config().GetString("app.name"),
		jaeger.NewConstSampler(true),
		jaeger.NewRemoteReporter(sender, jaeger.ReporterOptions.Logger(jaeger.StdLogger)),
	)

	opentracing.SetGlobalTracer(tracer)

	return tracer, closer
}
