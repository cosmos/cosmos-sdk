package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
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

	// NOTE: a check that the misbehaviour message data are not equal is done by
	// misbehaviour.ValidateBasic which is called by the 02-client keeper.

	// verify first signature
	if err := verifySignatureAndData(cdc, cs, soloMisbehaviour, soloMisbehaviour.SignatureOne); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to verify signature one")
	}

	// verify second signature
	if err := verifySignatureAndData(cdc, cs, soloMisbehaviour, soloMisbehaviour.SignatureTwo); err != nil {
		return nil, sdkerrors.Wrap(err, "failed to verify signature two")
	}

	cs.FrozenSequence = soloMisbehaviour.Sequence
	return cs, nil
}

// verifySignatureAndData verifies that the currently registered public key has signed
// over the provided data and that the data is valid. The data is valid if it can be
// unmarshaled into the specified data type or the timestamp of the signature is less
// than the consensus state timestamp.
func verifySignatureAndData(cdc codec.BinaryMarshaler, clientState ClientState, misbehaviour *Misbehaviour, sigAndData *SignatureAndData) error {
	// timestamp less than consensus state would always fail and not succeed in fooling the
	// light client
	if sigAndData.Timestamp < clientState.ConsensusState.Timestamp {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidMisbehaviour, "timestamp is less than consensus state timestamp (%d < %d)", sigAndData.Timestamp, clientState.ConsensusState.Timestamp)
	}

	// ensure data can be unmarshaled to the specified data type
	if _, err := UnmarshalDataByType(cdc, sigAndData.DataType, sigAndData.Data); err != nil {
		return err
	}

	data, err := MisbehaviourSignBytes(
		cdc,
		misbehaviour.Sequence, sigAndData.Timestamp,
		clientState.ConsensusState.Diversifier,
		sigAndData.DataType,
		sigAndData.Data,
	)
	if err != nil {
		return err
	}

	sigData, err := UnmarshalSignatureData(cdc, sigAndData.Signature)
	if err != nil {
		return err
	}

	if err := VerifySignature(clientState.ConsensusState.GetPubKey(), data, sigData); err != nil {
		return err
	}

	return nil

}
