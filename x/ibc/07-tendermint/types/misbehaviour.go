package types

import (
	"time"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	host "github.com/cosmos/cosmos-sdk/x/ibc/24-host"
)

// CheckMisbehaviourAndUpdateState determines whether or not two conflicting
// headers at the same height would have convinced the light client.
//
// NOTE: consensusState1 is the trusted consensus state that corresponds to the TrustedHeight
// of misbehaviour.Header1
// Similarly, consensusState2 is the trusted consensus state that corresponds
// to misbehaviour.Header2
func (cs *ClientState) CheckMisbehaviourAndUpdateState(
	ctx sdk.Context,
	cdc codec.BinaryMarshaler,
	clientStore sdk.KVStore,
	misbehaviour clientexported.Misbehaviour,
) (clientexported.ClientState, error) {

	// If client is already frozen at earlier height than evidence, return with error
	if cs.IsFrozen() && cs.FrozenHeight <= uint64(misbehaviour.GetHeight()) {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence,
			"client is already frozen at earlier height %d than misbehaviour height %d", cs.FrozenHeight, misbehaviour.GetHeight())
	}

	tmEvidence, ok := misbehaviour.(Evidence)
	if !ok {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "expected type %T, got %T", misbehaviour, Evidence{})
	}

	// Retrieve trusted consensus states for each Header in misbehaviour
	// and unmarshal from clientStore
	consBytes1 := clientStore.Get(host.KeyConsensusState(tmEvidence.Header1.TrustedHeight))
	consensusState1 := clienttypes.MustUnmarshalConsensusState(cdc, consBytes1)
	tmConsensusState1, ok := consensusState1.(*ConsensusState)
	if !ok {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "invalid consensus state type for first header: expected type %T, got %T", &ConsensusState{}, consensusState1)
	}
	consBytes2 := clientStore.Get(host.KeyConsensusState(tmEvidence.Header2.TrustedHeight))
	consensusState2 := clienttypes.MustUnmarshalConsensusState(cdc, consBytes2)
	tmConsensusState2, ok := consensusState2.(*ConsensusState)
	if !ok {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "invalid consensus state for second header: expected type %T, got %T", &ConsensusState{}, consensusState2)
	}

	// use earliest height of trusted consensus states to verify ageBlocks
	var height uint64
	if tmConsensusState1.Height < tmConsensusState2.Height {
		height = tmConsensusState1.Height
	} else {
		height = tmConsensusState2.Height
	}

	// calculate the age of the misbehaviour evidence
	infractionHeight := tmEvidence.GetHeight()
	infractionTime := tmEvidence.GetTime()
	ageDuration := ctx.BlockTime().Sub(infractionTime)
	ageBlocks := uint64(infractionHeight) - height

	consensusParams := ctx.ConsensusParams()

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
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence,
			"age duration (%s) and age blocks (%d) are greater than max consensus params for duration (%s) and block (%d)",
			ageDuration, ageBlocks, consensusParams.Evidence.MaxAgeDuration, consensusParams.Evidence.MaxAgeNumBlocks,
		)
	}

	// Check the validity of the two conflicting headers against their respective
	// trusted consensus states
	// NOTE: header height and commitment root assertions are checked in
	// evidence.ValidateBasic by the client keeper and msg.ValidateBasic
	// by the base application.
	if err := checkMisbehaviourHeader(
		cs, tmConsensusState1, tmEvidence.Header1, ctx.BlockTime(),
	); err != nil {
		return nil, sdkerrors.Wrap(err, "verifying Header1 in Evidence failed")
	}
	if err := checkMisbehaviourHeader(
		cs, tmConsensusState2, tmEvidence.Header2, ctx.BlockTime(),
	); err != nil {
		return nil, sdkerrors.Wrap(err, "verifying Header2 in Evidence failed")
	}

	cs.FrozenHeight = uint64(tmEvidence.GetHeight())
	return cs, nil
}

// checkMisbehaviourHeader checks that a Header in Misbehaviour is valid evidence given
// a trusted ConsensusState
func checkMisbehaviourHeader(
	clientState *ClientState, consState *ConsensusState, header Header, currentTimestamp time.Time,
) error {
	// check the trusted fields for the header against ConsensusState
	if err := checkTrustedHeader(header, consState); err != nil {
		return err
	}

	// assert that the timestamp is not from more than an unbonding period ago
	if currentTimestamp.Sub(consState.Timestamp) >= clientState.UnbondingPeriod {
		return sdkerrors.Wrapf(
			ErrUnbondingPeriodExpired,
			"current timestamp minus the latest consensus state timestamp is greater than or equal to the unbonding period (%s >= %s)",
			currentTimestamp.Sub(consState.Timestamp), clientState.UnbondingPeriod,
		)
	}

	// - ValidatorSet must have 2/3 similarity with trusted FromValidatorSet
	// - ValidatorSets on both headers are valid given the last trusted ValidatorSet
	if err := header.TrustedValidators.VerifyCommitLightTrusting(
		clientState.GetChainID(), header.Commit.BlockID, header.Height,
		header.Commit, clientState.TrustLevel.ToTendermint(),
	); err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "validator set in header has too much change from trusted validator set: %v", err)
	}
	return nil
}
