package cometbft

import (
	"context"
	"errors"
	"fmt"
	"strings"

	abci "github.com/cometbft/cometbft/abci/types"
	abciproto "github.com/cometbft/cometbft/api/cometbft/abci/v1"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/gogoproto/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	cmtv1beta1 "cosmossdk.io/api/cosmos/base/tendermint/v1beta1"
	"cosmossdk.io/core/server"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/log"
	storeserver "cosmossdk.io/server/v2/store"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/cmtservice"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
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
		txtypes.RegisterServiceServer(srv, txServer[T]{clientCtx, txCodec, app, consensus})
		nodeservice.RegisterServiceServer(srv, nodeServer[T]{cfg, cometBFTAppConfig, consensus})

		return nil
	}
}

type txServer[T transaction.Tx] struct {
	clientCtx client.Context
	txCodec   transaction.Codec[T]
	app       appSimulator[T]
	consensus abci.Application
}

// BroadcastTx implements tx.ServiceServer.
func (t txServer[T]) BroadcastTx(ctx context.Context, req *txtypes.BroadcastTxRequest) (*txtypes.BroadcastTxResponse, error) {
	return client.TxServiceBroadcast(ctx, t.clientCtx, req)
}

// GetBlockWithTxs implements tx.ServiceServer.
func (t txServer[T]) GetBlockWithTxs(ctx context.Context, req *txtypes.GetBlockWithTxsRequest) (*txtypes.GetBlockWithTxsResponse, error) {
	logger := log.NewNopLogger()
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	resp, err := t.consensus.Info(ctx, &abci.InfoRequest{})
	if err != nil {
		return nil, err
	}
	currentHeight := resp.LastBlockHeight

	if req.Height < 1 || req.Height > currentHeight {
		return nil, sdkerrors.ErrInvalidHeight.Wrapf("requested height %d but height must not be less than 1 "+
			"or greater than the current height %d", req.Height, currentHeight)
	}

	node, err := t.clientCtx.GetNode()
	if err != nil {
		return nil, err
	}

	blockID, block, err := cmtservice.GetProtoBlock(ctx, node, &req.Height)
	if err != nil {
		return nil, err
	}

	var offset, limit uint64
	if req.Pagination != nil {
		offset = req.Pagination.Offset
		limit = req.Pagination.Limit
	} else {
		offset = 0
		limit = query.DefaultLimit
	}

	blockTxs := block.Data.Txs
	blockTxsLn := uint64(len(blockTxs))
	txs := make([]*txtypes.Tx, 0, limit)
	if offset >= blockTxsLn && blockTxsLn != 0 {
		return nil, sdkerrors.ErrInvalidRequest.Wrapf("out of range: cannot paginate %d txs with offset %d and limit %d", blockTxsLn, offset, limit)
	}
	decodeTxAt := func(i uint64) error {
		tx := blockTxs[i]
		txb, err := t.txCodec.Decode(tx)
		if err != nil {
			return err
		}

		// txServer works only with sdk.Tx
		p, err := any(txb).(interface{ AsTx() (*txtypes.Tx, error) }).AsTx()
		if err != nil {
			return err
		}
		txs = append(txs, p)
		return nil
	}
	if req.Pagination != nil && req.Pagination.Reverse {
		for i, count := offset, uint64(0); i > 0 && count != limit; i, count = i-1, count+1 {
			if err = decodeTxAt(i); err != nil {
				logger.Error("failed to decode tx", "error", err)
			}
		}
	} else {
		for i, count := offset, uint64(0); i < blockTxsLn && count != limit; i, count = i+1, count+1 {
			if err = decodeTxAt(i); err != nil {
				logger.Error("failed to decode tx", "error", err)
			}
		}
	}

	return &txtypes.GetBlockWithTxsResponse{
		Txs:     txs,
		BlockId: &blockID,
		Block:   block,
		Pagination: &query.PageResponse{
			Total: blockTxsLn,
		},
	}, nil
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
func (t txServer[T]) GetTxsEvent(ctx context.Context, req *txtypes.GetTxsEventRequest) (*txtypes.GetTxsEventResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
	}

	orderBy := parseOrderBy(req.OrderBy)

	result, err := authtx.QueryTxsByEvents(t.clientCtx, int(req.Page), int(req.Limit), req.Query, orderBy)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	txsList := make([]*txtypes.Tx, len(result.Txs))
	for i, tx := range result.Txs {
		protoTx, ok := tx.Tx.GetCachedValue().(*txtypes.Tx)
		if !ok {
			return nil, status.Errorf(codes.Internal, "getting cached value failed expected %T, got %T", txtypes.Tx{}, tx.Tx.GetCachedValue())
		}

		txsList[i] = protoTx
	}

	return &txtypes.GetTxsEventResponse{
		Txs:         txsList,
		TxResponses: result.Txs,
		Total:       result.TotalCount,
	}, nil
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

	event, err := intoABCIEvents(txResult.Events, map[string]struct{}{}, false)
	if err != nil {
		return nil, status.Errorf(codes.Unknown, "failed to convert events: %v", err)
	}

	return &txtypes.SimulateResponse{
		GasInfo: &sdk.GasInfo{
			GasUsed:   txResult.GasUsed,
			GasWanted: txResult.GasWanted,
		},
		Result: &sdk.Result{
			MsgResponses: msgResponses,
			Events:       event,
		},
	}, nil
}

// TxDecode implements tx.ServiceServer.
func (t txServer[T]) TxDecode(ctx context.Context, req *txtypes.TxDecodeRequest) (*txtypes.TxDecodeResponse, error) {
	if req.TxBytes == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx bytes")
	}

	txb, err := t.txCodec.Decode(req.TxBytes)
	if err != nil {
		return nil, err
	}

	// txServer works only with sdk.Tx
	tx, err := any(txb).(interface{ AsTx() (*txtypes.Tx, error) }).AsTx()
	if err != nil {
		return nil, err
	}

	return &txtypes.TxDecodeResponse{
		Tx: tx,
	}, nil
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
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("invalid request %s", err))
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

func parseOrderBy(orderBy txtypes.OrderBy) string {
	switch orderBy {
	case txtypes.OrderBy_ORDER_BY_ASC:
		return "asc"
	case txtypes.OrderBy_ORDER_BY_DESC:
		return "desc"
	default:
		return "" // Defaults to CometBFT's default, which is `asc` now.
	}
}

func (c *consensus[T]) maybeHandleExternalServices(ctx context.Context, req *abci.QueryRequest) (transaction.Msg, error) {
	// Handle comet service
	if strings.HasPrefix(req.Path, "/cosmos.base.tendermint.v1beta1.Service") {
		rpcClient, _ := rpchttp.New(c.cfg.ConfigTomlConfig.RPC.ListenAddress)

		cometQServer := cmtservice.NewQueryServer(rpcClient, c.Query, c.appCodecs.ConsensusAddressCodec)
		paths := strings.Split(req.Path, "/")
		if len(paths) <= 2 {
			return nil, fmt.Errorf("invalid request path: %s", req.Path)
		}

		var resp transaction.Msg
		var err error
		switch paths[2] {
		case "GetNodeInfo":
			resp, err = handleExternalService(ctx, req, cometQServer.GetNodeInfo)
		case "GetSyncing":
			resp, err = handleExternalService(ctx, req, cometQServer.GetSyncing)
		case "GetLatestBlock":
			resp, err = handleExternalService(ctx, req, cometQServer.GetLatestBlock)
		case "GetBlockByHeight":
			resp, err = handleExternalService(ctx, req, cometQServer.GetBlockByHeight)
		case "GetLatestValidatorSet":
			resp, err = handleExternalService(ctx, req, cometQServer.GetLatestValidatorSet)
		case "GetValidatorSetByHeight":
			resp, err = handleExternalService(ctx, req, cometQServer.GetValidatorSetByHeight)
		case "ABCIQuery":
			resp, err = handleExternalService(ctx, req, cometQServer.ABCIQuery)
		}

		return resp, err
	}

	// Handle node service
	if strings.HasPrefix(req.Path, "/cosmos.base.node.v1beta1.Service") {
		nodeQService := nodeServer[T]{c.cfgMap, c.cfg.AppTomlConfig, c}
		paths := strings.Split(req.Path, "/")
		if len(paths) <= 2 {
			return nil, fmt.Errorf("invalid request path: %s", req.Path)
		}

		var resp transaction.Msg
		var err error
		switch paths[2] {
		case "Config":
			resp, err = handleExternalService(ctx, req, nodeQService.Config)
		case "Status":
			resp, err = handleExternalService(ctx, req, nodeQService.Status)
		}

		return resp, err
	}

	// Handle tx service
	if strings.HasPrefix(req.Path, "/cosmos.tx.v1beta1.Service") {
		rpcClient, _ := client.NewClientFromNode(c.cfg.AppTomlConfig.Address)

		txConfig := authtx.NewTxConfig(
			c.appCodecs.AppCodec,
			c.appCodecs.AppCodec.InterfaceRegistry().SigningContext().AddressCodec(),
			c.appCodecs.AppCodec.InterfaceRegistry().SigningContext().ValidatorAddressCodec(),
			authtx.DefaultSignModes,
		)

		// init simple client context
		clientCtx := client.Context{}.
			WithLegacyAmino(c.appCodecs.LegacyAmino.(*codec.LegacyAmino)).
			WithCodec(c.appCodecs.AppCodec).
			WithNodeURI(c.cfg.AppTomlConfig.Address).
			WithClient(rpcClient).
			WithTxConfig(txConfig)

		txService := txServer[T]{
			clientCtx: clientCtx,
			txCodec:   c.appCodecs.TxCodec,
			app:       c.app,
			consensus: c,
		}
		paths := strings.Split(req.Path, "/")
		if len(paths) <= 2 {
			return nil, fmt.Errorf("invalid request path: %s", req.Path)
		}

		var resp transaction.Msg
		var err error
		switch paths[2] {
		case "Simulate":
			resp, err = handleExternalService(ctx, req, txService.Simulate)
		case "GetTx":
			resp, err = handleExternalService(ctx, req, txService.GetTx)
		case "BroadcastTx":
			return nil, errors.New("can't route a broadcast tx message")
		case "GetTxsEvent":
			resp, err = handleExternalService(ctx, req, txService.GetTxsEvent)
		case "GetBlockWithTxs":
			resp, err = handleExternalService(ctx, req, txService.GetBlockWithTxs)
		case "TxDecode":
			resp, err = handleExternalService(ctx, req, txService.TxDecode)
		case "TxEncode":
			resp, err = handleExternalService(ctx, req, txService.TxEncode)
		case "TxEncodeAmino":
			resp, err = handleExternalService(ctx, req, txService.TxEncodeAmino)
		case "TxDecodeAmino":
			resp, err = handleExternalService(ctx, req, txService.TxDecodeAmino)
		}

		return resp, err
	}

	return nil, nil
}

func handleExternalService[T any, PT interface {
	*T
	proto.Message
},
	U any, UT interface {
		*U
		proto.Message
	}](
	ctx context.Context,
	rawReq *abciproto.QueryRequest,
	handler func(ctx context.Context, msg PT) (UT, error),
) (transaction.Msg, error) {
	req := PT(new(T))
	err := proto.Unmarshal(rawReq.Data, req)
	if err != nil {
		return nil, err
	}
	typedResp, err := handler(ctx, req)
	if err != nil {
		return nil, err
	}
	return typedResp, nil
}
