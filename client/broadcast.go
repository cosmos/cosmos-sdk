package client

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cometbft/cometbft/mempool"
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

	// To prevent races, first subscribe to a query for the transaction's
	// inclusion in a block.  Clean up when we are done.
	txHash := fmt.Sprintf("%X", tmhash.Sum(txBytes))
	waitTxCh := WaitTx(cctx, ctx.NodeURI, txHash)

	// Broadcast the transaction.
	res, err := node.BroadcastTxSync(context.Background(), txBytes)
	if errRes := CheckTendermintError(err, txBytes); errRes != nil {
		return errRes, nil
	}

	// Check for an error in the broadcast.
	syncRes := sdk.NewResponseFormatBroadcastTx(res)
	if syncRes.Code != 0 {
		return syncRes, err
	}

	// Wait for the transaction to be committed.
	waitTx := <-waitTxCh

	// Process the event data and return the commit response.
	if evt := waitTx.BlockInclusion; evt != nil {
		evtHash := fmt.Sprintf("%X", tmtypes.Tx(evt.Tx).Hash())
		if evtHash != txHash {
			return syncRes, fmt.Errorf("tx hash did not match: got %s, expected %s", evtHash, txHash)
		}

		parsedLogs, parseErr := sdk.ParseABCILogs(evt.Result.Log)
		commitRes := &sdk.TxResponse{
			TxHash:    evtHash,
			Height:    evt.Height,
			Codespace: evt.Result.Codespace,
			Code:      evt.Result.Code,
			Data:      strings.ToUpper(hex.EncodeToString(evt.Result.Data)),
			RawLog:    evt.Result.Log,
			Logs:      parsedLogs,
			Info:      evt.Result.Info,
			GasWanted: evt.Result.GasWanted,
			GasUsed:   evt.Result.GasUsed,
			Events:    evt.Result.Events,
		}
		if !evt.Result.IsOK() {
			return commitRes, fmt.Errorf("unexpected result code %d", evt.Result.Code)
		}
		return commitRes, parseErr
	}

	if waitTx.Err != nil {
		return syncRes, waitTx.Err
	}
	return syncRes, sdkerrors.ErrLogic.Wrapf("tx block inclusion not detected")
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

type WaitTxResult struct {
	BlockInclusion *tmtypes.EventDataTx
	Err            error
}

// [AGORIC]: WaitTx subscribes to the transaction result event raised
// when the transaction matching the hash either errors or is included in a block.
// It returns a channel for the response.
func WaitTx(ctx context.Context, nodeURI, hash string) <-chan WaitTxResult {
	// Guestimate the capacity of the deferrals, no big deal if we're wrong
	deferrals := make([]func(), 0, 2)
	resultCh := make(chan WaitTxResult, 1)

	// We wrap the return results in a function to ensure that we will run the
	// deferrals only when we are producing the result channel.
	result := func(res *tmtypes.EventDataTx, err error) <-chan WaitTxResult {
		// Ensure our deferrals are run before the caller gets its hands on the
		// result channel.
		defer func() {
			for i := len(deferrals) - 1; i >= 0; i-- {
				deferrals[i]()
			}
		}()
		resultCh <- WaitTxResult{res, err}
		close(resultCh)
		return resultCh
	}

	// We track our own deferrals to ensure that we clean up the client and
	// subscription regardless of whether we return synchronously, or from within
	// the goroutine.
	deferral := func(deferred func()) {
		deferrals = append(deferrals, deferred)
	}

	c, err := rpchttp.New(nodeURI, "/websocket")
	if err != nil {
		return result(nil, err)
	}
	if err := c.Start(); err != nil {
		return result(nil, err)
	}
	deferral(func() { c.Stop() }) //nolint:errcheck // ignore stop error

	// Subscribe to the tx event.
	query := fmt.Sprintf("%s='%s' AND %s='%s'", tmtypes.EventTypeKey, tmtypes.EventTx, tmtypes.TxHashKey, hash)
	const subscriber = "subscriber"
	eventCh, err := c.Subscribe(ctx, subscriber, query)
	if err != nil {
		return result(nil, fmt.Errorf("failed to subscribe to tx: %w", err))
	}
	deferral(func() { c.UnsubscribeAll(context.Background(), subscriber) }) //nolint:errcheck // ignore unsubscribe error

	// Since we're now fully subscribed, we can return the channel and wait for
	// the event or context deadline in a background goroutine.
	go func() <-chan WaitTxResult {
		select {
		case evt := <-eventCh:
			if txe, ok := evt.Data.(tmtypes.EventDataTx); ok {
				return result(&txe, nil)
			}
			return result(nil, sdkerrors.ErrLogic.Wrapf("unsupported event data type %T", evt.Data))
		case <-ctx.Done():
			return result(nil, sdkerrors.ErrLogic.Wrapf("timed out waiting for event"))
		}
	}()
	return resultCh
}
