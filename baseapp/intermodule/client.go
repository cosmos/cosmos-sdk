package intermodule

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/types/address"
)

type interModuleClient struct {
	module         string
	address        []byte
	path           []byte
	invokerFactory InvokerFactory
}

func newInterModuleClient(module string, path []byte, invokerFactory InvokerFactory) *interModuleClient {
	return &interModuleClient{
		module:         module,
		path:           path,
		invokerFactory: invokerFactory,
		address:        address.Module(module, path),
	}
}

func (c *interModuleClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	invoker, err := c.invokerFactory(CallInfo{
		Method:      method,
		DerivedPath: c.path,
	})
	if err != nil {
		return err
	}

	return invoker(ctx, args, reply, opts...)
}

func (c *interModuleClient) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("unsupported")
}

func (c *interModuleClient) Address() []byte {
	return c.address
}

var _ appmodule.InterModuleClient = &interModuleClient{}

type rootInterModuleClient struct {
	*interModuleClient
}

func NewRootInterModuleClient(module string, invokerFactory InvokerFactory) appmodule.RootInterModuleClient {
	return &rootInterModuleClient{newInterModuleClient(module, nil, invokerFactory)}
}

func (r *rootInterModuleClient) DerivedClient(key []byte) appmodule.InterModuleClient {
	return newInterModuleClient(r.module, key, r.invokerFactory)
}

var _ appmodule.RootInterModuleClient = &rootInterModuleClient{}
