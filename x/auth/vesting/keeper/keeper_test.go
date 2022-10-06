package keeper_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/baseapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/exported"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/keeper"
	vestingtestutil "github.com/cosmos/cosmos-sdk/x/auth/vesting/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting/types"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
)

const (
	feeDenom   = "fee"
	stakeDenom = "stake"
	holder     = "holder"
	multiPerm  = "multiple permissions account"
	randomPerm = "random permission"
)

type KeeperTestSuite struct {
	suite.Suite

	now           time.Time
	endTime       time.Time
	encCfg        moduletestutil.TestEncodingConfig
	vestingKeeper keeper.VestingKeeper
	accountKeeper *vestingtestutil.MockAccountKeeper
	bankKeeper    *vestingtestutil.MockBankKeeper
	ctx           sdk.Context

	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (suite *KeeperTestSuite) SetupTest() {
	key := sdk.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, sdk.NewTransientStoreKey("transient_test"))
	testCtx.CMS.MountStoreWithDB(sdk.NewKVStoreKey(authtypes.StoreKey), storetypes.StoreTypeIAVL, testCtx.DB)
	suite.ctx = testCtx.Ctx.WithBlockHeader(tmproto.Header{})

	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})

	suite.now = time.Now()
	suite.endTime = suite.now.Add(24 * time.Hour)

	ctrl := gomock.NewController(suite.T())
	accountKeeper := vestingtestutil.NewMockAccountKeeper(ctrl)
	for _, addr := range []string{"Alice", "Bob", "Carol"} {
		accAddress := sdk.AccAddress(addr)

		bacc, origCoins := initBaseAccount(accAddress)
		account := types.NewContinuousVestingAccount(bacc, origCoins, suite.now.Unix(), suite.endTime.Unix())
		accountKeeper.EXPECT().GetAccount(suite.ctx, accAddress).Return(account).AnyTimes()
	}

	suite.accountKeeper = accountKeeper
	suite.bankKeeper = vestingtestutil.NewMockBankKeeper(ctrl)
	suite.vestingKeeper = keeper.NewVestingKeeper(
		suite.accountKeeper,
		suite.bankKeeper,
		key,
	)

	types.RegisterInterfaces(encCfg.InterfaceRegistry)

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, keeper.Querier{&suite.vestingKeeper})
	queryClient := types.NewQueryClient(queryHelper)

	suite.queryClient = queryClient
	suite.encCfg = encCfg
	suite.msgServer = keeper.NewMsgServerImpl(&suite.vestingKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) TestAddVestingAccount() {
	ctx := suite.ctx

	// vesting accouts slice should be empty before we add vesting account.
	var vestingAccounts []exported.VestingAccount
	suite.vestingKeeper.IterateVestingAccounts(ctx, func(account exported.VestingAccount) (stop bool) {
		vestingAccounts = append(vestingAccounts, account)
		return false
	})
	suite.Require().Empty(vestingAccounts)

	// add vesting accounts.
	for _, addr := range []string{"Alice", "Bob", "Carol"} {
		bacc, origCoins := initBaseAccount(sdk.AccAddress(addr))
		cva := types.NewContinuousVestingAccount(bacc, origCoins, suite.now.Unix(), suite.endTime.Unix())

		suite.accountKeeper.EXPECT().SetAccount(ctx, cva)
		suite.accountKeeper.SetAccount(ctx, cva)
		suite.vestingKeeper.AddVestingAccount(ctx, cva.GetAddress())
	}

	// we should fetch three vesting accounts.
	suite.vestingKeeper.IterateVestingAccounts(ctx, func(account exported.VestingAccount) (stop bool) {
		vestingAccounts = append(vestingAccounts, account)
		return false
	})
	suite.Require().Equal(3, len(vestingAccounts))
}

func (suite *KeeperTestSuite) TestIterateVestingAccounts() {
	ctx := suite.ctx
	for _, addr := range []string{"Alice", "Bob", "Carol"} {
		bacc, origCoins := initBaseAccount(sdk.AccAddress(addr))
		cva := types.NewContinuousVestingAccount(bacc, origCoins, suite.now.Unix(), suite.endTime.Unix())

		suite.accountKeeper.EXPECT().SetAccount(ctx, cva)
		suite.accountKeeper.SetAccount(ctx, cva)
		suite.vestingKeeper.AddVestingAccount(ctx, cva.GetAddress())
	}

	var totalLockedCoins sdk.Coins
	suite.vestingKeeper.IterateVestingAccounts(ctx, func(account exported.VestingAccount) (stop bool) {
		totalLockedCoins = totalLockedCoins.Add(account.LockedCoins(ctx.BlockTime())...)
		return false
	})
	suite.Require().Equal(
		sdk.Coins{sdk.NewInt64Coin(feeDenom, 3000), sdk.NewInt64Coin(stakeDenom, 300)},
		totalLockedCoins,
	)
}

func initBaseAccount(addr sdk.AccAddress) (*authtypes.BaseAccount, sdk.Coins) {
	origCoins := sdk.Coins{sdk.NewInt64Coin(feeDenom, 1000), sdk.NewInt64Coin(stakeDenom, 100)}
	bacc := authtypes.NewBaseAccountWithAddress(addr)

	return bacc, origCoins
}
