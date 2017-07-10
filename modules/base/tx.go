package base

import (
	"regexp"

	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
)

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

// MultiTx - a transaction containing multiple transactions
type MultiTx struct {
	Txs []basecoin.Tx `json:"txs"`
}

var _ basecoin.TxInner = &MultiTx{}

//nolint - TxInner Functions
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
	ChainID   string      `json:"chain_id"`   // name of chain, must be [A-Za-z0-9_-]+
	ExpiresAt uint64      `json:"expires_at"` // block height at which it is no longer valid
	Tx        basecoin.Tx `json:"tx"`
}

var _ basecoin.TxInner = &ChainTx{}

var (
	chainPattern = regexp.MustCompile("^[A-Za-z0-9_-]+$")
)

//nolint - TxInner Functions
func NewChainTx(chainID string, expires uint64, tx basecoin.Tx) basecoin.Tx {
	return (ChainTx{Tx: tx, ChainID: chainID}).Wrap()
}
func (c ChainTx) Wrap() basecoin.Tx {
	return basecoin.Tx{c}
}
func (c ChainTx) ValidateBasic() error {
	if c.ChainID == "" {
		return errors.ErrNoChain()
	}
	if !chainPattern.MatchString(c.ChainID) {
		return errors.ErrWrongChain(c.ChainID)
	}
	if c.Tx.Empty() {
		return errors.ErrUnknownTxType(c.Tx)
	}
	// TODO: more checks? chainID?
	return c.Tx.ValidateBasic()
}
