package cometbft

import (
	"context"

	"cosmossdk.io/server/v2/core/event"
	"cosmossdk.io/server/v2/core/store"
	"cosmossdk.io/server/v2/streaming"
)

// streamDeliverBlockChanges will stream all the changes happened during deliver block.
func (c *Consensus[T]) streamDeliverBlockChanges(
	ctx context.Context,
	height int64,
	events []event.Event,
	stateChanges []store.AccountStateChanges,
) error {
	for _, streamingListener := range c.streaming.Listeners {
		if err := streamingListener.ListenDeliverBlock(ctx, streaming.ListenDeliverBlockRequest{
			BlockHeight: height,
			// Txs:         req.Txs, TODO: see how to map txs
			Events: streaming.IntoStreamingEvents(events),
		}); err != nil {
			c.logger.Error("ListenDeliverBlock listening hook failed", "height", height, "err", err)
		}

		if err := streamingListener.ListenStateChanges(ctx, intoStreamingKVPairs(stateChanges)); err != nil {
			c.logger.Error("ListenStateChanges listening hook failed", "height", height, "err", err)
		}
	}
	return nil
}

func intoStreamingKVPairs(stateChanges []store.AccountStateChanges) []*streaming.StoreKVPair {
	// Calculate the total number of KV pairs to preallocate the slice with the required capacity.
	totalKvPairs := 0
	for _, accounts := range stateChanges {
		totalKvPairs += len(accounts.StateChanges)
	}

	// Preallocate the slice with the required capacity.
	streamKvPairs := make([]*streaming.StoreKVPair, 0, totalKvPairs)

	for _, accounts := range stateChanges {
		// Reducing the scope of address variable.
		address := accounts.Account

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
