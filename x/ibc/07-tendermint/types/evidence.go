package types

import (
	"math"
	"time"

	yaml "gopkg.in/yaml.v2"

	"github.com/tendermint/tendermint/crypto/tmhash"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
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

// NewEvidence creates a new Evidence instance.
func NewEvidence(clientID, chainID string, header1, header2 *Header) *Evidence {
	return &Evidence{
		ClientId: clientID,
		ChainId:  chainID,
		Header1:  header1,
		Header2:  header2,
	}

}

// ClientType is Tendermint light client
func (ev Evidence) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (ev Evidence) GetClientID() string {
	return ev.ClientId
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
	bz := SubModuleCdc.MustMarshalBinaryBare(&ev)
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
	minTime := int64(math.Max(float64(ev.Header1.GetTime().UnixNano()), float64(ev.Header2.GetTime().UnixNano())))
	return time.Unix(0, minTime)
}

// ValidateBasic implements Evidence interface
func (ev Evidence) ValidateBasic() error {
	if ev.Header1 == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "evidence Header1 cannot be nil")
	}
	if ev.Header2 == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "evidence Header2 cannot be nil")
	}
	if ev.Header1.TrustedHeight == 0 {
		return sdkerrors.Wrap(ErrInvalidHeaderHeight, "evidence Header1 must have non-zero trusted height")
	}
	if ev.Header2.TrustedHeight == 0 {
		return sdkerrors.Wrap(ErrInvalidHeaderHeight, "evidence Header2 must have non-zero trusted height")
	}
	if ev.Header1.TrustedValidators == nil {
		return sdkerrors.Wrap(ErrInvalidValidatorSet, "trusted validator set in Header1 cannot be empty")
	}
	if ev.Header2.TrustedValidators == nil {
		return sdkerrors.Wrap(ErrInvalidValidatorSet, "trusted validator set in Header2 cannot be empty")
	}

	if err := host.ClientIdentifierValidator(ev.ClientId); err != nil {
		return sdkerrors.Wrap(err, "evidence client ID is invalid")
	}

	// ValidateBasic on both validators
	if err := ev.Header1.ValidateBasic(ev.ChainId); err != nil {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidEvidence,
			sdkerrors.Wrap(err, "header 1 failed validation").Error(),
		)
	}
	if err := ev.Header2.ValidateBasic(ev.ChainId); err != nil {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidEvidence,
			sdkerrors.Wrap(err, "header 2 failed validation").Error(),
		)
	}
	// Ensure that Heights are the same
	if !ev.Header1.Height.EQ(ev.Header2.Height) {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "headers in evidence are on different heights (%v ≠ %v)", ev.Header1.Height, ev.Header2.Height)
	}

	blockID1, err := tmtypes.BlockIDFromProto(&ev.Header1.SignedHeader.Commit.BlockID)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid block ID from header 1 in evidence")
	}
	blockID2, err := tmtypes.BlockIDFromProto(&ev.Header2.SignedHeader.Commit.BlockID)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid block ID from header 2 in evidence")
	}

	// Ensure that Commit Hashes are different
	if blockID1.Equals(*blockID2) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "headers blockIDs are not equal")
	}
	if err := ValidCommit(ev.ChainId, ev.Header1.Commit, ev.Header1.ValidatorSet); err != nil {
		return err
	}
	if err := ValidCommit(ev.ChainId, ev.Header2.Commit, ev.Header2.ValidatorSet); err != nil {
		return err
	}
	return nil
}

// ValidCommit checks if the given commit is a valid commit from the passed-in validatorset
//
// CommitToVoteSet will panic if the commit cannot be converted to a valid voteset given the validatorset
// This implies that someone tried to submit evidence that wasn't actually committed by the validatorset
// thus we should return an error here and reject the evidence rather than panicing.
func ValidCommit(chainID string, commit *tmproto.Commit, valSet *tmproto.ValidatorSet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "invalid commit: %v", r)
		}
	}()

	tmCommit, err := tmtypes.CommitFromProto(commit)
	if err != nil {
		return sdkerrors.Wrap(err, "commit is not tendermint commit type")
	}
	tmValset, err := tmtypes.ValidatorSetFromProto(valSet)
	if err != nil {
		return sdkerrors.Wrap(err, "validator set is not tendermint validator set type")
	}

	// Convert commits to vote-sets given the validator set so we can check if they both have 2/3 power
	voteSet := tmtypes.CommitToVoteSet(chainID, tmCommit, tmValset)

	blockID, ok := voteSet.TwoThirdsMajority()

	// Check that ValidatorSet did indeed commit to blockID in Commit
	if !ok || !blockID.Equals(tmCommit.BlockID) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "validator set did not commit to header")
	}

	return nil
}
