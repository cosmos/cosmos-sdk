package simulate

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type simulateServer struct {
	app         baseapp.BaseApp
	pubkeyCodec cryptotypes.PublicKeyCodec
	txConfig    client.TxConfig
}

// NewSimulateServer creates a new SimulateServer.
func NewSimulateServer(app baseapp.BaseApp, pubkeyCodec cryptotypes.PublicKeyCodec, txConfig client.TxConfig) SimulateServiceServer {
	return simulateServer{
		app:         app,
		pubkeyCodec: pubkeyCodec,
		txConfig:    txConfig,
	}
}

var _ SimulateServiceServer = simulateServer{}

// Simulate implements the SimulateService.Simulate RPC method.
func (s simulateServer) Simulate(ctx context.Context, req *SimulateRequest) (*SimulateResponse, error) {
	if req.Tx == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	err := req.Tx.UnpackInterfaces(s.app.GRPCQueryRouter().InterfaceRegistry())
	if err != nil {
		return nil, err
	}
	txBuilder := authtx.WrapTxBuilder(req.Tx, s.pubkeyCodec)

	txBytes, err := req.Tx.Marshal()
	if err != nil {
		return nil, err
	}

	gasInfo, result, err := s.app.Simulate(txBytes, txBuilder.GetTx())
	if err != nil {
		return nil, err
	}

	return &SimulateResponse{
		GasInfo: &gasInfo,
		Result:  result,
	}, nil
}
