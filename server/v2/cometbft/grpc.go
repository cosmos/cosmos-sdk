package cometbft

import (
	"context"
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	abciproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cmtv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	storeserver "cosmossdk.io/server/v2/store"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	txtypes "github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
)

type appSimulator[T transaction.Tx] interface {
	Simulate(ctx context.Context, tx T) (server.TxResult, corestore.WriterMap, error)
}

// gRPCServiceRegistrar returns a function that registers the CometBFT gRPC service
// Those services are defined for backward compatibility.
// Eventually, they will be removed in favor of the new gRPC services.
func gRPCServiceRegistrar[T transaction.Tx](
	clientCtx client.Context,
	cfg server.ConfigMap,
	cometBFTAppConfig *AppTomlConfig,
	txCodec transaction.Codec[T],
	consensus abci.Application,
	app appSimulator[T],
) func(srv *grpc.Server) error {
	return func(srv *grpc.Server) error {
		cmtservice.RegisterServiceServer(srv, cmtservice.NewQueryServer(clientCtx.Client, consensus.Query, clientCtx.ConsensusAddressCodec))
		txtypes.RegisterServiceServer(srv, txServer[T]{clientCtx, txCodec, app})
		nodeservice.RegisterServiceServer(srv, nodeServer[T]{cfg, cometBFTAppConfig, consensus})

		return nil
	}
}

type txServer[T transaction.Tx] struct {
	clientCtx client.Context
	txCodec   transaction.Codec[T]
	app       appSimulator[T]
}

// BroadcastTx implements tx.ServiceServer.
func (t txServer[T]) BroadcastTx(ctx context.Context, req *txtypes.BroadcastTxRequest) (*txtypes.BroadcastTxResponse, error) {
	return client.TxServiceBroadcast(ctx, t.clientCtx, req)
}

// GetBlockWithTxs implements tx.ServiceServer.
func (t txServer[T]) GetBlockWithTxs(context.Context, *txtypes.GetBlockWithTxsRequest) (*txtypes.GetBlockWithTxsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// GetTx implements tx.ServiceServer.
func (t txServer[T]) GetTx(ctx context.Context, req *txtypes.GetTxRequest) (*txtypes.GetTxResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	if len(req.Hash) == 0 {
		return nil, status.Error(codes.InvalidArgument, "tx hash cannot be empty")
	}

	result, err := authtx.QueryTx(t.clientCtx, req.Hash)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return nil, status.Errorf(codes.NotFound, "tx not found: %s", req.Hash)
		}

		return nil, err
	}

	protoTx, ok := result.Tx.GetCachedValue().(*txtypes.Tx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "expected %T, got %T", txtypes.Tx{}, result.Tx.GetCachedValue())
	}

	return &txtypes.GetTxResponse{
		Tx:         protoTx,
		TxResponse: result,
	}, nil
}

// GetTxsEvent implements tx.ServiceServer.
func (t txServer[T]) GetTxsEvent(context.Context, *txtypes.GetTxsEventRequest) (*txtypes.GetTxsEventResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// Simulate implements tx.ServiceServer.
func (t txServer[T]) Simulate(ctx context.Context, req *txtypes.SimulateRequest) (*txtypes.SimulateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	txBytes := req.TxBytes
	if txBytes == nil && req.Tx != nil {
		// This block is for backwards-compatibility.
		// We used to support passing a `Tx` in req. But if we do that, sig
		// verification might not pass, because the .Marshal() below might not
		// be the same marshaling done by the client.
		var err error
		txBytes, err = proto.Marshal(req.Tx)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid tx; %v", err)
		}
	}

	if txBytes == nil {
		return nil, status.Errorf(codes.InvalidArgument, "empty txBytes is not allowed")
	}

	tx, err := t.txCodec.Decode(txBytes)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to decode tx: %v", err)
	}

	txResult, _, err := t.app.Simulate(ctx, tx)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "%v with gas used: '%d'", err, txResult.GasUsed)
	}

	msgResponses := make([]*codectypes.Any, 0, len(txResult.Resp))
	// pack the messages into Any
	for _, msg := range txResult.Resp {
		anyMsg, err := codectypes.NewAnyWithValue(msg)
		if err != nil {
			return nil, status.Errorf(codes.Unknown, "failed to pack message response: %v", err)
		}
		msgResponses = append(msgResponses, anyMsg)
	}

	return &txtypes.SimulateResponse{
		GasInfo: &sdk.GasInfo{
			GasUsed:   txResult.GasUsed,
			GasWanted: txResult.GasWanted,
		},
		Result: &sdk.Result{
			MsgResponses: msgResponses,
		},
	}, nil
}

// TxDecode implements tx.ServiceServer.
func (t txServer[T]) TxDecode(context.Context, *txtypes.TxDecodeRequest) (*txtypes.TxDecodeResponse, error) {
	return nil, status.Error(codes.Unimplemented, "not implemented")
}

// TxDecodeAmino implements tx.ServiceServer.
func (t txServer[T]) TxDecodeAmino(_ context.Context, req *txtypes.TxDecodeAminoRequest) (*txtypes.TxDecodeAminoResponse, error) {
	if req.AminoBinary == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx bytes")
	}

	var stdTx legacytx.StdTx
	err := t.clientCtx.LegacyAmino.Unmarshal(req.AminoBinary, &stdTx)
	if err != nil {
		return nil, err
	}

	res, err := t.clientCtx.LegacyAmino.MarshalJSON(stdTx)
	if err != nil {
		return nil, err
	}

	return &txtypes.TxDecodeAminoResponse{
		AminoJson: string(res),
	}, nil
}

// TxEncode implements tx.ServiceServer.
func (t txServer[T]) TxEncode(_ context.Context, req *txtypes.TxEncodeRequest) (*txtypes.TxEncodeResponse, error) {
	if req.Tx == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	bodyBytes, err := t.clientCtx.Codec.Marshal(req.Tx.Body)
	if err != nil {
		return nil, err
	}

	authInfoBytes, err := t.clientCtx.Codec.Marshal(req.Tx.AuthInfo)
	if err != nil {
		return nil, err
	}

	raw := &txtypes.TxRaw{
		BodyBytes:     bodyBytes,
		AuthInfoBytes: authInfoBytes,
		Signatures:    req.Tx.Signatures,
	}

	encodedBytes, err := t.clientCtx.Codec.Marshal(raw)
	if err != nil {
		return nil, err
	}

	return &txtypes.TxEncodeResponse{
		TxBytes: encodedBytes,
	}, nil
}

// TxEncodeAmino implements tx.ServiceServer.
func (t txServer[T]) TxEncodeAmino(_ context.Context, req *txtypes.TxEncodeAminoRequest) (*txtypes.TxEncodeAminoResponse, error) {
	if req.AminoJson == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx json")
	}

	var stdTx legacytx.StdTx
	err := t.clientCtx.LegacyAmino.UnmarshalJSON([]byte(req.AminoJson), &stdTx)
	if err != nil {
		return nil, err
	}

	encodedBytes, err := t.clientCtx.LegacyAmino.Marshal(stdTx)
	if err != nil {
		return nil, err
	}

	return &txtypes.TxEncodeAminoResponse{
		AminoBinary: encodedBytes,
	}, nil
}

var _ txtypes.ServiceServer = txServer[transaction.Tx]{}

type nodeServer[T transaction.Tx] struct {
	cfg               server.ConfigMap
	cometBFTAppConfig *AppTomlConfig
	consensus         abci.Application
}

func (s nodeServer[T]) Config(ctx context.Context, _ *nodeservice.ConfigRequest) (*nodeservice.ConfigResponse, error) {
	minGasPricesStr := ""
	minGasPrices, ok := s.cfg["server"].(map[string]interface{})["minimum-gas-prices"]
	if ok {
		minGasPricesStr = minGasPrices.(string)
	}

	storeCfg, err := storeserver.UnmarshalConfig(s.cfg)
	if err != nil {
		return nil, err
	}

	return &nodeservice.ConfigResponse{
		MinimumGasPrice:   minGasPricesStr,
		PruningKeepRecent: fmt.Sprintf("%d", storeCfg.Options.SCPruningOption.KeepRecent),
		PruningInterval:   fmt.Sprintf("%d", storeCfg.Options.SCPruningOption.Interval),
		HaltHeight:        s.cometBFTAppConfig.HaltHeight,
	}, nil
}

func (s nodeServer[T]) Status(ctx context.Context, _ *nodeservice.StatusRequest) (*nodeservice.StatusResponse, error) {
	nodeInfo, err := s.consensus.Info(ctx, &abciproto.InfoRequest{})
	if err != nil {
		return nil, err
	}

	return &nodeservice.StatusResponse{
		Height:        uint64(nodeInfo.LastBlockHeight),
		Timestamp:     nil,
		AppHash:       nil,
		ValidatorHash: nodeInfo.LastBlockAppHash,
	}, nil
}

// CometBFTAutoCLIDescriptor is the auto-generated CLI descriptor for the CometBFT service
var CometBFTAutoCLIDescriptor = &autocliv1.ServiceCommandDescriptor{
	Service: cmtv1beta1.Service_ServiceDesc.ServiceName,
	RpcCommandOptions: []*autocliv1.RpcCommandOptions{
		{
			RpcMethod: "GetNodeInfo",
			Use:       "node-info",
			Short:     "Query the current node info",
		},
		{
			RpcMethod: "GetSyncing",
			Use:       "syncing",
			Short:     "Query node syncing status",
		},
		{
			RpcMethod: "GetLatestBlock",
			Use:       "block-latest",
			Short:     "Query for the latest committed block",
		},
		{
			RpcMethod:      "GetBlockByHeight",
			Use:            "block-by-height <height>",
			Short:          "Query for a committed block by height",
			Long:           "Query for a specific committed block using the CometBFT RPC `block_by_height` method",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "height"}},
		},
		{
			RpcMethod: "GetLatestValidatorSet",
			Use:       "validator-set",
			Alias:     []string{"validator-set-latest", "comet-validator-set", "cometbft-validator-set", "tendermint-validator-set"},
			Short:     "Query for the latest validator set",
		},
		{
			RpcMethod:      "GetValidatorSetByHeight",
			Use:            "validator-set-by-height <height>",
			Short:          "Query for a validator set by height",
			PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "height"}},
		},
		{
			RpcMethod: "ABCIQuery",
			Skip:      true,
		},
	},
}
