package systemtests

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"testing"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	client "github.com/cometbft/cometbft/rpc/client/http"
	cmtypes "github.com/cometbft/cometbft/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
)

// RPCClient is a test helper to interact with a node via the RPC endpoint.
type RPCClient struct {
	client *client.HTTP
	t      *testing.T
}

// NewRPCClient constructor
func NewRPCClient(t *testing.T, addr string) RPCClient {
	t.Helper()
	httpClient, err := client.New(addr, "/websocket")
	require.NoError(t, err)
	require.NoError(t, httpClient.Start())
	t.Cleanup(func() { _ = httpClient.Stop() })
	return RPCClient{client: httpClient, t: t}
}

// Validators returns list of validators
func (r RPCClient) Validators() []*cmtypes.Validator {
	v, err := r.client.Validators(context.Background(), nil, nil, nil)
	require.NoError(r.t, err)
	return v.Validators
}

func (r RPCClient) Invoke(ctx context.Context, method string, req, reply interface{}, opts ...grpc.CallOption) error {
	if reflect.ValueOf(req).IsNil() {
		return errors.New("request cannot be nil")
	}

	ir := types.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir).GRPCCodec()

	reqBz, err := cdc.Marshal(req)
	if err != nil {
		return err
	}

	var height int64
	md, _ := metadata.FromOutgoingContext(ctx)
	if heights := md.Get(grpctypes.GRPCBlockHeightHeader); len(heights) > 0 {
		height, err := strconv.ParseInt(heights[0], 10, 64)
		if err != nil {
			return err
		}
		if height < 0 {
			return errors.New("height must be greater than or equal to 0")
		}
	}

	abciReq := abci.QueryRequest{
		Path:   method,
		Data:   reqBz,
		Height: height,
	}

	abciOpts := rpcclient.ABCIQueryOptions{
		Height: height,
		Prove:  abciReq.Prove,
	}

	result, err := r.client.ABCIQueryWithOptions(ctx, abciReq.Path, abciReq.Data, abciOpts)
	if err != nil {
		return err
	}

	if !result.Response.IsOK() {
		return errors.New(result.Response.String())
	}

	err = cdc.Unmarshal(result.Response.Value, reply)
	if err != nil {
		return err
	}

	md = metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(result.Response.Height, 10))
	for _, callOpt := range opts {
		header, ok := callOpt.(grpc.HeaderCallOption)
		if !ok {
			continue
		}

		*header.HeaderAddr = md
	}

	return types.UnpackInterfaces(reply, ir)
}

func (r RPCClient) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("not implemented")
}
