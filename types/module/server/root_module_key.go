package server

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	sdkStore "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type RootModuleKey interface {
	ModuleKey
	sdkStore.StoreKey
	Derive(path []byte) DerivedModuleKey
}

type rootModuleKey struct {
	moduleName     string
	invokerFactory InvokerFactory
}

var _ RootModuleKey = &rootModuleKey{}

func (r *rootModuleKey) Name() string {
	return r.moduleName
}

func (r *rootModuleKey) String() string {
	return fmt.Sprintf("rootModuleKey{%p, %s}", r, r.moduleName)
}

func (r *rootModuleKey) Invoker(methodName string) (Invoker, error) {
	return r.invokerFactory(CallInfo{
		Method: methodName,
		Caller: r.ModuleID(),
	})
}

func (r *rootModuleKey) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, _ ...grpc.CallOption) error {
	invoker, err := r.Invoker(method)
	if err != nil {
		return err
	}

	return invoker(ctx, args, reply)
}

func (r *rootModuleKey) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("unsupported")
}

func (r *rootModuleKey) ModuleID() ModuleID {
	return ModuleID{ModuleName: r.moduleName}
}

func (r *rootModuleKey) Address() sdk.AccAddress {
	return r.ModuleID().Address()
}

func (r *rootModuleKey) Derive(path []byte) DerivedModuleKey {
	return DerivedModuleKey{
		moduleName:     r.moduleName,
		path:           path,
		invokerFactory: r.invokerFactory,
	}
}
