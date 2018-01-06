package eyes

import (
	wire "github.com/tendermint/go-wire"

	eyesmod "github.com/cosmos/cosmos-sdk/modules/eyes"
	"github.com/cosmos/cosmos-sdk/util"
)

// Tx is what is submitted to the chain.
// This embeds the tx data along with any info we want for
// decorators (just chain for now to demo)
type Tx struct {
	Tx    eyesmod.EyesTx `json:"tx"`
	Chain util.ChainData `json:"chain"`
}

// GetTx gets the tx info
func (e Tx) GetTx() interface{} {
	return e.Tx
}

// GetChain gets the chain we wish to perform the tx on
// (info for decorators)
func (e Tx) GetChain() util.ChainData {
	return e.Chain
}

// LoadTx parses the input data into our blockchain tx structure
func LoadTx(data []byte) (tx Tx, err error) {
	err = wire.ReadBinaryBytes(data, &tx)
	return
}
