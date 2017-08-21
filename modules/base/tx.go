package base

import (
	"regexp"

	sdk "github.com/cosmos/cosmos-sdk"
	"github.com/cosmos/cosmos-sdk/errors"
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
	sdk.TxMapper.
		RegisterImplementation(MultiTx{}, TypeMultiTx, ByteMultiTx).
		RegisterImplementation(ChainTx{}, TypeChainTx, ByteChainTx)
}

/**** MultiTx  ******/

// MultiTx - a transaction containing multiple transactions
type MultiTx struct {
	Txs []sdk.Tx `json:"txs"`
}

var _ sdk.TxInner = &MultiTx{}

//nolint - TxInner Functions
func NewMultiTx(txs ...sdk.Tx) sdk.Tx {
	return (MultiTx{Txs: txs}).Wrap()
}
func (mt MultiTx) Wrap() sdk.Tx {
	return sdk.Tx{mt}
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
	// name of chain, must be [A-Za-z0-9_-]+
	ChainID string `json:"chain_id"`
	// block height at which it is no longer valid, 0 means no expiration
	ExpiresAt uint64      `json:"expires_at"`
	Tx        sdk.Tx `json:"tx"`
}

var _ sdk.TxInner = &ChainTx{}

var (
	chainPattern = regexp.MustCompile("^[A-Za-z0-9_-]+$")
)

// NewChainTx wraps a particular tx with the ChainTx wrapper,
// to enforce chain and height
func NewChainTx(chainID string, expires uint64, tx sdk.Tx) sdk.Tx {
	c := ChainTx{
		ChainID:   chainID,
		ExpiresAt: expires,
		Tx:        tx,
	}
	return c.Wrap()
}

//nolint - TxInner Functions
func (c ChainTx) Wrap() sdk.Tx {
	return sdk.Tx{c}
}
func (c ChainTx) ValidateBasic() error {
	if c.ChainID == "" {
		return ErrNoChain()
	}
	if !chainPattern.MatchString(c.ChainID) {
		return ErrWrongChain(c.ChainID)
	}
	if c.Tx.Empty() {
		return errors.ErrUnknownTxType(c.Tx)
	}
	// TODO: more checks? chainID?
	return c.Tx.ValidateBasic()
}
