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
	_ evidenceexported.Evidence   = Misbehaviour{}
	_ clientexported.Misbehaviour = Misbehaviour{}
)

// NewMisbehaviour creates a new Misbehaviour instance.
func NewMisbehaviour(clientID, chainID string, header1, header2 *Header) *Misbehaviour {
	return &Misbehaviour{
		ClientId: clientID,
		ChainId:  chainID,
		Header1:  header1,
		Header2:  header2,
	}

}

// ClientType is Tendermint light client
func (misbehaviour Misbehaviour) ClientType() clientexported.ClientType {
	return clientexported.Tendermint
}

// GetClientID returns the ID of the client that committed a misbehaviour.
func (misbehaviour Misbehaviour) GetClientID() string {
	return misbehaviour.ClientId
}

// Route implements Misbehaviour interface
func (misbehaviour Misbehaviour) Route() string {
	return clienttypes.SubModuleName
}

// Type implements Misbehaviour interface
func (misbehaviour Misbehaviour) Type() string {
	return clientexported.TypeEvidenceClientMisbehaviour
}

// String implements Misbehaviour interface
func (misbehaviour Misbehaviour) String() string {
	// FIXME: implement custom marshaller
	bz, err := yaml.Marshal(misbehaviour)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// Hash implements Misbehaviour interface
func (misbehaviour Misbehaviour) Hash() tmbytes.HexBytes {
	bz := SubModuleCdc.MustMarshalBinaryBare(&misbehaviour)
	return tmhash.Sum(bz)
}

// GetHeight returns the height at which misbehaviour occurred
//
// NOTE: assumes that misbehaviour headers have the same height
func (misbehaviour Misbehaviour) GetHeight() int64 {
	return int64(math.Min(float64(misbehaviour.Header1.GetHeight()), float64(misbehaviour.Header2.GetHeight())))
}

// GetTime returns the timestamp at which misbehaviour occurred. It uses the
// maximum value from both headers to prevent producing an invalid header outside
// of the misbehaviour age range.
func (misbehaviour Misbehaviour) GetTime() time.Time {
	minTime := int64(math.Max(float64(misbehaviour.Header1.GetTime().UnixNano()), float64(misbehaviour.Header2.GetTime().UnixNano())))
	return time.Unix(0, minTime)
}

// ValidateBasic implements Misbehaviour interface
func (misbehaviour Misbehaviour) ValidateBasic() error {
	if misbehaviour.Header1 == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "misbehaviour Header1 cannot be nil")
	}
	if misbehaviour.Header2 == nil {
		return sdkerrors.Wrap(ErrInvalidHeader, "misbehaviour Header2 cannot be nil")
	}
	if !misbehaviour.Header1.TrustedHeight.IsValid() {
		return sdkerrors.Wrapf(ErrInvalidHeaderHeight, "misbehaviour Header1 has invalid trusted height: %v", misbehaviour.Header1.TrustedHeight)
	}
	if !misbehaviour.Header2.TrustedHeight.IsValid() {
		return sdkerrors.Wrapf(ErrInvalidHeaderHeight, "misbehaviour Header2 has invalid trusted height:: %v", misbehaviour.Header2.TrustedHeight)
	}
	if misbehaviour.Header1.TrustedValidators == nil {
		return sdkerrors.Wrap(ErrInvalidValidatorSet, "trusted validator set in Header1 cannot be empty")
	}
	if misbehaviour.Header2.TrustedValidators == nil {
		return sdkerrors.Wrap(ErrInvalidValidatorSet, "trusted validator set in Header2 cannot be empty")
	}

	if err := host.ClientIdentifierValidator(misbehaviour.ClientId); err != nil {
		return sdkerrors.Wrap(err, "misbehaviour client ID is invalid")
	}

	// ValidateBasic on both validators
	if err := misbehaviour.Header1.ValidateBasic(misbehaviour.ChainId); err != nil {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidMisbehaviour,
			sdkerrors.Wrap(err, "header 1 failed validation").Error(),
		)
	}
	if err := misbehaviour.Header2.ValidateBasic(misbehaviour.ChainId); err != nil {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidMisbehaviour,
			sdkerrors.Wrap(err, "header 2 failed validation").Error(),
		)
	}
	// Ensure that Heights are the same
	if misbehaviour.Header1.GetHeight() != misbehaviour.Header2.GetHeight() {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidMisbehaviour, "headers in misbehaviour are on different heights (%d ≠ %d)", misbehaviour.Header1.GetHeight(), misbehaviour.Header2.GetHeight())
	}

	blockID1, err := tmtypes.BlockIDFromProto(&misbehaviour.Header1.SignedHeader.Commit.BlockID)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid block ID from header 1 in misbehaviour")
	}
	blockID2, err := tmtypes.BlockIDFromProto(&misbehaviour.Header2.SignedHeader.Commit.BlockID)
	if err != nil {
		return sdkerrors.Wrap(err, "invalid block ID from header 2 in misbehaviour")
	}

	// Ensure that Commit Hashes are different
	if blockID1.Equals(*blockID2) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidMisbehaviour, "headers blockIDs are not equal")
	}
	if err := ValidCommit(misbehaviour.ChainId, misbehaviour.Header1.Commit, misbehaviour.Header1.ValidatorSet); err != nil {
		return err
	}
	if err := ValidCommit(misbehaviour.ChainId, misbehaviour.Header2.Commit, misbehaviour.Header2.ValidatorSet); err != nil {
		return err
	}
	return nil
}

// ValidCommit checks if the given commit is a valid commit from the passed-in validatorset
//
// CommitToVoteSet will panic if the commit cannot be converted to a valid voteset given the validatorset
// This implies that someone tried to submit misbehaviour that wasn't actually committed by the validatorset
// thus we should return an error here and reject the misbehaviour rather than panicing.
func ValidCommit(chainID string, commit *tmproto.Commit, valSet *tmproto.ValidatorSet) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = sdkerrors.Wrapf(clienttypes.ErrInvalidMisbehaviour, "invalid commit: %v", r)
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
		return sdkerrors.Wrap(clienttypes.ErrInvalidMisbehaviour, "validator set did not commit to header")
	}

	return nil
}
