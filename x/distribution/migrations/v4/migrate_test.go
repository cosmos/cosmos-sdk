package v4

import (
	"testing"

	"cosmossdk.io/collections"
	collcodec "cosmossdk.io/collections/codec"
	storetypes "cosmossdk.io/store/types"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/distribution"
)

func TestMigration(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(distribution.AppModuleBasic{}).Codec
	storeKey := storetypes.NewKVStoreKey("distribution")
	storeService := runtime.NewKVStoreService(storeKey)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	addr1 := secp256k1.GenPrivKey().PubKey().Address()
	consAddr1 := types.ConsAddress(addr1)

	err := SetPreviousProposerConsAddr(ctx, storeService, cdc, consAddr1)
	require.NoError(t, err)

	gotAddr, err := GetPreviousProposerConsAddr(ctx, storeService, cdc)
	require.NoError(t, err)
	require.Equal(t, consAddr1, gotAddr)

	err = MigrateStore(ctx, storeService, cdc)
	require.NoError(t, err)

	sb := collections.NewSchemaBuilder(storeService)
	prevProposer := collections.NewItem(sb, NewProposerKey, "previous_proposer", collcodec.KeyToValueCodec(sdk.ConsAddressKey))
	_, err = sb.Build()
	require.NoError(t, err)

	newAddr, err := prevProposer.Get(ctx)
	require.NoError(t, err)

	require.Equal(t, consAddr1, newAddr)
}
