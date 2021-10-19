package smt

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"hash"

	"github.com/cosmos/cosmos-sdk/store/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/lazyledger/smt"
	"github.com/tendermint/tendermint/crypto/merkle"
	tmmerkle "github.com/tendermint/tendermint/proto/tendermint/crypto"
)

type HasherType byte

const (
	SHA256 HasherType = iota
)

const (
	ProofType = "smt"
)

type ProofOp struct {
	Root   []byte
	Key    []byte
	Hasher HasherType
	Proof  smt.SparseMerkleProof
}

var _ merkle.ProofOperator = (*ProofOp)(nil)

// NewProofOp returns a ProofOp for a SparseMerkleProof.
func NewProofOp(root, key []byte, hasher HasherType, proof smt.SparseMerkleProof) *ProofOp {
	return &ProofOp{
		Root:   root,
		Key:    key,
		Hasher: hasher,
		Proof:  proof,
	}
}

func (p *ProofOp) Run(args [][]byte) ([][]byte, error) {
	switch len(args) {
	case 0: // non-membership proof
		if !smt.VerifyProof(p.Proof, p.Root, p.Key, []byte{}, getHasher(p.Hasher)) {
			return nil, sdkerrors.Wrapf(types.ErrInvalidProof, "proof did not verify absence of key: %s", p.Key)
		}
	case 1: // membership proof
		if !smt.VerifyProof(p.Proof, p.Root, p.Key, args[0], getHasher(p.Hasher)) {
			return nil, sdkerrors.Wrapf(types.ErrInvalidProof, "proof did not verify existence of key %s with given value %x", p.Key, args[0])
		}
	default:
		return nil, sdkerrors.Wrapf(types.ErrInvalidProof, "args must be length 0 or 1, got: %d", len(args))
	}
	return [][]byte{p.Root}, nil
}

func (p *ProofOp) GetKey() []byte {
	return p.Key
}

func (p *ProofOp) ProofOp() tmmerkle.ProofOp {
	var data bytes.Buffer
	enc := gob.NewEncoder(&data)
	enc.Encode(p)
	return tmmerkle.ProofOp{
		Type: "smt",
		Key:  p.Key,
		Data: data.Bytes(),
	}
}

func ProofDecoder(pop tmmerkle.ProofOp) (merkle.ProofOperator, error) {
	dec := gob.NewDecoder(bytes.NewBuffer(pop.Data))
	var proof ProofOp
	err := dec.Decode(&proof)
	if err != nil {
		return nil, err
	}
	return &proof, nil
}

func getHasher(hasher HasherType) hash.Hash {
	switch hasher {
	case SHA256:
		return sha256.New()
	default:
		return nil
	}
}
