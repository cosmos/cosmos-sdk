package txnrunner

import (
	"github.com/cosmos/cosmos-sdk/internal/blockstm"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var _ sdk.TxRunner = &STMRunner{}

// STMRunner is a public export of the internal implementation of a BlockSTM TxRunner.
type STMRunner = blockstm.STMRunner

// NewSTMRunner is a public export of the internal implementation of a BlockSTM TxRunner constructor.
func NewSTMRunner(
	txDecoder sdk.TxDecoder,
	stores []storetypes.StoreKey,
	workers int,
	estimate bool,
	coinDenom func(storetypes.MultiStore) string,
) *STMRunner {
	return blockstm.NewSTMRunner(txDecoder, stores, workers, estimate, coinDenom)
}
