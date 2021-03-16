package client

import (
	"context"
	"errors"
	"fmt"
	"log"

	tmrpc "github.com/tendermint/tendermint/rpc/client"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/client/reflection/codec"
	"github.com/cosmos/cosmos-sdk/client/reflection/descriptor"
	"github.com/cosmos/cosmos-sdk/client/reflection/tx"
	"github.com/cosmos/cosmos-sdk/client/reflection/unstructured"
	"github.com/cosmos/cosmos-sdk/types"
)

// Config defines Client configurations
type Config struct {
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

// Client defines a dynamic cosmos-sdk client, that can be used to query
// different cosmos sdk versions with different messages and available
// queries. It is gonna build the required codec in a dynamic fashion.
type Client struct {
	sdk                 reflection.ReflectionServiceClient
	tm                  tmrpc.Client
	cdc                 *codec.Codec // chain specific codec generated at run time
	accountInfoProvider AccountInfoProvider
	chainDesc           descriptor.Chain
}

// NewClient inits a client given the provided configurations
func NewClient(ctx context.Context, conf Config) (*Client, error) {
	c := &Client{
		sdk:                 conf.SDKReflectionClient,
		tm:                  conf.TMClient,
		cdc:                 codec.NewCodec(conf.ProtoImporter),
		accountInfoProvider: conf.AuthInfoProvider,
	}

	err := c.init(ctx)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// DialContext is going to create a new client by dialing to the tendermint and gRPC endpoints of the provided application.
func DialContext(ctx context.Context, grpcEndpoint, tmEndpoint string, accountInfoProvider AccountInfoProvider) (*Client, error) {
	conn, err := grpc.Dial(grpcEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	sdkReflect := reflection.NewReflectionServiceClient(conn)
	fetcher := protoDownloader{client: sdkReflect}

	tmRPC, err := tmhttp.New(tmEndpoint, "")
	if err != nil {
		return nil, err
	}

	return NewClient(ctx, Config{
		ProtoImporter:       fetcher,
		SDKReflectionClient: sdkReflect,
		TMClient:            tmRPC,
		AuthInfoProvider:    accountInfoProvider,
	})
}

// Codec exposes the client specific codec
func (c *Client) Codec() *codec.Codec {
	return c.cdc
}

func (c *Client) ChainDescriptor() descriptor.Chain {
	return c.chainDesc
}

func (c *Client) Query(ctx context.Context, method string, request proto.Message) (resp proto.Message, err error) {
	desc := c.chainDesc.Queriers().ByTMName(method)
	if desc == nil {
		return nil, fmt.Errorf("unknown method: %s", method)
	}

	reqBytes, err := c.cdc.Marshal(request)
	if err != nil {
		return nil, err
	}

	tmResp, err := c.tm.ABCIQuery(ctx, method, reqBytes)
	if err != nil {
		return nil, err
	}

	resp = dynamicpb.NewMessage(desc.Descriptor().Output())
	return resp, c.cdc.Unmarshal(tmResp.Response.Value, resp)
}

func (c *Client) QueryUnstructured(ctx context.Context, method string, request unstructured.Map) (resp proto.Message, err error) {
	desc := c.chainDesc.Queriers().ByTMName(method)
	if desc == nil {
		return nil, fmt.Errorf("unknown method: %s", method)
	}

	reqProto, err := request.Marshal(desc.Descriptor().Input())
	if err != nil {
		return nil, fmt.Errorf("unable to marshal request to proto message: %w", err)
	}

	b, err := c.cdc.Marshal(reqProto)
	if err != nil {
		return nil, err
	}

	tmResp, err := c.tm.ABCIQuery(ctx, method, b)
	if err != nil {
		return nil, err
	}

	resp = dynamicpb.NewMessage(desc.Descriptor().Output())
	return resp, c.cdc.Unmarshal(tmResp.Response.Value, resp)
}

func (c *Client) Tx(ctx context.Context, method string, request unstructured.Map, signerInfo tx.SignerInfo) (resp *ctypes.ResultBroadcastTxCommit, err error) {
	msgDesc := c.chainDesc.Deliverables().ByName(method)
	if msgDesc == nil {
		return nil, fmt.Errorf("deliverable not found: %s", method)
	}
	// marshal unstructured to proto.Message type
	pb, err := request.Marshal(msgDesc.Descriptor())
	if err != nil {
		return nil, err
	}

	pbJs, err := protojson.Marshal(pb)
	if err != nil {
		panic(err)
	}
	log.Printf("%s", pbJs)
	uBuilder := tx.NewUnsignedTxBuilder()
	uBuilder.AddMsg(pb)
	uBuilder.AddSigner(signerInfo)
	uBuilder.SetChainID("testing")
	uBuilder.SetFeePayer("cosmos1ujtnemf6jmfm995j000qdry064n5lq854gfe3j")
	uBuilder.SetFees(types.NewCoins(types.NewInt64Coin("stake", 10)))
	uBuilder.SetGasLimit(2500000)

	sBuilder, err := uBuilder.SignedBuilder()
	if err != nil {
		return nil, err
	}

	bytesToSign, err := sBuilder.BytesToSign(signerInfo.PubKey)
	if err != nil {
		return nil, err
	}

	signedBytes, err := c.accountInfoProvider.Sign(ctx, signerInfo.PubKey, bytesToSign)

	err = sBuilder.SetSignature(signerInfo.PubKey, signedBytes)
	if err != nil {
		return nil, err
	}

	txBytes, err := sBuilder.Bytes()
	if err != nil {
		return nil, err
	}

	tmResp, err := c.tm.BroadcastTxCommit(ctx, txBytes)
	return tmResp, err
}

func (c *Client) init(ctx context.Context) error {
	chainDescBuilder := descriptor.NewBuilder()
	err := c.buildQueries(ctx, chainDescBuilder)
	if err != nil {
		return err
	}

	err = c.buildDeliverables(ctx, chainDescBuilder)
	if err != nil {
		return err
	}

	err = c.resolveAnys(ctx)
	if err != nil {
		return err
	}

	c.chainDesc, err = chainDescBuilder.Build()
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) buildDeliverables(ctx context.Context, builder *descriptor.Builder) error {
	// get msg implementers
	msgImplementers, err := c.sdk.ListImplementations(ctx, &reflection.ListImplementationsRequest{InterfaceName: types.MsgInterfaceName})
	if err != nil {
		return err
	}
	// get service msg implementers
	svcMsgImplementers, err := c.sdk.ListImplementations(ctx, &reflection.ListImplementationsRequest{InterfaceName: types.ServiceMsgInterfaceName})
	if err != nil {
		return err
	}
	// join implementations as deliverables
	deliverablesProtoNames := make([]string, 0, len(msgImplementers.ImplementationMessageProtoNames)+len(svcMsgImplementers.ImplementationMessageProtoNames))
	deliverablesProtoNames = append(deliverablesProtoNames, msgImplementers.ImplementationMessageProtoNames...)
	deliverablesProtoNames = append(deliverablesProtoNames, svcMsgImplementers.ImplementationMessageProtoNames...)

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
		rptResp, err := c.sdk.ResolveProtoType(ctx, &reflection.ResolveProtoTypeRequest{Name: name})
		if err != nil {
			return err
		}
		desc, err := c.cdc.RegisterRawFileDescriptor(ctx, rptResp.RawDescriptor)
		// TODO: we should most likely check if error is file already registered and if it is
		// skip it as some people might define a module into a single proto file which we might have imported already
		if err != nil {
			return err
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
			err = builder.RegisterDeliverable(msgDesc)
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

func (c *Client) buildQueries(ctx context.Context, descBuilder *descriptor.Builder) error {
	queries, err := c.sdk.ListQueryServices(ctx, nil)
	if err != nil {
		return err
	}

	svcPerFile := make(map[string][]string)

	for _, q := range queries.Queries {
		_, exists := svcPerFile[q.ProtoFile]
		if !exists {
			svcPerFile[q.ProtoFile] = nil
		}

		svcPerFile[q.ProtoFile] = append(svcPerFile[q.ProtoFile], q.ServiceName)
	}

	svcDescriptors := make([][]byte, 0, len(svcPerFile))

	for file := range svcPerFile {
		rawDesc, err := c.sdk.ResolveService(ctx, &reflection.ResolveServiceRequest{FileName: file})
		if err != nil {
			return err
		}

		svcDescriptors = append(svcDescriptors, rawDesc.RawDescriptor)
	}

	for _, rawDesc := range svcDescriptors {
		fileDesc, err := c.cdc.RegisterRawFileDescriptor(ctx, rawDesc)
		if err != nil && !errors.Is(err, codec.ErrFileRegistered) {
			return err
		}
		err = registerServices(descBuilder, fileDesc)
		if err != nil {
			return err
		}

	}

	return nil
}

func (c *Client) resolveAnys(ctx context.Context) error {
	availableInterfaces, err := c.sdk.ListAllInterfaces(ctx, &reflection.ListAllInterfacesRequest{})
	if err != nil {
		return err
	}

	for _, implementation := range availableInterfaces.InterfaceNames {
		implsResp, err := c.sdk.ListImplementations(ctx, &reflection.ListImplementationsRequest{InterfaceName: implementation})
		if err != nil {
			return err
		}

		for _, implementer := range implsResp.ImplementationMessageProtoNames {
			err = c.resolveAny(ctx, implementation, implementer)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) resolveAny(ctx context.Context, implementation, implementer string) error {
	// if it's known skip
	if c.cdc.KnownMessage(implementer) {
		return nil
	}
	// if it's unknown then solve
	desc, err := c.sdk.ResolveProtoType(ctx, &reflection.ResolveProtoTypeRequest{Name: implementer})
	if err != nil {
		return err
	}
	_, err = c.cdc.RegisterRawFileDescriptor(ctx, desc.RawDescriptor)
	if err != nil {
		return err
	}

	return nil
}

func registerServices(builder *descriptor.Builder, file protoreflect.FileDescriptor) error {
	services := file.Services()

	for i := 0; i < services.Len(); i++ {
		svcDesc := services.Get(i)
		err := builder.RegisterQueryService(svcDesc)
		if err != nil {
			return err
		}
	}

	return nil
}
