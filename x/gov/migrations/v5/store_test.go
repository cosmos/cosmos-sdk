package v5_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	v5 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v5"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(gov.AppModuleBasic{}, bank.AppModuleBasic{}).Codec
	govKey := storetypes.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, storetypes.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)
	storeService := runtime.NewKVStoreService(govKey)
	sb := collections.NewSchemaBuilder(storeService)
	constitutionCollection := collections.NewItem(sb, v5.ConstitutionKey, "constitution", collections.StringValue)
	participationEMACollection := collections.NewItem(sb, v5.ParticipationEMAKey, "participation_ema", sdk.LegacyDecValue)
	constitutionAmendmentParticipationEMACollection := collections.NewItem(sb, v5.ConstitutionAmendmentParticipationEMAKey, "constitution_amendment_participation_ema", sdk.LegacyDecValue)
	lawParticipationEMACollection := collections.NewItem(sb, v5.LawParticipationEMAKey, "law_participation_ema", sdk.LegacyDecValue)

	var params v1.Params
	bz := store.Get(v5.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)

	// Run migrations.
	err := v5.MigrateStore(ctx, storeService, cdc, constitutionCollection, participationEMACollection, constitutionAmendmentParticipationEMACollection, lawParticipationEMACollection)
	require.NoError(t, err)

	// Check params
	bz = store.Get(v5.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, v1.DefaultParams().MinDepositRatio, params.MinDepositRatio)

	// Check constitution
	result, err := constitutionCollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, "This chain has no constitution.", result)

	// Check participation EMA values
	expectedEMA := math.LegacyNewDecWithPrec(12, 2)
	participationEMA, err := participationEMACollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedEMA, participationEMA)

	constitutionAmendmentEMA, err := constitutionAmendmentParticipationEMACollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedEMA, constitutionAmendmentEMA)

	lawEMA, err := lawParticipationEMACollection.Get(ctx)
	require.NoError(t, err)
	require.Equal(t, expectedEMA, lawEMA)
}
