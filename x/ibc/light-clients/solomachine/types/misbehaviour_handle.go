package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// CheckMisbehaviourAndUpdateState determines whether or not the currently registered
// public key signed over two different messages with the same sequence. If this is true
// the client state is updated to a frozen status.
func (cs ClientState) CheckMisbehaviourAndUpdateState(
	ctx sdk.Context,
	cdc codec.BinaryMarshaler,
	clientStore sdk.KVStore,
	misbehaviour exported.Misbehaviour,
) (exported.ClientState, error) {

	soloMisbehaviour, ok := misbehaviour.(*Misbehaviour)
	if !ok {
		return nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidClientType,
			"misbehaviour type %T, expected %T", misbehaviour, &Misbehaviour{},
		)
	}

	if cs.IsFrozen() {
		return nil, sdkerrors.Wrapf(clienttypes.ErrClientFrozen, "client is already frozen")
	}

	if err := checkMisbehaviour(cdc, cs, soloMisbehaviour); err != nil {
		return nil, err
	}

	cs.FrozenSequence = soloMisbehaviour.Sequence
	return cs, nil
}

// checkMisbehaviour checks if the currently registered public key has signed
// over two different messages at the same sequence.
//
// NOTE: a check that the misbehaviour message data are not equal is done by
// misbehaviour.ValidateBasic which is called by the 02-client keeper.
func checkMisbehaviour(cdc codec.BinaryMarshaler, clientState ClientState, soloMisbehaviour *Misbehaviour) error {
	pubKey := clientState.ConsensusState.GetPubKey()

	data, err := MisbehaviourSignBytes(
		cdc,
		soloMisbehaviour.Sequence, clientState.ConsensusState.Timestamp,
		clientState.ConsensusState.Diversifier,
		soloMisbehaviour.SignatureOne.Data,
	)
	if err != nil {
		return err
	}

	// check first signature
	if err := VerifySignature(pubKey, data, soloMisbehaviour.SignatureOne.Signature); err != nil {
		return sdkerrors.Wrap(err, "misbehaviour signature one failed to be verified")
	}

	data, err = MisbehaviourSignBytes(
		cdc,
		soloMisbehaviour.Sequence, clientState.ConsensusState.Timestamp,
		clientState.ConsensusState.Diversifier,
		soloMisbehaviour.SignatureTwo.Data,
	)
	if err != nil {
		return err
	}

	// check second signature
	if err := VerifySignature(pubKey, data, soloMisbehaviour.SignatureTwo.Signature); err != nil {
		return sdkerrors.Wrap(err, "misbehaviour signature two failed to be verified")
	}

	return nil
}
