package baseapp

import (
	"errors"

	"github.com/cometbft/cometbft/abci/types"

	"cosmossdk.io/core/genesis"
)

// Ensure BaseApp implements the genesis.TxHandler interface
var _ genesis.TxHandler = (*BaseApp)(nil)

// ExecuteGenesisTx implements the genesis.TxHandler interface from
// cosmossdk.io/core/genesis. It executes a transaction during genesis
// to set up the initial state of the blockchain.
func (ba *BaseApp) ExecuteGenesisTx(tx []byte) error {
	// Execute the transaction using the deliverTx method
	res := ba.deliverTx(tx)

	// Check if the transaction execution was successful
	if res.Code != types.CodeTypeOK {
		return errors.New(res.Log)
	}

	return nil
}
