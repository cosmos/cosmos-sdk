package baseapp

import (
	"io"
	"sync"

	abci "github.com/tendermint/tendermint/abci/types"

	store "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/types"
)

// ABCIListener interface used to hook into the ABCI message processing of the BaseApp
type ABCIListener interface {
	// ListenBeginBlock updates the streaming service with the latest BeginBlock messages
	ListenBeginBlock(ctx types.Context, req abci.RequestBeginBlock, res abci.ResponseBeginBlock) error
	// ListenEndBlock updates the steaming service with the latest EndBlock messages
	ListenEndBlock(ctx types.Context, req abci.RequestEndBlock, res abci.ResponseEndBlock) error
	// ListenDeliverTx updates the steaming service with the latest DeliverTx messages
	ListenDeliverTx(ctx types.Context, req abci.RequestDeliverTx, res abci.ResponseDeliverTx) error
}

// StreamingService interface for registering WriteListeners with the BaseApp and updating the service with the ABCI messages using the hooks
type StreamingService interface {
	// Stream is the streaming service loop, awaits kv pairs and writes them to some destination stream or file
	Stream(wg *sync.WaitGroup) error
	// Listeners returns the streaming service's listeners for the BaseApp to register
	Listeners() map[store.StoreKey][]store.WriteListener
	// ABCIListener interface for hooking into the ABCI messages from inside the BaseApp
	ABCIListener
	// Closer interface
	io.Closer
}
