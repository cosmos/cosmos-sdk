package intermodule

import (
	"context"

	"google.golang.org/grpc"
)

type InvokerFactory func(callInfo CallInfo) (Invoker, error)

type Invoker func(ctx context.Context, request, response interface{}, opts ...grpc.CallOption) error

type CallInfo struct {
	Method      string
	DerivedPath []byte
}
