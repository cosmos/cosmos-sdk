package docs

import (
	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var BalancesPrefix = collections.NewPrefix(1)

type Keeper struct {
	Schema   collections.Schema
	Balances collections.Map[collections.Pair[sdk.AccAddress, string], math.Int]
}

func NewKeeper(storeKey *storetypes.KVStoreKey) Keeper {
	sb := collections.NewSchemaBuilder(sdk.OpenKVStore(storeKey))
	return Keeper{
		Balances: collections.NewMap(
			sb, BalancesPrefix, "balances",
			collections.PairKeyCodec(sdk.AccAddressKey, collections.StringKey),
			math.IntValue,
		),
	}
}

func (k Keeper) SetBalance(ctx sdk.Context, address sdk.AccAddress, denom string, amount math.Int) error {
	return k.Balances.Set(ctx, collections.Join(address, denom), amount)
}

func (k Keeper) GetBalance(ctx sdk.Context, address sdk.AccAddress, denom string) (math.Int, error) {
	return k.Balances.Get(ctx, collections.Join(address, denom))
}

func (k Keeper) GetAllAddressBalances(ctx sdk.Context, address sdk.AccAddress) (sdk.Coins, error) {
	balances := sdk.NewCoins()

	rng := collections.NewPrefixedPairRange[sdk.AccAddress, string](address)

	iter, err := k.Balances.Iterate(ctx, rng)
	if err != nil {
		return nil, err
	}

	kvs, err := iter.KeyValues()
	if err != nil {
		return nil, err
	}

	for _, kv := range kvs {
		balances = balances.Add(sdk.NewCoin(kv.Key.K2(), kv.Value))
	}
	return balances, nil
}
