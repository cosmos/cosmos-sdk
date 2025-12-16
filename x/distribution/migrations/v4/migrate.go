package v4

import (
	"cosmossdk.io/collections"
	corestoretypes "cosmossdk.io/core/store"
	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dstrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
)

const (
	ModuleName = "distribution"
)

var (
	// ParamsKey is the key of x/distribution params
	ParamsKey = collections.NewPrefix(9)

	// NakamotoBonusKey is the key of x/distribution nakamoto bonus
	NakamotoBonusKey = collections.NewPrefix(10)

	// DefaultNakamotoBonus is the ADR's initial value: 3% (0.03)
	DefaultNakamotoBonus = math.LegacyNewDecWithPrec(3, 2) // 0.03
)

// MigrateStore migrates the x/distribution module state to version 4.
func MigrateStore(
	ctx sdk.Context,
	storeService corestoretypes.KVStoreService,
	cdc codec.BinaryCodec,
	nakamotoBonus collections.Item[math.LegacyDec],
) error {
	// Open the KVStore
	store := storeService.OpenKVStore(ctx)

	paramsBz, err := store.Get(ParamsKey)
	if err != nil {
		return err
	}

	var params dstrtypes.Params
	if err = cdc.Unmarshal(paramsBz, &params); err != nil {
		return err
	}

	defaultParams := dstrtypes.DefaultParams()
	params.NakamotoBonus = defaultParams.NakamotoBonus

	bz, err := cdc.Marshal(&params)
	if err != nil {
		return err
	}

	if err := store.Set(ParamsKey, bz); err != nil {
		return err
	}

	defaultNakamotoBonus := DefaultNakamotoBonus
	if ok, err := nakamotoBonus.Has(ctx); !ok || err != nil {
		if err := nakamotoBonus.Set(ctx, defaultNakamotoBonus); err != nil {
			return err
		}
	}

	return nil
}
