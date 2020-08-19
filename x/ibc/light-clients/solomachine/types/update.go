package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// CheckHeaderAndUpdateState checks if the provided header is valid and updates
// the consensus state if appropriate. It returns an error if:
// - the client or header provided are not parseable to solo machine types
// - the currently registered public key did not provide the update signature
func (cs ClientState) CheckHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	header clientexported.Header,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	smHeader, ok := header.(Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "header type %T is not solomachine", header,
		)
	}

	if err := checkHeader(cs, smHeader); err != nil {
		return nil, nil, err
	}

	clientState, consensusState := update(cs, smHeader)
	return clientState, consensusState, nil
}

// checkHeader checks if the Solo Machine update signature is valid.
func checkHeader(clientState ClientState, header Header) error {
	// assert update sequence is current sequence
	if header.Sequence != clientState.ConsensusState.Sequence {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"sequence provided in the header does not match the client state sequence (%d != %d)", header.Sequence, clientState.ConsensusState.Sequence,
		)
	}

	// assert currently registered public key signed over the new public key with correct sequence
	data := HeaderSignBytes(header)
	if err := CheckSignature(clientState.ConsensusState.PubKey, data, header.Signature); err != nil {
		return sdkerrors.Wrap(ErrInvalidHeader, err.Error())
	}

	return nil
}

// update the consensus state to the new public key and an incremented sequence
func update(clientState ClientState, header Header) (ClientState, ConsensusState) {
	consensusState := ConsensusState{
		// increment sequence number
		Sequence: clientState.ConsensusState.Sequence + 1,
		PubKey:   header.NewPubKey,
	}

	clientState.ConsensusState = consensusState
	return clientState, consensusState
}
