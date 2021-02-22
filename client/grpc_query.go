package client

import (
	gocontext "context"
	"fmt"
	"reflect"
	"strconv"

	gogogrpc "github.com/gogo/protobuf/grpc"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/encoding/proto"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

var _ gogogrpc.ClientConn = Context{}

var protoCodec = encoding.GetCodec(proto.Name)

// Invoke implements the grpc ClientConn.Invoke method
func (ctx Context) Invoke(grpcCtx gocontext.Context, method string, req, reply interface{}, opts ...grpc.CallOption) (err error) {
	// Two things can happen here:
	// 1. either we're broadcasting a Tx, in which call we call Tendermint's broadcast endpoint directly,
	// 2. or we are querying for state, in which case we call ABCI's Query.

	// In both cases, we don't allow empty request req (it will panic unexpectedly).
	if reflect.ValueOf(req).IsNil() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "request cannot be nil")
	}

	// Case 1. Broadcasting a Tx.
	if reqProto, ok := req.(*tx.BroadcastTxRequest); ok {
		if !ok {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected %T, got %T", (*tx.BroadcastTxRequest)(nil), req)
		}
		resProto, ok := reply.(*tx.BroadcastTxResponse)
		if !ok {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected %T, got %T", (*tx.BroadcastTxResponse)(nil), req)
		}

		broadcastRes, err := TxServiceBroadcast(grpcCtx, ctx, reqProto)
		if err != nil {
			return err
		}
		*resProto = *broadcastRes

		return err
	}

	// Case 2. Querying state.
	inMd, _ := metadata.FromOutgoingContext(grpcCtx)
	abciRes, outMd, err := RunGRPCQuery(ctx, grpcCtx, method, req, inMd)
	if err != nil {
		return err
	}

	err = protoCodec.Unmarshal(abciRes.Value, reply)
	if err != nil {
		return err
	}

	for _, callOpt := range opts {
		header, ok := callOpt.(grpc.HeaderCallOption)
		if !ok {
			continue
		}

		*header.HeaderAddr = outMd
	}

	if ctx.InterfaceRegistry != nil {
		return types.UnpackInterfaces(reply, ctx.InterfaceRegistry)
	}

	return nil
}

// NewStream implements the grpc ClientConn.NewStream method
func (Context) NewStream(gocontext.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming rpc not supported")
}

// RunGRPCQuery runs a gRPC query from the clientCtx, given all necessary
// arguments for the gRPC method, and returns the ABCI response. It is used
// to factorize code between client (Invoke) and server (RegisterGRPCServer)
// gRPC handlers.
func RunGRPCQuery(ctx Context, grpcCtx gocontext.Context, method string, req interface{}, md metadata.MD) (abci.ResponseQuery, metadata.MD, error) {
	reqBz, err := protoCodec.Marshal(req)
	if err != nil {
		return abci.ResponseQuery{}, nil, err
	}

	// parse height header
	if heights := md.Get(grpctypes.GRPCBlockHeightHeader); len(heights) > 0 {
		height, err := strconv.ParseInt(heights[0], 10, 64)
		if err != nil {
			return abci.ResponseQuery{}, nil, err
		}
		if height < 0 {
			return abci.ResponseQuery{}, nil, sdkerrors.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"client.Context.Invoke: height (%d) from %q must be >= 0", height, grpctypes.GRPCBlockHeightHeader)
		}

		ctx = ctx.WithHeight(height)
	}

	abciReq := abci.RequestQuery{
		Path: method,
		Data: reqBz,
	}

	abciRes, err := ctx.QueryABCI(abciReq)
	if err != nil {
		return abci.ResponseQuery{}, nil, err
	}

	// Create header metadata. For now the headers contain:
	// - block height
	// We then parse all the call options, if the call option is a
	// HeaderCallOption, then we manually set the value of that header to the
	// metadata.
	md = metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(abciRes.Height, 10))

	return abciRes, md, nil
}
