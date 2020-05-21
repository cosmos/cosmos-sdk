package solomachine

import (
	"bytes"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// CheckMisbehaviourAndUpdateState determines whether or not the currently registered
// public key signed over two different messages with the same sequence.
func CheckMisbehaviourAndUpdateState(
	clientState clientexported.ClientState,
	_ clientexported.ConsensusState,
	misbehaviour clientexported.Misbehaviour,
) (clientexported.ClientState, error) {
	// cast the interface to specific types before checking for misbehaviour
	smClientState, ok := clientState.(ClientState)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClientType, "client state type is not solo machine")
	}

	if smClientState.IsFrozen() {
		return nil, sdkerrors.Wrapf(clienttypes.ErrInvalidEvidence, "client is already frozen")
	}

	evidence, ok := misbehaviour.(Evidence)
	if !ok {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidClientType, "evidence type is not solo machine")
	}

	if err := checkMisbehaviour(smClientState, evidence); err != nil {
		return nil, sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, err.Error())
	}

	smClientState.Frozen = true
	return smClientState, nil
}

// checkMisbehaviour checks if the currently registered public key has signed
// over two different messages at the same sequence.
func checkMisbehaviour(clientState ClientState, evidence Evidence) error {
	pubKey := clientState.ConsensusState.PubKey

	// assert that provided signature data are different
	if bytes.Equal(evidence.SignatureOne.Data, evidence.SignatureTwo.Data) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "evidence signatures have identical data messages")
	}

	data := append(sdk.Uint64ToBigEndian(evidence.Sequence), evidence.SignatureOne.Data...)

	// check first signature
	if !pubKey.VerifyBytes(data, evidence.SignatureOne.Signature) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "evidence signature one failed to be verified by currently registered public key")
	}

	data = append(sdk.Uint64ToBigEndian(evidence.Sequence), evidence.SignatureTwo.Data...)

	// check second signature
	if !pubKey.VerifyBytes(data, evidence.SignatureTwo.Signature) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidEvidence, "evidence signature two failed to be verified by currently registered public key")
	}

	return nil
}
