package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
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
	if !upgradedClient.GetLatestHeight().GT(cs.GetLatestHeight()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "upgrade client height %s must be greater than current client height %s",
			upgradedClient.GetLatestHeight(), cs.GetLatestHeight())
	}

	tmClient, ok := upgradedClient.(*ClientState)
	if !ok {
		return sdkerrors.Wrapf(
			clienttypes.ErrInvalidClient, "upgrade client must be a Tendermint client. expected %T, got %T", &ClientState{}, upgradedClient,
		)
	}

	if proofUpgrade == nil {
		return sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof of upgrade is nil")
	}
	var merkleProof commitmenttypes.MerkleProof
	if err := cdc.UnmarshalBinaryBare(proofUpgrade, &merkleProof); err != nil {
		return sdkerrors.Wrapf(commitmenttypes.ErrInvalidProof, "could not unmarshal proof. error: %v", err)
	}

	// construct committed client from upgradedClient by setting all chain-specified parameters
	// with values from upgraded client, and ommitting all client-specified parameters
	committedClient := ClientState{
		ChainId:         tmClient.ChainId,
		UnbondingPeriod: tmClient.UnbondingPeriod,
		LatestHeight:    tmClient.LatestHeight,
		ProofSpecs:      tmClient.ProofSpecs,
		UpgradePath:     tmClient.UpgradePath,
	}

	bz, err := codec.MarshalAny(cdc, committedClient)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "could not marshal clientstate. error: %v", err)
	}

	// Must prove against latest consensus state to ensure we are verifying against latest upgrade plan
	consState, err := GetConsensusState(clientStore, cdc, cs.GetLatestHeight())
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve latest consensus state")
	}

	return merkleProof.VerifyMembership(cs.ProofSpecs, consState.GetRoot(), cs.UpgradePath, bz)
}
