package simulate

import (
	"context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type simulateServer struct {
	app baseapp.BaseApp
}

// NewSimulateServer creates a new SimulateServer.
func NewSimulateServer(app baseapp.BaseApp) simulateServer {
	return simulateServer{
		app: app,
	}
}

var _ SimulateServiceServer = simulateServer{}

// Simulate implements the SimulateService.Simulate RPC method.
func (s simulateServer) Simulate(ctx context.Context, req *SimulateRequest) (*SimulateResponse, error) {
	if req.Tx == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	req.Tx.UnpackInterfaces(s.app.GRPCQueryRouter().AnyUnpacker())

	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, err
	}

	gasInfo, result, err := s.app.Simulate(txBytes, req.Tx)
	if err != nil {
		return nil, err
	}

	return &SimulateResponse{
		GasInfo: &gasInfo,
		Result:  result,
	}, nil
}
