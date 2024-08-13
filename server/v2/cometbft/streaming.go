package cometbft

import (
	"context"
	"fmt"

	coreappmgr "cosmossdk.io/core/app"
	"cosmossdk.io/core/event"
	"cosmossdk.io/core/store"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/server/v2/streaming"
)

// streamDeliverBlockChanges will stream all the changes happened during deliver block.
func (c *Consensus[T]) streamDeliverBlockChanges(
	ctx context.Context,
	height int64,
	txs [][]byte,
	txResults []coreappmgr.TxResult,
	events []event.Event,
	stateChanges []store.StateChanges,
) error {
	if c.listener.OnKVPair != nil {
		for _, change := range stateChanges {
			err := c.listener.OnKVPair(appdata.KVPairData{
				Updates: []appdata.ModuleKVPairUpdate{
					{
						ModuleName: fmt.Sprintf("0x%x", change.Actor), // TODO: make human readable
						Update:     change.StateChanges,
					},
				},
			})
			if err != nil {
				return err
			}
		}
	}

	// convert txresults to streaming txresults
	streamingTxResults := make([]*streaming.ExecTxResult, len(txResults))
	for i, txResult := range txResults {
		streamingTxResults[i] = &streaming.ExecTxResult{
			Code:      txResult.Code,
			Data:      txResult.Data,
			Log:       txResult.Log,
			Info:      txResult.Info,
			GasWanted: uint64ToInt64(txResult.GasWanted),
			GasUsed:   uint64ToInt64(txResult.GasUsed),
			Events:    streaming.IntoStreamingEvents(txResult.Events),
			Codespace: txResult.Codespace,
		}
	}

	for _, streamingListener := range c.streaming.Listeners {
		if err := streamingListener.ListenDeliverBlock(ctx, streaming.ListenDeliverBlockRequest{
			BlockHeight: height,
			Txs:         txs,
			TxResults:   streamingTxResults,
			Events:      streaming.IntoStreamingEvents(events),
		}); err != nil {
			c.logger.Error("ListenDeliverBlock listening hook failed", "height", height, "err", err)
		}

		if err := streamingListener.ListenStateChanges(ctx, intoStreamingKVPairs(stateChanges)); err != nil {
			c.logger.Error("ListenStateChanges listening hook failed", "height", height, "err", err)
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
