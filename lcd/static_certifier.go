package lcd

import (
	"bytes"

	"github.com/pkg/errors"

	"github.com/tendermint/tendermint/types"

	lcdErr "github.com/cosmos/cosmos-sdk/lcd/errors"
)

var _ Certifier = (*StaticCertifier)(nil)

// StaticCertifier assumes a static set of validators, set on
// initilization and checks against them.
// The signatures on every header is checked for > 2/3 votes
// against the known validator set upon Certify
//
// Good for testing or really simple chains.  Building block
// to support real-world functionality.
type StaticCertifier struct {
	chainID string
	vSet    *types.ValidatorSet
	vhash   []byte
}

// NewStaticCertifier returns a new certifier with a static validator set.
func NewStaticCertifier(chainID string, vals *types.ValidatorSet) *StaticCertifier {
	return &StaticCertifier{
		chainID: chainID,
		vSet:    vals,
	}
}

// ChainID returns the chain id.
// Implements Certifier.
func (sc *StaticCertifier) ChainID() string {
	return sc.chainID
}

// Validators returns the validator set.
func (sc *StaticCertifier) Validators() *types.ValidatorSet {
	return sc.vSet
}

// Hash returns the hash of the validator set.
func (sc *StaticCertifier) Hash() []byte {
	if len(sc.vhash) == 0 {
		sc.vhash = sc.vSet.Hash()
	}
	return sc.vhash
}

// Certify makes sure that the commit is valid.
// Implements Certifier.
func (sc *StaticCertifier) Certify(commit Commit) error {
	// do basic sanity checks
	err := commit.ValidateBasic(sc.chainID)
	if err != nil {
		return err
	}

	// make sure it has the same validator set we have (static means static)
	if !bytes.Equal(sc.Hash(), commit.Header.ValidatorsHash) {
		return lcdErr.ErrValidatorsChanged()
	}

	// then make sure we have the proper signatures for this
	err = sc.vSet.VerifyCommit(sc.chainID, commit.Commit.BlockID,
		commit.Header.Height, commit.Commit)
	return errors.WithStack(err)
}
