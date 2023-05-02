package baseapp

import (
	"errors"

	"cosmossdk.io/core/genesis"
	"github.com/cometbft/cometbft/abci/types"
)

var _ genesis.GenesisTxHandler = (*BaseApp)(nil)

// ExecuteGenesisTx implements genesis.GenesisState from
// cosmossdk.io/core/genesis to set initial state in genesis
func (ba BaseApp) ExecuteGenesisTx(tx []byte) error {
	res := ba.DeliverTx(types.RequestDeliverTx{Tx: tx})

	if res.Code != types.CodeTypeOK {
		return errors.New(res.Log)
	}

	return nil
}
