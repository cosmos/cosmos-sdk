package types

import (
	"math"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evidenceexported "github.com/cosmos/cosmos-sdk/x/evidence/exported"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

var (
	_ evidenceexported.Evidence   = Evidence{}
	_ clientexported.Misbehaviour = Evidence{}
)

// Evidence is a wrapper over two conflicting headers
// that implements Evidence interface expected by ICS-02
type Evidence struct {
	ClientID string `json:"client_id" yaml:"client_id"`
	Header1  Header `json:"header1" yaml:"header1"`
	Header2  Header `json:"header2" yaml:"header2"`
	ChainID  string `json:"chain_id" yaml:"chain_id"`
}

// ClientType is Tendermint light client
func (ev Evidence) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (ev Evidence) GetClientID() string {
	return ev.ClientID
}

// Route implements Evidence interface
func (ev Evidence) Route() string {
	return clienttypes.SubModuleName
}

// Type implements Evidence interface
func (ev Evidence) Type() string {
	return "client_misbehaviour"
}

// String implements Evidence interface
func (ev Evidence) String() string {
	// FIXME: implement custom marshaller
	bz, err := yaml.Marshal(ev)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// Hash implements Evidence interface
func (ev Evidence) Hash() tmbytes.HexBytes {
	bz := SubModuleCdc.MustMarshalBinaryBare(ev)
	return tmhash.Sum(bz)
}

// GetHeight returns the height at which misbehaviour occurred
//
// NOTE: assumes that evidence headers have the same height
func (ev Evidence) GetHeight() int64 {
	return int64(math.Min(float64(ev.Header1.Height.EpochHeight), float64(ev.Header2.Height.EpochHeight)))
}

// GetIBCHeight returns the Height at which misbehaviour occurred
//
// NOTE: evidence headers must have same height
func (ev Evidence) GetIBCHeight() clientexported.Height {
	return ev.Header1.Height
}

// GetTime returns the timestamp at which misbehaviour occurred. It uses the
// maximum value from both headers to prevent producing an invalid header outside
// of the evidence age range.
func (ev Evidence) GetTime() time.Time {
	minTime := int64(math.Max(float64(ev.Header1.Time.UnixNano()), float64(ev.Header2.Time.UnixNano())))
	return time.Unix(0, minTime)
}

// ValidateBasic implements Evidence interface
func (ev Evidence) ValidateBasic() error {
	if err := host.ClientIdentifierValidator(ev.ClientID); err != nil {
		return sdkerrors.Wrap(err, "evidence client ID is invalid")
	}

	// ValidateBasic on both validators
	if err := ev.Header1.ValidateBasic(ev.ChainID); err != nil {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidEvidence,
			sdkerrors.Wrap(err, "header 1 failed validation").Error(),
		)
	}
	if err := ev.Header2.ValidateBasic(ev.ChainID); err != nil {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidEvidence,
			sdkerrors.Wrap(err, "header 2 failed validation").Error(),
		)
	}
	// Ensure that Heights are the same
	if cmp := ev.Header1.Height.Compare(ev.Header2.Height); cmp != 0 {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "headers in evidence are on different heights (%v ≠ %v)", ev.Header1.Height, ev.Header2.Height)
	}
	// Ensure that Commit Hashes are different
	if ev.Header1.Commit.BlockID.Equals(ev.Header2.Commit.BlockID) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "headers commit to same blockID")
	}
	if err := ValidCommit(ev.ChainID, ev.Header1.Commit, ev.Header1.ValidatorSet); err != nil {
		return err
	}
	if err := ValidCommit(ev.ChainID, ev.Header2.Commit, ev.Header2.ValidatorSet); err != nil {
		return err
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
			err = sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "invalid commit: %v", r)
		}
	}()

	// Convert commits to vote-sets given the validator set so we can check if they both have 2/3 power
	voteSet := tmtypes.CommitToVoteSet(chainID, commit, valSet)

	blockID, ok := voteSet.TwoThirdsMajority()

	// Check that ValidatorSet did indeed commit to blockID in Commit
	if !ok || !blockID.Equals(commit.BlockID) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "validator set did not commit to header 1")
	}

	return nil
}
