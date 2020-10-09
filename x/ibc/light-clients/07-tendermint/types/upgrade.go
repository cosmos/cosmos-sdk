package types

import (
	"fmt"
	"net/url"
	"reflect"
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
	upgradePath, err := constructUpgradeMerklePath(cs.UpgradePath, upgradeHeight)
	if err != nil {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidUpgradeClient, "cannot upgrade client, unescaping key with URL format failed: %v", err)
	}

	// UpgradeHeight must be in same version as client state height
	if cs.GetLatestHeight().GetVersionNumber() != upgradeHeight.GetVersionNumber() {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidHeight, "version at which upgrade occurs must be same as current client version. expected version %d, got %d",
			cs.GetLatestHeight().GetVersionNumber(), upgradeHeight.GetVersionNumber())
	}

	tmClient, ok := upgradedClient.(*ClientState)
	if !ok {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClientType, "upgraded client must be Tendermint client. expected: %T got: %T",
			&ClientState{}, upgradedClient)
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
	// This verifies that upgrade is intended for the provided version, since committed client must exist
	// at this consensus state
	consState, err := GetConsensusState(clientStore, cdc, upgradeHeight)
	if err != nil {
		return sdkerrors.Wrap(err, "could not retrieve consensus state for upgradeHeight")
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
		cs.MaxClockDrift, tmCommittedClient.LatestHeight, tmCommittedClient.ConsensusParams, tmCommittedClient.ProofSpecs, tmCommittedClient.UpgradePath,
		cs.AllowUpdateAfterExpiry, cs.AllowUpdateAfterMisbehaviour,
	)
	if !reflect.DeepEqual(expectedClient, tmClient) {
		return sdkerrors.Wrapf(clienttypes.ErrInvalidClient, "upgraded client does not maintain previous chosen parameters. expected: %v got: %v",
			expectedClient, tmClient)
	}

	return merkleProof.VerifyMembership(cs.ProofSpecs, consState.GetRoot(), upgradePath, bz)
}

// construct MerklePath from upgradePath
func constructUpgradeMerklePath(upgradePath string, upgradeHeight exported.Height) (commitmenttypes.MerklePath, error) {
	// assume that all keys here are separated by `/` and
	// any `/` within a merkle key is correctly escaped
	upgradeKeys := strings.Split(upgradePath, "/")
	// unescape the last key so that we can append `/{height}` to the last key
	lastKey, err := url.PathUnescape(upgradeKeys[len(upgradeKeys)-1])
	if err != nil {
		return commitmenttypes.MerklePath{}, err
	}
	// append upgradeHeight to last key in merkle path
	// this will create the IAVL key that is used to store client in upgrade store
	upgradeKeys[len(upgradeKeys)-1] = fmt.Sprintf("%s/%d", lastKey, upgradeHeight.GetVersionHeight())
	return commitmenttypes.NewMerklePath(upgradeKeys), nil
}
