package baseapp

import (
	"errors"

	abci "github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/genesis"
)

var _ genesis.TxHandler = (*BaseApp)(nil)

// ExecuteGenesisTx implements genesis.TxHandler from
// cosmossdk.io/core/genesis to set initial state in genesis.
func (ba *BaseApp) ExecuteGenesisTx(tx []byte) error {
	res := ba.deliverTx(tx, nil, nil, -1, nil)
	if res.Code != abci.CodeTypeOK {
		return errors.New(res.Log)
	}

	ba.genesisEvents = append(ba.genesisEvents, res.Events...)

	return nil
}
