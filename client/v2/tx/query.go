package tx

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"fmt"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	"github.com/cosmos/cosmos-sdk/types/tx"
	gogogrpc "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/metadata"
	"reflect"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	abci "github.com/cometbft/cometbft/abci/types"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/store/rootmulti"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ gogogrpc.ClientConn = txContext{}

	// fallBackCodec is used by Context in case Codec is not set.
	// it can process every gRPC type, except the ones which contain
	// interfaces in their types.
	fallBackCodec = codec.NewProtoCodec(types.NewInterfaceRegistry())
)

// GetNode returns an RPC client. If the context's client is not defined, an
// error is returned.
func (ctx txContext) GetNode() (CometRPC, error) {
	if ctx.Client == nil {
		return nil, errors.New("no RPC client is defined in offline mode")
	}

	return ctx.Client, nil
}

// Query performs a query to a CometBFT node with the provided path.
// It returns the result and height of the query upon success or an error if
// the query fails.
func (ctx txContext) Query(path string) ([]byte, int64, error) {
	return ctx.query(path, nil)
}

// QueryWithData performs a query to a CometBFT node with the provided path
// and a data payload. It returns the result and height of the query upon success
// or an error if the query fails.
func (ctx txContext) QueryWithData(path string, data []byte) ([]byte, int64, error) {
	return ctx.query(path, data)
}

// QueryStore performs a query to a CometBFT node with the provided key and
// store name. It returns the result and height of the query upon success
// or an error if the query fails.
func (ctx txContext) QueryStore(key []byte, storeName string) ([]byte, int64, error) {
	return ctx.queryStore(key, storeName, "key")
}

// QueryABCI performs a query to a CometBFT node with the provide RequestQuery.
// It returns the ResultQuery obtained from the query. The height used to perform
// the query is the RequestQuery Height if it is non-zero, otherwise the context
// height is used.
func (ctx txContext) QueryABCI(req abci.RequestQuery) (abci.ResponseQuery, error) {
	return ctx.queryABCI(req)
}

// GetFromAddress returns the from address from the context's name.
func (ctx txContext) GetFromAddress() sdk.AccAddress {
	return ctx.FromAddress
}

// Invoke implements the grpc ClientConn.Invoke method
func (ctx txContext) Invoke(grpcCtx context.Context, method string, req, reply interface{}, opts ...grpc.CallOption) (err error) {
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
		return ctx.GRPCClient.Invoke(grpcCtx, method, req, reply, opts...)
	}

	// Case 2-2. Querying state via abci query.
	reqBz, err := ctx.gRPCCodec().Marshal(req)
	if err != nil {
		return err
	}

	// parse height header
	md, _ := metadata.FromOutgoingContext(grpcCtx)
	if heights := md.Get(grpctypes.GRPCBlockHeightHeader); len(heights) > 0 {
		height, err := strconv.ParseInt(heights[0], 10, 64)
		if err != nil {
			return err
		}
		if height < 0 {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidRequest,
				"client.Context.Invoke: height (%d) from %q must be >= 0", height, grpctypes.GRPCBlockHeightHeader)
		}

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
	md = metadata.Pairs(grpctypes.GRPCBlockHeightHeader, strconv.FormatInt(res.Height, 10))
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
func (txContext) NewStream(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error) {
	return nil, fmt.Errorf("streaming rpc not supported")
}

// gRPCCodec checks if Context's Codec is codec.GRPCCodecProvider
// otherwise it returns fallBackCodec.
func (ctx txContext) gRPCCodec() encoding.Codec {
	if ctx.Codec == nil {
		return fallBackCodec.GRPCCodec()
	}

	pc, ok := ctx.Codec.(codec.GRPCCodecProvider)
	if !ok {
		return fallBackCodec.GRPCCodec()
	}

	return pc.GRPCCodec()
}

func (ctx txContext) queryABCI(req abci.RequestQuery) (abci.ResponseQuery, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	var queryHeight int64
	if req.Height != 0 {
		queryHeight = req.Height
	} else {
		// fallback on the context height
		queryHeight = ctx.Height
	}

	opts := rpcclient.ABCIQueryOptions{
		Height: queryHeight,
		Prove:  req.Prove,
	}

	result, err := node.ABCIQueryWithOptions(context.Background(), req.Path, req.Data, opts)
	if err != nil {
		return abci.ResponseQuery{}, err
	}

	if !result.Response.IsOK() {
		return abci.ResponseQuery{}, sdkErrorToGRPCError(result.Response)
	}

	// data from trusted node or subspace query doesn't need verification
	if !opts.Prove || !isQueryStoreWithProof(req.Path) {
		return result.Response, nil
	}

	return result.Response, nil
}

// TxServiceBroadcast is a helper function to broadcast a Tx with the correct gRPC types
// from the tx service. Calls `clientCtx.BroadcastTx` under the hood.
func TxServiceBroadcast(_ context.Context, clientCtx txContext, req *tx.BroadcastTxRequest) (*tx.BroadcastTxResponse, error) {
	if req == nil || req.TxBytes == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	clientCtx = clientCtx.WithBroadcastMode(normalizeBroadcastMode(req.Mode))
	resp, err := clientCtx.BroadcastTx(req.TxBytes)
	if err != nil {
		return nil, err
	}

	return &tx.BroadcastTxResponse{
		TxResponse: resp,
	}, nil
}

func sdkErrorToGRPCError(resp abci.ResponseQuery) error {
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

// query performs a query to a CometBFT node with the provided store name
// and path. It returns the result and height of the query upon success
// or an error if the query fails.
func (ctx txContext) query(path string, key []byte) ([]byte, int64, error) {
	resp, err := ctx.queryABCI(abci.RequestQuery{
		Path:   path,
		Data:   key,
		Height: ctx.Height,
	})
	if err != nil {
		return nil, 0, err
	}

	return resp.Value, resp.Height, nil
}

// queryStore performs a query to a CometBFT node with the provided a store
// name and path. It returns the result and height of the query upon success
// or an error if the query fails.
func (ctx txContext) queryStore(key []byte, storeName, endPath string) ([]byte, int64, error) {
	path := fmt.Sprintf("/store/%s/%s", storeName, endPath)
	return ctx.query(path, key)
}

// isQueryStoreWithProof expects a format like /<queryType>/<storeName>/<subpath>
// queryType must be "store" and subpath must be "key" to require a proof.
func isQueryStoreWithProof(path string) bool {
	if !strings.HasPrefix(path, "/") {
		return false
	}

	paths := strings.SplitN(path[1:], "/", 3)

	switch {
	case len(paths) != 3:
		return false
	case paths[0] != "store":
		return false
	case rootmulti.RequireProof("/" + paths[2]):
		return true
	}

	return false
}
