package tx

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BaseAppSimulateFn is the signature of the Baseapp#Simulate function.
type BaseAppSimulateFn func(txBytes []byte) (sdk.GasInfo, *sdk.Result, error)

type txServer struct {
	simulate          BaseAppSimulateFn
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewTxServer creates a new TxService server.
func NewTxServer(simulate BaseAppSimulateFn, interfaceRegistry codectypes.InterfaceRegistry) ServiceServer {
	return txServer{
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

// TxByHash implements the ServiceServer.TxByHash RPC method.
func (s txServer) TxByHash(ctx context.Context, req *TxByHashRequest) (*TxByHashResponse, error) {
	// TODO
	return nil, nil
}
