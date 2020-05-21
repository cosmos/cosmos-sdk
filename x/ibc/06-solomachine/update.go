package solomachine

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// CheckValidityAndUpdateState checks if the provided header is valid and updates
// the consensus state if appropriate. It returns an error if:
// - the client or header provided are not parseable to solo machine types
// - the currently registered public key did not provide the update signature
func CheckValidityAndUpdateState(
	clientState clientexported.ClientState, header clientexported.Header,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	// cast the client state to solo machine
	smClientState, ok := clientState.(ClientState)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidClientType, "client state type %T is not solomachine", clientState,
		)
	}

	smHeader, ok := header.(Header)
	if !ok {
		return nil, nil, sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader, "header type %T is not solomachine", header,
		)
	}

	if err := checkValidity(smClientState, smHeader); err != nil {
		return nil, nil, err
	}

	smClientState, consensusState := update(smClientState, smHeader)
	return smClientState, consensusState, nil
}

// checkValidity checks if the Solo Machine update signature is valid.
func checkValidity(clientState ClientState, header Header) error {
	// assert update sequence is current sequence
	if header.Sequence != clientState.ConsensusState.Sequence {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"sequence provided in the header does not match the client state sequence (%d != %d)", header.Sequence, clientState.ConsensusState.Sequence,
		)
	}

	// assert currently registered public key signed over the new public key with correct sequence
	value := append(
		sdk.Uint64ToBigEndian(header.Sequence),
		header.NewPubKey.Bytes()...,
	)

	if !clientState.ConsensusState.PubKey.VerifyBytes(value, header.Signature) {
		return sdkerrors.Wrap(
			clienttypes.ErrInvalidHeader,
			"header signature verification failed",
		)
	}

	return nil
}

// update the consensus state to the new public key and an incremented sequence
func update(clientState ClientState, header Header) (ClientState, ConsensusState) {
	consensusState := ConsensusState{
		// increment sequence number
		Sequence: header.Sequence + 1,
		PubKey:   header.NewPubKey,
	}

	clientState.ConsensusState = consensusState
	return clientState, consensusState
}
