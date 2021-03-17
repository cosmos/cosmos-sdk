package client

import (
	"context"
	"fmt"
	"log"

	tmrpc "github.com/tendermint/tendermint/rpc/client"
	tmhttp "github.com/tendermint/tendermint/rpc/client/http"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/dynamicpb"

	"github.com/cosmos/cosmos-sdk/client/grpc/reflection"
	"github.com/cosmos/cosmos-sdk/client/reflection/codec"
	"github.com/cosmos/cosmos-sdk/client/reflection/descriptor"
	"github.com/cosmos/cosmos-sdk/client/reflection/tx"
	"github.com/cosmos/cosmos-sdk/client/reflection/unstructured"
	"github.com/cosmos/cosmos-sdk/types"
)

// Client defines a dynamic cosmos-sdk client, that can be used to query
// different cosmos sdk versions with different messages and available
// queries.
type Client struct {
	tm                  tmrpc.Client
	cdc                 *codec.Codec // chain specific codec generated at run time
	accountInfoProvider AccountInfoProvider
	chainDesc           descriptor.Chain // chain specific descriptor generated at runtime
}

// Dial is going to create a new client by dialing to the tendermint and gRPC endpoints of the provided application.
func Dial(ctx context.Context, grpcEndpoint, tmEndpoint string, accountInfoProvider AccountInfoProvider) (*Client, error) {
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

	builder := NewBuilder(BuilderConfig{
		ProtoImporter:       fetcher,
		SDKReflectionClient: sdkReflect,
		TMClient:            tmRPC,
		AuthInfoProvider:    accountInfoProvider,
	})
	c, err := builder.Build(ctx)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// Codec exposes the client specific codec
func (c *Client) Codec() *codec.Codec {
	return c.cdc
}

func (c *Client) ChainDescriptor() descriptor.Chain {
	return c.chainDesc
}

// QueryTM routes the query via tendermint abci.Query, given the tendermint full query name
func (c *Client) QueryTM(ctx context.Context, method string, request proto.Message) (resp proto.Message, err error) {
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

func (c *Client) Query(ctx context.Context, request proto.Message) (resp proto.Message, err error) {
	desc := c.chainDesc.Queriers().ByInput(request)
	if desc == nil {
		return nil, fmt.Errorf("unknown input: %s", request.ProtoReflect().Descriptor().FullName())
	}
	return c.QueryTM(ctx, desc.TMQueryPath(), request)
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

func (c *Client) Tx() *Tx {
	return NewTx()
}

func (c *Client) TxBeta(ctx context.Context, method string, request unstructured.Map, signerInfo tx.SignerInfo) (resp *ctypes.ResultBroadcastTxCommit, err error) {
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
