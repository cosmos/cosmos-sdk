package types

import (
	"reflect"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
)

// VerifyUpgrade checks if the upgraded client has been committed by the current client
// It will zero out all client-specific fields (e.g. TrustingPeriod and verify all data
// in client state that must be the same across all valid Tendermint clients for the new chain.
// VerifyUpgrade will return an error if:
// - the upgradedClient is not a Tendermint ClientState
// - the height of upgraded client is not greater than that of current client
// - the latest height of the new client does not match the height in committed client
// - any Tendermint chain specified parameter in upgraded client such as ChainID, UnbondingPeriod,
//   and ProofSpecs do not match parameters set by committed client
func (cs ClientState) VerifyUpgrade(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	upgradedClient exported.ClientState, proofUpgrade []byte,
) error {
	if cs.UpgradePath == nil {
		return sdkerrors.Wrap(clienttypes.ErrInvalidUpgradeClient, "cannot upgrade client, no upgrade path set")
	}
	tmClient, ok := upgradedClient.(*ClientState)
	if !ok {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "upgraded client must be Tendermint client. expected: %T got: %T",
			&ClientState{}, upgradedClient)
	}

	if !upgradedClient.GetLatestHeight().GT(cs.GetLatestHeight()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "upgrade client height %s must be greater than current client height %s",
			upgradedClient.GetLatestHeight(), cs.GetLatestHeight())
	}

	if len(proofUpgrade) == 0 {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof of upgrade is empty")
	}

	var merkleProof commitmenttypes.MerkleProof
	if err := cdc.UnmarshalBinaryBare(proofUpgrade, &merkleProof); err != nil {
		return sdkerrors.Wrapf(commitmenttypes.ErrInvalidProof, "could not unmarshal merkle proof: %v", err)
	}

	// counterparty chain must commit the upgraded client with all client-customizable fields zeroed out
	// at the upgrade path specified by current client
	committedClient := upgradedClient.ZeroCustomFields()
	bz, err := codec.MarshalAny(cdc, committedClient)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "could not marshal client state: %v", err)
	}

	// Must prove against latest consensus state to ensure we are verifying against latest upgrade plan
	consState, err := GetConsensusState(clientStore, cdc, cs.GetLatestHeight())
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve latest consensus state")
	}

	if cs.IsExpired(consState.Timestamp, ctx.BlockTime()) {
		return sdkerrors.Wrap(clienttypes.ErrInvalidClient, "cannot upgrade an expired client")
	}

	tmCommittedClient, ok := committedClient.(*ClientState)
	if !ok {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "upgraded client must be Tendermint client. expected: %T got: %T",
			&ClientState{}, upgradedClient)
	}

	// Relayer must keep all client-chosen parameters the same as the previous client.
	// Compare relayer-provided client state against expected client state.
	// All chain-chosen parameters come from committed client, all client-chosen parameters
	// come from current client
	expectedClient := NewClientState(
		tmCommittedClient.ChainId, cs.TrustLevel, cs.TrustingPeriod, tmCommittedClient.UnbondingPeriod,
		cs.MaxClockDrift, tmCommittedClient.LatestHeight, tmCommittedClient.ProofSpecs, tmCommittedClient.UpgradePath,
		cs.AllowUpdateAfterExpiry, cs.AllowUpdateAfterMisbehaviour,
	)
	if !reflect.DeepEqual(expectedClient, tmClient) {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "upgraded client does not maintain previous chosen parameters. expected: %v got: %v",
			expectedClient, tmClient)
	}

	return merkleProof.VerifyMembership(cs.ProofSpecs, consState.GetRoot(), *cs.UpgradePath, bz)
}
