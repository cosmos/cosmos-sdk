package proofs

import (
	"github.com/pkg/errors"

	wire "github.com/tendermint/go-wire"
	data "github.com/tendermint/go-wire/data"
	"github.com/tendermint/iavl"

	"github.com/tendermint/tendermint/rpc/client"

	"github.com/tendermint/tendermint/certifiers"
	certerr "github.com/tendermint/tendermint/certifiers/errors"
)

var _ Prover = AppProver{}
var _ Proof = AppProof{}

// we limit proofs to 1MB to stop overflow attacks
const appLimit = 1000 * 1000

// AppProver provides positive proofs of key-value pairs in the abciapp.
//
// TODO: also support negative proofs (this key is not set)
type AppProver struct {
	node client.Client
}

func NewAppProver(node client.Client) AppProver {
	return AppProver{node: node}
}

// Get tries to download a merkle hash for app state on this key from
// the tendermint node.
func (a AppProver) Get(key []byte, h uint64) (Proof, error) {
	resp, err := a.node.ABCIQuery("/key", key)
	if err != nil {
		return nil, err
	}

	// make sure the proof is the proper height
	if !resp.Code.IsOK() {
		return nil, errors.Errorf("Query error %d: %s", resp.Code, resp.Code.String())
	}
	if len(resp.Key) == 0 || len(resp.Value) == 0 || len(resp.Proof) == 0 {
		return nil, ErrNoData()
	}
	if resp.Height == 0 {
		resp.Height = h
	}
	if h != 0 && h != resp.Height {
		return nil, certerr.ErrHeightMismatch(int(h), int(resp.Height))
	}
	proof := AppProof{
		Height: resp.Height,
		Key:    resp.Key,
		Value:  resp.Value,
		Proof:  resp.Proof,
	}
	return proof, nil
}

func (a AppProver) Unmarshal(data []byte) (Proof, error) {
	var proof AppProof
	err := errors.WithStack(wire.ReadBinaryBytes(data, &proof))
	return proof, err
}

// AppProof containts a key-value pair at a given height.
// It also contains the merkle proof from that key-value pair to the root hash,
// which can be verified against a signed header.
type AppProof struct {
	Height uint64
	Key    data.Bytes
	Value  data.Bytes
	Proof  data.Bytes
}

func (p AppProof) Data() []byte {
	return p.Value
}

func (p AppProof) BlockHeight() uint64 {
	return p.Height
}

func (p AppProof) Validate(check certifiers.Commit) error {
	if uint64(check.Height()) != p.Height {
		return certerr.ErrHeightMismatch(int(p.Height), check.Height())
	}

	proof, err := iavl.ReadKeyExistsProof(p.Proof)
	if err != nil {
		return errors.WithStack(err)
	}

	err = proof.Verify(p.Key, p.Value, check.Header.AppHash)
	if err != nil {
		return err
	}

	// LGTM!
	return nil
}

func (p AppProof) Marshal() ([]byte, error) {
	data := wire.BinaryBytes(p)
	return data, nil
}
