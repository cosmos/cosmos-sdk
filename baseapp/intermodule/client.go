package intermodule

import (
	"context"
	"fmt"

	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	protov2 "google.golang.org/protobuf/proto"

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

func (c *interModuleClient) InvokerByMethod(method string) (appmodule.InterModuleInvoker, error) {
	return c.invokerFactory(CallInfo{
		Method:      method,
		DerivedPath: c.path,
	})
}

func (c *interModuleClient) InvokerByRequest(request interface{}) (appmodule.InterModuleInvoker, error) {
	var method string
	if msg, ok := request.(protov2.Message); ok {
		method = string(msg.ProtoReflect().Descriptor().FullName())
	} else if msg, ok := request.(proto.Message); ok {
		method = proto.MessageName(msg)
	} else {
		return nil, fmt.Errorf("expected a proto message, got %T", request)
	}

	return c.InvokerByMethod(method)
}

func (c *interModuleClient) Invoke(ctx context.Context, method string, args interface{}, reply interface{}, opts ...grpc.CallOption) error {
	var invoker appmodule.InterModuleInvoker
	var err error
	if method == "" {
		invoker, err = c.InvokerByRequest(args)
	} else {
		invoker, err = c.InvokerByRequest(method)
	}
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
