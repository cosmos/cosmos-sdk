package cometbft

import (
	"context"
	"encoding/json"
	"fmt"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	errorsmod "cosmossdk.io/errors/v2"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/server/v2/streaming"
)

// streamDeliverBlockChanges will stream all the changes happened during deliver block.
func (c *consensus[T]) streamDeliverBlockChanges(
	ctx context.Context,
	height int64,
	txs [][]byte,
	decodedTxs []T,
	blockResp server.BlockResponse,
	stateChanges []store.StateChanges,
) error {
	return StreamOut(ctx, height, txs, decodedTxs, blockResp, stateChanges, c.streamingManager, c.listener, c.cfg.AppTomlConfig.Trace, c.logger.Error)
}

// StreamOut stream all the changes happened during deliver block.
func StreamOut[T transaction.Tx](
	ctx context.Context,
	height int64,
	rawTXs [][]byte,
	decodedTXs []T,
	blockRsp server.BlockResponse,
	stateChanges []store.StateChanges,
	streamingManager streaming.Manager,
	listener *appdata.Listener,
	traceErrs bool,
	logErrFn func(msg string, keyVals ...any),
) error {
	var events []event.Event
	events = append(events, blockRsp.PreBlockEvents...)
	events = append(events, blockRsp.BeginBlockEvents...)
	for _, tx := range blockRsp.TxResults {
		events = append(events, tx.Events...)
	}
	events = append(events, blockRsp.EndBlockEvents...)
	txResults := blockRsp.TxResults

	err := doServeStreamListeners(
		ctx,
		height,
		rawTXs,
		txResults,
		traceErrs,
		streamingManager,
		events,
		logErrFn,
		stateChanges,
	)
	if err != nil {
		return err
	}
	return doServeHookListener(listener, height, rawTXs, decodedTXs, events, stateChanges)
}

func doServeStreamListeners(
	ctx context.Context,
	height int64,
	rawTXs [][]byte,
	txResults []server.TxResult,
	traceErrs bool,
	streamingManager streaming.Manager,
	events []event.Event,
	logErrFn func(msg string, keyVals ...any),
	stateChanges []store.StateChanges,
) error {
	if len(streamingManager.Listeners) == 0 {
		return nil
	}
	// convert txresults to streaming txresults
	streamingTxResults := make([]*streaming.ExecTxResult, len(txResults))
	for i, txResult := range txResults {
		space, code, log := errorsmod.ABCIInfo(txResult.Error, traceErrs)

		events, err := streaming.IntoStreamingEvents(txResult.Events)
		if err != nil {
			return err
		}

		streamingTxResults[i] = &streaming.ExecTxResult{
			Code:      code,
			Codespace: space,
			Log:       log,
			GasWanted: uint64ToInt64(txResult.GasWanted),
			GasUsed:   uint64ToInt64(txResult.GasUsed),
			Events:    events,
		}
	}

	for _, streamingListener := range streamingManager.Listeners {
		events, err := streaming.IntoStreamingEvents(events)
		if err != nil {
			return err
		}
		if err := streamingListener.ListenDeliverBlock(ctx, streaming.ListenDeliverBlockRequest{
			BlockHeight: height,
			Txs:         rawTXs,
			TxResults:   streamingTxResults,
			Events:      events,
		}); err != nil {
			if streamingManager.StopNodeOnErr {
				return fmt.Errorf("listen deliver block: %w", err)
			}
			logErrFn("ListenDeliverBlock listening hook failed", "height", height, "err", err)
		}

		if err := streamingListener.ListenStateChanges(ctx, intoStreamingKVPairs(stateChanges)); err != nil {
			if streamingManager.StopNodeOnErr {
				return fmt.Errorf("listen state changes: %w", err)
			}
			logErrFn("ListenStateChanges listening hook failed", "height", height, "err", err)
		}
	}
	return nil
}

func doServeHookListener[T transaction.Tx](
	listener *appdata.Listener,
	height int64,
	rawTXs [][]byte,
	decodedTXs []T,
	events []event.Event,
	stateChanges []store.StateChanges,
) error {
	if listener == nil {
		return nil
	}
	// stream the StartBlockData to the listener.
	if listener.StartBlock != nil {
		if err := listener.StartBlock(appdata.StartBlockData{
			Height:      uint64(height),
			HeaderBytes: nil, // TODO: https://github.com/cosmos/cosmos-sdk/issues/22009
			HeaderJSON:  nil, // TODO: https://github.com/cosmos/cosmos-sdk/issues/22009
		}); err != nil {
			return err
		}
	}
	// stream the TxData to the listener.
	if listener.OnTx != nil {
		for i, tx := range rawTXs {
			if err := listener.OnTx(appdata.TxData{
				BlockNumber: uint64(height),
				TxIndex:     int32(i),
				Bytes:       func() ([]byte, error) { return tx, nil },
				JSON: func() (json.RawMessage, error) {
					return json.Marshal(decodedTXs[i])
				},
			}); err != nil {
				return err
			}
		}
	}
	// stream the EventData to the listener.
	if listener.OnEvent != nil {
		if err := listener.OnEvent(appdata.EventData{Events: events}); err != nil {
			return err
		}
	}
	// stream the KVPairData to the listener.
	if listener.OnKVPair != nil {
		if err := listener.OnKVPair(appdata.KVPairData{Updates: stateChanges}); err != nil {
			return err
		}
	}
	// stream the CommitData to the listener.
	if listener.Commit != nil {
		if completionCallback, err := listener.Commit(appdata.CommitData{}); err != nil {
			return err
		} else if completionCallback != nil {
			if err := completionCallback(); err != nil {
				return err
			}
		}
	}

	return nil
}

func intoStreamingKVPairs(stateChanges []store.StateChanges) []*streaming.StoreKVPair {
	// Calculate the total number of KV pairs to preallocate the slice with the required capacity.
	totalKvPairs := 0
	for _, accounts := range stateChanges {
		totalKvPairs += len(accounts.StateChanges)
	}

	// Preallocate the slice with the required capacity.
	streamKvPairs := make([]*streaming.StoreKVPair, 0, totalKvPairs)

	for _, accounts := range stateChanges {
		// Reducing the scope of address variable.
		address := accounts.Actor

		for _, kv := range accounts.StateChanges {
			streamKvPairs = append(streamKvPairs, &streaming.StoreKVPair{
				Address: address,
				Key:     kv.Key,
				Value:   kv.Value,
				Delete:  kv.Remove,
			})
		}
	}
	return streamKvPairs
}
