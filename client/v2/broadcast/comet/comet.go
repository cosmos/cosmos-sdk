package comet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/mempool"
	rpcclient "github.com/cometbft/cometbft/rpc/client"
	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"

	apiacbci "cosmossdk.io/api/cosmos/base/abci/v1beta1"
	"cosmossdk.io/client/v2/broadcast"

	"github.com/cosmos/cosmos-sdk/codec"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const (
	// BroadcastSync defines a tx broadcasting mode where the client waits for
	// a CheckTx execution response only.
	BroadcastSync = "sync"
	// BroadcastAsync defines a tx broadcasting mode where the client returns
	// immediately.
	BroadcastAsync = "async"

	// cometBftConsensus is the identifier for the CometBFT consensus engine.
	cometBFTConsensus = "comet"
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

var _ broadcast.Broadcaster = &CometBFTBroadcaster{}

// CometBFTBroadcaster implements the Broadcaster interface for CometBFT consensus engine.
type CometBFTBroadcaster struct {
	rpcClient CometRPC
	mode      string
	cdc       codec.Codec
}

// NewCometBFTBroadcaster creates a new CometBFTBroadcaster.
func NewCometBFTBroadcaster(rpcURL, mode string, cdc codec.Codec) (*CometBFTBroadcaster, error) {
	if cdc == nil {
		return nil, errors.New("codec can't be nil")
	}

	if mode == "" {
		mode = BroadcastSync
	}

	rpcClient, err := rpchttp.New(rpcURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create CometBft RPC client: %w", err)
	}

	return &CometBFTBroadcaster{
		rpcClient: rpcClient,
		mode:      mode,
		cdc:       cdc,
	}, nil
}

// Consensus returns the consensus engine name used by the broadcaster.
// It always returns "comet" for CometBFTBroadcaster.
func (c *CometBFTBroadcaster) Consensus() string {
	return cometBFTConsensus
}

// Broadcast sends a transaction to the network and returns the result.
// returns a byte slice containing the JSON-encoded result and an error if the broadcast failed.
func (c *CometBFTBroadcaster) Broadcast(ctx context.Context, txBytes []byte) ([]byte, error) {
	if c.cdc == nil {
		return []byte{}, fmt.Errorf("JSON codec is not initialized")
	}

	var broadcastFunc func(ctx context.Context, tx cmttypes.Tx) (*coretypes.ResultBroadcastTx, error)
	switch c.mode {
	case BroadcastSync:
		broadcastFunc = c.rpcClient.BroadcastTxSync
	case BroadcastAsync:
		broadcastFunc = c.rpcClient.BroadcastTxAsync
	default:
		return []byte{}, fmt.Errorf("unknown broadcast mode: %s", c.mode)
	}

	res, err := c.broadcast(ctx, txBytes, broadcastFunc)
	if err != nil {
		return []byte{}, err
	}

	return c.cdc.MarshalJSON(res)
}

// broadcast sends a transaction to the CometBFT network using the provided function.
func (c *CometBFTBroadcaster) broadcast(ctx context.Context, txBytes []byte,
	fn func(ctx context.Context, tx cmttypes.Tx) (*coretypes.ResultBroadcastTx, error),
) (*apiacbci.TxResponse, error) {
	res, err := fn(ctx, txBytes)
	if errRes := checkCometError(err, txBytes); errRes != nil {
		return errRes, nil
	}

	if res == nil {
		return nil, err
	}

	parsedLogs, _ := parseABCILogs(res.Log)
	return &apiacbci.TxResponse{
		Code:      res.Code,
		Codespace: res.Codespace,
		Data:      res.Data.String(),
		RawLog:    res.Log,
		Logs:      parsedLogs,
		Txhash:    res.Hash.String(),
	}, err
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
	case strings.Contains(errStr, strings.ToLower(mempool.ErrTxInCache.Error())) ||
		strings.Contains(errStr, strings.ToLower(sdkerrors.ErrTxInMempoolCache.Error())):
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

	case strings.Contains(errStr, "no signatures supplied"):
		return &apiacbci.TxResponse{
			Code:      sdkerrors.ErrNoSignatures.ABCICode(),
			Codespace: sdkerrors.ErrNoSignatures.Codespace(),
			Txhash:    txHash,
		}

	default:
		return nil
	}
}

// parseABCILogs attempts to parse a stringified ABCI tx log into a slice of
// ABCIMessageLog types. It returns an error upon JSON decoding failure.
func parseABCILogs(logs string) (res []*apiacbci.ABCIMessageLog, err error) {
	err = json.Unmarshal([]byte(logs), &res)
	return res, err
}
