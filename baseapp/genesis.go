package baseapp

import (
	"errors"

	"github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/genesis"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ genesis.TxHandler = (*BaseApp)(nil)

// ExecuteGenesisTx implements genesis.GenesisState from
// cosmossdk.io/core/genesis to set initial state in genesis.
//
// On success the events produced by the genesis transaction's message
// handlers are re-published on the finalize-block state's EventManager so
// that subscribers running during InitChain (event indexers, ABCI listeners
// and any custom post-genesis hook that walks the manager) can observe
// them. Without this, gen-tx events are confined to the per-tx result and
// silently dropped because ResponseInitChain has no top-level events
// field. Refs #25984.
func (ba *BaseApp) ExecuteGenesisTx(tx []byte) error {
	res := ba.deliverTx(tx, nil, nil, -1, nil)

	if res.Code != types.CodeTypeOK {
		return errors.New(res.Log)
	}

	if len(res.Events) > 0 {
		if finalizeState := ba.stateManager.GetState(execModeFinalize); finalizeState != nil {
			events := make(sdk.Events, len(res.Events))
			for i, ev := range res.Events {
				events[i] = sdk.Event(ev)
			}
			finalizeState.Context().EventManager().EmitEvents(events)
		}
	}

	return nil
}
