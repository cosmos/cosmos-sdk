package lcd

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/types"

	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
)

// Certifier checks the votes to make sure the block really is signed properly.
// Certifier must know the current set of validitors by some other means.
type Certifier interface {
	Certify(check Commit) error
	ChainID() string
}

// Commit is basically the rpc /commit response, but extended
//
// This is the basepoint for proving anything on the blockchain. It contains
// a signed header.  If the signatures are valid and > 2/3 of the known set,
// we can store this checkpoint and use it to prove any number of aspects of
// the system: such as txs, abci state, validator sets, etc...
type Commit types.SignedHeader

// FullCommit is a commit and the actual validator set,
// the base info you need to update to a given point,
// assuming knowledge of some previous validator set
type FullCommit struct {
	Commit     `json:"commit"`
	Validators *types.ValidatorSet `json:"validator_set"`
}

// NewFullCommit returns a new FullCommit.
func NewFullCommit(commit Commit, vals *types.ValidatorSet) FullCommit {
	return FullCommit{
		Commit:     commit,
		Validators: vals,
	}
}

// Height returns the height of the header.
func (c Commit) Height() int64 {
	if c.Header == nil {
		return 0
	}
	return c.Header.Height
}

// ValidatorsHash returns the hash of the validator set.
func (c Commit) ValidatorsHash() []byte {
	if c.Header == nil {
		return nil
	}
	return c.Header.ValidatorsHash
}

// ValidateBasic does basic consistency checks and makes sure the headers
// and commits are all consistent and refer to our chain.
//
// Make sure to use a Verifier to validate the signatures actually provide
// a significantly strong proof for this header's validity.
func (c Commit) ValidateBasic(chainID string) error {
	// make sure the header is reasonable
	if c.Header == nil {
		return errors.New("Commit missing header")
	}
	if c.Header.ChainID != chainID {
		return errors.Errorf("Header belongs to another chain '%s' not '%s'",
			c.Header.ChainID, chainID)
	}

	if c.Commit == nil {
		return errors.New("Commit missing signatures")
	}

	// make sure the header and commit match (height and hash)
	if c.Commit.Height() != c.Header.Height {
		return lcdErr.ErrHeightMismatch(c.Commit.Height(), c.Header.Height)
	}
	hhash := c.Header.Hash()
	chash := c.Commit.BlockID.Hash
	if !bytes.Equal(hhash, chash) {
		return errors.Errorf("Commits sign block %X header is block %X",
			chash, hhash)
	}

	// make sure the commit is reasonable
	err := c.Commit.ValidateBasic()
	if err != nil {
		return errors.WithStack(err)
	}

	// looks good, we just need to make sure the signatures are really from
	// empowered validators
	return nil
}
