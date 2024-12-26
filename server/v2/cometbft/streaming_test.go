package cometbft

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/event"
	"cosmossdk.io/core/server"
	"cosmossdk.io/core/store"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/server/v2/streaming"
)

func TestIntoStreamingKVPairs(t *testing.T) {
	tests := []struct {
		name         string
		stateChanges []store.StateChanges
		expected     []*streaming.StoreKVPair
	}{
		{
			name:         "empty state changes",
			stateChanges: []store.StateChanges{},
			expected:     []*streaming.StoreKVPair{},
		},
		{
			name: "single state change",
			stateChanges: []store.StateChanges{
				{
					Actor: []byte("actor1"),
					StateChanges: []store.KVPair{
						{
							Key:    []byte("key1"),
							Value:  []byte("value1"),
							Remove: false,
						},
					},
				},
			},
			expected: []*streaming.StoreKVPair{
				{
					Address: []byte("actor1"),
					Key:     []byte("key1"),
					Value:   []byte("value1"),
					Delete:  false,
				},
			},
		},
		{
			name: "multiple state changes with delete",
			stateChanges: []store.StateChanges{
				{
					Actor: []byte("actor1"),
					StateChanges: []store.KVPair{
						{
							Key:    []byte("key1"),
							Value:  []byte("value1"),
							Remove: false,
						},
						{
							Key:    []byte("key2"),
							Value:  []byte{},
							Remove: true,
						},
					},
				},
				{
					Actor: []byte("actor2"),
					StateChanges: []store.KVPair{
						{
							Key:    []byte("key3"),
							Value:  []byte("value3"),
							Remove: false,
						},
					},
				},
			},
			expected: []*streaming.StoreKVPair{
				{
					Address: []byte("actor1"),
					Key:     []byte("key1"),
					Value:   []byte("value1"),
					Delete:  false,
				},
				{
					Address: []byte("actor1"),
					Key:     []byte("key2"),
					Value:   []byte{},
					Delete:  true,
				},
				{
					Address: []byte("actor2"),
					Key:     []byte("key3"),
					Value:   []byte("value3"),
					Delete:  false,
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := intoStreamingKVPairs(tc.stateChanges)
			require.Equal(t, tc.expected, result)
		})
	}
}

func TestStreamDeliverBlockChanges(t *testing.T) {
	mockConsensus := &consensus[transaction.Tx]{
		streamingManager: streaming.Manager{
			Listeners: []streaming.Listener{
				&mockStreamingListener{},
			},
		},
	}

	mockConsensus.cfg = Config{AppTomlConfig: DefaultAppTomlConfig()}

	ctx := context.Background()
	height := int64(100)
	txs := [][]byte{[]byte("tx1"), []byte("tx2")}

	decodedTxs := []transaction.Tx{InjectedTx([]byte("decoded1")), InjectedTx([]byte("decoded2"))}
	txResults := []server.TxResult{
		{
			Error:     nil,
			GasWanted: 1000,
			GasUsed:   500,
			Events:    []event.Event{},
		},
	}
	events := []event.Event{}
	stateChanges := []store.StateChanges{
		{
			Actor: []byte("actor1"),
			StateChanges: []store.KVPair{
				{
					Key:    []byte("key1"),
					Value:  []byte("value1"),
					Remove: false,
				},
			},
		},
	}

	err := mockConsensus.streamDeliverBlockChanges(ctx, height, txs, decodedTxs, txResults, events, stateChanges)
	require.NoError(t, err)
}

type mockStreamingListener struct{}

func (m *mockStreamingListener) ListenDeliverBlock(ctx context.Context, req streaming.ListenDeliverBlockRequest) error {
	return nil
}

func (m *mockStreamingListener) ListenStateChanges(ctx context.Context, kvPairs []*streaming.StoreKVPair) error {
	return nil
}
