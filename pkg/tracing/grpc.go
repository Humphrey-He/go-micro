package tracing

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

const (
	GRPCTraceIDKey = "x-trace-id"
	GRPCSpanIDKey  = "x-span-id"
)

func GRPCServerInterceptor() grpc.ServerOption {
	return grpc.ChainUnaryInterceptor(serverUnaryInterceptor())
}

func serverUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(nil)
		}

		propagator := propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)

		ctx = propagator.Extract(ctx, MetadataCarrier{md})

		spanName := info.FullMethod
		ctx, span := Tracer().Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				attribute.String("rpc.method", info.FullMethod),
				attribute.String("rpc.service", info.FullMethod),
			),
		)
		defer span.End()

		resp, err := handler(ctx, req)

		if err != nil {
			span.SetAttributes(attribute.String("error", err.Error()))
			span.RecordError(err)
		}

		return resp, err
	}
}

func GRPCClientInterceptor() grpc.DialOption {
	return grpc.WithChainUnaryInterceptor(clientUnaryInterceptor())
}

func clientUnaryInterceptor() grpc.UnaryClientInterceptor {
	return func(
		ctx context.Context,
		method string,
		req interface{},
		reply interface{},
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

type MetadataCarrier struct {
	metadata.MD
}

func (c MetadataCarrier) Get(key string) string {
	if v := c.MD.Get(key); len(v) > 0 {
		return v[0]
	}
	return ""
}

func (c MetadataCarrier) Set(key, value string) {
	c.MD.Set(key, value)
}

func (c MetadataCarrier) Keys() []string {
	out := make([]string, 0, len(c.MD))
	for k := range c.MD {
		out = append(out, k)
	}
	return out
}
