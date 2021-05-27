package app

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"google.golang.org/grpc"
)

type ModuleKey interface {
	InvokerConn

	ModuleID() ModuleID
	Address() []byte
}

type RootModuleKey interface {
	ModuleKey
	sdk.StoreKey
}

type InvokerConn interface {
	grpc.ClientConnInterface
	Invoker(methodName string) (Invoker, error)
}

type Invoker func(ctx context.Context, request, response interface{}, opts ...interface{}) error

type ModuleID struct {
	ModuleName string
	Path       []byte
}
