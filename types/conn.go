package types

import (
	"context"

	"google.golang.org/grpc"
)

type InvokerConn interface {
	grpc.ClientConnInterface
	Invoker(methodName string) (Invoker, error)
}

type Invoker func(ctx context.Context, request, response interface{}, opts ...interface{}) error
