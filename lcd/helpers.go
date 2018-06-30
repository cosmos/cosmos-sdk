package lcd

import (
	"time"

	crypto "github.com/tendermint/go-crypto"

	"github.com/tendermint/tendermint/types"
)

// ValKeys is a helper for testing.
//
// It lets us simulate signing with many keys, either ed25519 or secp256k1.
// The main use case is to create a set, and call GenCommit
// to get properly signed header for testing.
//
// You can set different weights of validators each time you call
// ToValidators, and can optionally extend the validator set later
// with Extend or ExtendSecp
type ValKeys []crypto.PrivKey

// GenValKeys produces an array of private keys to generate commits.
func GenValKeys(n int) ValKeys {
	res := make(ValKeys, n)
	for i := range res {
		res[i] = crypto.GenPrivKeyEd25519()
	}
	return res
}

// Change replaces the key at index i.
func (v ValKeys) Change(i int) ValKeys {
	res := make(ValKeys, len(v))
	copy(res, v)
	res[i] = crypto.GenPrivKeyEd25519()
	return res
}

// Extend adds n more keys (to remove, just take a slice).
func (v ValKeys) Extend(n int) ValKeys {
	extra := GenValKeys(n)
	return append(v, extra...)
}

// GenSecpValKeys produces an array of secp256k1 private keys to generate commits.
func GenSecpValKeys(n int) ValKeys {
	res := make(ValKeys, n)
	for i := range res {
		res[i] = crypto.GenPrivKeySecp256k1()
	}
	return res
}

// ExtendSecp adds n more secp256k1 keys (to remove, just take a slice).
func (v ValKeys) ExtendSecp(n int) ValKeys {
	extra := GenSecpValKeys(n)
	return append(v, extra...)
}

// ToValidators produces a list of validators from the set of keys
// The first key has weight `init` and it increases by `inc` every step
// so we can have all the same weight, or a simple linear distribution
// (should be enough for testing).
func (v ValKeys) ToValidators(init, inc int64) *types.ValidatorSet {
	res := make([]*types.Validator, len(v))
	for i, k := range v {
		res[i] = types.NewValidator(k.PubKey(), init+int64(i)*inc)
	}
	return types.NewValidatorSet(res)
}

// signHeader properly signs the header with all keys from first to last exclusive.
func (v ValKeys) signHeader(header *types.Header, first, last int) *types.Commit {
	votes := make([]*types.Vote, len(v))

	// we need this list to keep the ordering...
	vset := v.ToValidators(1, 0)

	// fill in the votes we want
	for i := first; i < last && i < len(v); i++ {
		vote := makeVote(header, vset, v[i])
		votes[vote.ValidatorIndex] = vote
	}

	res := &types.Commit{
		BlockID:    types.BlockID{Hash: header.Hash()},
		Precommits: votes,
	}
	return res
}

func makeVote(header *types.Header, vals *types.ValidatorSet, key crypto.PrivKey) *types.Vote {
	addr := key.PubKey().Address()
	idx, _ := vals.GetByAddress(addr)
	vote := &types.Vote{
		ValidatorAddress: addr,
		ValidatorIndex:   idx,
		Height:           header.Height,
		Round:            1,
		Timestamp:        time.Now().UTC(),
		Type:             types.VoteTypePrecommit,
		BlockID:          types.BlockID{Hash: header.Hash()},
	}
	// Sign it
	signBytes := vote.SignBytes(header.ChainID)
	vote.Signature = key.Sign(signBytes)
	return vote
}

// Silences warning that vals can also be merkle.Hashable
// nolint: interfacer
func genHeader(chainID string, height int64, txs types.Txs,
	vals *types.ValidatorSet, appHash, consHash, resHash []byte) *types.Header {

	return &types.Header{
		ChainID:  chainID,
		Height:   height,
		Time:     time.Now(),
		NumTxs:   int64(len(txs)),
		TotalTxs: int64(len(txs)),
		// LastBlockID
		// LastCommitHash
		ValidatorsHash:  vals.Hash(),
		DataHash:        txs.Hash(),
		AppHash:         appHash,
		ConsensusHash:   consHash,
		LastResultsHash: resHash,
	}
}

// GenCommit calls genHeader and signHeader and combines them into a Commit.
func (v ValKeys) GenCommit(chainID string, height int64, txs types.Txs,
	vals *types.ValidatorSet, appHash, consHash, resHash []byte, first, last int) Commit {

	header := genHeader(chainID, height, txs, vals, appHash, consHash, resHash)
	check := Commit{
		Header: header,
		Commit: v.signHeader(header, first, last),
	}
	return check
}

// GenFullCommit calls genHeader and signHeader and combines them into a Commit.
func (v ValKeys) GenFullCommit(chainID string, height int64, txs types.Txs,
	vals *types.ValidatorSet, appHash, consHash, resHash []byte, first, last int) FullCommit {

	header := genHeader(chainID, height, txs, vals, appHash, consHash, resHash)
	commit := Commit{
		Header: header,
		Commit: v.signHeader(header, first, last),
	}
	return NewFullCommit(commit, vals)
}
