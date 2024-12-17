package comet

import (
	"context"
	"errors"
	"strconv"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const grpcBlockHeightHeader = "x-cosmos-block-height"

var (
	_ gogogrpc.ClientConn      = &CometBFTBroadcaster{}
	_ grpc.ClientConnInterface = &CometBFTBroadcaster{}
)

func (c *CometBFTBroadcaster) NewStream(_ context.Context, _ *grpc.StreamDesc, _ string, _ ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, errors.New("not implemented")
}

// Invoke implements the gRPC ClientConn interface by forwarding the RPC call to CometBFT's ABCI Query.
// It marshals the request, sends it as an ABCI query, and unmarshals the response.
func (c *CometBFTBroadcaster) Invoke(ctx context.Context, method string, req, reply interface{}, opts ...grpc.CallOption) (err error) {
	reqBz, err := c.getRPCCodec().Marshal(req)
	if err != nil {
		return err
	}

	// parse height header
	md, _ := metadata.FromOutgoingContext(ctx)
	var height int64
	if heights := md.Get(grpcBlockHeightHeader); len(heights) > 0 {
		height, err = strconv.ParseInt(heights[0], 10, 64)
		if err != nil {
			return err
		}
		if height < 0 {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"client.Context.Invoke: height (%d) from %q must be >= 0", height, grpcBlockHeightHeader)
		}
	}

	abciR := abci.QueryRequest{
		Path:   method,
		Data:   reqBz,
		Height: height,
	}

	res, err := c.queryABCI(ctx, abciR)
	if err != nil {
		return err
	}

	err = c.getRPCCodec().Unmarshal(res.Value, reply)
	if err != nil {
		return err
	}

	// Create header metadata. For now the headers contain:
	// - block height
	// We then parse all the call options, if the call option is a
	// HeaderCallOption, then we manually set the value of that header to the
	// metadata.
	md = metadata.Pairs(grpcBlockHeightHeader, strconv.FormatInt(res.Height, 10))
	for _, callOpt := range opts {
		header, ok := callOpt.(grpc.HeaderCallOption)
		if !ok {
			continue
		}

		*header.HeaderAddr = md
	}

	if c.cdc.InterfaceRegistry() != nil {
		return types.UnpackInterfaces(reply, c.cdc.InterfaceRegistry())
	}

	return nil
}

// queryABCI performs an ABCI query request to the CometBFT RPC client.
// If the RPC query fails or returns a non-OK response, it will return an error.
// The response is converted from ABCI error codes to gRPC status errors.
func (c *CometBFTBroadcaster) queryABCI(ctx context.Context, req abci.QueryRequest) (abci.QueryResponse, error) {
	opts := rpcclient.ABCIQueryOptions{
		Height: req.Height,
		Prove:  req.Prove,
	}

	result, err := c.rpcClient.ABCIQueryWithOptions(ctx, req.Path, req.Data, opts)
	if err != nil {
		return abci.QueryResponse{}, err
	}

	if !result.Response.IsOK() {
		return abci.QueryResponse{}, sdkErrorToGRPCError(result.Response)
	}

	return result.Response, nil
}

// sdkErrorToGRPCError converts an ABCI query response error code to an appropriate gRPC status error.
// It maps common SDK error codes to their gRPC equivalents:
// - ErrInvalidRequest -> InvalidArgument
// - ErrUnauthorized -> Unauthenticated
// - ErrKeyNotFound -> NotFound
// Any other error codes are mapped to Unknown.
func sdkErrorToGRPCError(resp abci.QueryResponse) error {
	switch resp.Code {
	case sdkerrors.ErrInvalidRequest.ABCICode():
		return status.Error(codes.InvalidArgument, resp.Log)
	case sdkerrors.ErrUnauthorized.ABCICode():
		return status.Error(codes.Unauthenticated, resp.Log)
	case sdkerrors.ErrKeyNotFound.ABCICode():
		return status.Error(codes.NotFound, resp.Log)
	default:
		return status.Error(codes.Unknown, resp.Log)
	}
}

// getRPCCodec returns the gRPC codec for the CometBFT broadcaster.
// If the broadcaster's codec implements GRPCCodecProvider, it returns its gRPC codec.
// Otherwise, it creates a new ProtoCodec with the broadcaster's interface registry and returns its gRPC codec.
func (c *CometBFTBroadcaster) getRPCCodec() encoding.Codec {
	cdc, ok := c.cdc.(codec.GRPCCodecProvider)
	if !ok {
		return codec.NewProtoCodec(c.cdc.InterfaceRegistry()).GRPCCodec()
	}

	return cdc.GRPCCodec()
}
