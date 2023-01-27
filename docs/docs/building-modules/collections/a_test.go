package collections

import (
	"cosmossdk.io/collections"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

var AccountsPrefix = collections.NewPrefix(0)

type Keeper struct {
	Schema   collections.Schema
	Accounts collections.Map[uint64, authtypes.BaseAccount]
}

func NewKeeper(storeKey *storetypes.KVStoreKey, cdc codec.BinaryCodec) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Accounts: collections.NewMap(sb, AccountsPrefix, "params", collections.Uint64Key, codec.CollValue[authtypes.BaseAccount](cdc)),
	}
}

func (k Keeper) GetAllAccounts(ctx sdk.Context) ([]authtypes.BaseAccount, error) {
	// passing a nil Ranger equals to: iterate over every possible key
	iter, err := k.Accounts.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	accounts, err := iter.Values()
	if err != nil {
		return nil, err
	}

	return accounts, err
}

func (k Keeper) IterateAccountsBetween(ctx sdk.Context, start, end uint64) ([]authtypes.BaseAccount, error) {
	// The collections.Range API offers a lot of capabilities
	// like defining where the iteration starts or ends.
	rng := new(collections.Range[uint64]).
		StartInclusive(start).
		EndExclusive(end).
		Descending()

	iter, err := k.Accounts.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}
	accounts, err := iter.Values()
	if err != nil {
		return nil, err
	}

	return accounts, nil
}

func (k Keeper) IterateAccounts(ctx sdk.Context, do func(id uint64, acc authtypes.BaseAccount) (stop bool)) error {
	iter, err := k.Accounts.Iterate(ctx, nil)
	if err != nil {
		return err
	}
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		kv, err := iter.KeyValue()
		if err != nil {
			return err
		}

		if do(kv.Key, kv.Value) {
			break
		}
	}
	return nil
}
