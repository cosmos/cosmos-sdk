package types

import (
	"fmt"
)

var (
	ParamStoreKeyTendermintParams = []byte("tendermintparams")
)

type TendermintParams struct {
	MaxTxNumPerBlock int `json:"max_tx_num_per_block" yaml:"max_tx_num_per_block"`
}

// NewMaxTxNumParams creates a new TendermintParams object
func NewTendermintParams(maxTxNumPerBlock int) TendermintParams {
	return TendermintParams{
		MaxTxNumPerBlock: maxTxNumPerBlock,
	}
}

func (tp TendermintParams) String() string {
	return fmt.Sprintf(`Tendermint Params:
  MaxTxNumPerBlock: %d`, tp.MaxTxNumPerBlock)
}
