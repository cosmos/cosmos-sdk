package simulate

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

// BaseAppSimulateFn is the signature of the Baseapp#Simulate function.
type BaseAppSimulateFn func(txBytes []byte, txtypes sdk.Tx) (sdk.GasInfo, *sdk.Result, error)

type simulateServer struct {
	simulate          BaseAppSimulateFn
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewSimulateServer creates a new SimulateServer.
func NewSimulateServer(simulate BaseAppSimulateFn, interfaceRegistry codectypes.InterfaceRegistry) SimulateServiceServer {
	return simulateServer{
		simulate:          simulate,
		interfaceRegistry: interfaceRegistry,
	}
}

var _ SimulateServiceServer = simulateServer{}

// Simulate implements the SimulateService.Simulate RPC method.
func (s simulateServer) Simulate(ctx context.Context, req *SimulateRequest) (*SimulateResponse, error) {
	if req.Tx == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	err := req.Tx.UnpackInterfaces(s.interfaceRegistry)
	if err != nil {
		return nil, err
	}
	txBuilder := authtx.WrapTx(req.Tx)
	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, err
	}

	gasInfo, result, err := s.simulate(txBytes, txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return &SimulateResponse{
		GasInfo: &gasInfo,
		Result:  result,
	}, nil
}
