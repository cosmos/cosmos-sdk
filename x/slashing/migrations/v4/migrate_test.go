package v4_test

import (
	"testing"

	"github.com/bits-and-blooms/bitset"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	v4 "github.com/cosmos/cosmos-sdk/x/slashing/migrations/v4"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
)

var consAddr = sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))

func TestMigrate(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(slashing.AppModuleBasic{}).Codec
	storeKey := storetypes.NewKVStoreKey(slashingtypes.ModuleName)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)
	store := ctx.KVStore(storeKey)
	params := slashingtypes.Params{SignedBlocksWindow: 100}

	// store old signing info and bitmap entries
	bz := cdc.MustMarshal(&slashingtypes.ValidatorSigningInfo{Address: consAddr.String()})
	store.Set(v4.ValidatorSigningInfoKey(consAddr), bz)

	for i := int64(0); i < params.SignedBlocksWindow; i++ {
		// all even blocks are missed
		missed := &gogotypes.BoolValue{Value: i%2 == 0}
		bz := cdc.MustMarshal(missed)
		store.Set(v4.ValidatorMissedBlockBitArrayKey(consAddr, i), bz)
	}

	err := v4.Migrate(ctx, cdc, store, params)
	require.NoError(t, err)

	// ensure old entries no longer exist and new bitmap chunk entries exist
	entries := v4.GetValidatorMissedBlocks(ctx, cdc, store, consAddr, params)
	require.Empty(t, entries)

	for i := int64(0); i < params.SignedBlocksWindow; i++ {
		chunkIndex := i / v4.MissedBlockBitmapChunkSize
		chunk := store.Get(v4.ValidatorMissedBlockBitmapKey(consAddr, chunkIndex))
		require.NotNil(t, chunk)

		bs := bitset.New(uint(v4.MissedBlockBitmapChunkSize))
		require.NoError(t, bs.UnmarshalBinary(chunk))

		// ensure all even blocks are missed
		bitIndex := uint(i % v4.MissedBlockBitmapChunkSize)
		require.Equal(t, i%2 == 0, bs.Test(bitIndex))
		require.Equal(t, i%2 == 1, !bs.Test(bitIndex))
	}

	// ensure there's only one chunk for a window of size 100
	chunk := store.Get(v4.ValidatorMissedBlockBitmapKey(consAddr, 1))
	require.Nil(t, chunk)
}
