package types

import (
	"context"

	abci "github.com/cometbft/cometbft/api/cometbft/abci/v1"

	"cosmossdk.io/schema/appdata"
)

// ABCIListener is the interface that we're exposing as a streaming service.
// It hooks into the ABCI message processing of the BaseApp.
// The error results are propagated to consensus state machine,
// if you don't want to affect consensus, handle the errors internally and always return `nil` in these APIs.
type ABCIListener interface {
	// ListenFinalizeBlock updates the streaming service with the latest FinalizeBlock messages
	ListenFinalizeBlock(ctx context.Context, req abci.FinalizeBlockRequest, res abci.FinalizeBlockResponse) error
	// ListenCommit updates the steaming service with the latest Commit messages and state changes
	ListenCommit(ctx context.Context, res abci.CommitResponse, changeSet []*StoreKVPair) error
}

// StreamingManager is the struct that maintains a list of ABCIListeners and configuration settings.
type StreamingManager struct {
	// ABCIListeners for hooking into the ABCI message processing of the BaseApp
	// and exposing the requests and responses to external consumers
	ABCIListeners []ABCIListener

	// StopNodeOnErr halts the node when ABCI streaming service listening results in an error.
	StopNodeOnErr bool
}

func FromSchemaListener(listener appdata.Listener) ABCIListener {
	return &listenerWrapper{listener: listener}
}

type listenerWrapper struct {
	listener appdata.Listener
}

func (p listenerWrapper) ListenFinalizeBlock(_ context.Context, req abci.FinalizeBlockRequest, res abci.FinalizeBlockResponse) error {
	if p.listener.StartBlock != nil {
		err := p.listener.StartBlock(appdata.StartBlockData{
			Height: uint64(req.Height),
		})
		if err != nil {
			return err
		}
	}

	//// TODO txs, events

	return nil
}

func (p listenerWrapper) ListenCommit(ctx context.Context, res abci.CommitResponse, changeSet []*StoreKVPair) error {
	if p.listener.OnKVPair != nil {
		for _, _ = range changeSet {
			//err := p.listener.OnKVPair(
			//	kv.StoreKey, kv.Key, kv.Value, kv.Delete)
			//if err != nil {
			//	return err
			//}
			panic("TODO")
		}
	}
	//
	//if p.listener.Commit != nil {
	//	err := p.listener.Commit()
	//	if err != nil {
	//		return err
	//	}
	//}
	return nil
}
