package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/mempool"
	coretypes "github.com/cometbft/cometbft/rpc/core/types"
	cmttypes "github.com/cometbft/cometbft/types"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	rpchttp "github.com/cometbft/cometbft/rpc/client/http"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// BroadcastTx broadcasts a transactions either synchronously or asynchronously
// based on the context parameters. The result of the broadcast is parsed into
// an intermediate structure which is logged if the context has a logger
// defined.
func (ctx Context) BroadcastTx(txBytes []byte) (res *sdk.TxResponse, err error) {
	switch ctx.BroadcastMode {
	case flags.BroadcastSync:
		res, err = ctx.BroadcastTxSync(txBytes)

	case flags.BroadcastAsync:
		res, err = ctx.BroadcastTxAsync(txBytes)

	// [AGORIC]: We provide BroadcastTxCommit to ensure that the transaction is
	// included in a block.
	case flags.BroadcastBlock:
		res, err = ctx.BroadcastTxCommit(txBytes)

	default:
		return nil, fmt.Errorf("unsupported return type %s; supported types: sync, async", ctx.BroadcastMode)
	}

	return res, err
}

// Deprecated: Use CheckCometError instead.
func CheckTendermintError(err error, tx cmttypes.Tx) *sdk.TxResponse {
	return CheckCometError(err, tx)
}

// CheckCometError checks if the error returned from BroadcastTx is a
// CometBFT error that is returned before the tx is submitted due to
// precondition checks that failed. If an CometBFT error is detected, this
// function returns the correct code back in TxResponse.
//
// TODO: Avoid brittle string matching in favor of error matching. This requires
// a change to CometBFT's RPCError type to allow retrieval or matching against
// a concrete error type.
func CheckCometError(err error, tx cmttypes.Tx) *sdk.TxResponse {
	if err == nil {
		return nil
	}

	errStr := strings.ToLower(err.Error())
	txHash := fmt.Sprintf("%X", tx.Hash())

	switch {
	case strings.Contains(errStr, strings.ToLower(mempool.ErrTxInCache.Error())):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrTxInMempoolCache.ABCICode(),
			Codespace: sdkerrors.ErrTxInMempoolCache.Codespace(),
			TxHash:    txHash,
		}

	case strings.Contains(errStr, "mempool is full"):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrMempoolIsFull.ABCICode(),
			Codespace: sdkerrors.ErrMempoolIsFull.Codespace(),
			TxHash:    txHash,
		}

	case strings.Contains(errStr, "tx too large"):
		return &sdk.TxResponse{
			Code:      sdkerrors.ErrTxTooLarge.ABCICode(),
			Codespace: sdkerrors.ErrTxTooLarge.Codespace(),
			TxHash:    txHash,
		}

	default:
		return nil
	}
}

// WaitTxAsBroadcastTxCommit waits for a transaction to be included in a block by subscribing
// to the CometBFT WebSocket connection and waiting for a transaction event with
// the given hash. It returns a ResultBroadcastTxCommit which contains the
// transaction result and the block height in which the transaction was included.
func (ctx Context) WaitTxAsBroadcastTxCommit(cctx context.Context, txSubmitted func() (*sdk.TxResponse, []byte, error)) (*coretypes.ResultBroadcastTxCommit, error) {
	c, err := rpchttp.New(ctx.NodeURI, "/websocket")
	if err != nil {
		return nil, err
	}
	if err := c.Start(); err != nil {
		return nil, err
	}
	defer c.Stop() //nolint:errcheck // ignore stop error

	var hash []byte
	submitRes, hash, err := txSubmitted()
	if err != nil {
		return nil, err
	}
	if submitRes != nil && submitRes.Code != 0 {
		return nil, sdkerrors.ErrLogic.Wrapf("transaction %X submission failed with code %d: %s", hash, submitRes.Code, submitRes.RawLog)
	}
	txHash := fmt.Sprintf("%X", hash)

	// subscribe to websocket events
	blockQuery := fmt.Sprintf("%s='%s'", cmttypes.EventTypeKey, cmttypes.EventNewBlockHeader)
	const subscriber = "subscriber"
	blockEventCh, err := c.Subscribe(cctx, subscriber, blockQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to new blocks: %w", err)
	}
	txQuery := fmt.Sprintf("%s='%s' AND %s='%s'", cmttypes.EventTypeKey, cmttypes.EventTx, cmttypes.TxHashKey, txHash)
	txEventCh, err := c.Subscribe(cctx, subscriber, txQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to tx: %w", err)
	}
	defer c.UnsubscribeAll(context.Background(), subscriber) //nolint:errcheck // ignore unsubscribe error

	for err = cctx.Err(); err == nil; err = cctx.Err() {
		// return immediately if tx is already included in a block
		if res, err := c.Tx(cctx, hash, false); err == nil {
			res := &coretypes.ResultBroadcastTxCommit{
				TxResult: res.TxResult,
				Hash:     res.Hash,
				Height:   res.Height,
			}
			return res, nil
		}

		// tx not yet included in a block, wait for new blocks on websocket or context timeout
		select {
		case evt := <-txEventCh:
			if txe, ok := evt.Data.(cmttypes.EventDataTx); ok {
				evtHashBytes := cmttypes.Tx(txe.Tx).Hash()
				evtHash := fmt.Sprintf("%X", evtHashBytes)
				if evtHash == txHash {
					return &coretypes.ResultBroadcastTxCommit{
						TxResult: txe.Result,
						Hash:     evtHashBytes,
						Height:   txe.Height,
					}, nil
				}
			}
		case _, ok := <-blockEventCh:
			if !ok {
				return nil, sdkerrors.ErrIO.Wrapf("could not find transaction %X included in a block; subscription closed", hash)
			}
			// Poll again to check if the transaction is included in the new block.
		case <-cctx.Done():
			// If the context is done, exit the loop via the ctx.Err() condition.
		}
	}

	return nil, sdkerrors.ErrLogic.Wrapf("could not find transaction %X included in a block: %s", hash, err.Error())
}

// WaitTx waits for a transaction to be included in a block by subscribing
// to the CometBFT WebSocket connection and waiting for a transaction event with
// the given hash. It returns a TxResponse which contains the transaction result
// and the block height in which the transaction was included.
func (ctx Context) WaitTx(cctx context.Context, ensureTxSubmitted func() (*sdk.TxResponse, []byte, error)) (*sdk.TxResponse, error) {
	res, err := ctx.WaitTxAsBroadcastTxCommit(cctx, ensureTxSubmitted)
	if err != nil {
		return nil, err
	}

	return newResponseFormatBroadcastTxCommit(res), nil
}

// BroadcastTxCommit broadcasts transaction bytes to a Tendermint node and
// waits for a commit. An error is only returned if there is no RPC node
// connection or if broadcasting fails.
//
// [AGORIC]: This function subscribes to the transaction's inclusion in a block,
// and then use BroadcastTxSync to broadcast the transaction.  This will block
// potentially forever if the transaction is never included in a block.  And so,
// it is up to the caller to ensure that proper timeout/retry logic is
// implemented.
func (ctx Context) BroadcastTxCommit(txBytes []byte) (*sdk.TxResponse, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	cctx := context.Background()
	// This is where we would allow for a timeout on the commit.  We want to
	// default to no timeout, and accept a specific timeout as a CLI flag (as is
	// done in v0.50), but hardcoding an arbitrary timeout (like in v0.47 and
	// before) is just rude.
	/*
		cctx, cancel := context.WithTimeout(cctx, <specified time.Duration>)
		defer cancel()
	*/

	txSubmitted := func() (*sdk.TxResponse, []byte, error) {
		// Broadcast the transaction.
		hash := tmhash.Sum(txBytes)

		res, err := node.BroadcastTxSync(context.Background(), txBytes)
		if errRes := CheckCometError(err, txBytes); errRes != nil {
			return errRes, hash, nil
		}

		// Check for an error in the broadcast.
		syncRes := sdk.NewResponseFormatBroadcastTx(res)
		return syncRes, hash, err
	}

	// To prevent races, first subscribe to a query for the transaction's
	// inclusion in a block.  Clean up when we are done.
	res, err := ctx.WaitTx(cctx, txSubmitted)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func newTxResponseCheckTx(res *coretypes.ResultBroadcastTxCommit) *sdk.TxResponse {
	if res == nil {
		return nil
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	parsedLogs, _ := sdk.ParseABCILogs(res.CheckTx.Log)

	return &sdk.TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Codespace: res.CheckTx.Codespace,
		Code:      res.CheckTx.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.CheckTx.Data)),
		RawLog:    res.CheckTx.Log,
		Logs:      parsedLogs,
		Info:      res.CheckTx.Info,
		GasWanted: res.CheckTx.GasWanted,
		GasUsed:   res.CheckTx.GasUsed,
		Events:    res.CheckTx.Events,
	}
}

func newTxResponseDeliverTx(res *coretypes.ResultBroadcastTxCommit) *sdk.TxResponse {
	if res == nil {
		return nil
	}

	var txHash string
	if res.Hash != nil {
		txHash = res.Hash.String()
	}

	parsedLogs, _ := sdk.ParseABCILogs(res.TxResult.Log)

	return &sdk.TxResponse{
		Height:    res.Height,
		TxHash:    txHash,
		Codespace: res.TxResult.Codespace,
		Code:      res.TxResult.Code,
		Data:      strings.ToUpper(hex.EncodeToString(res.TxResult.Data)),
		RawLog:    res.TxResult.Log,
		Logs:      parsedLogs,
		Info:      res.TxResult.Info,
		GasWanted: res.TxResult.GasWanted,
		GasUsed:   res.TxResult.GasUsed,
		Events:    res.TxResult.Events,
	}
}

func newResponseFormatBroadcastTxCommit(res *coretypes.ResultBroadcastTxCommit) *sdk.TxResponse {
	if res == nil {
		return nil
	}

	if !res.CheckTx.IsOK() {
		return newTxResponseCheckTx(res)
	}

	return newTxResponseDeliverTx(res)
}

// BroadcastTxSync broadcasts transaction bytes to a Tendermint node
// synchronously (i.e. returns after CheckTx execution).
func (ctx Context) BroadcastTxSync(txBytes []byte) (*sdk.TxResponse, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxSync(context.Background(), txBytes)
	if errRes := CheckCometError(err, txBytes); errRes != nil {
		return errRes, nil
	}

	return sdk.NewResponseFormatBroadcastTx(res), err
}

// BroadcastTxAsync broadcasts transaction bytes to a CometBFT node
// asynchronously (i.e. returns immediately).
func (ctx Context) BroadcastTxAsync(txBytes []byte) (*sdk.TxResponse, error) {
	node, err := ctx.GetNode()
	if err != nil {
		return nil, err
	}

	res, err := node.BroadcastTxAsync(context.Background(), txBytes)
	if errRes := CheckCometError(err, txBytes); errRes != nil {
		return errRes, nil
	}

	return sdk.NewResponseFormatBroadcastTx(res), err
}

// TxServiceBroadcast is a helper function to broadcast a Tx with the correct gRPC types
// from the tx service. Calls `clientCtx.BroadcastTx` under the hood.
func TxServiceBroadcast(_ context.Context, clientCtx Context, req *tx.BroadcastTxRequest) (*tx.BroadcastTxResponse, error) {
	if req == nil || req.TxBytes == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid empty tx")
	}

	clientCtx = clientCtx.WithBroadcastMode(normalizeBroadcastMode(req.Mode))
	resp, err := clientCtx.BroadcastTx(req.TxBytes)
	if err != nil {
		return nil, err
	}

	return &tx.BroadcastTxResponse{
		TxResponse: resp,
	}, nil
}

// normalizeBroadcastMode converts a broadcast mode into a normalized string
// to be passed into the clientCtx.
func normalizeBroadcastMode(mode tx.BroadcastMode) string {
	switch mode {
	case tx.BroadcastMode_BROADCAST_MODE_ASYNC:
		return "async"
	case tx.BroadcastMode_BROADCAST_MODE_SYNC:
		return "sync"
	// [AGORIC]: We reimplemented block mode as a wrapper for wait-tx.
	case tx.BroadcastMode_BROADCAST_MODE_BLOCK:
		return "block"
	default:
		return "unspecified"
	}
}
