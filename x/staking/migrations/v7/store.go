package v7

import (
	"context"
	"fmt"

	"cosmossdk.io/core/address"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/staking/types"

	"github.com/cosmos/cosmos-sdk/codec"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MigrateStore performs in-place store migrations from v6 to v7.
// It updates all validators to set the consensus address.
func MigrateStore(_ context.Context, store storetypes.KVStore, cdc codec.BinaryCodec, consensusAddressCodec address.Codec) error {
	validatorIter := store.Iterator(types.ValidatorsKey, storetypes.PrefixEndBytes(types.ValidatorsKey))
	defer validatorIter.Close()

	for ; validatorIter.Valid(); validatorIter.Next() {
		var validator types.Validator
		if err := cdc.Unmarshal(validatorIter.Value(), &validator); err != nil {
			return fmt.Errorf("can't unmarshal validator: %w", err)
		}

		setConsensusAddress(&validator, consensusAddressCodec)
		store.Set(validatorIter.Key(), cdc.MustMarshal(&validator))
	}

	return nil
}

// setConsensusAddress sets the ConsensusAddress field for the given validator
func setConsensusAddress(validator *types.Validator, consensusAddressCodec address.Codec) {
	cpk, ok := validator.ConsensusPubkey.GetCachedValue().(cryptotypes.PubKey)
	// Best-effort way
	if ok {
		consAddr := sdk.ConsAddress(cpk.Address())
		validator.ConsensusAddress, _ = consensusAddressCodec.BytesToString(consAddr)
	}
}
