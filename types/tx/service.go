package tx

import (
	"context"
	fmt "fmt"

	rpcclient "github.com/tendermint/tendermint/rpc/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BaseAppSimulateFn is the signature of the Baseapp#Simulate function.
type BaseAppSimulateFn func(txBytes []byte) (sdk.GasInfo, *sdk.Result, error)

type txServer struct {
	rpcClient         rpcclient.Client
	simulate          BaseAppSimulateFn
	interfaceRegistry codectypes.InterfaceRegistry
}

// NewTxServer creates a new TxService server.
func NewTxServer(rpcClient rpcclient.Client, simulate BaseAppSimulateFn, interfaceRegistry codectypes.InterfaceRegistry) ServiceServer {
	return txServer{
		rpcClient:         rpcClient,
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
	// TODO We should also check the proof flag in gRPC header.
	// https://github.com/cosmos/cosmos-sdk/issues/7036.
	result, err := s.rpcClient.Tx(ctx, req.Hash, false)
	if err != nil {
		return nil, err
	}

	fmt.Println(result)
	return nil, nil
}
