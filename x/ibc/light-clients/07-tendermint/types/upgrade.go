package types

import (
	"fmt"
	"strings"

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
	upgradedClient exported.ClientState, upgradeHeight exported.Height, proofUpgrade []byte,
) error {
	if cs.UpgradePath == "" {
		return sdkerrors.Wrap(clienttypes.ErrInvalidUpgradeClient, "cannot upgrade client, no upgrade path set")
	}
	// construct MerklePath from upgradePath
	upgradeKeys := strings.Split(cs.UpgradePath, "/")
	// append upgradeHeight to last key in merkle path
	upgradeKeys[len(upgradeKeys)-1] = fmt.Sprintf("%s/%d", upgradeKeys[len(upgradeKeys)-1], upgradeHeight.GetEpochHeight())
	upgradePath := commitmenttypes.NewMerklePath(upgradeKeys)

	if cs.GetLatestHeight().GetEpochNumber() != upgradeHeight.GetEpochNumber() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "epoch at which upgrade occurs must be same as current client epoch. expected epoch %d, got %d",
			cs.GetLatestHeight().GetEpochNumber(), upgradeHeight.GetEpochNumber())
	}

	if !upgradedClient.GetLatestHeight().GT(cs.GetLatestHeight()) {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "upgraded client height %s must be greater than current client height %s",
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
	consState, err := GetConsensusState(clientStore, cdc, upgradeHeight)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve consensus state for upgradeHeight")
	}

	return merkleProof.VerifyMembership(cs.ProofSpecs, consState.GetRoot(), upgradePath, bz)
}
