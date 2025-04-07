package keeper

import (
	"bytes"
	"fmt"

	"cosmossdk.io/store/prefix"

	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// MigrateDelegationsByValidatorIndex is a migration that runs over multiple blocks,
// this is necessary as to build the reverse index we need to iterate over a large set
// of delegations.
func (k Keeper) MigrateDelegationsByValidatorIndex(ctx sdk.Context, iterationLimit int) error {
	store := runtime.KVStoreAdapter(k.storeService.OpenKVStore(ctx))
	valStore := prefix.NewStore(store, types.DelegationKey)

	// Check the store to see if there is a value stored under the key
	key := store.Get(types.NextMigrateDelegationsByValidatorIndexKey)
	if key == nil {
		return nil
	}

	// Initialize the counter to 0
	iterationCounter := 0

	// Start the iterator from the key that is in the store
	iterator := valStore.Iterator(key, nil)
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		key := iterator.Key()

		// Parse the index to use setting the reverse index
		del, val, err := ParseDelegationKey(key)
		if err != nil {
			return err
		}

		// Set the reverse index in the store
		store.Set(types.GetDelegationsByValKey(val, del), []byte{})

		iterationCounter++
		if iterationCounter >= iterationLimit {
			ctx.Logger().Info(fmt.Sprintf("Migrated %d delegations, next key %x", iterationLimit, key))

			// Set the key in the store after it has been processed
			store.Set(types.NextMigrateDelegationsByValidatorIndexKey, key)
			break
		}
	}

	// If the iterator is invalid we have processed the full store
	if !iterator.Valid() {
		ctx.Logger().Info("successfully completed migration for delegation keys")
		store.Delete(types.NextMigrateDelegationsByValidatorIndexKey)
	}

	return nil
}

// ParseDelegationKey parses given key and returns delagator, validator address bytes
//
// input should not contain the DelegationKey prefix.
func ParseDelegationKey(input []byte) (sdk.AccAddress, sdk.ValAddress, error) {
	if len(input) == 0 {
		return nil, nil, fmt.Errorf("no bytes left to parse delegator length: %X", input)
	}

	if bytes.HasPrefix(input, types.DelegationKey) {
		return nil, nil, fmt.Errorf("input should not contain the DelegationKey prefix: %X", input)
	}

	delegatorLen := input[0]
	input = input[1:] // remove the length byte of delegator address.
	if len(input) == 0 {
		return nil, nil, fmt.Errorf("no bytes left to parse delegator address: %X", input)
	}

	delegator := input[:int(delegatorLen)]
	input = input[int(delegatorLen):] // remove the delegator address.
	if len(input) == 0 {
		return nil, nil, fmt.Errorf("no bytes left to parse validator length: %X", input)
	}

	input = input[1:] // remove the length byte of the validator address.
	if len(input) == 0 {
		return nil, nil, fmt.Errorf("no bytes left to parse validator address: %X", input)
	}

	validator := input
	return delegator, validator, nil
}
