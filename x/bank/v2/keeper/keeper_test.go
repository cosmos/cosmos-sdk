package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/header"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	banktypes "cosmossdk.io/x/bank/types"
	"cosmossdk.io/x/bank/v2/keeper"
	banktestutil "cosmossdk.io/x/bank/v2/testutil"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

const (
	fooDenom            = "foo"
	barDenom            = "bar"
	ibcPath             = "transfer/channel-0"
	ibcBaseDenom        = "farboo"
	metaDataDescription = "IBC Token from %s"
	initialPower        = int64(100)
	holder              = "holder"
	multiPerm           = "multiple permissions account"
	randomPerm          = "random permission"
)

var (
	holderAcc    = authtypes.NewEmptyModuleAccount(holder)
	randomAcc    = authtypes.NewEmptyModuleAccount(randomPerm)
	burnerAcc    = authtypes.NewEmptyModuleAccount(authtypes.Burner, authtypes.Burner, authtypes.Staking)
	minterAcc    = authtypes.NewEmptyModuleAccount(authtypes.Minter, authtypes.Minter)
	mintAcc      = authtypes.NewEmptyModuleAccount(banktypes.MintModuleName, authtypes.Minter)
	multiPermAcc = authtypes.NewEmptyModuleAccount(multiPerm, authtypes.Burner, authtypes.Minter, authtypes.Staking)

	baseAcc = authtypes.NewBaseAccountWithAddress(sdk.AccAddress([]byte("baseAcc")))

	accAddrs = []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
		sdk.AccAddress([]byte("addr2_______________")),
		sdk.AccAddress([]byte("addr3_______________")),
		sdk.AccAddress([]byte("addr4_______________")),
		sdk.AccAddress([]byte("addr5_______________")),
	}

	// The default power validators are initialized to have within tests
	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))
)

func newFooCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(fooDenom, amt)
}

func newBarCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(barDenom, amt)
}

func (suite *KeeperTestSuite) mockMintCoins(moduleAcc *authtypes.ModuleAccount) {
	suite.authKeeper.EXPECT().GetModuleAccount(suite.ctx, moduleAcc.Name).Return(moduleAcc)
}

func (suite *KeeperTestSuite) mockSendCoinsFromModuleToAccount(moduleAcc *authtypes.ModuleAccount, _ sdk.AccAddress) {
	suite.authKeeper.EXPECT().GetModuleAddress(moduleAcc.Name).Return(moduleAcc.GetAddress())
	suite.authKeeper.EXPECT().GetAccount(suite.ctx, moduleAcc.GetAddress()).Return(moduleAcc)
}

func (suite *KeeperTestSuite) mockFundAccount(receiver sdk.AccAddress) {
	suite.mockMintCoins(mintAcc)
	suite.mockSendCoinsFromModuleToAccount(mintAcc, receiver)
}

func (suite *KeeperTestSuite) mockSendCoins(ctx context.Context, sender sdk.AccountI, _ sdk.AccAddress) {
	suite.authKeeper.EXPECT().GetAccount(ctx, sender.GetAddress()).Return(sender)
}

type KeeperTestSuite struct {
	suite.Suite

	ctx        context.Context
	bankKeeper keeper.Keeper
	authKeeper *banktestutil.MockAccountKeeper
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (suite *KeeperTestSuite) SetupTest() {
	key := storetypes.NewKVStoreKey(banktypes.StoreKey)
	testCtx := testutil.DefaultContextWithDB(suite.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	ctx := testCtx.Ctx.WithHeaderInfo(header.Info{Time: time.Now()})
	encCfg := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})

	env := runtime.NewEnvironment(runtime.NewKVStoreService(key), coretesting.NewNopLogger())

	ac := codectestutil.CodecOptions{}.GetAddressCodec()
	authority, err := ac.BytesToString(authtypes.NewModuleAddress("gov"))
	suite.Require().NoError(err)

	// gomock initializations
	ctrl := gomock.NewController(suite.T())
	authKeeper := banktestutil.NewMockAccountKeeper(ctrl)
	suite.ctx = ctx
	suite.authKeeper = authKeeper
	suite.bankKeeper = *keeper.NewKeeper(
		[]byte(authority),
		ac,
		env,
		encCfg.Codec,
		authKeeper,
	)
}

func (suite *KeeperTestSuite) TestSendCoins() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	sendAmt := sdk.NewCoins(newFooCoin(10), newBarCoin(10))

	acc0 := authtypes.NewBaseAccountWithAddress(accAddrs[0])

	// Try send with empty balances
	suite.authKeeper.EXPECT().GetAccount(suite.ctx, accAddrs[0]).Return(acc0)
	err := suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sendAmt)
	require.Error(err)

	// Set balances for acc0 and then try send to acc1
	suite.mockFundAccount(accAddrs[0])
	require.NoError(banktestutil.FundAccount(ctx, suite.bankKeeper, accAddrs[0], balances))
	suite.mockSendCoins(ctx, acc0, accAddrs[1])
	require.NoError(suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sendAmt))

	// Check balances
	acc0FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], fooDenom)
	require.Equal(acc0FooBalance.Amount, math.NewInt(90))
	acc0BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], barDenom)
	require.Equal(acc0BarBalance.Amount, math.NewInt(40))
	acc1FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[1], fooDenom)
	require.Equal(acc1FooBalance.Amount, math.NewInt(10))
	acc1BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[1], barDenom)
	require.Equal(acc1BarBalance.Amount, math.NewInt(10))

}
