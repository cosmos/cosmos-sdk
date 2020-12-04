package tx

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	pagination "github.com/cosmos/cosmos-sdk/types/query"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
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
func (s txServer) GetTxsEvent(ctx context.Context, req *txtypes.GetTxsEventRequest) (*txtypes.GetTxsEventResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	page, limit, err := pagination.ParsePagination(req.Pagination)
	if err != nil {
		return nil, err
	}

	if len(req.Events) == 0 {
		return nil, status.Error(codes.InvalidArgument, "must declare at least one event to search")
	}

	tmEvents := make([]string, len(req.Events))
	for i, event := range req.Events {
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
	txRespList := make([]*sdk.TxResponse, len(result.Txs))
	txsList := make([]*txtypes.Tx, len(result.Txs))

	for i, tx := range result.Txs {
		txResp := txResultToTxResponse(&tx.TxResult)
		txResp.Height = tx.Height
		txResp.TxHash = tx.Hash.String()
		txRespList[i] = txResp

		var protoTx txtypes.Tx
		if err := cdc.UnmarshalBinaryBare(tx.Tx, &protoTx); err != nil {
			return nil, err
		}
		txsList[i] = &protoTx
	}

	return &txtypes.GetTxsEventResponse{
		Txs:         txsList,
		TxResponses: txRespList,
		Pagination: &pagination.PageResponse{
			Total: uint64(result.TotalCount),
		},
	}, nil

}

// Simulate implements the ServiceServer.Simulate RPC method.
func (s txServer) Simulate(ctx context.Context, req *txtypes.SimulateRequest) (*txtypes.SimulateResponse, error) {
	if req == nil || req.Tx == nil {
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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

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

	var protoTx txtypes.Tx
	if err := cdc.UnmarshalBinaryBare(result.Tx, &protoTx); err != nil {
		return nil, err
	}

	txResp := txResultToTxResponse(&result.TxResult)
	txResp.Height = result.Height
	txResp.TxHash = result.Hash.String()

	return &txtypes.GetTxResponse{
		Tx:         &protoTx,
		TxResponse: txResp,
	}, nil
}

func (s txServer) BroadcastTx(ctx context.Context, req *txtypes.BroadcastTxRequest) (*txtypes.BroadcastTxResponse, error) {
	return client.TxServiceBroadcast(ctx, s.clientCtx, req)
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

func txResultToTxResponse(respTx *abci.ResponseDeliverTx) *sdk.TxResponse {
	logs, _ := sdk.ParseABCILogs(respTx.Log)
	return &sdk.TxResponse{
		Code:      respTx.Code,
		Codespace: respTx.Codespace,
		GasUsed:   respTx.GasUsed,
		GasWanted: respTx.GasWanted,
		Info:      respTx.Info,
		Logs:      logs,
	}
}
