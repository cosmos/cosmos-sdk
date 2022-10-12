package baseapp

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/types/address"
)

type invokerFactory func(callInfo callInfo) (invoker, error)

type invoker func(ctx context.Context, request, response interface{}, opts ...grpc.CallOption) error

type callInfo struct {
	method        string
	callingModule string
	derivedPath   []byte
}

type interModuleClient struct {
	module         string
	address        []byte
	path           []byte
	invokerFactory invokerFactory
}

func newInterModuleClient(module string, path []byte, invokerFactory invokerFactory) *interModuleClient {
	return &interModuleClient{
		module:         module,
		path:           path,
		invokerFactory: invokerFactory,
		address:        address.Module(module, path),
	}
}

func (c *interModuleClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	invoker, err := c.invokerFactory(callInfo{
		method:        method,
		callingModule: c.module,
		derivedPath:   c.path,
	})
	if err != nil {
		return err
	}

	return invoker(ctx, args, reply, opts...)
}

func (c *interModuleClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *interModuleClient) Address() []byte {
	return c.address
}

var _ appmodule.InterModuleClient = &interModuleClient{}

type rootInterModuleClient struct {
	*interModuleClient
}

func newRootInterModuleClient(module string, invokerFactory invokerFactory) *rootInterModuleClient {
	return &rootInterModuleClient{newInterModuleClient(module, nil, invokerFactory)}
}

func (r *rootInterModuleClient) DerivedClient(key []byte) appmodule.InterModuleClient {
	return newInterModuleClient(r.module, key, r.invokerFactory)
}

var _ appmodule.RootInterModuleClient = &rootInterModuleClient{}
