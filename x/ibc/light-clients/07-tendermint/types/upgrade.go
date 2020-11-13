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
// - the height of upgraded client is not greater than that of current client
// - the latest height of the new client does not match the height in committed client
// - any Tendermint chain specified parameter in upgraded client such as ChainID, UnbondingPeriod,
//   and ProofSpecs do not match parameters set by committed client
func (cs ClientState) VerifyUpgradeAndUpdateState(
	ctx sdk.Context, cdc codec.BinaryMarshaler, clientStore sdk.KVStore,
	upgradedClient exported.ClientState, upgradedConsState exported.ConsensusState,
	upgradeHeight exported.Height, proofUpgradeClient, proofUpgradeConsState []byte,
) (exported.ClientState, exported.ConsensusState, error) {
	if len(cs.UpgradePath) == 0 {
		return nil, nil, sdkerrors.Wrap(clienttypes.ErrInvalidUpgradeClient, "cannot upgrade client, no upgrade path set")
	}

	// UpgradeHeight must be in same version as client state height
	if cs.GetLatestHeight().GetVersionNumber() != upgradeHeight.GetVersionNumber() {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "version at which upgrade occurs must be same as current client version. expected version %d, got %d",
			cs.GetLatestHeight().GetVersionNumber(), upgradeHeight.GetVersionNumber())
	}

	if upgradedClient.GetLatestHeight().GetVersionNumber() <= cs.GetLatestHeight().GetVersionNumber() {
		return nil, nil, sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "upgraded client height %s must be at greater version than current client height %s",
			upgradedClient.GetLatestHeight(), cs.GetLatestHeight())
	}

	// counterparty chain must commit the upgraded client with all client-customizable fields zeroed out
	// at the upgrade path specified by current client
	// counterparty must also commit to the upgraded consensus state at a sub-path under the upgrade path specified
	committedClient := upgradedClient.ZeroCustomFields()
	tmCommittedClient, ok := committedClient.(*ClientState)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "upgraded client must be Tendermint client. expected: %T got: %T",
			&ClientState{}, upgradedClient)
	}
	tmUpgradeConsState, ok := upgradedConsState.(*ConsensusState)
	if !ok {
		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "upgraded consensus state must be Tendermint consensus state. expected %T, got: %T",
			&ConsensusState{}, tmUpgradeConsState)
	}

	// check proofs not empty
	if len(proofUpgradeClient) == 0 {
		return nil, nil, sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof of upgrade client is empty")
	}
	if len(proofUpgradeConsState) == 0 {
		return nil, nil, sdkerrors.Wrap(commitmenttypes.ErrInvalidProof, "proof of upgrade consensus state is empty")
	}

	// unmarshal proofs
	var merkleProofClient, merkleProofConsState commitmenttypes.MerkleProof
	if err := cdc.UnmarshalBinaryBare(proofUpgradeClient, &merkleProofClient); err != nil {
		return nil, nil, sdkerrors.Wrapf(commitmenttypes.ErrInvalidProof, "could not unmarshal client merkle proof: %v", err)
	}
	if err := cdc.UnmarshalBinaryBare(proofUpgradeClient, &merkleProofConsState); err != nil {
		return nil, nil, sdkerrors.Wrapf(commitmenttypes.ErrInvalidProof, "could not unmarshal consensus state merkle proof: %v", err)
	}

	// Must prove against latest consensus state to ensure we are verifying against latest upgrade plan
	// This verifies that upgrade is intended for the provided version, since committed client must exist
	// at this consensus state
	consState, err := GetConsensusState(clientStore, cdc, upgradeHeight)
	if err != nil {
		return nil, nil, sdkerrors.Wrap(err, "could not retrieve consensus state for upgradeHeight")
	}

	if cs.IsExpired(consState.Timestamp, ctx.BlockTime()) {
		return nil, nil, sdkerrors.Wrap(clienttypes.ErrInvalidClient, "cannot upgrade an expired client")
	}

	// Verify client proof
	bz, err := codec.MarshalAny(cdc, committedClient)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "could not marshal client state: %v", err)
	}
	// construct clientState Merkle path
	upgradeClientPath := constructUpgradeClientMerklePath(cs.UpgradePath[:], upgradeHeight)
	if err := merkleProofClient.VerifyMembership(cs.ProofSpecs, consState.GetRoot(), upgradeClientPath, bz); err != nil {
		return nil, nil, err
	}

	// Verify consensus state proof
	bz, err = codec.MarshalAny(cdc, upgradedConsState)
	if err != nil {
		return nil, nil, sdkerrors.Wrapf(clienttypes.ErrInvalidConsensus, "could not marshal consensus state: %v", err)
	}
	// construct consensus state Merkle path
	upgradeConsStatePath := constructUpgradeConsStateMerklePath(cs.UpgradePath[:], upgradeHeight)
	if err := merkleProofConsState.VerifyMembership(cs.ProofSpecs, consState.GetRoot(), upgradeConsStatePath, bz); err != nil {
		return nil, nil, err
	}

	// Construct new client state and consensus state
	// Relayer chosen client parameters are ignored.
	// All chain-chosen parameters come from committed client, all client-chosen parameters
	// come from current client.
	newClientState := NewClientState(
		tmCommittedClient.ChainId, cs.TrustLevel, cs.TrustingPeriod, tmCommittedClient.UnbondingPeriod,
		cs.MaxClockDrift, tmCommittedClient.LatestHeight, tmCommittedClient.ProofSpecs, tmCommittedClient.UpgradePath,
		cs.AllowUpdateAfterExpiry, cs.AllowUpdateAfterMisbehaviour,
	)

	if err := newClientState.Validate(); err != nil {
		return nil, nil, sdkerrors.Wrap(err, "updated client state failed basic validation")
	}

	// The new consensus state is merely used as a trusted kernel against which headers on the new
	// chain can be verified. The root is empty as it cannot be known in advance, thus no proof verification will pass.
	// The timestamp of the consensus state is also this chain's blocktime. This is because starting up a new chain
	// may take a long time, especially if there are unexpected issues and thus we do not want the new client to be
	// automatically expired due to unforeseen delays.
	newConsState := NewConsensusState(
		ctx.BlockTime(), commitmenttypes.MerkleRoot{}, tmUpgradeConsState.NextValidatorsHash,
	)

	return newClientState, newConsState, nil
}

// construct MerklePath for the committed client from upgradePath
func constructUpgradeClientMerklePath(upgradePath []string, upgradeHeight exported.Height) commitmenttypes.MerklePath {
	// append upgradeHeight and `upgradedClient` to last key in merkle path
	// this will create the IAVL key that is used to store client in upgrade store
	upgradePath[len(upgradePath)-1] = fmt.Sprintf("%s/%d/%s", upgradePath[len(upgradePath)-1], upgradeHeight.GetVersionHeight(), upgradetypes.KeyUpgradedClient)
	return commitmenttypes.NewMerklePath(upgradePath...)
}

// construct MerklePath for the committed consensus state from upgradePath
func constructUpgradeConsStateMerklePath(upgradePath []string, upgradeHeight exported.Height) commitmenttypes.MerklePath {
	// append upgradeHeight and `upgradedClient` to last key in merkle path
	// this will create the IAVL key that is used to store client in upgrade store
	upgradePath[len(upgradePath)-1] = fmt.Sprintf("%s/%d/%s", upgradePath[len(upgradePath)-1], upgradeHeight.GetVersionHeight(), upgradetypes.KeyUpgradedConsState)
	return commitmenttypes.NewMerklePath(upgradePath...)
}
