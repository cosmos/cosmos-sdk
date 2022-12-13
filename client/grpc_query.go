package client

import (
	gocontext "context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"github.com/cosmos/gogoproto/proto"
	abci "github.com/tendermint/tendermint/abci/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

var _ gogogrpc.ClientConn = Context{}

// fallBackCodec is used by Context in case Codec is not set.
// it can process every gRPC type, except the ones which contain
// interfaces in their types.
var fallBackCodec = codec.NewProtoCodec(failingInterfaceRegistry{})

// Invoke implements the grpc ClientConn.Invoke method
func (ctx Context) Invoke(grpcCtx gocontext.Context, method string, req, reply interface{}, opts ...grpc.CallOption) (err error) {
	// Two things can happen here:
	// 1. either we're broadcasting a Tx, in which call we call Tendermint's broadcast endpoint directly,
	// 2-1. or we are querying for state, in which case we call grpc if grpc client set.
	// 2-2. or we are querying for state, in which case we call ABCI's Query if grpc client not set.

	// In both cases, we don't allow empty request args (it will panic unexpectedly).
	if reflect.ValueOf(req).IsNil() {
		return sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "request cannot be nil")
	}

	// Case 1. Broadcasting a Tx.
	if reqProto, ok := req.(*tx.BroadcastTxRequest); ok {
		res, ok := reply.(*tx.BroadcastTxResponse)
		if !ok {
			return sdkerrors.Wrapf(sdkerrors.ErrInvalidRequest, "expected %T, got %T", (*tx.BroadcastTxResponse)(nil), req)
		}

		broadcastRes, err := TxServiceBroadcast(grpcCtx, ctx, reqProto)
		if err != nil {
			return err
		}
		*res = *broadcastRes

		return err
	}

	// Certain queries must not be be concurrent with ABCI to function correctly.
	// As a result, we direct them to the ABCI flow where they get syncronized.
	_, isSimulationRequest := req.(*tx.SimulateRequest)
	isTendermintQuery := strings.Contains(method, "tendermint")
	grpcConcurrentEnabled := ctx.GRPCConcurrency
	isGRPCAllowed := !isTendermintQuery && !isSimulationRequest && grpcConcurrentEnabled

	requestedHeight, err := selectHeight(ctx, grpcCtx)
	if err != nil {
		return err
	}

	if ctx.GRPCClient != nil && isGRPCAllowed {
		md := metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(requestedHeight, 10))
		context := metadata.NewOutgoingContext(grpcCtx, md)
		// Case 2-1. Invoke grpc.
		return ctx.GRPCClient.Invoke(context, method, req, reply, opts...)
	}

	// Case 2-2. Querying state via abci query.
	reqBz, err := ctx.gRPCCodec().Marshal(req)
	if err != nil {
		return err
	}

	abciReq := abci.RequestQuery{
		Path:   method,
		Data:   reqBz,
		Height: requestedHeight,
	}

	res, err := ctx.QueryABCI(abciReq)
	if err != nil {
		return err
	}

	if err := ctx.gRPCCodec().Unmarshal(res.Value, reply); err != nil {
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

// selectHeight returns the height chosen from client context and grpc context.
// If exists, height extracted from grpcCtx takes precedence.
func selectHeight(clientContext Context, grpcCtx gocontext.Context) (int64, error) {
	var height int64
	if clientContext.Height > 0 {
		height = clientContext.Height
	}

	md, _ := metadata.FromOutgoingContext(grpcCtx)
	if heights := md.Get(grpctypes.GRPCBlockHeightHeader); len(heights) > 0 {
		var err error
		height, err = strconv.ParseInt(heights[0], 10, 64)
		if err != nil {
			return 0, err
		}
	}
	return height, nil
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

var _ types.InterfaceRegistry = failingInterfaceRegistry{}

// failingInterfaceRegistry is used by the fallback codec
// in case Context's Codec is not set.
type failingInterfaceRegistry struct{}

// errCodecNotSet is return by failingInterfaceRegistry in case there are attempt to decode
// or encode a type which contains an interface field.
var errCodecNotSet = errors.New("client: cannot encode or decode type which requires the application specific codec")

func (f failingInterfaceRegistry) UnpackAny(any *types.Any, iface interface{}) error {
	return errCodecNotSet
}

func (f failingInterfaceRegistry) Resolve(typeUrl string) (proto.Message, error) {
	return nil, errCodecNotSet
}

func (f failingInterfaceRegistry) RegisterInterface(protoName string, iface interface{}, impls ...proto.Message) {
	panic("cannot be called")
}

func (f failingInterfaceRegistry) RegisterImplementations(iface interface{}, impls ...proto.Message) {
	panic("cannot be called")
}

func (f failingInterfaceRegistry) ListAllInterfaces() []string {
	panic("cannot be called")
}

func (f failingInterfaceRegistry) ListImplementations(ifaceTypeURL string) []string {
	panic("cannot be called")
}

func (f failingInterfaceRegistry) EnsureRegistered(iface interface{}) error {
	panic("cannot be called")
}
