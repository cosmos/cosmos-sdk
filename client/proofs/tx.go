package proofs

import (
	"github.com/pkg/errors"

	wire "github.com/tendermint/go-wire"

	"github.com/tendermint/tendermint/rpc/client"
	"github.com/tendermint/tendermint/types"

	"github.com/tendermint/tendermint/certifiers"
	certerr "github.com/tendermint/tendermint/certifiers/errors"
)

var _ Prover = TxProver{}
var _ Proof = TxProof{}

// we store up to 10MB as a proof, as we need an entire block! right now
const txLimit = 10 * 1000 * 1000

// TxProver provides positive proofs of key-value pairs in the abciapp.
//
// TODO: also support negative proofs (this key is not set)
type TxProver struct {
	node client.Client
}

func NewTxProver(node client.Client) TxProver {
	return TxProver{node: node}
}

// Get tries to download a merkle hash for app state on this key from
// the tendermint node.
//
// Important: key must be Tx.Hash()
// Height is completely ignored for now :(
func (t TxProver) Get(key []byte, h uint64) (Proof, error) {
	res, err := t.node.Tx(key, true)
	if err != nil {
		return nil, err
	}

	// and build a proof for lighter storage
	proof := TxProof{
		Height: uint64(res.Height),
		Proof:  res.Proof,
	}
	return proof, err
}

func (t TxProver) Unmarshal(data []byte) (pr Proof, err error) {
	var proof TxProof
	err = errors.WithStack(wire.ReadBinaryBytes(data, &proof))
	return proof, err
}

// TxProof checks ALL txs for one block... we need a better way!
type TxProof struct {
	Height uint64
	Proof  types.TxProof
}

func (p TxProof) Data() []byte {
	return p.Proof.Data
}

func (p TxProof) BlockHeight() uint64 {
	return p.Height
}

func (p TxProof) Validate(check certifiers.Commit) error {
	if uint64(check.Height()) != p.Height {
		return certerr.ErrHeightMismatch(int(p.Height), check.Height())
	}
	return p.Proof.Validate(check.Header.DataHash)
}

func (p TxProof) Marshal() ([]byte, error) {
	data := wire.BinaryBytes(p)
	return data, nil
}
