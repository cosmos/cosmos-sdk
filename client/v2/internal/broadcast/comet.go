package broadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/mempool"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"

	apiacbci "cosmossdk.io/api/cosmos/base/abci/v1beta1"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// CometRPC defines the interface of a CometBFT RPC client needed for
// queries and transaction handling.
type CometRPC interface {
	rpcclient.ABCIClient

	Validators(ctx context.Context, height *int64, page, perPage *int) (*coretypes.ResultValidators, error)
	Status(context.Context) (*coretypes.ResultStatus, error)
	Block(ctx context.Context, height *int64) (*coretypes.ResultBlock, error)
	BlockByHash(ctx context.Context, hash []byte) (*coretypes.ResultBlock, error)
	BlockResults(ctx context.Context, height *int64) (*coretypes.ResultBlockResults, error)
	BlockchainInfo(ctx context.Context, minHeight, maxHeight int64) (*coretypes.ResultBlockchainInfo, error)
	Commit(ctx context.Context, height *int64) (*coretypes.ResultCommit, error)
	Tx(ctx context.Context, hash []byte, prove bool) (*coretypes.ResultTx, error)
	TxSearch(
		ctx context.Context,
		query string,
		prove bool,
		page, perPage *int,
		orderBy string,
	) (*coretypes.ResultTxSearch, error)
	BlockSearch(
		ctx context.Context,
		query string,
		page, perPage *int,
		orderBy string,
	) (*coretypes.ResultBlockSearch, error)
}

var _ Broadcaster = CometBFTBroadcaster{}

// CometBFTBroadcaster implements the Broadcaster interface for CometBFT consensus engine.
type CometBFTBroadcaster struct {
	rpcClient CometRPC
	mode      string
	cdc       codec.JSONCodec
}

func withMode(mode string) func(broadcaster Broadcaster) {
	return func(b Broadcaster) {
		cbc, ok := b.(*CometBFTBroadcaster)
		if !ok {
			return
		}
		cbc.mode = mode
	}
}

func withJsonCodec(codec codec.JSONCodec) func(broadcaster Broadcaster) {
	return func(b Broadcaster) {
		cbc, ok := b.(*CometBFTBroadcaster)
		if !ok {
			return
		}
		cbc.cdc = codec
	}
}

// NewCometBFTBroadcaster creates a new CometBftBroadcaster.
func NewCometBFTBroadcaster(rpcURL string, opts ...Option) (*CometBFTBroadcaster, error) {
	rpcClient, err := rpchttp.New(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create CometBft RPC client: %w", err)
	}

	bc := &CometBFTBroadcaster{}
	for _, opt := range opts {
		opt(bc)
	}

	bc.rpcClient = *rpcClient
	return bc, nil
}

func (c CometBFTBroadcaster) Consensus() string {
	return cometBFTConsensus
}

// Broadcast sends a transaction to the network and returns the result.
// returns a byte slice containing the JSON-encoded result and an error if the broadcast failed.
func (c CometBFTBroadcaster) Broadcast(ctx context.Context, txBytes []byte) ([]byte, error) {
	var fn func(ctx context.Context, tx cmttypes.Tx) (*coretypes.ResultBroadcastTx, error)
	switch c.mode {
	case BroadcastSync:
		fn = c.rpcClient.BroadcastTxSync
	case BroadcastAsync:
		fn = c.rpcClient.BroadcastTxAsync
	default:
		return []byte{}, fmt.Errorf("unknown broadcast mode: %s", c.mode)
	}

	res, err := c.broadcast(ctx, txBytes, fn)
	if err != nil {
		return []byte{}, err
	}

	return c.cdc.MarshalJSON(res)
}

// broadcast sends a transaction to the CometBFT network using the provided function.
func (c CometBFTBroadcaster) broadcast(ctx context.Context, txbytes []byte,
	fn func(ctx context.Context, tx cmttypes.Tx) (*coretypes.ResultBroadcastTx, error),
) (*apiacbci.TxResponse, error) {
	bResult, err := fn(ctx, txbytes)
	if errRes := checkCometError(err, txbytes); err != nil {
		return errRes, nil
	}

	return newResponseFormatBroadcastTx(bResult), err
}

// checkCometError checks for errors returned by the CometBFT network and returns an appropriate TxResponse.
// It extracts error information and constructs a TxResponse with the error details.
func checkCometError(err error, tx cmttypes.Tx) *apiacbci.TxResponse {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	txHash := fmt.Sprintf("%X", tx.Hash())

	switch {
	case strings.Contains(errStr, strings.ToLower(mempool.ErrTxInCache.Error())):
		return &apiacbci.TxResponse{
			Code:      sdkerrors.ErrTxInMempoolCache.ABCICode(),
			Codespace: sdkerrors.ErrTxInMempoolCache.Codespace(),
			Txhash:    txHash,
		}

	case strings.Contains(errStr, "mempool is full"):
		return &apiacbci.TxResponse{
			Code:      sdkerrors.ErrMempoolIsFull.ABCICode(),
			Codespace: sdkerrors.ErrMempoolIsFull.Codespace(),
			Txhash:    txHash,
		}

	case strings.Contains(errStr, "tx too large"):
		return &apiacbci.TxResponse{
			Code:      sdkerrors.ErrTxTooLarge.ABCICode(),
			Codespace: sdkerrors.ErrTxTooLarge.Codespace(),
			Txhash:    txHash,
		}

	default:
		return nil
	}
}

// newResponseFormatBroadcastTx returns a TxResponse given a ResultBroadcastTx from cometbft
func newResponseFormatBroadcastTx(res *coretypes.ResultBroadcastTx) *apiacbci.TxResponse {
	if res == nil {
		return nil
	}

	parsedLogs, _ := parseABCILogs(res.Log)

	return &apiacbci.TxResponse{
		Code:      res.Code,
		Codespace: res.Codespace,
		Data:      res.Data.String(),
		RawLog:    res.Log,
		Logs:      parsedLogs,
		Txhash:    res.Hash.String(),
	}
}

// parseABCILogs attempts to parse a stringified ABCI tx log into a slice of
// ABCIMessageLog types. It returns an error upon JSON decoding failure.
func parseABCILogs(logs string) (res []*apiacbci.ABCIMessageLog, err error) {
	err = json.Unmarshal([]byte(logs), &res)
	return res, err
}
