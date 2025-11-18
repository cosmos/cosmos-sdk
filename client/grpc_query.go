package client

import (
	gocontext "context"
	"fmt"
	"reflect"
	"strconv"

	abci "github.com/cometbft/cometbft/abci/types"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"

	errorsmod "cosmossdk.io/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// GRPCConn provides a method to get the appropriate gRPC connection based on block height.
type GRPCConn interface {
	GetGRPCConn(height int64) *grpc.ClientConn
}

// GRPCConnProvider manages gRPC connections with optional backup connections for historical queries.
type GRPCConnProvider struct {
	// DefaultConn is the primary gRPC connection
	DefaultConn *grpc.ClientConn
	// BackupConns maps block ranges to backup gRPC connections for routing historical queries
	BackupConns config.BackupGRPCConnections
}

// NewGRPCConnProvider creates a new GRPCConnProvider with the given connections.
func NewGRPCConnProvider(defaultConn *grpc.ClientConn, backupConns config.BackupGRPCConnections) *GRPCConnProvider {
	if backupConns == nil {
		backupConns = make(config.BackupGRPCConnections)
	}
	return &GRPCConnProvider{
		DefaultConn: defaultConn,
		BackupConns: backupConns,
	}
}

// GetGRPCConn returns the appropriate gRPC connection based on the block height.
// For height <= 0 (latest block), it returns the default connection.
// For positive heights, it checks if a backup connection exists for that height range.
func (g *GRPCConnProvider) GetGRPCConn(height int64) *grpc.ClientConn {
	// height = 0 means latest block, use the default connection
	if height <= 0 {
		return g.DefaultConn
	}

	// Check if there's a backup connection for this height
	for blockRange, conn := range g.BackupConns {
		if int64(blockRange[0]) <= height && int64(blockRange[1]) >= height {
			return conn
		}
	}

	// Default to the primary connection if no backup matches
	return g.DefaultConn
}

var _ gogogrpc.ClientConn = Context{}

// fallBackCodec is used by Context in case Codec is not set.
// it can process every gRPC type, except the ones which contain
// interfaces in their types.
var fallBackCodec = codec.NewProtoCodec(types.NewInterfaceRegistry())

func getHeightFromMetadata(grpcCtx gocontext.Context) int64 {
	md, ok := metadata.FromOutgoingContext(grpcCtx)
	if !ok {
		return 0
	}
	heights := md.Get(grpctypes.GRPCBlockHeightHeader)
	if len(heights) == 0 {
		return 0
	}
	height, err := strconv.ParseInt(heights[0], 10, 64)
	if err != nil {
		return 0
	}
	if height < 0 {
		return 0
	}
	return height
}

// Invoke implements the grpc ClientConn.Invoke method
func (ctx Context) Invoke(grpcCtx gocontext.Context, method string, req, reply any, opts ...grpc.CallOption) (err error) {
	// Two things can happen here:
	// 1. either we're broadcasting a Tx, in which call we call CometBFT's broadcast endpoint directly,
	// 2-1. or we are querying for state, in which case we call grpc if grpc client set.
	// 2-2. or we are querying for state, in which case we call ABCI's Query if grpc client not set.

	// In both cases, we don't allow empty request args (it will panic unexpectedly).
	if reflect.ValueOf(req).IsNil() {
		return errorsmod.Wrap(sdkerrors.ErrInvalidRequest, "request cannot be nil")
	}

	// Case 1. Broadcasting a Tx.
	if reqProto, ok := req.(*tx.BroadcastTxRequest); ok {
		res, ok := reply.(*tx.BroadcastTxResponse)
		if !ok {
			return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "expected %T, got %T", (*tx.BroadcastTxResponse)(nil), req)
		}

		broadcastRes, err := TxServiceBroadcast(grpcCtx, ctx, reqProto)
		if err != nil {
			return err
		}
		*res = *broadcastRes

		return err
	}

	if ctx.GRPCClient != nil {
		// Case 2-1. Invoke grpc.
		grpcConn := ctx.GRPCClient
		if ctx.GRPCConnProvider != nil {
			height := ctx.Height
			if height <= 0 {
				height = getHeightFromMetadata(grpcCtx)
			}

			grpcConn = ctx.GRPCConnProvider.GetGRPCConn(height)
		}
		return grpcConn.Invoke(grpcCtx, method, req, reply, opts...)
	}

	// Case 2-2. Querying state via abci query.
	reqBz, err := ctx.gRPCCodec().Marshal(req)
	if err != nil {
		return err
	}

	// parse height header
	height := getHeightFromMetadata(grpcCtx)
	if height < 0 {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidRequest,
			"client.Context.Invoke: height (%d) from %q must be >= 0", height, grpctypes.GRPCBlockHeightHeader)
	}

	if height > 0 {
		ctx = ctx.WithHeight(height)
	}

	abciReq := abci.RequestQuery{
		Path:   method,
		Data:   reqBz,
		Height: ctx.Height,
	}

	res, err := ctx.QueryABCI(abciReq)
	if err != nil {
		return err
	}

	err = ctx.gRPCCodec().Unmarshal(res.Value, reply)
	if err != nil {
		return err
	}

	// Create header metadata. For now the headers contain:
	// - block height
	// We then parse all the call options, if the call option is a
	// HeaderCallOption, then we manually set the value of that header to the
	// metadata.
	md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(res.Height, 10))
	for _, callOpt := range opts {
		header, ok := callOpt.(grpc.HeaderCallOption)
		if !ok {
			continue
		}

		*header.HeaderAddr = md
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

// gRPCCodec checks if Context's Codec is codec.GRPCCodecProvider
// otherwise it returns fallBackCodec.
func (ctx Context) gRPCCodec() encoding.Codec {
	if ctx.Codec == nil {
		return fallBackCodec.GRPCCodec()
	}

	pc, ok := ctx.Codec.(codec.GRPCCodecProvider)
	if !ok {
		return fallBackCodec.GRPCCodec()
	}

	return pc.GRPCCodec()
}
