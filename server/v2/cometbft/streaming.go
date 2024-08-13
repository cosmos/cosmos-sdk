package cometbft

import (
	"context"
	"encoding/json"
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
	if c.listener.StartBlock != nil {
		err := c.listener.StartBlock(appdata.StartBlockData{
			Height:      uint64(height),
			HeaderBytes: nil, // TODO: missing this data
			HeaderJSON:  nil, // TODO: missing this data
		})
		if err != nil {
			return err
		}
	}
	if c.listener.OnTx != nil {
		for i, tx := range txs {
			err := c.listener.OnTx(appdata.TxData{
				TxIndex: int32(i),
				Bytes: func() ([]byte, error) {
					return tx, nil
				},
				JSON: nil, // TODO: missing this data
			})
			if err != nil {
				return err
			}
		}
	}
	if c.listener.OnEvent != nil {
		for i, result := range txResults {
			for j, e := range result.Events {
				err := c.listener.OnEvent(appdata.EventData{
					TxIndex:    int32(i),
					MsgIndex:   -1,       // TODO: missing this data
					EventIndex: int32(j), // TODO: this doesn't match the spec because it should be the index of the event in the message
					Type:       e.Type,
					Data: func() (json.RawMessage, error) {
						// TODO: this is unnecessarily lossy for typed events which have their own JSON encoding
						m := map[string]interface{}{}
						for _, attr := range e.Attributes {
							m[attr.Key] = attr.Value
						}
						return json.Marshal(m)
					},
				})
				if err != nil {
					return err
				}
			}
		}
		// TODO: begin/end block events are in the main event array bundled with tx events which we already sent, what to do??
	}
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
	if c.listener.Commit != nil {
		err := c.listener.Commit(appdata.CommitData{})
		if err != nil {
			return err
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
