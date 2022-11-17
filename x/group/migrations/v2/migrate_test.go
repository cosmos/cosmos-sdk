package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/cosmos/cosmos-sdk/x/group/internal/orm"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	v2 "github.com/cosmos/cosmos-sdk/x/group/migrations/v2"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
)

var (
	policyAddr  = sdk.MustAccAddressFromBech32("cosmos1q32tjg5qm3n9fj8wjgpd7gl98prefntrckjkyvh8tntp7q33zj0s5tkjrk")
	accountAddr = sdk.AccAddress("addr2_______________")
)

func TestMigrate(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, groupmodule.AppModuleBasic{}).Codec
	storeKey := sdk.NewKVStoreKey(v2.ModuleName)
	tKey := sdk.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	accountKeeper := createOldPolicyAccount(ctx, storeKey, cdc)
	groupPolicyTable, err := createGroupPolicies(ctx, storeKey, cdc)
	require.NoError(t, err)

	oldAcc := accountKeeper.GetAccount(ctx, policyAddr)

	require.NoError(t, v2.Migrate(ctx, storeKey, accountKeeper, orm.NewSequence(v2.GroupPolicyTableSeqPrefix), groupPolicyTable))
	newAcc := accountKeeper.GetAccount(ctx, policyAddr)

	require.NotEqual(t, oldAcc, newAcc)
	require.True(t, func() bool { _, ok := oldAcc.(*authtypes.ModuleAccount); return ok }())
	require.True(t, func() bool { _, ok := newAcc.(*authtypes.BaseAccount); return ok }())
	require.Equal(t, oldAcc.GetAddress(), newAcc.GetAddress())
	require.Equal(t, oldAcc.GetAccountNumber(), newAcc.GetAccountNumber())
	require.Equal(t, newAcc.GetPubKey().Address().Bytes(), newAcc.GetAddress().Bytes())
}

func createGroupPolicies(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.Codec) (orm.PrimaryKeyTable, error) {
	groupPolicyTable, err := orm.NewPrimaryKeyTable([2]byte{groupkeeper.GroupPolicyTablePrefix}, &group.GroupPolicyInfo{}, cdc)
	if err != nil {
		panic(err.Error())
	}

	groupPolicyInfo, err := group.NewGroupPolicyInfo(policyAddr, 1, accountAddr, "", 1, group.NewPercentageDecisionPolicy("1", 1, 1), ctx.BlockTime())
	if err != nil {
		return orm.PrimaryKeyTable{}, err
	}

	if err := groupPolicyTable.Create(ctx.KVStore(storeKey), &groupPolicyInfo); err != nil {
		return orm.PrimaryKeyTable{}, err
	}

	return *groupPolicyTable, nil
}

// createOldPolicyAccount re-creates the group policy account using a module account
func createOldPolicyAccount(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.Codec) group.AccountKeeper {
	accountKeeper := authkeeper.NewAccountKeeper(cdc, storeKey, authtypes.ProtoBaseAccount, nil, sdk.Bech32MainPrefix, accountAddr.String())
	acc := accountKeeper.NewAccount(ctx, &authtypes.ModuleAccount{
		BaseAccount: &authtypes.BaseAccount{
			Address: policyAddr.String(),
		},
		Name: policyAddr.String(),
	})
	accountKeeper.SetAccount(ctx, acc)

	return accountKeeper
}
