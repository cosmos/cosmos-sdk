package tendermint

import (
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
	Header1 Header
	Header2 Header
	ChainID string
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
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, err.Error())
	}
	if err := ev.Header2.ValidateBasic(ev.ChainID); err != nil {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, err.Error())
	}
	// Ensure that Heights are the same
	if ev.Header1.Height != ev.Header2.Height {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "headers in evidence are on different heights")
	}
	// Ensure that Commit Hashes are different
	if ev.Header1.Commit.BlockID.Equals(ev.Header2.Commit.BlockID) {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "Headers commit to same blockID")
	}

	// Convert commits to vote-sets given the validator set so we can check if they both have 2/3 power
	voteSet1 := tmtypes.CommitToVoteSet(ev.ChainID, ev.Header1.Commit, ev.Header1.ValidatorSet)
	voteSet2 := tmtypes.CommitToVoteSet(ev.ChainID, ev.Header2.Commit, ev.Header2.ValidatorSet)

	blockID1, ok1 := voteSet1.TwoThirdsMajority()
	blockID2, ok2 := voteSet2.TwoThirdsMajority()

	// Check that ValidatorSet did indeed commit to two different headers
	if !ok1 || !blockID1.Equals(ev.Header1.Commit.BlockID) {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "ValidatorSet did not commit to Header1")
	}
	if !ok2 || !blockID2.Equals(ev.Header2.Commit.BlockID) {
		return errors.ErrInvalidEvidence(errors.DefaultCodespace, "ValidatorSet did not commit to Header2")
	}

	return nil
}
