package txs

import (
	"github.com/tendermint/basecoin"
	"github.com/tendermint/basecoin/errors"
	"github.com/tendermint/go-wire/data"
)

const (
	// for utils...
	ByteRaw   = 0x1
	ByteMulti = 0x2
	ByteChain = 0x3

	// for signatures
	ByteSig      = 0x16
	ByteMultiSig = 0x17
)

const (
	// for utils...
	TypeRaw   = "raw"
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
		return errors.ErrTooLarge()
	}
	return nil
}

func NewRaw(d []byte) basecoin.Tx {
	return Raw{data.Bytes(d)}.Wrap()
}

/**** MultiTx  ******/
type MultiTx struct {
	Txs []basecoin.Tx `json:"txs"`
}

func NewMultiTx(txs ...basecoin.Tx) basecoin.Tx {
	return (&MultiTx{Txs: txs}).Wrap()
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

func NewChain(chainID string, tx basecoin.Tx) basecoin.Tx {
	return (&Chain{Tx: tx, ChainID: chainID}).Wrap()
}

func (c *Chain) Wrap() basecoin.Tx {
	return basecoin.Tx{c}
}

func (c *Chain) ValidateBasic() error {
	// TODO: more checks? chainID?
	return c.Tx.ValidateBasic()
}
