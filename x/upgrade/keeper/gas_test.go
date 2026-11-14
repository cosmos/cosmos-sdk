package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/runtime"
	storetypes "github.com/cosmos/cosmos-sdk/store/v2/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func TestGetProtocolVersionGasChargesOneStoreRead(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig()
	key := storetypes.NewKVStoreKey(types.StoreKey)
	storeService := runtime.NewKVStoreService(key)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithGasMeter(storetypes.NewGasMeter(100000))

	k := NewKeeper(map[int64]bool{}, storeService, encCfg.Codec, t.TempDir(), nil, authtypes.NewModuleAddress(govtypes.ModuleName).String())

	require.NoError(t, k.setProtocolVersion(ctx, 7))

	start := ctx.GasMeter().GasConsumed()
	pv, err := k.getProtocolVersion(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(7), pv)
	charged := ctx.GasMeter().GasConsumed() - start

	// One Get of the 8-byte protocol version: ReadCostFlat 1000 + ReadCostPerByte 3 * value bytes.
	// Update this constant if the store gas schedule changes.
	require.Equal(t, uint64(1027), charged)
	t.Logf("getProtocolVersion charged %d gas", charged)
}
