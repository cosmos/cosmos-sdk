package types

import (
	"sync"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// Hook interface used to hook into the ABCI message processing of the BaseApp
type Hook interface {
	// update the streaming service with the latest BeginBlock messages
	ListenBeginBlock(ctx Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error
	// update the steaming service with the latest EndBlock messages
	ListenEndBlock(ctx Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error
	// update the steaming service with the latest DeliverTx messages
	ListenDeliverTx(ctx Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error
}

// StreamingService interface for registering WriteListeners with the BaseApp and updating the service with the ABCI messages using the hooks
type StreamingService interface {
	// streaming service loop, awaits kv pairs and writes them to some destination stream or file
	Stream(wg *sync.WaitGroup, quitChan <-chan struct{})
	// returns the streaming service's listeners for the BaseApp to register
	Listeners() map[StoreKey][]types.WriteListener
	// interface for hooking into the ABCI messages from inside the BaseApp
	Hook
}
