package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// CheckProposedHeaderAndUpdateState updates the consensus state to the header's sequence and
// public key. An error is returned if the header cannot be casted to a solo machine header.
func (cs ClientState) CheckProposedHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	header clientexported.Header,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	smHeader, ok := header.(*Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "header type %T, expected  %T", header, &Header{},
		)
	}

	consensusState := &ConsensusState{
		Sequence:  smHeader.Sequence,
		PublicKey: smHeader.NewPublicKey,
	}

	cs.ConsensusState = consensusState

	return cs, consensusState, nil
}
