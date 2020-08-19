package types

import (
	"bytes"

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

	evidence, ok := misbehaviour.(Evidence)
	if !ok {
		return nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidClientType,
			"evidence type %T, expected %T", misbehaviour, Evidence{},
		)
	}

	if cs.IsFrozen() {
		return nil, sdkerrors.Wrapf(clienttypes.ErrClientFrozen, "client is already frozen")
	}

	if err := checkMisbehaviour(cs, evidence); err != nil {
		return nil, err
	}

	cs.FrozenHeight = uint64(evidence.GetHeight())
	return cs, nil
}

// checkMisbehaviour checks if the currently registered public key has signed
// over two different messages at the same sequence.
func checkMisbehaviour(clientState ClientState, evidence Evidence) error {
	pubKey := clientState.ConsensusState.GetPubKey()

	// assert that provided signature data are different
	if bytes.Equal(evidence.SignatureOne.Data, evidence.SignatureTwo.Data) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "evidence signatures have identical data messages")
	}

	data := EvidenceSignBytes(evidence.Sequence, evidence.SignatureOne.Data)

	// check first signature
	if err := VerifySignature(pubKey, data, evidence.SignatureOne.Signature); err != nil {
		return sdkerrors.Wrap(err, "evidence signature one failed to be verified")
	}

	data = EvidenceSignBytes(evidence.Sequence, evidence.SignatureTwo.Data)

	// check second signature
	if err := VerifySignature(pubKey, data, evidence.SignatureTwo.Signature); err != nil {
		return sdkerrors.Wrap(err, "evidence signature two failed to be verified")
	}

	return nil
}
