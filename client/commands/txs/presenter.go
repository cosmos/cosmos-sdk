package txs

import (
	"github.com/pkg/errors"

	wire "github.com/tendermint/go-wire"
	"github.com/tendermint/light-client/proofs"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/tendermint/basecoin"
)

// BaseTxPresenter this decodes all basecoin tx
type BaseTxPresenter struct {
	proofs.RawPresenter // this handles MakeKey as hex bytes
}

// ParseData - unmarshal raw bytes to a basecoin tx
func (BaseTxPresenter) ParseData(raw []byte) (interface{}, error) {
	var tx basecoin.Tx
	err := wire.ReadBinaryBytes(raw, &tx)
	return tx, err
}

// ValidateResult returns an appropriate error if the server rejected the
// tx in CheckTx or DeliverTx
func ValidateResult(res *ctypes.ResultBroadcastTxCommit) error {
	if res.CheckTx.IsErr() {
		return errors.Errorf("CheckTx: (%d): %s", res.CheckTx.Code, res.CheckTx.Log)
	}
	if res.DeliverTx.IsErr() {
		return errors.Errorf("DeliverTx: (%d): %s", res.DeliverTx.Code, res.DeliverTx.Log)
	}
	return nil
}
