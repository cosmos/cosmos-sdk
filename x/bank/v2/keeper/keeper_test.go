package keeper_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

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
	fooDenom = "foo"
	barDenom = "bar"
)

var (
	burnerAcc = authtypes.NewEmptyModuleAccount(authtypes.Burner, authtypes.Burner, authtypes.Staking)
	mintAcc   = authtypes.NewEmptyModuleAccount(banktypes.MintModuleName, authtypes.Minter)

	accAddrs = []sdk.AccAddress{
		sdk.AccAddress([]byte("addr1_______________")),
		sdk.AccAddress([]byte("addr2_______________")),
		sdk.AccAddress([]byte("addr3_______________")),
		sdk.AccAddress([]byte("addr4_______________")),
		sdk.AccAddress([]byte("addr5_______________")),
	}
)

func newFooCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(fooDenom, amt)
}

func newBarCoin(amt int64) sdk.Coin {
	return sdk.NewInt64Coin(barDenom, amt)
}

type KeeperTestSuite struct {
	suite.Suite

	ctx          context.Context
	bankKeeper   keeper.Keeper
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
	authority := authtypes.NewModuleAddress("gov")

	suite.ctx = ctx
	suite.bankKeeper = *keeper.NewKeeper(
		authority,
		ac,
		env,
		encCfg.Codec,
	)
	suite.addressCodec = ac
}

func (suite *KeeperTestSuite) TestSendCoins_Acount_To_Account() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	sendAmt := sdk.NewCoins(newFooCoin(10), newBarCoin(10))

	// Try send with empty balances
	err := suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sendAmt)
	require.Error(err)

	// Set balances for acc0 and then try send to acc1
	require.NoError(banktestutil.FundAccount(ctx, suite.bankKeeper, accAddrs[0], balances))
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

func (suite *KeeperTestSuite) TestSendCoins_Acount_To_Module() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	sendAmt := sdk.NewCoins(newFooCoin(10), newBarCoin(10))

	// Try send with empty balances
	err := suite.bankKeeper.SendCoins(ctx, accAddrs[0], burnerAcc.GetAddress(), sendAmt)
	require.Error(err)

	// Set balances for acc0 and then try send to acc1
	require.NoError(banktestutil.FundAccount(ctx, suite.bankKeeper, accAddrs[0], balances))
	require.NoError(suite.bankKeeper.SendCoins(ctx, accAddrs[0], burnerAcc.GetAddress(), sendAmt))

	// Check balances
	acc0FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], fooDenom)
	require.Equal(acc0FooBalance.Amount, math.NewInt(90))
	acc0BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], barDenom)
	require.Equal(acc0BarBalance.Amount, math.NewInt(40))
	burnerFooBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), fooDenom)
	require.Equal(burnerFooBalance.Amount, math.NewInt(10))
	burnerBarBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), barDenom)
	require.Equal(burnerBarBalance.Amount, math.NewInt(10))
}

func (suite *KeeperTestSuite) TestSendCoins_Module_To_Account() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	require.NoError(suite.bankKeeper.MintCoins(ctx, mintAcc.GetAddress(), balances))

	// Try send from burner module
	err := suite.bankKeeper.SendCoins(ctx, burnerAcc.GetAddress(), accAddrs[4], balances)
	require.Error(err)

	// Send from mint module
	err = suite.bankKeeper.SendCoins(ctx, mintAcc.GetAddress(), accAddrs[4], balances)
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
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	require.NoError(suite.bankKeeper.MintCoins(ctx, mintAcc.GetAddress(), balances))

	// Try send from burner module
	err := suite.bankKeeper.SendCoins(ctx, burnerAcc.GetAddress(), mintAcc.GetAddress(), sdk.NewCoins(newFooCoin(100), newBarCoin(50)))
	require.Error(err)

	// Send from mint module to burn module
	err = suite.bankKeeper.SendCoins(ctx, mintAcc.GetAddress(), burnerAcc.GetAddress(), sdk.NewCoins(newFooCoin(100), newBarCoin(50)))
	require.NoError(err)

	// Check balances
	burnerFooBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), fooDenom)
	require.Equal(burnerFooBalance.Amount, math.NewInt(100))
	burnerBarBalance := suite.bankKeeper.GetBalance(ctx, burnerAcc.GetAddress(), barDenom)
	require.Equal(burnerBarBalance.Amount, math.NewInt(50))
	mintFooBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), fooDenom)
	require.Equal(mintFooBalance.Amount, math.NewInt(0))
	mintBarBalance := suite.bankKeeper.GetBalance(ctx, mintAcc.GetAddress(), barDenom)
	require.Equal(mintBarBalance.Amount, math.NewInt(0))
}

func (suite *KeeperTestSuite) TestSendCoins_WithRestriction() {
	ctx := suite.ctx
	require := suite.Require()
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	sendAmt := sdk.NewCoins(newFooCoin(10), newBarCoin(10))

	require.NoError(banktestutil.FundAccount(ctx, suite.bankKeeper, accAddrs[0], balances))

	// Add first restriction
	addrRestrictFunc := func(ctx context.Context, from, to []byte, amount sdk.Coins) ([]byte, error) {
		if bytes.Equal(from, to) {
			return nil, fmt.Errorf("Can not send to same address")
		}
		return to, nil
	}
	suite.bankKeeper.AppendGlobalSendRestriction(addrRestrictFunc)

	err := suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[0], sendAmt)
	require.Error(err)
	require.Contains(err.Error(), "Can not send to same address")

	// Add second restriction
	amtRestrictFunc := func(ctx context.Context, from, to []byte, amount sdk.Coins) ([]byte, error) {
		if len(amount) > 1 {
			return nil, fmt.Errorf("Allow only one denom per one send")
		}
		return to, nil
	}
	suite.bankKeeper.AppendGlobalSendRestriction(amtRestrictFunc)

	// Pass the 1st but failt at the 2nd
	err = suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sendAmt)
	require.Error(err)
	require.Contains(err.Error(), "Allow only one denom per one send")

	// Pass both 2 restrictions
	err = suite.bankKeeper.SendCoins(ctx, accAddrs[0], accAddrs[1], sdk.NewCoins(newFooCoin(10)))
	require.NoError(err)

	// Check balances
	acc0FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], fooDenom)
	require.Equal(acc0FooBalance.Amount, math.NewInt(90))
	acc0BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[0], barDenom)
	require.Equal(acc0BarBalance.Amount, math.NewInt(50))
	acc1FooBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[1], fooDenom)
	require.Equal(acc1FooBalance.Amount, math.NewInt(10))
	acc1BarBalance := suite.bankKeeper.GetBalance(ctx, accAddrs[1], barDenom)
	require.Equal(acc1BarBalance.Amount, math.ZeroInt())
}
