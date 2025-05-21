package baseapp

import (
	"errors"

	"github.com/cometbft/cometbft/abci/types"

	"github.com/cosmos/cosmos-sdk/core/v2/genesis"
)

var _ genesis.TxHandler = (*BaseApp)(nil)

// ExecuteGenesisTx implements genesis.GenesisState from
// github.com/cosmos/cosmos-sdk/core/v2/genesis to set initial state in genesis
func (ba *BaseApp) ExecuteGenesisTx(tx []byte) error {
	res := ba.deliverTx(tx)

	if res.Code != types.CodeTypeOK {
		return errors.New(res.Log)
	}

	return nil
}
