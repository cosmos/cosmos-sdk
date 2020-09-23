package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
)

// CheckHeaderAndUpdateState checks if the provided header is valid and updates
// the consensus state if appropriate. It returns an error if:
// - the client or header provided are not parseable to solo machine types
// - the currently registered public key did not provide the update signature
func (cs ClientState) CheckHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	header exported.Header,
) (exported.ClientState, exported.ConsensusState, error) {
	smHeader, ok := header.(*Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "header type %T, expected  %T", header, &Header{},
		)
	}

	if err := checkHeader(cdc, &cs, smHeader); err != nil {
		return nil, nil, err
	}

	clientState, consensusState := update(&cs, smHeader)
	return clientState, consensusState, nil
}

// checkHeader checks if the Solo Machine update signature is valid.
func checkHeader(cdc codec.BinaryMarshaler, clientState *ClientState, header *Header) error {
	// assert update sequence is current sequence
	if header.Sequence != clientState.Sequence {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"sequence provided in the header does not match the client state sequence (%d != %d)", header.Sequence, clientState.Sequence,
		)
	}

	// assert currently registered public key signed over the new public key with correct sequence
	data, err := HeaderSignBytes(cdc, header)
	if err != nil {
		return err
	}

	if err := VerifySignature(clientState.ConsensusState.GetPubKey(), data, header.Signature); err != nil {
		return sdkerrors.Wrap(ErrInvalidHeader, err.Error())
	}

	return nil
}

// update the consensus state to the new public key and an incremented sequence
func update(clientState *ClientState, header *Header) (*ClientState, *ConsensusState) {
	consensusState := &ConsensusState{
		PublicKey:   header.NewPublicKey,
		Diversifier: header.NewDiversifier,
		Timestamp:   header.Timestamp,
	}

	// increment sequence number
	clientState.Sequence++
	clientState.ConsensusState = consensusState
	return clientState, consensusState
}
