package v2_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/log"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/auth"
	authkeeper "cosmossdk.io/x/auth/keeper"
	authtypes "cosmossdk.io/x/auth/types"
	"cosmossdk.io/x/group"
	"cosmossdk.io/x/group/internal/orm"
	groupkeeper "cosmossdk.io/x/group/keeper"
	v2 "cosmossdk.io/x/group/migrations/v2"
	groupmodule "cosmossdk.io/x/group/module"

	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
)

var (
	policies      = []sdk.AccAddress{policyAddr1, policyAddr2, policyAddr3}
	policyAddr1   = sdk.MustAccAddressFromBech32("cosmos1q32tjg5qm3n9fj8wjgpd7gl98prefntrckjkyvh8tntp7q33zj0s5tkjrk")
	policyAddr2   = sdk.MustAccAddressFromBech32("cosmos1afk9zr2hn2jsac63h4hm60vl9z3e5u69gndzf7c99cqge3vzwjzsfwkgpd")
	policyAddr3   = sdk.MustAccAddressFromBech32("cosmos1dlszg2sst9r69my4f84l3mj66zxcf3umcgujys30t84srg95dgvsmn3jeu")
	authorityAddr = sdk.AccAddress("authority")
)

func TestMigrate(t *testing.T) {
	cdc := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{}, auth.AppModule{}, groupmodule.AppModule{}).Codec
	storeKey := storetypes.NewKVStoreKey(v2.ModuleName)
	storeService := runtime.NewKVStoreService(storeKey)
	tKey := storetypes.NewTransientStoreKey("transient_test")
	ctx := testutil.DefaultContext(storeKey, tKey)

	oldAccs, accountKeeper := createOldPolicyAccount(ctx, storeKey, cdc, policies)
	groupPolicyTable, groupPolicySeq, err := createGroupPolicies(ctx, storeService, cdc, policies)
	require.NoError(t, err)

	require.NoError(t, v2.Migrate(ctx, storeService, accountKeeper, groupPolicySeq, groupPolicyTable))
	for i, policyAddr := range policies {
		oldAcc := oldAccs[i]
		newAcc := accountKeeper.GetAccount(ctx, policyAddr)
		require.NotEqual(t, oldAcc, newAcc)
		require.True(t, func() bool { _, ok := newAcc.(*authtypes.BaseAccount); return ok }())
		require.Equal(t, oldAcc.GetAddress(), newAcc.GetAddress())
		require.Equal(t, int(oldAcc.GetAccountNumber()), int(newAcc.GetAccountNumber()))
		require.Equal(t, newAcc.GetPubKey().Address().Bytes(), newAcc.GetAddress().Bytes())
	}
}

func createGroupPolicies(ctx sdk.Context, storeService corestore.KVStoreService, cdc codec.Codec, policies []sdk.AccAddress) (orm.PrimaryKeyTable, orm.Sequence, error) {
	groupPolicyTable, err := orm.NewPrimaryKeyTable([2]byte{groupkeeper.GroupPolicyTablePrefix}, &group.GroupPolicyInfo{}, cdc)
	if err != nil {
		panic(err.Error())
	}

	groupPolicySeq := orm.NewSequence(v2.GroupPolicyTableSeqPrefix)
	kvStore := storeService.OpenKVStore(ctx)

	for _, policyAddr := range policies {
		groupPolicyInfo, err := group.NewGroupPolicyInfo(policyAddr, 1, authorityAddr, "", 1, group.NewPercentageDecisionPolicy("1", 1, 1), ctx.HeaderInfo().Time)
		if err != nil {
			return orm.PrimaryKeyTable{}, orm.Sequence{}, err
		}

		if err := groupPolicyTable.Create(kvStore, &groupPolicyInfo); err != nil {
			return orm.PrimaryKeyTable{}, orm.Sequence{}, err
		}

		groupPolicySeq.NextVal(kvStore)
	}

	return *groupPolicyTable, groupPolicySeq, nil
}

// createOldPolicyAccount re-creates the group policy account using a module account
func createOldPolicyAccount(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.Codec, policies []sdk.AccAddress) ([]*authtypes.ModuleAccount, group.AccountKeeper) {
	accountKeeper := authkeeper.NewAccountKeeper(runtime.NewEnvironment(runtime.NewKVStoreService(storeKey.(*storetypes.KVStoreKey)), log.NewNopLogger()), cdc, authtypes.ProtoBaseAccount, nil, addresscodec.NewBech32Codec(sdk.Bech32MainPrefix), sdk.Bech32MainPrefix, authorityAddr.String())

	oldPolicyAccounts := make([]*authtypes.ModuleAccount, len(policies))
	for i, policyAddr := range policies {
		acc := accountKeeper.NewAccount(ctx, &authtypes.ModuleAccount{
			BaseAccount: &authtypes.BaseAccount{
				Address: policyAddr.String(),
			},
			Name: policyAddr.String(),
		})
		accountKeeper.SetAccount(ctx, acc)
		oldPolicyAccounts[i] = acc.(*authtypes.ModuleAccount)
	}

	return oldPolicyAccounts, accountKeeper
}
