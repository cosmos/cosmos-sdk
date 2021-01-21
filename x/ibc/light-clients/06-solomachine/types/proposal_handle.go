package types

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// CheckSubstituteAndUpdateState verifies that the subject is allowed to be updated by
// a governance proposal and that the substitute client is a solo machine.
// updates the consensus state to the header's sequence and
// public key. An error is returned if the client has been disallowed to be updated by a
// governance proposal, the header cannot be casted to a solo machine header, or the current
// public key equals the new public key.
func (cs ClientState) CheckSubstituteAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, subjectClientStore,
	_ sdk.KVStore, substituteClient exported.ClientState,
	inittialHeight exported.Height,
) (exported.ClientState, error) {

	if !cs.AllowUpdateAfterProposal {
		return nil, sdkerrors.Wrapf(
			clienttypes.ErrUpdateClientFailed,
			"solo machine client is not allowed to updated with a proposal",
		)
	}

	substituteClientState, ok := substituteClient.(*ClientState)
	if !ok {
		return nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidClientType, "client state type %T, expected  %T", substituteClient, &ClientState{},
		)
	}

	consensusPublicKey, err := cs.ConsensusState.GetPubKey()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get consensus public key")
	}

	substitutePublicKey, err := substituteClientState.ConsensusState.GetPubKey()
	if err != nil {
		return nil, sdkerrors.Wrap(err, "failed to get substitute client public key")
	}

	if reflect.DeepEqual(consensusPublicKey, substitutePublicKey) {
		return nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "new public key in substitute equals current public key",
		)
	}

	clientState := &cs

	// update to substitute parameters
	clientState.Sequence = substituteClientState.Sequence
	clientState.ConsensusState = substituteClientState.ConsensusState
	clientState.FrozenSequence = 0

	return clientState, nil
}
