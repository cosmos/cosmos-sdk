package tx

import (
	"context"
	"encoding/hex"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// BaseAppSimulateFn is the signature of the Baseapp#Simulate function.
type BaseAppSimulateFn func(txBytes []byte) (sdk.GasInfo, *sdk.Result, error)

type txServer struct {
	clientCtx         client.Context
	simulate          BaseAppSimulateFn
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewTxServer creates a new TxService server.
func NewTxServer(clientCtx client.Context, simulate BaseAppSimulateFn, interfaceRegistry codectypes.InterfaceRegistry) ServiceServer {
	return txServer{
		clientCtx:         clientCtx,
		simulate:          simulate,
		interfaceRegistry: interfaceRegistry,
	}
}

var _ ServiceServer = txServer{}

// Simulate implements the ServiceServer.Simulate RPC method.
func (s txServer) Simulate(ctx context.Context, req *SimulateRequest) (*SimulateResponse, error) {
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

	return &SimulateResponse{
		GasInfo: &gasInfo,
		Result:  result,
	}, nil
}

// GetTx implements the ServiceServer.GetTx RPC method.
func (s txServer) GetTx(ctx context.Context, req *GetTxRequest) (*GetTxResponse, error) {
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

	sdkTx, err := s.clientCtx.TxConfig.TxDecoder()(result.Tx)
	if err != nil {
		return nil, err
	}

	txBuilder, ok := sdkTx.(client.TxBuilder)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (client.TxBuilder)(nil), sdkTx)
	}
	// Convert the txBuilder to a tx.Tx.
	protoTx, err := TxBuilderToProtoTx(txBuilder)
	if err != nil {
		return nil, err
	}

	return &GetTxResponse{
		Tx: protoTx,
	}, nil
}

// TxBuilderToProtoTx convert a txBuilder into a proto tx.Tx.
func TxBuilderToProtoTx(txBuilder client.TxBuilder) (*Tx, error) { // nolint
	intoAnyTx, ok := txBuilder.(codectypes.IntoAny)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (codectypes.IntoAny)(nil), intoAnyTx)
	}

	any := intoAnyTx.AsAny().GetCachedValue()
	if any == nil {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "any's cached value is empty")
	}

	protoTx, ok := any.(*Tx)
	if !ok {
		return nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidType, "expected %T, got %T", (codectypes.IntoAny)(nil), intoAnyTx)
	}

	return protoTx, nil
}

// RegisterGRPCGatewayRoutes mounts the tx service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	RegisterServiceHandlerClient(context.Background(), mux, NewServiceClient(clientCtx))
}
