package baseapp

import (
	"errors"

	"github.com/cometbft/cometbft/abci/types"
)

// ExecuteGenesis implements a genesis TxHandler used to execute a genTxs (from genutil).
func (ba *BaseApp) ExecuteGenesisTx(tx []byte) error {
	res := ba.deliverTx(tx)

	if res.Code != types.CodeTypeOK {
		return errors.New(res.Log)
	}

	return nil
}
