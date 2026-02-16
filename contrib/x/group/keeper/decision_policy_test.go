package keeper

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	coreaddress "cosmossdk.io/core/address"
	"cosmossdk.io/log/v2"
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	group "github.com/cosmos/cosmos-sdk/contrib/x/group"
	"github.com/cosmos/cosmos-sdk/contrib/x/group/internal/orm"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank"
)

// mockAccountKeeperForDecisionPolicyTest implements group.AccountKeeper for testing.
// GetAccount returns nil so CreateGroupPolicy can create new group policy accounts
// (it loops until it finds an address with no existing account).
type mockAccountKeeperForDecisionPolicyTest struct {
	addressCodec coreaddress.Codec
}

func (m mockAccountKeeperForDecisionPolicyTest) AddressCodec() coreaddress.Codec {
	return m.addressCodec
}

func (m mockAccountKeeperForDecisionPolicyTest) NewAccount(ctx context.Context, i sdk.AccountI) sdk.AccountI {
	return i
}

func (m mockAccountKeeperForDecisionPolicyTest) GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI {
	return nil
}

func (m mockAccountKeeperForDecisionPolicyTest) SetAccount(ctx context.Context, i sdk.AccountI) {}

func (m mockAccountKeeperForDecisionPolicyTest) RemoveAccount(ctx context.Context, acc sdk.AccountI) {
}

var _ group.AccountKeeper = mockAccountKeeperForDecisionPolicyTest{}

// TestLeaveGroup_MalformedPolicyUnpacking simulates malformed policy in store:
// a group policy whose DecisionPolicy Any contains wrong type. LeaveGroup
// triggers validateDecisionPolicies which calls GetDecisionPolicy; we expect
// defensive error return instead of panic.
func TestLeaveGroup_MalformedPolicyUnpacking(t *testing.T) {
	key := storetypes.NewKVStoreKey(group.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))
	encCfg := moduletestutil.MakeTestEncodingConfig(bank.AppModuleBasic{})
	group.RegisterInterfaces(encCfg.InterfaceRegistry)

	addrs := simtestutil.CreateIncrementalAccounts(4)
	accountKeeper := mockAccountKeeperForDecisionPolicyTest{
		addressCodec: address.NewBech32Codec("cosmos"),
	}

	bApp := baseapp.NewBaseApp("group", log.NewNopLogger(), testCtx.DB, encCfg.TxConfig.TxDecoder())
	bApp.SetInterfaceRegistry(encCfg.InterfaceRegistry)

	k := NewKeeper(key, encCfg.Codec, bApp.MsgServiceRouter(), accountKeeper, group.DefaultConfig())
	ctx := testCtx.Ctx
	sdkCtx := sdk.UnwrapSDKContext(ctx)

	// Create group with two members
	members := []group.MemberRequest{
		{Address: addrs[0].String(), Weight: "1"},
		{Address: addrs[1].String(), Weight: "1"},
	}
	groupRes, err := k.CreateGroup(ctx, &group.MsgCreateGroup{
		Admin:   addrs[0].String(),
		Members: members,
	})
	require.NoError(t, err)
	groupID := groupRes.GroupId

	// Create group policy
	policy := group.NewThresholdDecisionPolicy("1", time.Second, 5*time.Second)
	policyReq := &group.MsgCreateGroupPolicy{
		Admin:   addrs[0].String(),
		GroupId: groupID,
	}
	require.NoError(t, policyReq.SetDecisionPolicy(policy))

	nextAccVal := k.GetGroupPolicySeq(sdkCtx) + 1
	derivationKey := make([]byte, 8)
	binary.BigEndian.PutUint64(derivationKey, nextAccVal)
	policyRes, err := k.CreateGroupPolicy(ctx, policyReq)
	require.NoError(t, err)
	policyAddr := policyRes.Address

	// Load policy from store, corrupt its DecisionPolicy, save back.
	// We must write directly to the store because the ORM's Update validates
	// the policy (calls ValidateBasic->GetDecisionPolicy) and would reject our corrupt data.
	var policyInfo group.GroupPolicyInfo
	rowID := orm.PrimaryKey(&group.GroupPolicyInfo{Address: policyAddr})
	err = k.groupPolicyTable.GetOne(sdkCtx.KVStore(k.key), rowID, &policyInfo)
	require.NoError(t, err)

	coin := sdk.NewInt64Coin("stake", 1)
	wrongAny, err := codectypes.NewAnyWithValue(&coin)
	require.NoError(t, err)
	policyInfo.DecisionPolicy = wrongAny

	encoded, err := k.cdc.Marshal(&policyInfo)
	require.NoError(t, err)
	// Match the table's 2-byte prefix
	tablePrefix := [2]byte{GroupPolicyTablePrefix}
	pStore := prefix.NewStore(sdkCtx.KVStore(k.key), tablePrefix[:])
	pStore.Set(rowID, encoded)

	// LeaveGroup triggers validateDecisionPolicies, which loads policies and calls GetDecisionPolicy.
	// With corrupted policy (wrong type in DecisionPolicy Any), we expect defensive error instead of panic.
	_, err = k.LeaveGroup(ctx, &group.MsgLeaveGroup{
		Address: addrs[1].String(),
		GroupId: groupID,
	})
	require.Error(t, err)
}
