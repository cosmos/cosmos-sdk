package tx

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	pagination "github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// baseAppSimulateFn is the signature of the Baseapp#Simulate function.
type baseAppSimulateFn func(txBytes []byte) (sdk.GasInfo, *sdk.Result, error)

// txServer is the server for the protobuf Tx service.
type txServer struct {
	clientCtx         client.Context
	simulate          baseAppSimulateFn
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewTxServer creates a new Tx service server.
func NewTxServer(clientCtx client.Context, simulate baseAppSimulateFn, interfaceRegistry codectypes.InterfaceRegistry) txtypes.ServiceServer {
	return txServer{
		clientCtx:         clientCtx,
		simulate:          simulate,
		interfaceRegistry: interfaceRegistry,
	}
}

var _ txtypes.ServiceServer = txServer{}

const (
	eventFormat = "{eventType}.{eventAttribute}={value}"
)

// TxsByEvents implements the ServiceServer.TxsByEvents RPC method.
func (s txServer) TxsByEvents(ctx context.Context, req *txtypes.GetTxsEventRequest) (*txtypes.TxsByEventsResponse, error) {

	offset := int(req.Pagination.Offset)
	limit := int(req.Pagination.Limit)
	page := offset/limit + 1

	if page < 0 {
		return nil, status.Error(codes.InvalidArgument, "offset must greater than 0")
	}
	if limit < 0 {
		return nil, status.Error(codes.InvalidArgument, "limit must greater than 0")
	}
	if len(req.Event) == 0 {
		return nil, status.Error(codes.InvalidArgument, "must declare at least one event to search")
	}
	var events []string
	if strings.Contains(req.Event, "&") {
		events = strings.Split(req.Event, "&")
	} else {
		events = append(events, req.Event)
	}
	tmEvents := make([]string, len(events))

	for i, event := range events {
		if !strings.Contains(event, "=") {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid event; event %s should be of the format: %s", event, eventFormat))
		} else if strings.Count(event, "=") > 1 {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid event; event %s should be of the format: %s", event, eventFormat))
		}

		tokens := strings.Split(event, "=")
		if tokens[0] == tmtypes.TxHeightKey {
			event = fmt.Sprintf("%s=%s", tokens[0], tokens[1])
		} else {
			event = fmt.Sprintf("%s='%s'", tokens[0], tokens[1])
		}

		tmEvents[i] = event
	}

	query := strings.Join(tmEvents, " AND ")

	result, err := s.clientCtx.Client.TxSearch(ctx, query, false, &page, &limit, "")

	if err != nil {
		return nil, err
	}
	// Create a proto codec, we need it to unmarshal the tx bytes.
	cdc := codec.NewProtoCodec(s.clientCtx.InterfaceRegistry)
	res := make([]*txtypes.TxRes, len(result.Txs))

	for i, tx := range result.Txs {
		result, err := txResponseToTxRes(cdc, tx)
		if err != nil {
			return nil, err
		}
		res[i] = result
	}

	return &txtypes.TxsByEventsResponse{
		Txs: res,
		Pagination: &pagination.PageResponse{
			Total: uint64(result.TotalCount),
		},
	}, nil

}

// Simulate implements the ServiceServer.Simulate RPC method.
func (s txServer) Simulate(ctx context.Context, req *txtypes.SimulateRequest) (*txtypes.SimulateResponse, error) {
	if req.Tx == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	err := req.Tx.UnpackInterfaces(s.interfaceRegistry)
	if err != nil {
		return nil, err
	}
	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, err
	}

	gasInfo, result, err := s.simulate(txBytes)
	if err != nil {
		return nil, err
	}

	return &txtypes.SimulateResponse{
		GasInfo: &gasInfo,
		Result:  result,
	}, nil
}

// GetTx implements the ServiceServer.GetTx RPC method.
func (s txServer) GetTx(ctx context.Context, req *txtypes.GetTxRequest) (*txtypes.GetTxResponse, error) {
	// We get hash as a hex string in the request, convert it to bytes.
	hash, err := hex.DecodeString(req.Hash)
	if err != nil {
		return nil, err
	}

	// TODO We should also check the proof flag in gRPC header.
	// https://github.com/cosmos/cosmos-sdk/issues/7036.
	result, err := s.clientCtx.Client.Tx(ctx, hash, false)
	if err != nil {
		return nil, err
	}

	// Create a proto codec, we need it to unmarshal the tx bytes.
	cdc := codec.NewProtoCodec(s.clientCtx.InterfaceRegistry)

	resp, err := txResponseToTxRes(cdc, result)
	if err != nil {
		return nil, err
	}

	return &txtypes.GetTxResponse{
		Result: resp,
	}, nil
}

// RegisterTxService registers the tx service on the gRPC router.
func RegisterTxService(
	qrt gogogrpc.Server,
	clientCtx client.Context,
	simulateFn baseAppSimulateFn,
	interfaceRegistry codectypes.InterfaceRegistry,
) {
	txtypes.RegisterServiceServer(
		qrt,
		NewTxServer(clientCtx, simulateFn, interfaceRegistry),
	)
}

// RegisterGRPCGatewayRoutes mounts the tx service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	txtypes.RegisterServiceHandlerClient(context.Background(), mux, txtypes.NewServiceClient(clientConn))
}

func txResponseToTxRes(cdc *codec.ProtoCodec, result *coretypes.ResultTx) (*txtypes.TxRes, error) {
	var protoTx txtypes.Tx

	if err := cdc.UnmarshalBinaryBare(result.Tx, &protoTx); err != nil {
		return nil, err
	}
	proof := result.Proof.ToProto()

	return &txtypes.TxRes{
		Code:      result.TxResult.Code,
		Codespace: result.TxResult.Codespace,
		GasUsed:   result.TxResult.GasUsed,
		GasWanted: result.TxResult.GasWanted,
		Height:    result.Height,
		Info:      result.TxResult.Info,
		RawLog:    result.TxResult.Log,
		TxHash:    result.Hash.String(),
		Tx:        &protoTx,
		Proof:     &proof,
		Index:     result.Index,
	}, nil
}
