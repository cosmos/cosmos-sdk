package types

import (
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// CheckProposedHeaderAndUpdateState will try to update the client with the new header if and
// only if the proposal passes and one of the following two conditions is satisfied:
// 		1) AllowUpdateAfterExpiry=true and Expire(ctx.BlockTime) = true
// 		2) AllowUpdateAfterMisbehaviour and IsFrozen() = true
// In case 2) before trying to update the client, the client will be unfrozen by resetting
// the FrozenHeight to the zero Height. If AllowUpdateAfterMisbehaviour is set to true,
// expired clients will also be updated even if AllowUpdateAfterExpiry is set to false.
// Note, that even if the update happens, it may not be successful. The header may fail
// validation checks and an error will be returned in that case.
func (cs ClientState) CheckProposedHeaderAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	header exported.Header,
) (exported.ClientState, exported.ConsensusState, error) {
	tmHeader, ok := header.(*Header)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader, "expected type %T, got %T", &Header{}, header,
		)
	}

	// get consensus state corresponding to client state to check if the client is expired
	consensusState, err := GetConsensusState(clientStore, cdc, cs.GetLatestHeight())
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(
			err, "could not get consensus state from clientstore at height: %d", cs.GetLatestHeight(),
		)
	}

	switch {

	case cs.IsFrozen():
		if !cs.AllowUpdateAfterMisbehaviour {
			return nil, nil, sdkerrors.Wrap(clienttypes.ErrUpdateClientFailed, "client is not allowed to be unfrozen")
		}

		// unfreeze the client
		cs.FrozenHeight = clienttypes.ZeroHeight()

		// if the client is expired we unexpire the client using softer validation, otherwise
		// full validation on the header is performed.
		if cs.IsExpired(consensusState.Timestamp, ctx.BlockTime()) {
			return cs.unexpireClient(ctx, clientStore, consensusState, tmHeader, ctx.BlockTime())
		}

		// NOTE: the client may be frozen again since the misbehaviour evidence may
		// not be expired yet
		return cs.CheckHeaderAndUpdateState(ctx, cdc, clientStore, header)

	case cs.AllowUpdateAfterExpiry && cs.IsExpired(consensusState.Timestamp, ctx.BlockTime()):
		return cs.unexpireClient(ctx, clientStore, consensusState, tmHeader, ctx.BlockTime())

	default:
		return nil, nil, sdkerrors.Wrap(clienttypes.ErrUpdateClientFailed, "client cannot be updated with proposal")
	}

}

// unexpireClient checks if the proposed header is sufficient to update an expired client.
// The client is updated if no error occurs.
func (cs ClientState) unexpireClient(
	ctx sdk.Context, clientStore sdk.KVStore, consensusState *ConsensusState, header *Header, currentTimestamp time.Time,
) (exported.ClientState, exported.ConsensusState, error) {

	// the client is expired and either AllowUpdateAfterMisbehaviour or AllowUpdateAfterExpiry
	// is set to true so light validation of the header is executed
	if err := cs.checkProposedHeader(consensusState, header, currentTimestamp); err != nil {
		return nil, nil, err
	}

	newClientState, consensusState := update(ctx, clientStore, &cs, header)
	return newClientState, consensusState, nil
}

// checkProposedHeader checks if the Tendermint header is valid for updating a client after
// a passed proposal.
// It returns an error if:
// - the header provided is not parseable to tendermint types
// - header height is less than or equal to the latest client state height
// - signed tendermint header is invalid
// - header timestamp is less than or equal to the latest consensus state timestamp
// - header timestamp is expired
// NOTE: header.ValidateBasic is called in the 02-client proposal handler. Additional checks
// on the validator set and the validator set hash are done in header.ValidateBasic.
func (cs ClientState) checkProposedHeader(consensusState *ConsensusState, header *Header, currentTimestamp time.Time) error {
	tmSignedHeader, err := tmtypes.SignedHeaderFromProto(header.SignedHeader)
	if err != nil {
		return sdkerrors.Wrap(err, "signed header in not tendermint signed header type")
	}

	if !header.GetTime().After(consensusState.Timestamp) {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header timestamp is less than or equal to latest consensus state timestamp (%s ≤ %s)", header.GetTime(), consensusState.Timestamp)
	}

	// assert header height is newer than latest client state
	if header.GetHeight().LTE(cs.GetLatestHeight()) {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidHeader,
			"header height ≤ consensus state height (%s ≤ %s)", header.GetHeight(), cs.GetLatestHeight(),
		)
	}

	if err := tmSignedHeader.ValidateBasic(cs.GetChainID()); err != nil {
		return sdkerrors.Wrap(err, "signed header failed basic validation")
	}

	if cs.IsExpired(header.GetTime(), currentTimestamp) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidHeader, "header timestamp is already expired")
	}

	return nil
}
