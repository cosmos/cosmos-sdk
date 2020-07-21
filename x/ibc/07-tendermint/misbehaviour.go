package tendermint

import (
	"time"

	abci "github.com/tendermint/tendermint/abci/types"
	tmtypes "github.com/tendermint/tendermint/types"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
)

// CheckMisbehaviourAndUpdateState determines whether or not two conflicting
// headers at the same height would have convinced the light client.
//
// NOTE: assumes provided height is the height at which the consensusState is
// stored.
func CheckMisbehaviourAndUpdateState(
	clientState clientexported.ClientState,
	consensusState clientexported.ConsensusState,
	misbehaviour clientexported.Misbehaviour,
	height uint64, // height at which the consensus state was loaded
	currentTimestamp time.Time,
	consensusParams *abci.ConsensusParams,
) (clientexported.ClientState, error) {

	// cast the interface to specific types before checking for misbehaviour
	tmClientState, ok := clientState.(types.ClientState)
	if !ok {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "expected type %T, got %T", types.ClientState{}, clientState)
	}

	// If client is already frozen at earlier height than evidence, return with error
	if tmClientState.IsFrozen() && tmClientState.FrozenHeight <= uint64(misbehaviour.GetHeight()) {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence,
			"client is already frozen at earlier height %d than misbehaviour height %d", tmClientState.FrozenHeight, misbehaviour.GetHeight())
	}

	tmConsensusState, ok := consensusState.(types.ConsensusState)
	if !ok {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "expected type %T, got %T", consensusState, types.ConsensusState{})
	}

	tmEvidence, ok := misbehaviour.(types.Evidence)
	if !ok {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "expected type %T, got %T", misbehaviour, types.Evidence{})
	}

	if err := checkMisbehaviour(
		tmClientState, tmConsensusState, tmEvidence, height, currentTimestamp, consensusParams,
	); err != nil {
		return nil, err
	}

	tmClientState.FrozenHeight = uint64(tmEvidence.GetHeight())
	return tmClientState, nil
}

// checkMisbehaviour checks if the evidence provided is a valid light client misbehaviour
func checkMisbehaviour(
	clientState types.ClientState, consensusState types.ConsensusState, evidence types.Evidence,
	height uint64, currentTimestamp time.Time, consensusParams *abci.ConsensusParams,
) error {
	// calculate the age of the misbehaviour evidence
	infractionHeight := evidence.GetHeight()
	infractionTime := evidence.GetTime()
	ageDuration := currentTimestamp.Sub(infractionTime)
	ageBlocks := height - uint64(infractionHeight)

	// Reject misbehaviour if the age is too old. Evidence is considered stale
	// if the difference in time and number of blocks is greater than the allowed
	// parameters defined.
	//
	// NOTE: The first condition is a safety check as the consensus params cannot
	// be nil since the previous param values will be used in case they can't be
	// retrieved. If they are not set during initialization, Tendermint will always
	// use the default values.
	if consensusParams != nil &&
		consensusParams.Evidence != nil &&
		ageDuration > consensusParams.Evidence.MaxAgeDuration &&
		ageBlocks > uint64(consensusParams.Evidence.MaxAgeNumBlocks) {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence,
			"age duration (%s) and age blocks (%d) are greater than max consensus params for duration (%s) and block (%d)",
			ageDuration, ageBlocks, consensusParams.Evidence.MaxAgeDuration, consensusParams.Evidence.MaxAgeNumBlocks,
		)
	}

	// check if provided height matches the headers' height
	if height > uint64(evidence.GetHeight()) {
		return sdkerrors.Wrapf(
			sdkerrors.ErrInvalidHeight,
			"height > evidence header height (%d > %d)", height, evidence.GetHeight(),
		)
	}

	// NOTE: header height and commitment root assertions are checked with the
	// evidence and msg ValidateBasic functions at the AnteHandler level.

	// assert that the timestamp is not from more than an unbonding period ago
	if currentTimestamp.Sub(consensusState.Timestamp) >= clientState.UnbondingPeriod {
		return sdkerrors.Wrapf(
			types.ErrUnbondingPeriodExpired,
			"current timestamp minus the latest consensus state timestamp is greater than or equal to the unbonding period (%s >= %s)",
			currentTimestamp.Sub(consensusState.Timestamp), clientState.UnbondingPeriod,
		)
	}

	valset, err := tmtypes.ValidatorSetFromProto(consensusState.ValidatorSet)
	if err != nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	signedHeader1, err := tmtypes.SignedHeaderFromProto(&evidence.Header1.SignedHeader)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "invalid signed header 1: %s", err.Error())
	}

	signedHeader2, err := tmtypes.SignedHeaderFromProto(&evidence.Header2.SignedHeader)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "invalid signed header 2: %s", err.Error())
	}

	// TODO: Evidence must be within trusting period
	// Blocked on https://github.com/cosmos/ics/issues/379

	// - ValidatorSet must have (1-trustLevel) similarity with trusted FromValidatorSet
	// - ValidatorSets on both headers are valid given the last trusted ValidatorSet
	if err := valset.VerifyCommitLightTrusting(
		evidence.ChainID, signedHeader1.Commit.BlockID, signedHeader1.Height,
		signedHeader1.Commit, clientState.TrustLevel.ToTendermint(),
	); err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "validator set in header 1 has too much change from last known validator set: %v", err)
	}

	if err := valset.VerifyCommitLightTrusting(
		evidence.ChainID, signedHeader2.Commit.BlockID, signedHeader2.Height,
		signedHeader2.Commit, clientState.TrustLevel.ToTendermint(),
	); err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "validator set in header 2 has too much change from last known validator set: %v", err)
	}

	return nil
}
