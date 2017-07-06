package base

import "github.com/tendermint/basecoin"

// nolint
const (
	// for utils...
	ByteMultiTx = 0x2
	ByteChainTx = 0x3
)

//nolint
const (
	TypeMultiTx = NameMultiplexer + "/tx"
	TypeChainTx = NameChain + "/tx"
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(MultiTx{}, TypeMultiTx, ByteMultiTx).
		RegisterImplementation(ChainTx{}, TypeChainTx, ByteChainTx)
}

/**** MultiTx  ******/
type MultiTx struct {
	Txs []basecoin.Tx `json:"txs"`
}

func NewMultiTx(txs ...basecoin.Tx) basecoin.Tx {
	return (MultiTx{Txs: txs}).Wrap()
}

func (mt MultiTx) Wrap() basecoin.Tx {
	return basecoin.Tx{mt}
}

func (mt MultiTx) ValidateBasic() error {
	for _, t := range mt.Txs {
		err := t.ValidateBasic()
		if err != nil {
			return err
		}
	}
	return nil
}

/*** ChainTx ****/

// ChainTx locks this tx to one chainTx, wrap with this before signing
type ChainTx struct {
	Tx      basecoin.Tx `json:"tx"`
	ChainID string      `json:"chain_id"`
}

func NewChainTx(chainID string, tx basecoin.Tx) basecoin.Tx {
	return (ChainTx{Tx: tx, ChainID: chainID}).Wrap()
}

func (c ChainTx) Wrap() basecoin.Tx {
	return basecoin.Tx{c}
}

func (c ChainTx) ValidateBasic() error {
	// TODO: more checks? chainID?
	return c.Tx.ValidateBasic()
}
