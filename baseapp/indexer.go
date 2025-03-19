package baseapp

import (
	"context"
	"cosmossdk.io/log"
	"cosmossdk.io/schema"
	"cosmossdk.io/schema/appdata"
	"cosmossdk.io/schema/decoding"
	"cosmossdk.io/schema/indexer"
	storetypes "cosmossdk.io/store/types"
	"encoding/json"
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
)

// EnableIndexer enables the built-in indexer with the provided options (usually from the app.toml indexer key),
// kv-store keys, and app modules. Using the built-in indexer framework is mutually exclusive from using other
// types of streaming listeners.
func (app *BaseApp) EnableIndexer(indexerOpts interface{}, keys map[string]*storetypes.KVStoreKey, appModules map[string]any) error {
	listener, err := indexer.StartIndexing(indexer.IndexingOptions{
		Config:       indexerOpts,
		Resolver:     decoding.ModuleSetDecoderResolver(appModules),
		Logger:       app.logger.With(log.ModuleKey, "indexer"),
		SyncSource:   nil, // TODO: Support catch-up syncs
		AddressCodec: app.interfaceRegistry.SigningContext().AddressCodec(),
	})
	if err != nil {
		return err
	}

	exposedKeys := exposeStoreKeysSorted([]string{"*"}, keys)
	app.cms.AddListeners(exposedKeys)

	app.streamingManager = storetypes.StreamingManager{
		ABCIListeners: []storetypes.ABCIListener{listenerWrapper{listener.Listener, app.txDecoder}},
		StopNodeOnErr: true,
	}

	return nil
}

func eventToAppDataEvent(event abci.Event, height int64) (appdata.Event, error) {
	appdataEvent := appdata.Event{
		BlockNumber: uint64(height),
		Type:        event.Type,
		Attributes: func() ([]appdata.EventAttribute, error) {
			attrs := make([]appdata.EventAttribute, len(event.Attributes))
			for j, attr := range event.Attributes {
				attrs[j] = appdata.EventAttribute{
					Key:   attr.Key,
					Value: attr.Value,
				}
			}
			return attrs, nil
		},
	}

	for _, attr := range event.Attributes {
		if attr.Key == "mode" {
			switch attr.Value {
			case "PreBlock":
				appdataEvent.BlockStage = appdata.PreBlockStage
			case "BeginBlock":
				appdataEvent.BlockStage = appdata.BeginBlockStage
			case "EndBlock":
				appdataEvent.BlockStage = appdata.EndBlockStage
			default:
				appdataEvent.BlockStage = appdata.UnknownBlockStage
			}
		} else if attr.Key == "tx_index" {
			txIndex, err := strconv.Atoi(attr.Value)
			if err != nil {
				return appdata.Event{}, err
			}
			appdataEvent.TxIndex = int32(txIndex + 1)
			appdataEvent.BlockStage = appdata.TxProcessingStage
		} else if attr.Key == "msg_index" {
			msgIndex, err := strconv.Atoi(attr.Value)
			if err != nil {
				return appdata.Event{}, err
			}
			appdataEvent.MsgIndex = int32(msgIndex + 1)
		} else if attr.Key == "event_index" {
			eventIndex, err := strconv.Atoi(attr.Value)
			if err != nil {
				return appdata.Event{}, err
			}
			appdataEvent.EventIndex = int32(eventIndex + 1)
		}
	}

	return appdataEvent, nil
}

type listenerWrapper struct {
	listener  appdata.Listener
	txDecoder sdk.TxDecoder
}

// NewListenerWrapper creates a new ABCIListener that wraps an appdata.Listener.
// This is primarily intended for testing purposes, although you could use
// this for a custom indexing setup.
// Generally, you should use BaseApp.EnableIndexer to enable the built-in indexer.
func NewListenerWrapper(listener appdata.Listener, txDecoder sdk.TxDecoder) storetypes.ABCIListener {
	return listenerWrapper{listener: listener, txDecoder: txDecoder}
}

func (p listenerWrapper) ListenFinalizeBlock(_ context.Context, req abci.RequestFinalizeBlock, res abci.ResponseFinalizeBlock) error {
	if p.listener.StartBlock != nil {
		// clean up redundant data
		reqWithoutTxs := req
		reqWithoutTxs.Txs = nil

		if err := p.listener.StartBlock(appdata.StartBlockData{
			Height:      uint64(req.Height),
			HeaderBytes: nil, // TODO: https://github.com/cosmos/cosmos-sdk/issues/22009
			HeaderJSON: func() (json.RawMessage, error) {
				return json.Marshal(reqWithoutTxs)
			},
		}); err != nil {
			return err
		}
	}
	if p.listener.OnTx != nil {
		for i, tx := range req.Txs {
			if err := p.listener.OnTx(appdata.TxData{
				BlockNumber: uint64(req.Height),
				TxIndex:     int32(i),
				Bytes:       func() ([]byte, error) { return tx, nil },
				JSON: func() (json.RawMessage, error) {
					sdkTx, err := p.txDecoder(tx)
					if err != nil {
						// if the transaction cannot be decoded, return the error as JSON
						// as there are some txs that might not be decodeable by the txDecoder
						return json.Marshal(err)
					}
					return json.Marshal(sdkTx)
				},
			}); err != nil {
				return err
			}
		}
	}
	if p.listener.OnEvent != nil {
		events := make([]appdata.Event, len(res.Events))
		var err error
		for i, event := range res.Events {
			events[i], err = eventToAppDataEvent(event, req.Height)
			if err != nil {
				return err
			}
		}
		for _, txResult := range res.TxResults {
			for _, event := range txResult.Events {
				appdataEvent, err := eventToAppDataEvent(event, req.Height)
				if err != nil {
					return err
				}
				events = append(events, appdataEvent)
			}
		}
		if err := p.listener.OnEvent(appdata.EventData{Events: events}); err != nil {
			return err
		}
	}

	return nil
}

func (p listenerWrapper) ListenCommit(ctx context.Context, res abci.ResponseCommit, changeSet []*storetypes.StoreKVPair) error {
	if cb := p.listener.OnKVPair; cb != nil {
		updates := make([]appdata.ActorKVPairUpdate, len(changeSet))
		for i, pair := range changeSet {
			updates[i] = appdata.ActorKVPairUpdate{
				Actor: []byte(pair.StoreKey),
				StateChanges: []schema.KVPairUpdate{
					{
						Key:    pair.Key,
						Value:  pair.Value,
						Remove: pair.Delete,
					},
				},
			}
		}
		err := cb(appdata.KVPairData{Updates: updates})
		if err != nil {
			return err
		}
	}

	if p.listener.Commit != nil {
		commitCb, err := p.listener.Commit(appdata.CommitData{})
		if err != nil {
			return err
		}
		if commitCb != nil {
			err := commitCb()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
