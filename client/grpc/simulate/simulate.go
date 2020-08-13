package simulate

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
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
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	fmt.Println("sdkCtx.TxBytes()=", sdkCtx.TxBytes())

	req.Tx.UnpackInterfaces(s.app.GRPCQueryRouter().AnyUnpacker())

	if req.Tx == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, err
	}
	fmt.Println("txBytes=", txBytes)

	gasInfo, result, err := s.app.Simulate(txBytes, req.Tx)
	if err != nil {
		return nil, err
	}

	return &SimulateResponse{
		GasInfo: &gasInfo,
		Result:  result,
	}, nil
}
