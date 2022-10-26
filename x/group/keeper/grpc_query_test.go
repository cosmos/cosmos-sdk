package keeper_test

import (
	"context"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	"github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/golang/mock/gomock"
	"github.com/tendermint/tendermint/libs/log"

	grouptestutil "github.com/cosmos/cosmos-sdk/x/group/testutil"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/testutil"
)

func TestQueryGroupsByMember(t *testing.T) {
	var (
		groupKeeper       groupkeeper.Keeper
		interfaceRegistry codectypes.InterfaceRegistry
	)

	key := sdk.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, sdk.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(module.AppModuleBasic{})

	ctx := testCtx.Ctx

	sdkCtx := sdk.WrapSDKContext(ctx)

	bApp := baseapp.NewBaseApp(
		"group",
		log.NewNopLogger(),
		testCtx.DB,
		encCfg.TxConfig.TxDecoder(),
	)

	banktypes.RegisterInterfaces(encCfg.InterfaceRegistry)

	addrs := simtestutil.CreateIncrementalAccounts(6)
	ctrl := gomock.NewController(t)
	accountKeeper := grouptestutil.NewMockAccountKeeper(ctrl)
	accountKeeper.EXPECT().GetAccount(gomock.Any(), addrs[0]).Return(authtypes.NewBaseAccountWithAddress(addrs[0])).AnyTimes()
	accountKeeper.EXPECT().GetAccount(gomock.Any(), addrs[1]).Return(authtypes.NewBaseAccountWithAddress(addrs[1])).AnyTimes()
	accountKeeper.EXPECT().GetAccount(gomock.Any(), addrs[2]).Return(authtypes.NewBaseAccountWithAddress(addrs[2])).AnyTimes()
	accountKeeper.EXPECT().GetAccount(gomock.Any(), addrs[3]).Return(authtypes.NewBaseAccountWithAddress(addrs[3])).AnyTimes()
	accountKeeper.EXPECT().GetAccount(gomock.Any(), addrs[4]).Return(authtypes.NewBaseAccountWithAddress(addrs[4])).AnyTimes()
	accountKeeper.EXPECT().GetAccount(gomock.Any(), addrs[5]).Return(authtypes.NewBaseAccountWithAddress(addrs[5])).AnyTimes()

	groupKeeper = groupkeeper.NewKeeper(key, encCfg.Codec, bApp.MsgServiceRouter(), accountKeeper, group.DefaultConfig())

	// Initial group, group policy and balance setup
	members := []group.MemberRequest{
		{Address: addrs[2].String(), Weight: "1"}, {Address: addrs[3].String(), Weight: "2"},
	}

	_, err := groupKeeper.CreateGroup(sdkCtx, &group.MsgCreateGroup{
		Admin:   addrs[0].String(),
		Members: members,
	})
	require.NoError(t, err)

	members = []group.MemberRequest{
		{Address: addrs[3].String(), Weight: "1"}, {Address: addrs[4].String(), Weight: "2"},
	}
	_, err = groupKeeper.CreateGroup(sdkCtx, &group.MsgCreateGroup{
		Admin:   addrs[1].String(),
		Members: members,
	})
	require.NoError(t, err)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, interfaceRegistry)
	group.RegisterQueryServer(queryHelper, groupKeeper)
	queryClient := group.NewQueryClient(queryHelper)

	// not part of any group
	resp, err := queryClient.GroupsByMember(context.Background(), &group.QueryGroupsByMemberRequest{
		Address: addrs[5].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 0)

	// expect one group
	resp, err = queryClient.GroupsByMember(context.Background(), &group.QueryGroupsByMemberRequest{
		Address: addrs[4].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 1)

	// expect two groups
	resp, err = queryClient.GroupsByMember(context.Background(), &group.QueryGroupsByMemberRequest{
		Address: addrs[3].String(),
	})
	require.NoError(t, err)
	require.Len(t, resp.Groups, 2)
}
