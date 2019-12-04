package tendermint

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/tmhash"
	cmn "github.com/tendermint/tendermint/libs/common"
	tmtypes "github.com/tendermint/tendermint/types"

	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types/errors"
)

var _ evidenceexported.Evidence = Evidence{}

// Evidence is a wrapper over tendermint's DuplicateVoteEvidence
// that implements Evidence interface expected by ICS-02
type Evidence struct {
	Header1 Header `json:"header1" yaml:"header1"`
	Header2 Header `json:"header2" yaml:"header2"`
	ChainID string `json:"chain_id" yaml:"chain_id"`
}

// Route implements Evidence interface
func (ev Evidence) Route() string {
	return "client"
}

// Type implements Evidence interface
func (ev Evidence) Type() string {
	return "client_misbehaviour"
}

// String implements Evidence interface
func (ev Evidence) String() string {
	bz, err := yaml.Marshal(ev)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// Hash implements Evidence interface
func (ev Evidence) Hash() cmn.HexBytes {
	return tmhash.Sum(SubModuleCdc.MustMarshalBinaryBare(ev))
}

// GetHeight returns the height at which misbehaviour occurred
func (ev Evidence) GetHeight() int64 {
	return ev.Header1.Height
}

// ValidateBasic implements Evidence interface
func (ev Evidence) ValidateBasic() error {
	// ValidateBasic on both validators
	if err := ev.Header1.ValidateBasic(ev.ChainID); err != nil {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, fmt.Sprintf("Header1 failed ValidateBasic: %v", err))
	}
	if err := ev.Header2.ValidateBasic(ev.ChainID); err != nil {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, fmt.Sprintf("Header2 failed ValidateBasic: %v", err))
	}
	// Ensure that Heights are the same
	if ev.Header1.Height != ev.Header2.Height {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "headers in evidence are on different heights")
	}
	// Ensure that Commit Hashes are different
	if ev.Header1.Commit.BlockID.Equals(ev.Header2.Commit.BlockID) {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "Headers commit to same blockID")
	}

	if err1 := ValidCommit(ev.ChainID, ev.Header1.Commit, ev.Header1.ValidatorSet); err1 != nil {
		return err1
	}
	if err2 := ValidCommit(ev.ChainID, ev.Header2.Commit, ev.Header2.ValidatorSet); err2 != nil {
		return err2
	}

	return nil
}

// ValidCommit checks if the given commit is a valid commit from the passed-in validatorset
//
// CommitToVoteSet will panic if the commit cannot be converted to a valid voteset given the validatorset
// This implies that someone tried to submit evidence that wasn't actually committed by the validatorset
// thus we should return an error here and reject the evidence rather than panicing.
func ValidCommit(chainID string, commit *tmtypes.Commit, valSet *tmtypes.ValidatorSet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.ErrInvalidEvidence(errors.DefaultCodespace, fmt.Sprintf("invalid commit: %v", r))
		}
	}()

	// Convert commits to vote-sets given the validator set so we can check if they both have 2/3 power
	voteSet := tmtypes.CommitToVoteSet(chainID, commit, valSet)

	blockID, ok := voteSet.TwoThirdsMajority()

	// Check that ValidatorSet did indeed commit to blockID in Commit
	if !ok || !blockID.Equals(commit.BlockID) {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "ValidatorSet did not commit to Header1")
	}

	return nil
}
