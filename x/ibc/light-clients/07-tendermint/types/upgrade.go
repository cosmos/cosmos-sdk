package types

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/core/23-commitment/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

// VerifyUpgradeAndUpdateState checks if the upgraded client has been committed by the current client
// It will zero out all client-specific fields (e.g. TrustingPeriod and verify all data
// in client state that must be the same across all valid Tendermint clients for the new chain.
// VerifyUpgrade will return an error if:
// - the upgradedClient is not a Tendermint ClientState
// - the lastest height of the client state does not have the same revision number or has a greater
// height than the committed client.
// - the height of upgraded client is not greater than that of current client
// - the latest height of the new client does not match or is greater than the height in committed client
// - any Tendermint chain specified parameter in upgraded client such as ChainID, UnbondingPeriod,
//   and ProofSpecs do not match parameters set by committed client
func (cs ClientState) VerifyUpgradeAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	upgradedClient exported.ClientState, upgradedConsState exported.ConsensusState,
	proofUpgradeClient, proofUpgradeConsState []byte,
) (exported.ClientState, exported.ConsensusState, error) {
	if len(cs.UpgradePath) == 0 {
		return nil, nil, sdkerrors.Wrap(clienttypes.ErrInvalidUpgradeClient, "cannot upgrade client, no upgrade path set")
	}

	// last height of current counterparty chain must be client's latest height
	lastHeight := cs.GetLatestHeight()

	if !upgradedClient.GetLatestHeight().GT(lastHeight) {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "upgraded client height %s must be at greater than current client height %s",
			upgradedClient.GetLatestHeight(), lastHeight)
	}

	// counterparty chain must commit the upgraded client with all client-customizable fields zeroed out
	// at the upgrade path specified by current client
	// counterparty must also commit to the upgraded consensus state at a sub-path under the upgrade path specified
	tmUpgradeClient, ok := upgradedClient.(*ClientState)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "upgraded client must be Tendermint client. expected: %T got: %T",
			&ClientState{}, upgradedClient)
	}
	tmUpgradeConsState, ok := upgradedConsState.(*ConsensusState)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "upgraded consensus state must be Tendermint consensus state. expected %T, got: %T",
			&ConsensusState{}, upgradedConsState)
	}

	// unmarshal proofs
	var merkleProofClient, merkleProofConsState commitmenttypes.MerkleProof
	if err := cdc.UnmarshalBinaryBare(proofUpgradeClient, &merkleProofClient); err != nil {
		return nil, nil, sdkerrors.Wrapf(commitmenttypes.ErrInvalidProof, "could not unmarshal client merkle proof: %v", err)
	}
	if err := cdc.UnmarshalBinaryBare(proofUpgradeConsState, &merkleProofConsState); err != nil {
		return nil, nil, sdkerrors.Wrapf(commitmenttypes.ErrInvalidProof, "could not unmarshal consensus state merkle proof: %v", err)
	}

	// Must prove against latest consensus state to ensure we are verifying against latest upgrade plan
	// This verifies that upgrade is intended for the provided revision, since committed client must exist
	// at this consensus state
	consState, err := GetConsensusState(clientStore, cdc, lastHeight)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "could not retrieve consensus state for lastHeight")
	}

	if cs.IsExpired(consState.Timestamp, ctx.BlockTime()) {
		return nil, nil, sdkerrors.Wrap(clienttypes.ErrInvalidClient, "cannot upgrade an expired client")
	}

	// Verify client proof
	bz, err := cdc.MarshalInterface(upgradedClient)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "could not marshal client state: %v", err)
	}
	// construct clientState Merkle path
	upgradeClientPath := constructUpgradeClientMerklePath(cs.UpgradePath, lastHeight)
	if err := merkleProofClient.VerifyMembership(cs.ProofSpecs, consState.GetRoot(), upgradeClientPath, bz); err != nil {
		return nil, nil, sdkerrors.Wrapf(err, "client state proof failed. Path: %s", upgradeClientPath.Pretty())
	}

	// Verify consensus state proof
	bz, err = cdc.MarshalInterface(upgradedConsState)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "could not marshal consensus state: %v", err)
	}
	// construct consensus state Merkle path
	upgradeConsStatePath := constructUpgradeConsStateMerklePath(cs.UpgradePath, lastHeight)
	if err := merkleProofConsState.VerifyMembership(cs.ProofSpecs, consState.GetRoot(), upgradeConsStatePath, bz); err != nil {
		return nil, nil, sdkerrors.Wrapf(err, "consensus state proof failed. Path: %s", upgradeConsStatePath.Pretty())
	}

	// Construct new client state and consensus state
	// Relayer chosen client parameters are ignored.
	// All chain-chosen parameters come from committed client, all client-chosen parameters
	// come from current client.
	newClientState := NewClientState(
		tmUpgradeClient.ChainId, cs.TrustLevel, cs.TrustingPeriod, tmUpgradeClient.UnbondingPeriod,
		cs.MaxClockDrift, tmUpgradeClient.LatestHeight, tmUpgradeClient.ProofSpecs, tmUpgradeClient.UpgradePath,
		cs.AllowUpdateAfterExpiry, cs.AllowUpdateAfterMisbehaviour,
	)

	if err := newClientState.Validate(); err != nil {
		return nil, nil, sdkerrors.Wrap(err, "updated client state failed basic validation")
	}

	// The new consensus state is merely used as a trusted kernel against which headers on the new
	// chain can be verified. The root is empty as it cannot be known in advance, thus no proof verification will pass.
	// The timestamp and the NextValidatorsHash of the consensus state is the blocktime and NextValidatorsHash
	// of the last block committed by the old chain. This will allow the first block of the new chain to be verified against
	// the last validators of the old chain so long as it is submitted within the TrustingPeriod of this client.
	// NOTE: We do not set processed time for this consensus state since this consensus state should not be used for packet verification
	// as the root is empty. The next consensus state submitted using update will be usable for packet-verification.
	newConsState := NewConsensusState(
		tmUpgradeConsState.Timestamp, commitmenttypes.MerkleRoot{}, tmUpgradeConsState.NextValidatorsHash,
	)

	return newClientState, newConsState, nil
}

// construct MerklePath for the committed client from upgradePath
func constructUpgradeClientMerklePath(upgradePath []string, lastHeight exported.Height) commitmenttypes.MerklePath {
	// copy all elements from upgradePath except final element
	clientPath := make([]string, len(upgradePath)-1)
	copy(clientPath, upgradePath)

	// append lastHeight and `upgradedClient` to last key of upgradePath and use as lastKey of clientPath
	// this will create the IAVL key that is used to store client in upgrade store
	lastKey := upgradePath[len(upgradePath)-1]
	appendedKey := fmt.Sprintf("%s/%d/%s", lastKey, lastHeight.GetRevisionHeight(), upgradetypes.KeyUpgradedClient)

	clientPath = append(clientPath, appendedKey)
	return commitmenttypes.NewMerklePath(clientPath...)
}

// construct MerklePath for the committed consensus state from upgradePath
func constructUpgradeConsStateMerklePath(upgradePath []string, lastHeight exported.Height) commitmenttypes.MerklePath {
	// copy all elements from upgradePath except final element
	consPath := make([]string, len(upgradePath)-1)
	copy(consPath, upgradePath)

	// append lastHeight and `upgradedClient` to last key of upgradePath and use as lastKey of clientPath
	// this will create the IAVL key that is used to store client in upgrade store
	lastKey := upgradePath[len(upgradePath)-1]
	appendedKey := fmt.Sprintf("%s/%d/%s", lastKey, lastHeight.GetRevisionHeight(), upgradetypes.KeyUpgradedConsState)

	consPath = append(consPath, appendedKey)
	return commitmenttypes.NewMerklePath(consPath...)
}
