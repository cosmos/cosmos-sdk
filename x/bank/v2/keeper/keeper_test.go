package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"cosmossdk.io/core/address"
	"cosmossdk.io/core/header"
	coretesting "cosmossdk.io/core/testing"
	"cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"cosmossdk.io/x/bank/v2/keeper"
	banktestutil "cosmossdk.io/x/bank/v2/testutil"
	banktypes "cosmossdk.io/x/bank/v2/types"

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

func (suite *KeeperTestSuite) mockSendCoinsFromModuleToAccount(moduleAcc *authtypes.ModuleAccount, recv sdk.AccAddress) {
	suite.authKeeper.EXPECT().GetModuleAddress(moduleAcc.Name).Return(moduleAcc.GetAddress())
}

func (suite *KeeperTestSuite) mockFundAccount(receiver sdk.AccAddress) {
	suite.mockMintCoins(mintAcc)
	suite.mockSendCoinsFromModuleToAccount(mintAcc, receiver)
}

type KeeperTestSuite struct {
	suite.Suite

	ctx          context.Context
	bankKeeper   keeper.Keeper
	authKeeper   *banktestutil.MockAccountKeeper
	addressCodec address.Codec
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
	suite.addressCodec = ac
}

func (suite *KeeperTestSuite) TestSendCoins_Acount_To_Account() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	sendAmt := sdk.NewCoins(newFooCoin(10), newBarCoin(10))

	// Try send with empty balances
	suite.authKeeper.EXPECT().GetModuleAddress(string(accAddrs[0])).Return(nil)
	suite.authKeeper.EXPECT().GetModuleAddress(string(accAddrs[1])).Return(nil)
	err := suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sendAmt)
	require.Error(err)

	// Set balances for acc0 and then try send to acc1
	suite.mockFundAccount(accAddrs[0])
	suite.authKeeper.EXPECT().GetModuleAddress(string(accAddrs[0])).Return(nil)
	require.NoError(banktestutil.FundAccount(ctx, suite.bankKeeper, accAddrs[0], balances))

	suite.authKeeper.EXPECT().GetModuleAddress(string(accAddrs[0])).Return(nil)
	suite.authKeeper.EXPECT().GetModuleAddress(string(accAddrs[1])).Return(nil)
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

func (suite *KeeperTestSuite) TestSendCoins_Module_To_Account() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	acc4Str, _ := suite.addressCodec.BytesToString(accAddrs[4])

	suite.mockMintCoins(mintAcc)
	require.NoError(suite.bankKeeper.MintCoins(ctx, banktypes.MintModuleName, balances))

	// Try send from invalid module
	suite.authKeeper.EXPECT().GetModuleAddress("invalid").Return(nil)
	suite.authKeeper.EXPECT().GetModuleAddress(acc4Str).Return(nil)
	err := suite.bankKeeper.SendCoins(ctx, []byte("invalid"), []byte(acc4Str), balances)
	require.Error(err)

	// Send from mint module
	suite.authKeeper.EXPECT().GetModuleAddress(mintAcc.Name).Return(mintAcc.GetAddress())
	suite.authKeeper.EXPECT().GetModuleAddress(string(accAddrs[4])).Return(nil)
	err = suite.bankKeeper.SendCoins(ctx, []byte(banktypes.MintModuleName), accAddrs[4], balances)
	require.NoError(err)

	// Check balances
	acc4FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[4], fooDenom)
	require.Equal(acc4FooBalance.Amount, math.NewInt(100))
	acc4BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[4], barDenom)
	require.Equal(acc4BarBalance.Amount, math.NewInt(50))
	mintFooBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), fooDenom)
	require.Equal(mintFooBalance.Amount, math.NewInt(0))
	mintBarBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), barDenom)
	require.Equal(mintBarBalance.Amount, math.NewInt(0))
}

func (suite *KeeperTestSuite) TestSendCoins_Module_To_Module() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(200), newBarCoin(100))

	suite.mockMintCoins(mintAcc)
	require.NoError(suite.bankKeeper.MintCoins(ctx, banktypes.MintModuleName, balances))

	// Try send to invalid module
	// In this case it will create a new account
	suite.authKeeper.EXPECT().GetModuleAddress(mintAcc.Name).Return(mintAcc.GetAddress())
	suite.authKeeper.EXPECT().GetModuleAddress("invalid").Return(nil)
	err := suite.bankKeeper.SendCoins(ctx, []byte(mintAcc.Name), []byte("invalid"), sdk.NewCoins(newFooCoin(100), newBarCoin(50)))
	require.NoError(err)

	// Send from mint module to burn module
	suite.authKeeper.EXPECT().GetModuleAddress(mintAcc.Name).Return(mintAcc.GetAddress())
	suite.authKeeper.EXPECT().GetModuleAddress(burnerAcc.Name).Return(burnerAcc.GetAddress())
	err = suite.bankKeeper.SendCoins(ctx, []byte(mintAcc.Name), []byte(burnerAcc.Name), sdk.NewCoins(newFooCoin(100), newBarCoin(50)))
	require.NoError(err)

	// Check balances
	burnerFooBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), fooDenom)
	require.Equal(burnerFooBalance.Amount, math.NewInt(100))
	burnerBarBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), barDenom)
	require.Equal(burnerBarBalance.Amount, math.NewInt(50))
	invalidFooBalance := suite.bankKeeper.GetBalance(ctx, []byte("invalid"), fooDenom)
	require.Equal(invalidFooBalance.Amount, math.NewInt(100))
	invalidBarBalance := suite.bankKeeper.GetBalance(ctx, []byte("invalid"), barDenom)
	require.Equal(invalidBarBalance.Amount, math.NewInt(50))
	mintFooBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), fooDenom)
	require.Equal(mintFooBalance.Amount, math.NewInt(0))
	mintBarBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), barDenom)
	require.Equal(mintBarBalance.Amount, math.NewInt(0))
}
