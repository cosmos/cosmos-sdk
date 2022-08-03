package v4_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/gov"
	v4 "github.com/cosmos/cosmos-sdk/x/gov/migrations/v4"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
)

type mockSubspace struct {
	dp v1.DepositParams
	vp v1.VotingParams
	tp v1.TallyParams
}

func newMockSubspace(p v1.Params) mockSubspace {
	return mockSubspace{
		dp: v1.DepositParams{
			MinDeposit:       p.MinDeposit,
			MaxDepositPeriod: p.MaxDepositPeriod,
		},
		vp: v1.VotingParams{
			VotingPeriod: p.VotingPeriod,
		},
		tp: v1.TallyParams{
			Quorum:        p.Quorum,
			Threshold:     p.Threshold,
			VetoThreshold: p.VetoThreshold,
		},
	}
}

func (ms mockSubspace) Get(ctx sdk.Context, key []byte, ptr interface{}) {
	switch string(key) {
	case string(v1.ParamStoreKeyDepositParams):
		*ptr.(*v1.DepositParams) = ms.dp
	case string(v1.ParamStoreKeyVotingParams):
		*ptr.(*v1.VotingParams) = ms.vp
	case string(v1.ParamStoreKeyTallyParams):
		*ptr.(*v1.TallyParams) = ms.tp
	}
}

func TestMigrateStore(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(upgrade.AppModuleBasic{}, gov.AppModuleBasic{}).Codec
	govKey := sdk.NewKVStoreKey("gov")
	ctx := testutil.DefaultContext(govKey, sdk.NewTransientStoreKey("transient_test"))
	store := ctx.KVStore(govKey)

	legacySubspace := newMockSubspace(v1.DefaultParams())
	// Run migrations.
	err := v4.MigrateStore(ctx, govKey, legacySubspace, cdc)
	require.NoError(t, err)

	var params v1.Params
	bz := store.Get(v4.ParamsKey)
	require.NoError(t, cdc.Unmarshal(bz, &params))
	require.NotNil(t, params)
	require.Equal(t, legacySubspace.dp.MinDeposit, params.MinDeposit)
	require.Equal(t, legacySubspace.dp.MaxDepositPeriod, params.MaxDepositPeriod)
	require.Equal(t, legacySubspace.vp.VotingPeriod, params.VotingPeriod)
	require.Equal(t, legacySubspace.tp.Quorum, params.Quorum)
	require.Equal(t, legacySubspace.tp.Threshold, params.Threshold)
	require.Equal(t, legacySubspace.tp.VetoThreshold, params.VetoThreshold)
	require.Equal(t, sdk.ZeroDec().String(), params.MinInitialDepositRatio)
}
