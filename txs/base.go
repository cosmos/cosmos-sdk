package txs

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/basecoin/types"
	"github.com/tendermint/go-wire/data"
)

const (
	// for utils...
	ByteRaw = 0x1
	// TODO: move fees into a module, multiplexer is standard
	ByteFees  = 0x2
	ByteMulti = 0x3
	ByteChain = 0x4

	// for signatures
	ByteSig      = 0x16
	ByteMultiSig = 0x17
)

const (
	// for utils...
	TypeRaw   = "raw"
	TypeFees  = "fee"
	TypeMulti = "multi"
	TypeChain = "chain"

	// for signatures
	TypeSig      = "sig"
	TypeMultiSig = "multisig"
)

const (
	rawMaxSize = 2000 * 1000
)

func init() {
	basecoin.TxMapper.
		RegisterImplementation(Raw{}, TypeRaw, ByteRaw).
		RegisterImplementation(&Fee{}, TypeFees, ByteFees).
		RegisterImplementation(&MultiTx{}, TypeMulti, ByteMulti).
		RegisterImplementation(&Chain{}, TypeChain, ByteChain)
}

// Raw just contains bytes that can be hex-ified
type Raw struct {
	data.Bytes
}

func (r Raw) Wrap() basecoin.Tx {
	return basecoin.Tx{r}
}

func (r Raw) ValidateBasic() error {
	if len(r.Bytes) > rawMaxSize {
		return errors.TooLarge()
	}
	return nil
}

func NewRaw(d []byte) Raw {
	return Raw{data.Bytes(d)}
}

/**** Fee ****/

// Fee attaches a fee payment to the embedded tx
type Fee struct {
	Tx    basecoin.Tx    `json:"tx"`
	Fee   types.Coin     `json:"fee"`
	Payer basecoin.Actor `json:"payer"` // the address who pays the fee
	// Gas types.Coin `json:"gas"`  // ?????
}

func NewFee(tx basecoin.Tx, fee types.Coin, payer basecoin.Actor) *Fee {
	return &Fee{Tx: tx, Fee: fee, Payer: payer}
}

func (f *Fee) ValidateBasic() error {
	// TODO: more checks
	return f.Tx.ValidateBasic()
}

func (f *Fee) Wrap() basecoin.Tx {
	return basecoin.Tx{f}
}

func (f *Fee) Next() basecoin.Tx {
	return f.Tx
}

/**** MultiTx  ******/
type MultiTx struct {
	Txs []basecoin.Tx `json:"txs"`
}

func NewMultiTx(txs ...basecoin.Tx) *MultiTx {
	return &MultiTx{Txs: txs}
}

func (mt *MultiTx) Wrap() basecoin.Tx {
	return basecoin.Tx{mt}
}

func (mt *MultiTx) ValidateBasic() error {
	for _, t := range mt.Txs {
		err := t.ValidateBasic()
		if err != nil {
			return err
		}
	}
	return nil
}

/*** Chain ****/

// Chain locks this tx to one chain, wrap with this before signing
type Chain struct {
	Tx      basecoin.Tx `json:"tx"`
	ChainID string      `json:"chain_id"`
}

func NewChain(chainID string, tx basecoin.Tx) *Chain {
	return &Chain{Tx: tx, ChainID: chainID}
}

func (c *Chain) Wrap() basecoin.Tx {
	return basecoin.Tx{c}
}

func (c *Chain) ValidateBasic() error {
	// TODO: more checks? chainID?
	return c.Tx.ValidateBasic()
}
