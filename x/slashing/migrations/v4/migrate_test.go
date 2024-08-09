package v4_test

import (
	"testing"

	"github.com/bits-and-blooms/bitset"
	gogotypes "github.com/cosmos/gogoproto/types"
	"github.com/stretchr/testify/require"

	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/x/slashing"
	v4 "cosmossdk.io/x/slashing/migrations/v4"
	slashingtypes "cosmossdk.io/x/slashing/types"

	"github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var consAddr = sdk.ConsAddress(sdk.AccAddress([]byte("addr1_______________")))

func TestMigrate(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, slashing.AppModule{}).Codec
	ctx := coretesting.Context()
	store := coretesting.KVStoreService(ctx, slashingtypes.ModuleName).OpenKVStore(ctx)
	params := slashingtypes.Params{SignedBlocksWindow: 100}
	valCodec := address.NewBech32Codec("cosmosvalcons")
	consStrAddr, err := valCodec.BytesToString(consAddr)
	require.NoError(t, err)

	// store old signing info and bitmap entries
	bz := cdc.MustMarshal(&slashingtypes.ValidatorSigningInfo{Address: consStrAddr})
	err = store.Set(v4.ValidatorSigningInfoKey(consAddr), bz)
	require.NoError(t, err)

	for i := int64(0); i < params.SignedBlocksWindow; i++ {
		// all even blocks are missed
		missed := &gogotypes.BoolValue{Value: i%2 == 0}
		bz := cdc.MustMarshal(missed)
		err := store.Set(v4.ValidatorMissedBlockBitArrayKey(consAddr, i), bz)
		require.NoError(t, err)
	}

	err = v4.Migrate(ctx, cdc, store, params, valCodec)
	require.NoError(t, err)

	for i := int64(0); i < params.SignedBlocksWindow; i++ {
		chunkIndex := i / v4.MissedBlockBitmapChunkSize
		chunk, err := store.Get(v4.ValidatorMissedBlockBitmapKey(consAddr, chunkIndex))
		require.NoError(t, err)
		require.NotNil(t, chunk)

		bs := bitset.New(uint(v4.MissedBlockBitmapChunkSize))
		require.NoError(t, bs.UnmarshalBinary(chunk))

		// ensure all even blocks are missed
		bitIndex := uint(i % v4.MissedBlockBitmapChunkSize)
		require.Equal(t, i%2 == 0, bs.Test(bitIndex))
		require.Equal(t, i%2 == 1, !bs.Test(bitIndex))
	}

	// ensure there's only one chunk for a window of size 100
	chunk, err := store.Get(v4.ValidatorMissedBlockBitmapKey(consAddr, 1))
	require.NoError(t, err)
	require.Empty(t, chunk)
}
