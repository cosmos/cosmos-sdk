package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	v2 "github.com/cosmos/cosmos-sdk/x/upgrade/migrations/v2"
	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func TestMigrate(t *testing.T) {
	encCfg := simapp.MakeTestEncodingConfig()
	key := sdk.NewKVStoreKey("upgrade")
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(key, tKey)
	paramstore := paramtypes.NewSubspace(encCfg.Marshaler, encCfg.Amino, key, tKey, "upgrade")

	// Check no params
	require.False(t, paramstore.Has(ctx, types.KeyIsMainnet))

	err := v2.Migrate(ctx, paramstore)
	require.NoError(t, err)

	var result bool
	paramstore.Get(ctx, types.KeyIsMainnet, &result)
	require.True(t, result)
}
