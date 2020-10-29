package tx

import (
	"context"
	"encoding/hex"
	fmt "fmt"

	gogogrpc "github.com/gogo/protobuf/grpc"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

	// Create a proto codec, we need it to unmarshal the tx bytes.
	cdc := codec.NewProtoCodec(s.clientCtx.InterfaceRegistry)
	var protoTx Tx
	err = cdc.UnmarshalBinaryBare(result.Tx, &protoTx)
	if err != nil {
		return nil, err
	}
	fmt.Println(protoTx)

	return &GetTxResponse{
		Tx: &protoTx,
	}, nil
}

// RegisterGRPCGatewayRoutes mounts the tx service's GRPC-gateway routes on the
// given Mux.
func RegisterGRPCGatewayRoutes(clientConn gogogrpc.ClientConn, mux *runtime.ServeMux) {
	RegisterServiceHandlerClient(context.Background(), mux, NewServiceClient(clientConn))
}
