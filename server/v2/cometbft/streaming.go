package cometbft

import (
	"context"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	errorsmod "cosmossdk.io/errors/v2"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/server/v2/streaming"
)

// streamDeliverBlockChanges will stream all the changes happened during deliver block.
func (c *Consensus[T]) streamDeliverBlockChanges(
	ctx context.Context,
	height int64,
	txs [][]byte,
	txResults []server.TxResult,
	events []event.Event,
	stateChanges []store.StateChanges,
) error {
	// convert txresults to streaming txresults
	streamingTxResults := make([]*streaming.ExecTxResult, len(txResults))
	for i, txResult := range txResults {
		space, code, log := errorsmod.ABCIInfo(txResult.Error, c.cfg.AppTomlConfig.Trace)

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

	for _, streamingListener := range c.streaming.Listeners {
		events, err := streaming.IntoStreamingEvents(events)
		if err != nil {
			return err
		}
		if err := streamingListener.ListenDeliverBlock(ctx, streaming.ListenDeliverBlockRequest{
			BlockHeight: height,
			Txs:         txs,
			TxResults:   streamingTxResults,
			Events:      events,
		}); err != nil {
			c.logger.Error("ListenDeliverBlock listening hook failed", "height", height, "err", err)
		}

		if err := streamingListener.ListenStateChanges(ctx, intoStreamingKVPairs(stateChanges)); err != nil {
			c.logger.Error("ListenStateChanges listening hook failed", "height", height, "err", err)
		}
	}

	if c.listener == nil {
		return nil
	}
	// stream the StartBlockData to the listener.
	if c.listener.StartBlock != nil {
		if err := c.listener.StartBlock(appdata.StartBlockData{
			Height:      uint64(height),
			HeaderBytes: nil, // TODO: https://github.com/cosmos/cosmos-sdk/issues/22009
			HeaderJSON:  nil, // TODO: https://github.com/cosmos/cosmos-sdk/issues/22009
		}); err != nil {
			return err
		}
	}
	// stream the TxData to the listener.
	if c.listener.OnTx != nil {
		for i, tx := range txs {
			if err := c.listener.OnTx(appdata.TxData{
				TxIndex: int32(i),
				Bytes:   func() ([]byte, error) { return tx, nil },
				JSON:    nil, // TODO: https://github.com/cosmos/cosmos-sdk/issues/22009
			}); err != nil {
				return err
			}
		}
	}
	// stream the EventData to the listener.
	if c.listener.OnEvent != nil {
		if err := c.listener.OnEvent(appdata.EventData{Events: events}); err != nil {
			return err
		}
	}
	// stream the KVPairData to the listener.
	if c.listener.OnKVPair != nil {
		if err := c.listener.OnKVPair(appdata.KVPairData{Updates: stateChanges}); err != nil {
			return err
		}
	}
	// stream the CommitData to the listener.
	if c.listener.Commit != nil {
		if completionCallback, err := c.listener.Commit(appdata.CommitData{}); err != nil {
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
