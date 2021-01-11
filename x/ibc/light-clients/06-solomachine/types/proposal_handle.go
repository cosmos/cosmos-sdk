package types

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// CheckProposedHeaderAndUpdateState updates the consensus state to the header's sequence and
// public key. An error is returned if the client has been disallowed to be updated by a
// governance proposal, the header cannot be casted to a solo machine header, or the current
// public key equals the new public key.
func (cs ClientState) CheckProposedHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	header exported.Header,
) (exported.ClientState, exported.ConsensusState, error) {

	if !cs.AllowUpdateAfterProposal {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrUpdateClientFailed,
			"solo machine client is not allowed to updated with a proposal",
		)
	}

	smHeader, ok := header.(*Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "header type %T, expected  %T", header, &Header{},
		)
	}

	consensusPublicKey, err := cs.ConsensusState.GetPubKey()
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to get consensus public key")
	}

	headerPublicKey, err := smHeader.GetPubKey()
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "failed to get header public key")
	}

	if reflect.DeepEqual(consensusPublicKey, headerPublicKey) {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "new public key in header equals current public key",
		)
	}

	clientState := &cs

	consensusState := &ConsensusState{
		PublicKey:   smHeader.NewPublicKey,
		Diversifier: smHeader.NewDiversifier,
		Timestamp:   smHeader.Timestamp,
	}

	clientState.Sequence = smHeader.Sequence
	clientState.ConsensusState = consensusState
	clientState.FrozenSequence = 0

	return clientState, consensusState, nil
}
