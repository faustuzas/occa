package auth

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	pkgerrors "github.com/faustuzas/occa/src/pkg/errors"
	pkggrpc "github.com/faustuzas/occa/src/pkg/grpc"
	pkginstrument "github.com/faustuzas/occa/src/pkg/instrument"
)

func GRPCStreamTokenAuthorizationInterceptor(_ pkginstrument.Instrumentation, validator TokenValidator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		_, err := authorizeRequest(ss.Context(), validator)
		if err != nil {
			return pkggrpc.Error(pkgerrors.ErrUnauthorized(err))
		}

		return handler(srv, ss)
	}
}

func authorizeRequest(ctx context.Context, validator TokenValidator) (Principal, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return Principal{}, fmt.Errorf("missing gRPC metadata")
	}

	values := md.Get("authorization")
	if len(values) != 1 {
		return Principal{}, fmt.Errorf("missing Authorization token")
	}
	token := values[0]

	if strings.HasPrefix(token, "Bearer ") {
		token = token[len("Bearer "):]
	}

	principal, err := validator.Validate(ctx, token)
	if err != nil {
		return Principal{}, err
	}

	return principal, nil
}

func GRPCStreamNoopInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return handler(srv, ss)
	}
}
