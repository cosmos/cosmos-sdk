package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clientexported "github.com/cosmos/cosmos-sdk/x/ibc/02-client/exported"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
)

// CheckProposedHeaderAndUpdateState will try to update the client with the new header if and
// only if the proposal passes and one of the following two conditions is satisfied:
// 		1) AllowGovernanceOverrideAfterExpiry=true and Expire(ctx.BlockTime) = true
// 		2) AllowGovernanceOverrideAfterMinbehaviour and IsFrozen() = true
// In case 2) before trying to update the client, the client will be unfrozen by resetting
// the FrozenHeight to the zero Height. Note, that even if the update happens, it may not be
// successful. The header may fail validation checks and an error will be returned in that
// case.
func (cs ClientState) CheckProposedHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	header clientexported.Header,
) (clientexported.ClientState, clientexported.ConsensusState, error) {
	tmHeader, ok := header.(*Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "expected type %T, got %T", &Header{}, header,
		)
	}

	// get consensus state corresponding to client state to check if the client is expired
	tmConsState, err := GetConsensusState(clientStore, cdc, clientState.GetLatestHeight())
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(
			err, "could not get consensus state from clientstore at height: %d", cs.GetLatestHeight(),
		)
	}

	if cs.IsFrozen() && cs.AllowGovernanceOverrideAfterMisbehaviour {
		cs.FrozenHeight = 0
	} else if !(tmClientState.AllowGovernanceOverrideAfterExpiry && cs.Expired(consensusState, ctx.BlockTime())) {
		return nil, nil, sdkerrors.Wrap(clienttypes.ErrClientUpdateFailed, "client cannot be updated with proposal")
	}

	// TODO add header checks

	newClientState, consensusState := update(&cs, tmHeader)
	return newClientState, consensusState, nil
}
