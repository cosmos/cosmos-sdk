package baseapp

import (
	"errors"

	"github.com/cometbft/cometbft/abci/types"
)

var _ TxHandler = (*BaseApp)(nil)

// TxHandler is an interface that defines how genesis txs are handled.
// By default, BaseApp handles them using the deliverTx method.
type TxHandler interface {
	ExecuteGenesisTx([]byte) error
}

// ExecuteGenesis implements TxHandler. It executes a genesis tx.
func (ba *BaseApp) ExecuteGenesisTx(tx []byte) error {
	res := ba.deliverTx(tx)

	if res.Code != types.CodeTypeOK {
		return errors.New(res.Log)
	}

	return nil
}
