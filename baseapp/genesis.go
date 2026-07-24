package baseapp

import (
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/genesis"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ genesis.TxHandler = (*BaseApp)(nil)

// ExecuteGenesisTx implements genesis.TxHandler from
// cosmossdk.io/core/genesis to set initial state in genesis.
func (ba *BaseApp) ExecuteGenesisTx(tx []byte) error {
	res := ba.deliverTx(tx, nil, nil, -1, nil)

	if res.Code != abci.CodeTypeOK {
		return errors.New(res.Log)
	}

	// Emit events from the genesis tx so they are not silently dropped.
	// Without this, indexers that rely solely on events cannot track
	// state changes that happen during InitGenesis.
	finalizeState := ba.stateManager.GetState(execModeFinalize)
	if finalizeState != nil {
		// MarkEventsToIndex returns []abci.Event; convert to sdk.Events for EmitEvents.
		marked := sdk.MarkEventsToIndex(res.Events, ba.indexEvents)
		sdkEvents := make(sdk.Events, len(marked))
		for i, e := range marked {
			sdkEvents[i] = sdk.Event(e)
		}
		finalizeState.Context().EventManager().EmitEvents(sdkEvents)
	}

	return nil
}
