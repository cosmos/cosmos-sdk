package client

import (
	"context"
	"errors"
	"fmt"

	tmrpc "github.com/tendermint/tendermint/rpc/client"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/client/reflection/codec"
	"github.com/cosmos/cosmos-sdk/client/reflection/descriptor"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BuilderConfig defines the required *Builder configurations
type BuilderConfig struct {
	// ProtoImporter is used to import proto files dynamically
	ProtoImporter codec.ProtoImportsDownloader
	// SDKReflectionClient is the client used to build the codec
	// in a dynamic way, based on the chain the client is connected to.
	SDKReflectionClient reflection.ReflectionServiceClient
	// TMClient is the client used to interact with the tendermint endpoint
	// for queries and transaction posting
	TMClient tmrpc.Client
	// AuthInfoProvider takes care of providing authentication information
	// such as account sequence, number, address and signing capabilities.
	AuthInfoProvider AccountInfoProvider
}

// Builder is used to generate a *Client, it wraps all the complex logic
// required to build the *codec.Codec and the descriptor.Chain
type Builder struct {
	tm               tmrpc.Client
	sdk              reflection.ReflectionServiceClient
	chainDesc        *descriptor.Builder
	cdc              *codec.Codec
	authInfoProvider AccountInfoProvider
}

// NewBuilder instantiates a new *Client *Builder
func NewBuilder(opts BuilderConfig) *Builder {
	return &Builder{
		tm:               opts.TMClient,
		sdk:              opts.SDKReflectionClient,
		chainDesc:        descriptor.NewBuilder(),
		cdc:              codec.NewCodec(opts.ProtoImporter),
		authInfoProvider: opts.AuthInfoProvider,
	}
}

func (b *Builder) Build(ctx context.Context) (*Client, error) {
	err := b.queries(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to generate queries: %w", err)
	}

	err = b.deliverables(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to generate deliverables: %w", err)
	}

	err = b.resolveAnys(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to generate interfaces: %w", err)
	}

	chainDesc, err := b.chainDesc.Build()
	if err != nil {
		return nil, fmt.Errorf("unable to generate chain descriptor: %w", err)
	}
	return &Client{
		tm:                  b.tm,
		cdc:                 b.cdc,
		accountInfoProvider: b.authInfoProvider,
		chainDesc:           chainDesc,
	}, nil
}

// queries attempts to build all the available query service
func (b *Builder) queries(ctx context.Context) error {
	resp, err := b.sdk.ListQueryServices(ctx, nil)
	if err != nil {
		return err
	}

	// iterate over files to parse the descriptors
	for _, q := range resp.Queries {
		// get raw descriptor
		rawDesc, err := b.sdk.ResolveService(ctx, &reflection.ResolveServiceRequest{FileName: q.ProtoFile})
		if err != nil {
			return fmt.Errorf("unable to get file descriptor for %s: %w", q.ProtoFile, err)
		}
		// register proto file
		fd, err := b.cdc.RegisterRawFileDescriptor(ctx, rawDesc.RawDescriptor)
		// we ignore file already registered errors
		if err != nil && !errors.Is(err, codec.ErrFileRegistered) {
			return fmt.Errorf("unable to register descriptor for %s: %w", q.ProtoFile, err)
		}
		// register query services in chain descriptor Builder
		sds := fd.Services()
		for i := 0; i < sds.Len(); i++ {
			sd := sds.Get(i)
			err = b.chainDesc.RegisterQueryService(sd)
			if err != nil {
				return fmt.Errorf("unable to compute chain query descriptor for %s: %w", q.ProtoFile, err)
			}
		}
	}

	return nil
}

func (b *Builder) deliverables(ctx context.Context) error {
	// list sdk.Msg implementers
	msgImplsResp, err := b.sdk.ListImplementations(ctx, &reflection.ListImplementationsRequest{
		InterfaceName: sdk.MsgInterfaceName,
	})
	if err != nil {
		return fmt.Errorf("unable to list sdk.Msg implementations: %w", err)
	}

	// list sdk.ServiceMsg implementers
	svcMsgImplsResp, err := b.sdk.ListImplementations(ctx, &reflection.ListImplementationsRequest{
		InterfaceName: sdk.ServiceMsgInterfaceName,
	})
	if err != nil {
		return fmt.Errorf("unable to list sdk.ServiceMsg implementations: %w", err)
	}
	// join the implementations as deliverables
	deliverablesProtoNames := make(
		[]string,
		0,
		len(msgImplsResp.ImplementationMessageProtoNames)+len(svcMsgImplsResp.ImplementationMessageProtoNames),
	)

	deliverablesProtoNames = append(deliverablesProtoNames, msgImplsResp.ImplementationMessageProtoNames...)
	deliverablesProtoNames = append(deliverablesProtoNames, svcMsgImplsResp.ImplementationMessageProtoNames...)

	// we create a map which contains the message names that we expect to process
	// so in case one file contains multiple messages we need then we won't need
	// to resolve the same proto file multiple times :)
	expectedMsgs := make(map[string]struct{}, len(deliverablesProtoNames))
	foundMsgs := make(map[string]struct{}, len(deliverablesProtoNames))
	for _, name := range deliverablesProtoNames {
		expectedMsgs[name] = struct{}{}
	}

	// now resolve types
	for name := range expectedMsgs {
		// check if we already processed it
		if _, exists := foundMsgs[name]; exists {
			continue
		}
		rptResp, err := b.sdk.ResolveProtoType(ctx, &reflection.ResolveProtoTypeRequest{Name: name})
		if err != nil {
			return fmt.Errorf("unable to resolve proto type %s: %w", name, err)
		}
		desc, err := b.cdc.RegisterRawFileDescriptor(ctx, rptResp.RawDescriptor)
		// TODO: we should most likely check if error is file already registered and if it is
		// skip it as some people might define a module into a single proto file which we might have imported already
		if err != nil {
			return fmt.Errorf("unable to resolve proto type %s: %w", name, err)
		}
		// iterate over msgs
		found := false // we assume to always find our message in the file descriptor... but still
		for i := 0; i < desc.Messages().Len(); i++ {
			msgDesc := desc.Messages().Get(i)
			msgName := (string)(msgDesc.FullName())
			// check if msg is required
			if _, required := expectedMsgs[msgName]; !required {
				continue
			}
			// ok msg is required, so insert it in found list
			foundMsgs[msgName] = struct{}{}
			if msgName == name {
				found = true
			}
			// save in msgs
			err = b.chainDesc.RegisterDeliverable(msgDesc)
			if err != nil {
				return err
			}
		}
		if !found {
			return fmt.Errorf("unable to find message %s in resolved descriptor", name)
		}
	}
	return nil
}

// resolveAnys is gonna resolve all the concrete types we can have in *anypb.Any
// proto messages, we could do it in a different way by querying for the type registry
// but we do it in this way as in the future the sdk will provide interface identification
// for *anypb.Any field types which will allow this library to offer concrete type safety
// during marshalling, as now any proto.Message can be used to fill the type
func (b *Builder) resolveAnys(ctx context.Context) error {
	ifaces, err := b.sdk.ListAllInterfaces(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to get interfaces list: %s", err)
	}
	// we list all applications available interfaces
	for _, implementation := range ifaces.InterfaceNames {
		implementers, err := b.sdk.ListImplementations(ctx, &reflection.ListImplementationsRequest{
			InterfaceName: implementation,
		})
		if err != nil {
			return fmt.Errorf("unable to list implementations for %s: %w", implementation, err)
		}

		// register all implementers
		for _, implementer := range implementers.ImplementationMessageProtoNames {
			// if the type is known then we can skip
			if b.cdc.KnownMessage(implementer) {
				continue
			}
			// if unknown then solve
			rawDesc, err := b.sdk.ResolveProtoType(ctx, &reflection.ResolveProtoTypeRequest{Name: implementer})
			if err != nil {
				return fmt.Errorf("unable to resolve interface implemenenter %s concrete type %s: %w", implementation, implementer, err)
			}

			_, err = b.cdc.RegisterRawFileDescriptor(ctx, rawDesc.RawDescriptor)
			if err != nil {
				return fmt.Errorf("unable to register descriptor for implementer %s concrete type %s: %w", implementation, implementer, err)
			}
		}
	}

	return nil
}
