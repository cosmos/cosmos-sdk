package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// CheckMisbehaviourAndUpdateState determines whether or not the currently registered
// public key signed over two different messages with the same sequence. If this is true
// the client state is updated to a frozen status.
func (cs ClientState) CheckMisbehaviourAndUpdateState(
	ctx sdk.Context,
	cdc codec.BinaryMarshaler,
	clientStore sdk.KVStore,
	misbehaviour clientexported.Misbehaviour,
) (clientexported.ClientState, error) {

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

	if err := checkMisbehaviour(cs, soloMisbehaviour); err != nil {
		return nil, err
	}

	cs.FrozenSequence = soloMisbehaviour.GetHeight()
	return cs, nil
}

// checkMisbehaviour checks if the currently registered public key has signed
// over two different messages at the same sequence.
// NOTE: a check that the misbehaviour message data are not equal is done by
// misbehaviour.ValidateBasic which is called by the 02-client keeper.
func checkMisbehaviour(clientState ClientState, soloMisbehaviour *Misbehaviour) error {
	pubKey := clientState.ConsensusState.GetPubKey()

	data := EvidenceSignBytes(soloMisbehaviour.Sequence, soloMisbehaviour.SignatureOne.Data)

	// check first signature
	if err := VerifySignature(pubKey, data, soloMisbehaviour.SignatureOne.Signature); err != nil {
		return sdkerrors.Wrap(err, "misbehaviour signature one failed to be verified")
	}

	data = EvidenceSignBytes(soloMisbehaviour.Sequence, soloMisbehaviour.SignatureTwo.Data)

	// check second signature
	if err := VerifySignature(pubKey, data, soloMisbehaviour.SignatureTwo.Signature); err != nil {
		return sdkerrors.Wrap(err, "misbehaviour signature two failed to be verified")
	}

	return nil
}
