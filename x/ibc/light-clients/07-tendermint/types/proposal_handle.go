package types

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// CheckProposedHeaderAndUpdateState will try to update the client with the state of the
// substitute if and only if the proposal passes and one of the following conditions  are
// satisfied:
//		1) The substitute client is the same type as the subject client
//		2) The subject and substitute client states match in all parameters (expect frozen height, latest height, and chain-id)
// 		3) AllowUpdateAfterMisbehaviour and IsFrozen() = true
// 		4) AllowUpdateAfterExpiry=true and Expire(ctx.BlockTime) = true
// In case 3) before updating the client, the client will be unfrozen by resetting
// the FrozenHeight to the zero Height. If a client is frozen and AllowUpdateAfterMisbehaviour
// is set to true, the client will be unexpired even if AllowUpdateAfterExpiry is set to false.
// Note, that even if the subject is updated to the state of the substitute, an error may be
// returned if the updated client state is invalid or the client is expired.
func (cs ClientState) CheckSubstituteAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, subjectClientStore,
	substituteClientStore sdk.KVStore, substituteClient exported.ClientState,
	initialHeight exported.Height,
) (exported.ClientState, error) {
	substituteClientState, ok := substituteClient.(*ClientState)
	if !ok {
		return nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidClient, "expected type %T, got %T", &ClientState{}, substituteClient,
		)
	}

	if !IsMatchingClientState(cs, *substituteClientState) {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidSubstitute, "subject client state does not match substitute client state")
	}

	// get consensus state corresponding to client state to check if the client is expired
	consensusState, err := GetConsensusState(subjectClientStore, cdc, cs.GetLatestHeight())
	if err != nil {
		return nil, sdkerrors.Wrapf(
			err, "unexpected error: could not get consensus state from clientstore at height: %d", cs.GetLatestHeight(),
		)
	}

	switch {

	case cs.IsFrozen():
		if !cs.AllowUpdateAfterMisbehaviour {
			return nil, sdkerrors.Wrap(clienttypes.ErrUpdateClientFailed, "client is not allowed to be unfrozen")
		}

		// unfreeze the client
		cs.FrozenHeight = clienttypes.ZeroHeight()

	case cs.IsExpired(consensusState.Timestamp, ctx.BlockTime()):
		if !cs.AllowUpdateAfterExpiry {
			return nil, sdkerrors.Wrap(clienttypes.ErrUpdateClientFailed, "client is not allowed to be unexpired")
		}

	default:
		return nil, sdkerrors.Wrap(clienttypes.ErrUpdateClientFailed, "client cannot be updated with proposal")
	}

	// copy consensus states and processed time from substitute to subject
	// starting from initial height and ending on the latest height (inclusive)
	// CONTRACT: the revision number is same for substitute and subject
	// as cheked in 02-client.
	for i := initialHeight.GetRevisionHeight(); i <= substituteClientState.GetLatestHeight().GetRevisionHeight(); i++ {
		height := clienttypes.NewHeight(substituteClientState.GetLatestHeight().GetRevisionNumber(), i)

		consensusState, err := GetConsensusState(substituteClientStore, cdc, height)
		if err != nil {
			// not all consensus states will be filled in
			continue
		}
		SetConsensusState(subjectClientStore, cdc, consensusState, height)

		processedTime, found := GetProcessedTime(substituteClientStore, height)
		if !found {
			continue
		}
		SetProcessedTime(subjectClientStore, height, processedTime)

	}

	cs.LatestHeight = substituteClientState.LatestHeight

	// validate the updated client and ensure it isn't expired
	if err := cs.Validate(); err != nil {
		return nil, sdkerrors.Wrap(err, "unexpected error: updated subject client state is invalid")
	}

	latestConsensusState, err := GetConsensusState(subjectClientStore, cdc, cs.GetLatestHeight())
	if err != nil {
		return nil, sdkerrors.Wrapf(
			err, "unexpected error: could not get consensus state for updated subject client from clientstore at height: %d", cs.GetLatestHeight(),
		)
	}

	if cs.IsExpired(latestConsensusState.Timestamp, ctx.BlockTime()) {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClient, "updated subject client is expired")
	}

	return &cs, nil
}

// IsMatchingClientState returns true if all the client state parameters match
// except for frozen height, latest height, and chain-id.
func IsMatchingClientState(subject, substitute ClientState) bool {
	// zero out parameters which do not need to match
	subject.LatestHeight = clienttypes.ZeroHeight()
	subject.FrozenHeight = clienttypes.ZeroHeight()
	substitute.LatestHeight = clienttypes.ZeroHeight()
	substitute.FrozenHeight = clienttypes.ZeroHeight()
	subject.ChainId = ""
	substitute.ChainId = ""

	return reflect.DeepEqual(subject, substitute)
}
