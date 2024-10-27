package keeper_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	abci "github.com/cometbft/cometbft/abci/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtime "github.com/cometbft/cometbft/types/time"

	"cosmossdk.io/simapp"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
)

const (
	fooDenom     = "foo"
	barDenom     = "bar"
	initialPower = int64(100)
	holder       = "holder"
	multiPerm    = "multiple permissions account"
	randomPerm   = "random permission"
)

var (
	holderAcc     = authtypes.NewEmptyModuleAccount(holder)
	burnerAcc     = authtypes.NewEmptyModuleAccount(authtypes.Burner, authtypes.Burner)
	minterAcc     = authtypes.NewEmptyModuleAccount(authtypes.Minter, authtypes.Minter)
	multiPermAcc  = authtypes.NewEmptyModuleAccount(multiPerm, authtypes.Burner, authtypes.Minter, authtypes.Staking)
	randomPermAcc = authtypes.NewEmptyModuleAccount(randomPerm, "random")

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

// nolint: interfacer
func getCoinsByName(ctx sdk.Context, bk keeper.Keeper, ak types.AccountKeeper, moduleName string) sdk.Coins {
	moduleAddress := ak.GetModuleAddress(moduleName)
	macc := ak.GetAccount(ctx, moduleAddress)
	if macc == nil {
		return sdk.Coins(nil)
	}

	return bk.GetAllBalances(ctx, macc.GetAddress())
}

type IntegrationTestSuite struct {
	suite.Suite

	bankKeeper    keeper.BaseKeeper
	accountKeeper authkeeper.AccountKeeper
	stakingKeeper *stakingkeeper.Keeper
	ctx           sdk.Context
	appCodec      codec.Codec
	authConfig    *authmodulev1.Module

	queryClient   types.QueryClient
	msgServer     types.MsgServer
	fetchStoreKey func(string) storetypes.StoreKey
}

func (suite *IntegrationTestSuite) initKeepersWithmAccPerms(blockedAddrs map[string]bool) (authkeeper.AccountKeeper, keeper.BaseKeeper) {
	maccPerms := map[string][]string{}
	for _, permission := range suite.authConfig.ModuleAccountPermissions {
		maccPerms[permission.Account] = permission.Permissions
	}

	appCodec := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}).Codec

	maccPerms[holder] = nil
	maccPerms[authtypes.Burner] = []string{authtypes.Burner}
	maccPerms[authtypes.Minter] = []string{authtypes.Minter}
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}
	maccPerms[randomPerm] = []string{"random"}
	authKeeper := authkeeper.NewAccountKeeper(
		appCodec, suite.fetchStoreKey(types.StoreKey), authtypes.ProtoBaseAccount,
		maccPerms, sdk.Bech32MainPrefix, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	bankKeeper := keeper.NewBaseKeeper(
		appCodec, suite.fetchStoreKey(types.StoreKey), authKeeper, blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	return authKeeper, bankKeeper
}

func (suite *IntegrationTestSuite) SetupTest() {
	var interfaceRegistry codectypes.InterfaceRegistry

	app, err := sims.Setup(
		configurator.NewAppConfig(
			configurator.AuthModule(),
			configurator.BankModule(),
			configurator.StakingModule(),
			configurator.ParamsModule(),
			configurator.ConsensusModule(),
			configurator.VestingModule()),
		&suite.accountKeeper, &suite.bankKeeper, &suite.stakingKeeper,
		&interfaceRegistry, &suite.appCodec, &suite.authConfig)
	suite.NoError(err)

	suite.ctx = app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})
	suite.fetchStoreKey = app.UnsafeFindStoreKey

	queryHelper := baseapp.NewQueryServerTestHelper(suite.ctx, interfaceRegistry)
	types.RegisterQueryServer(queryHelper, suite.bankKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	types.RegisterInterfaces(interfaceRegistry)

	suite.queryClient = queryClient
	suite.msgServer = keeper.NewMsgServerImpl(suite.bankKeeper)
}

func (suite *IntegrationTestSuite) TestSupply() {
	ctx := suite.ctx

	require := suite.Require()

	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	genesisSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	require.NoError(err)

	initialPower := int64(100)
	initTokens := suite.stakingKeeper.TokensFromConsensusPower(ctx, initialPower)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))

	// set burnerAcc balance
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	require.NoError(keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	require.NoError(keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Minter, burnerAcc.GetAddress(), initCoins))

	total, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	require.NoError(err)

	expTotalSupply := initCoins.Add(genesisSupply...)
	require.Equal(expTotalSupply, total)

	// burning all supplied tokens
	err = keeper.BurnCoins(ctx, authtypes.Burner, initCoins)
	require.NoError(err)

	total, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	require.NoError(err)
	require.Equal(total, genesisSupply)
}

func (suite *IntegrationTestSuite) TestSendCoinsFromModuleToAccount_Blocklist() {
	ctx := suite.ctx

	// add module accounts to supply keeper
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	_, keeper := suite.initKeepersWithmAccPerms(map[string]bool{addr1.String(): true})

	suite.Require().NoError(keeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	suite.Require().Error(keeper.SendCoinsFromModuleToAccount(
		ctx, minttypes.ModuleName, addr1, initCoins,
	))
}

func (suite *IntegrationTestSuite) TestSupply_SendCoins() {
	ctx := suite.ctx

	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	baseAcc := authKeeper.NewAccountWithAddress(ctx, authtypes.NewModuleAddress("baseAcc"))

	// set initial balances
	suite.
		Require().
		NoError(keeper.MintCoins(ctx, minttypes.ModuleName, initCoins))

	suite.
		Require().
		NoError(keeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, holderAcc.GetAddress(), initCoins))

	authKeeper.SetModuleAccount(ctx, holderAcc)
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	authKeeper.SetAccount(ctx, baseAcc)

	suite.Require().Panics(func() {
		_ = keeper.SendCoinsFromModuleToModule(ctx, "", holderAcc.GetName(), initCoins) // nolint:errcheck
	})

	suite.Require().Panics(func() {
		_ = keeper.SendCoinsFromModuleToModule(ctx, authtypes.Burner, "", initCoins) // nolint:errcheck
	})

	suite.Require().Panics(func() {
		_ = keeper.SendCoinsFromModuleToAccount(ctx, "", baseAcc.GetAddress(), initCoins) // nolint:errcheck
	})

	suite.Require().Error(
		keeper.SendCoinsFromModuleToAccount(ctx, holderAcc.GetName(), baseAcc.GetAddress(), initCoins.Add(initCoins...)),
	)

	suite.Require().NoError(
		keeper.SendCoinsFromModuleToModule(ctx, holderAcc.GetName(), authtypes.Burner, initCoins),
	)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, holderAcc.GetName()).String())
	suite.Require().Equal(initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner))

	suite.Require().NoError(
		keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Burner, baseAcc.GetAddress(), initCoins),
	)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner).String())
	suite.Require().Equal(initCoins, keeper.GetAllBalances(ctx, baseAcc.GetAddress()))

	suite.Require().NoError(keeper.SendCoinsFromAccountToModule(ctx, baseAcc.GetAddress(), authtypes.Burner, initCoins))
	suite.Require().Equal(sdk.NewCoins().String(), keeper.GetAllBalances(ctx, baseAcc.GetAddress()).String())
	suite.Require().Equal(initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner))
}

func (suite *IntegrationTestSuite) TestSupply_MintCoins() {
	ctx := suite.ctx

	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	authKeeper.SetModuleAccount(ctx, burnerAcc)
	authKeeper.SetModuleAccount(ctx, minterAcc)
	authKeeper.SetModuleAccount(ctx, multiPermAcc)
	authKeeper.SetModuleAccount(ctx, randomPermAcc)

	initialSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)

	suite.Require().Panics(func() { keeper.MintCoins(ctx, "", initCoins) }, "no module account")                // nolint:errcheck
	suite.Require().Panics(func() { keeper.MintCoins(ctx, authtypes.Burner, initCoins) }, "invalid permission") // nolint:errcheck

	err = keeper.MintCoins(ctx, authtypes.Minter, sdk.Coins{sdk.Coin{Denom: "denom", Amount: sdk.NewInt(-10)}})
	suite.Require().Error(err, "insufficient coins")

	suite.Require().Panics(func() { keeper.MintCoins(ctx, randomPerm, initCoins) }) // nolint:errcheck

	err = keeper.MintCoins(ctx, authtypes.Minter, initCoins)
	suite.Require().NoError(err)

	suite.Require().Equal(initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Minter))
	totalSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)

	suite.Require().Equal(initialSupply.Add(initCoins...), totalSupply)

	// test same functionality on module account with multiple permissions
	initialSupply, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)

	err = keeper.MintCoins(ctx, multiPermAcc.GetName(), initCoins)
	suite.Require().NoError(err)

	totalSupply, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(initCoins, getCoinsByName(ctx, keeper, authKeeper, multiPermAcc.GetName()))
	suite.Require().Equal(initialSupply.Add(initCoins...), totalSupply)
	suite.Require().Panics(func() { keeper.MintCoins(ctx, authtypes.Burner, initCoins) }) // nolint:errcheck
}

func (suite *IntegrationTestSuite) TestSupply_BurnCoins() {
	ctx := suite.ctx
	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	// set burnerAcc balance
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	suite.
		Require().
		NoError(keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	suite.
		Require().
		NoError(keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Minter, burnerAcc.GetAddress(), initCoins))

	// inflate supply
	suite.
		Require().
		NoError(keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	supplyAfterInflation, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().Panics(func() { keeper.BurnCoins(ctx, "", initCoins) }, "no module account")                    // nolint:errcheck
	suite.Require().Panics(func() { keeper.BurnCoins(ctx, authtypes.Minter, initCoins) }, "invalid permission")     // nolint:errcheck
	suite.Require().Panics(func() { keeper.BurnCoins(ctx, randomPerm, supplyAfterInflation) }, "random permission") // nolint:errcheck
	err = keeper.BurnCoins(ctx, authtypes.Burner, supplyAfterInflation)
	suite.Require().Error(err, "insufficient coins")

	err = keeper.BurnCoins(ctx, authtypes.Burner, initCoins)
	suite.Require().NoError(err)
	supplyAfterBurn, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner).String())
	suite.Require().Equal(supplyAfterInflation.Sub(initCoins...), supplyAfterBurn)

	// test same functionality on module account with multiple permissions
	suite.
		Require().
		NoError(keeper.MintCoins(ctx, authtypes.Minter, initCoins))

	supplyAfterInflation, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().NoError(keeper.SendCoins(ctx, authtypes.NewModuleAddress(authtypes.Minter), multiPermAcc.GetAddress(), initCoins))
	authKeeper.SetModuleAccount(ctx, multiPermAcc)

	err = keeper.BurnCoins(ctx, multiPermAcc.GetName(), initCoins)
	suite.Require().NoError(err)
	supplyAfterBurn, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, multiPermAcc.GetName()).String())
	suite.Require().Equal(supplyAfterInflation.Sub(initCoins...), supplyAfterBurn)
}

func (suite *IntegrationTestSuite) TestSendCoinsNewAccount() {
	ctx := suite.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := suite.accountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.accountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, balances))

	acc1Balances := suite.bankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Nil(suite.accountKeeper.GetAccount(ctx, addr2))
	suite.bankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Empty(suite.bankKeeper.GetAllBalances(ctx, addr2))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(50))
	suite.Require().NoError(suite.bankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc2Balances := suite.bankKeeper.GetAllBalances(ctx, addr2)
	acc1Balances = suite.bankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(sendAmt, acc2Balances)
	updatedAcc1Bal := balances.Sub(sendAmt...)
	suite.Require().Len(acc1Balances, len(updatedAcc1Bal))
	suite.Require().Equal(acc1Balances, updatedAcc1Bal)
	suite.Require().NotNil(suite.accountKeeper.GetAccount(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestInputOutputNewAccount() {
	ctx := suite.ctx

	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := suite.accountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.accountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, balances))

	acc1Balances := suite.bankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Nil(suite.accountKeeper.GetAccount(ctx, addr2))
	suite.Require().Empty(suite.bankKeeper.GetAllBalances(ctx, addr2))

	inputs := []types.Input{
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}
	outputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	suite.Require().NoError(suite.bankKeeper.InputOutputCoins(ctx, inputs, outputs))

	expected := sdk.NewCoins(newFooCoin(30), newBarCoin(10))
	acc2Balances := suite.bankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Equal(expected, acc2Balances)
	suite.Require().NotNil(suite.accountKeeper.GetAccount(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestInputOutputCoins() {
	ctx := suite.ctx
	balances := sdk.NewCoins(newFooCoin(90), newBarCoin(30))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := suite.accountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.accountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	acc2 := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)
	suite.accountKeeper.SetAccount(ctx, acc2)

	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	acc3 := suite.accountKeeper.NewAccountWithAddress(ctx, addr3)
	suite.accountKeeper.SetAccount(ctx, acc3)

	input := []types.Input{
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(60), newBarCoin(20))},
	}
	outputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
		{Address: addr3.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	suite.Require().Error(suite.bankKeeper.InputOutputCoins(ctx, input, []types.Output{}))
	suite.Require().Error(suite.bankKeeper.InputOutputCoins(ctx, input, outputs))

	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, balances))

	insufficientInput := []types.Input{
		{
			Address: addr1.String(),
			Coins:   sdk.NewCoins(newFooCoin(300), newBarCoin(100)),
		},
	}
	insufficientOutputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
		{Address: addr3.String(), Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
	}
	suite.Require().Error(suite.bankKeeper.InputOutputCoins(ctx, insufficientInput, insufficientOutputs))
	suite.Require().NoError(suite.bankKeeper.InputOutputCoins(ctx, input, outputs))

	acc1Balances := suite.bankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(30), newBarCoin(10))
	suite.Require().Equal(expected, acc1Balances)

	acc2Balances := suite.bankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Equal(expected, acc2Balances)

	acc3Balances := suite.bankKeeper.GetAllBalances(ctx, addr3)
	suite.Require().Equal(expected, acc3Balances)
}

func (suite *IntegrationTestSuite) TestSendCoins() {
	ctx := suite.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress("addr1_______________")
	acc1 := suite.accountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.accountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress("addr2_______________")
	acc2 := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)
	suite.accountKeeper.SetAccount(ctx, acc2)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr2, balances))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().Error(suite.bankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, balances))
	suite.Require().NoError(suite.bankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc1Balances := suite.bankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().Equal(expected, acc1Balances)

	acc2Balances := suite.bankKeeper.GetAllBalances(ctx, addr2)
	expected = sdk.NewCoins(newFooCoin(150), newBarCoin(75))
	suite.Require().Equal(expected, acc2Balances)

	// we sent all foo coins to acc2, so foo balance should be deleted for acc1 and bar should be still there
	var coins []sdk.Coin
	suite.bankKeeper.IterateAccountBalances(ctx, addr1, func(c sdk.Coin) (stop bool) {
		coins = append(coins, c)
		return true
	})
	suite.Require().Len(coins, 1)
	suite.Require().Equal(newBarCoin(25), coins[0], "expected only bar coins in the account balance, got: %v", coins)
}

func (suite *IntegrationTestSuite) TestValidateBalance() {
	ctx := suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Error(suite.bankKeeper.ValidateBalance(ctx, addr1))

	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.accountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, balances))
	suite.Require().NoError(suite.bankKeeper.ValidateBalance(ctx, addr1))

	bacc := authtypes.NewBaseAccountWithAddress(addr2)
	vacc := vesting.NewContinuousVestingAccount(bacc, balances.Add(balances...), now.Unix(), endTime.Unix())

	suite.accountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr2, balances))
	suite.Require().Error(suite.bankKeeper.ValidateBalance(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestSendCoins_Invalid_SendLockedCoins() {
	ctx := suite.ctx
	balances := sdk.NewCoins(newFooCoin(50))
	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	acc0 := authtypes.NewBaseAccountWithAddress(addr)
	vacc := vesting.NewContinuousVestingAccount(acc0, origCoins, now.Unix(), endTime.Unix())
	suite.accountKeeper.SetAccount(ctx, vacc)

	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, suite.ctx, addr2, balances))
	suite.Require().Error(suite.bankKeeper.SendCoins(ctx, addr, addr2, sendCoins))
}

func (suite *IntegrationTestSuite) TestSendEnabled() {
	ctx := suite.ctx
	enabled := true
	params := types.DefaultParams()
	suite.Require().Equal(enabled, params.DefaultSendEnabled)

	suite.Require().NoError(suite.bankKeeper.SetParams(ctx, params))

	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, math.OneInt())
	fooCoin := sdk.NewCoin("foocoin", math.OneInt())
	barCoin := sdk.NewCoin("barcoin", math.OneInt())

	// assert with default (all denom) send enabled both Bar and Bond Denom are enabled
	suite.Require().Equal(enabled, suite.bankKeeper.IsSendEnabledCoin(ctx, barCoin))
	suite.Require().Equal(enabled, suite.bankKeeper.IsSendEnabledCoin(ctx, bondCoin))

	// Both coins should be send enabled.
	err := suite.bankKeeper.IsSendEnabledCoins(ctx, fooCoin, bondCoin)
	suite.Require().NoError(err)

	// Set default send_enabled to !enabled, add a foodenom that overrides default as enabled
	params.DefaultSendEnabled = !enabled
	suite.Require().NoError(suite.bankKeeper.SetParams(ctx, params))
	suite.bankKeeper.SetSendEnabled(ctx, fooCoin.Denom, enabled)

	// Expect our specific override to be enabled, others to be !enabled.
	suite.Require().Equal(enabled, suite.bankKeeper.IsSendEnabledCoin(ctx, fooCoin))
	suite.Require().Equal(!enabled, suite.bankKeeper.IsSendEnabledCoin(ctx, barCoin))
	suite.Require().Equal(!enabled, suite.bankKeeper.IsSendEnabledCoin(ctx, bondCoin))

	// Foo coin should be send enabled.
	err = suite.bankKeeper.IsSendEnabledCoins(ctx, fooCoin)
	suite.Require().NoError(err)

	// Expect an error when one coin is not send enabled.
	err = suite.bankKeeper.IsSendEnabledCoins(ctx, fooCoin, bondCoin)
	suite.Require().Error(err)

	// Expect an error when all coins are not send enabled.
	err = suite.bankKeeper.IsSendEnabledCoins(ctx, bondCoin, barCoin)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestHasBalance() {
	ctx := suite.ctx
	addr := sdk.AccAddress([]byte("addr1_______________"))

	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr)
	suite.accountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().False(suite.bankKeeper.HasBalance(ctx, addr, newFooCoin(99)))

	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr, balances))
	suite.Require().False(suite.bankKeeper.HasBalance(ctx, addr, newFooCoin(101)))
	suite.Require().True(suite.bankKeeper.HasBalance(ctx, addr, newFooCoin(100)))
	suite.Require().True(suite.bankKeeper.HasBalance(ctx, addr, newFooCoin(1)))
}

func (suite *IntegrationTestSuite) TestMsgSendEvents() {
	ctx := suite.ctx
	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr)

	suite.accountKeeper.SetAccount(ctx, acc)
	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr, newCoins))

	suite.Require().NoError(suite.bankKeeper.SendCoins(ctx, addr, addr2, newCoins))
	event1 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []abci.EventAttribute{},
	}
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: types.AttributeKeyRecipient, Value: addr2.String()},
	)
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: types.AttributeKeySender, Value: addr.String()},
	)
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: sdk.AttributeKeyAmount, Value: newCoins.String()},
	)

	event2 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event2.Attributes = append(
		event2.Attributes,
		abci.EventAttribute{Key: types.AttributeKeySender, Value: addr.String()},
	)

	// events are shifted due to the funding account events
	events := ctx.EventManager().ABCIEvents()
	suite.Require().Equal(10, len(events))
	suite.Require().Equal(abci.Event(event1), events[8])
	suite.Require().Equal(abci.Event(event2), events[9])
}

func (suite *IntegrationTestSuite) TestMsgMultiSendEvents() {
	ctx := suite.ctx

	suite.Require().NoError(suite.bankKeeper.SetParams(ctx, types.DefaultParams()))

	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	addr4 := sdk.AccAddress([]byte("addr4_______________"))
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr)
	acc2 := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)

	suite.accountKeeper.SetAccount(ctx, acc)
	suite.accountKeeper.SetAccount(ctx, acc2)

	coins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50), sdk.NewInt64Coin(barDenom, 100))
	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))
	newCoins2 := sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))
	input := []types.Input{
		{
			Address: addr.String(),
			Coins:   coins,
		},
	}
	outputs := []types.Output{
		{Address: addr3.String(), Coins: newCoins},
		{Address: addr4.String(), Coins: newCoins2},
	}

	suite.Require().Error(suite.bankKeeper.InputOutputCoins(ctx, input, outputs))

	events := ctx.EventManager().ABCIEvents()
	suite.Require().Equal(0, len(events))

	// Set addr's coins but not addr2's coins
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50), sdk.NewInt64Coin(barDenom, 100))))
	suite.Require().NoError(suite.bankKeeper.InputOutputCoins(ctx, input, outputs))

	events = ctx.EventManager().ABCIEvents()
	suite.Require().Equal(12, len(events)) // 12 events because account funding causes extra minting + coin_spent + coin_recv events

	event1 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: types.AttributeKeySender, Value: addr.String()},
	)
	suite.Require().Equal(abci.Event(event1), events[7])

	// Set addr's coins and addr2's coins
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))))
	newCoins = sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))

	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))))
	newCoins2 = sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))

	suite.Require().NoError(suite.bankKeeper.InputOutputCoins(ctx, input, outputs))

	events = ctx.EventManager().ABCIEvents()
	suite.Require().Equal(30, len(events)) // 27 due to account funding + coin_spent + coin_recv events

	event2 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []abci.EventAttribute{},
	}
	event2.Attributes = append(
		event2.Attributes,
		abci.EventAttribute{Key: types.AttributeKeyRecipient, Value: addr3.String()},
	)
	event2.Attributes = append(
		event2.Attributes,
		abci.EventAttribute{Key: sdk.AttributeKeyAmount, Value: newCoins.String()})
	event3 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []abci.EventAttribute{},
	}
	event3.Attributes = append(
		event3.Attributes,
		abci.EventAttribute{Key: types.AttributeKeyRecipient, Value: addr4.String()},
	)
	event3.Attributes = append(
		event3.Attributes,
		abci.EventAttribute{Key: sdk.AttributeKeyAmount, Value: newCoins2.String()},
	)
	// events are shifted due to the funding account events
	suite.Require().Equal(abci.Event(event1), events[25])
	suite.Require().Equal(abci.Event(event2), events[27])
	suite.Require().Equal(abci.Event(event3), events[29])
}

func (suite *IntegrationTestSuite) TestSpendableCoins() {
	ctx := suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	macc := suite.accountKeeper.NewAccountWithAddress(ctx, addrModule)
	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)

	suite.accountKeeper.SetAccount(ctx, macc)
	suite.accountKeeper.SetAccount(ctx, vacc)
	suite.accountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr2, origCoins))

	suite.Require().Equal(origCoins, suite.bankKeeper.SpendableCoins(ctx, addr2))
	suite.Require().Equal(origCoins[0], suite.bankKeeper.SpendableCoin(ctx, addr2, "stake"))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(suite.bankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	suite.Require().Equal(origCoins.Sub(delCoins...), suite.bankKeeper.SpendableCoins(ctx, addr1))
	suite.Require().Equal(origCoins.Sub(delCoins...)[0], suite.bankKeeper.SpendableCoin(ctx, addr1, "stake"))
}

func (suite *IntegrationTestSuite) TestVestingAccountSend() {
	ctx := suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	suite.accountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	suite.Require().Error(suite.bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))

	// receive some coins
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, sendCoins))
	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(suite.bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	suite.Require().Equal(origCoins, suite.bankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestPeriodicVestingAccountSend() {
	ctx := suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
	}

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewPeriodicVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), periods)

	suite.accountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	suite.Require().Error(suite.bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))

	// receive some coins
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, sendCoins))

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(suite.bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	suite.Require().Equal(origCoins, suite.bankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestVestingAccountReceive() {
	ctx := suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)

	suite.accountKeeper.SetAccount(ctx, vacc)
	suite.accountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr2, origCoins))

	// send some coins to the vesting account
	suite.Require().NoError(suite.bankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = suite.accountKeeper.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	balances := suite.bankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(origCoins.Add(sendCoins...), balances)
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now)...), sendCoins)

	// require coins are spendable plus any that have vested
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))...), origCoins)
}

func (suite *IntegrationTestSuite) TestPeriodicVestingAccountReceive() {
	ctx := suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	periods := vesting.Periods{
		vesting.Period{Length: int64(12 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 50)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
		vesting.Period{Length: int64(6 * 60 * 60), Amount: sdk.Coins{sdk.NewInt64Coin("stake", 25)}},
	}

	vacc := vesting.NewPeriodicVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), periods)
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)

	suite.accountKeeper.SetAccount(ctx, vacc)
	suite.accountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr2, origCoins))

	// send some coins to the vesting account
	suite.Require().NoError(suite.bankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = suite.accountKeeper.GetAccount(ctx, addr1).(*vesting.PeriodicVestingAccount)
	balances := suite.bankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(origCoins.Add(sendCoins...), balances)
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now)...), sendCoins)

	// require coins are spendable plus any that have vested
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))...), origCoins)
}

func (suite *IntegrationTestSuite) TestDelegateCoins() {
	ctx := suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	macc := suite.accountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)
	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())

	suite.accountKeeper.SetAccount(ctx, vacc)
	suite.accountKeeper.SetAccount(ctx, acc)
	suite.accountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	suite.Require().NoError(suite.bankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	suite.Require().Equal(origCoins.Sub(delCoins...), suite.bankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().Equal(delCoins, suite.bankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to delegate
	suite.Require().NoError(suite.bankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	suite.Require().Equal(delCoins, suite.bankKeeper.GetAllBalances(ctx, addr1))

	// require that delegated vesting amount is equal to what was delegated with DelegateCoins
	acc = suite.accountKeeper.GetAccount(ctx, addr1)
	vestingAcc, ok := acc.(types.VestingAccount)
	suite.Require().True(ok)
	suite.Require().Equal(delCoins, vestingAcc.GetDelegatedVesting())
}

func (suite *IntegrationTestSuite) TestDelegateCoins_Invalid() {
	ctx := suite.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))
	macc := suite.accountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr1)

	suite.Require().Error(suite.bankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	invalidCoins := sdk.Coins{sdk.Coin{Denom: "fooDenom", Amount: sdk.NewInt(-50)}}
	suite.Require().Error(suite.bankKeeper.DelegateCoins(ctx, addr1, addrModule, invalidCoins))

	suite.accountKeeper.SetAccount(ctx, macc)
	suite.Require().Error(suite.bankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	suite.accountKeeper.SetAccount(ctx, acc)
	suite.Require().Error(suite.bankKeeper.DelegateCoins(ctx, addr1, addrModule, origCoins.Add(origCoins...)))
}

func (suite *IntegrationTestSuite) TestUndelegateCoins() {
	ctx := suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	macc := suite.accountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing

	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr2)

	suite.accountKeeper.SetAccount(ctx, vacc)
	suite.accountKeeper.SetAccount(ctx, acc)
	suite.accountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := suite.bankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins)
	suite.Require().NoError(err)

	suite.Require().Equal(origCoins.Sub(delCoins...), suite.bankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().Equal(delCoins, suite.bankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a non-vesting account to undelegate
	suite.Require().NoError(suite.bankKeeper.UndelegateCoins(ctx, addrModule, addr2, delCoins))

	suite.Require().Equal(origCoins, suite.bankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().True(suite.bankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require the ability for a vesting account to delegate
	suite.Require().NoError(suite.bankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))

	suite.Require().Equal(origCoins.Sub(delCoins...), suite.bankKeeper.GetAllBalances(ctx, addr1))
	suite.Require().Equal(delCoins, suite.bankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to undelegate
	suite.Require().NoError(suite.bankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	suite.Require().Equal(origCoins, suite.bankKeeper.GetAllBalances(ctx, addr1))
	suite.Require().True(suite.bankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require that delegated vesting amount is completely empty, since they were completely undelegated
	acc = suite.accountKeeper.GetAccount(ctx, addr1)
	vestingAcc, ok := acc.(types.VestingAccount)
	suite.Require().True(ok)
	suite.Require().Empty(vestingAcc.GetDelegatedVesting())
}

func (suite *IntegrationTestSuite) TestUndelegateCoins_Invalid() {
	ctx := suite.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))
	macc := suite.accountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := suite.accountKeeper.NewAccountWithAddress(ctx, addr1)

	suite.Require().Error(suite.bankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	suite.accountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(suite.bankKeeper, ctx, addr1, origCoins))

	suite.Require().Error(suite.bankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))
	suite.accountKeeper.SetAccount(ctx, acc)

	suite.Require().Error(suite.bankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))
}

func (suite *IntegrationTestSuite) TestSetDenomMetaData() {
	ctx := suite.ctx

	metadata := suite.getTestMetadata()

	for i := range []int{1, 2} {
		suite.bankKeeper.SetDenomMetaData(ctx, metadata[i])
	}

	actualMetadata, found := suite.bankKeeper.GetDenomMetaData(ctx, metadata[1].Base)
	suite.Require().True(found)
	found = suite.bankKeeper.HasDenomMetaData(ctx, metadata[1].Base)
	suite.Require().True(found)
	suite.Require().Equal(metadata[1].GetBase(), actualMetadata.GetBase())
	suite.Require().Equal(metadata[1].GetDisplay(), actualMetadata.GetDisplay())
	suite.Require().Equal(metadata[1].GetDescription(), actualMetadata.GetDescription())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetDenom(), actualMetadata.GetDenomUnits()[1].GetDenom())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetExponent(), actualMetadata.GetDenomUnits()[1].GetExponent())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetAliases(), actualMetadata.GetDenomUnits()[1].GetAliases())
}

func (suite *IntegrationTestSuite) TestIterateAllDenomMetaData() {
	ctx := suite.ctx

	expectedMetadata := suite.getTestMetadata()
	// set metadata
	for i := range []int{1, 2} {
		suite.bankKeeper.SetDenomMetaData(ctx, expectedMetadata[i])
	}
	// retrieve metadata
	actualMetadata := make([]types.Metadata, 0)
	suite.bankKeeper.IterateAllDenomMetaData(ctx, func(metadata types.Metadata) bool {
		actualMetadata = append(actualMetadata, metadata)
		return false
	})
	// execute checks
	for i := range []int{1, 2} {
		suite.Require().Equal(expectedMetadata[i].GetBase(), actualMetadata[i].GetBase())
		suite.Require().Equal(expectedMetadata[i].GetDisplay(), actualMetadata[i].GetDisplay())
		suite.Require().Equal(expectedMetadata[i].GetDescription(), actualMetadata[i].GetDescription())
		suite.Require().Equal(expectedMetadata[i].GetDenomUnits()[1].GetDenom(), actualMetadata[i].GetDenomUnits()[1].GetDenom())
		suite.Require().Equal(expectedMetadata[i].GetDenomUnits()[1].GetExponent(), actualMetadata[i].GetDenomUnits()[1].GetExponent())
		suite.Require().Equal(expectedMetadata[i].GetDenomUnits()[1].GetAliases(), actualMetadata[i].GetDenomUnits()[1].GetAliases())
	}
}

func (suite *IntegrationTestSuite) TestBalanceTrackingEvents() {
	// replace account keeper and bank keeper otherwise the account keeper won't be aware of the
	// existence of the new module account because GetModuleAccount checks for the existence via
	// permissions map and not via state... weird
	maccPerms := simapp.GetMaccPerms()
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}

	suite.accountKeeper = authkeeper.NewAccountKeeper(
		suite.appCodec, suite.fetchStoreKey(authtypes.StoreKey),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	suite.bankKeeper = keeper.NewBaseKeeper(suite.appCodec, suite.fetchStoreKey(types.StoreKey),
		suite.accountKeeper, nil, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// set account with multiple permissions
	suite.accountKeeper.SetModuleAccount(suite.ctx, multiPermAcc)
	// mint coins
	suite.Require().NoError(
		suite.bankKeeper.MintCoins(
			suite.ctx,
			multiPermAcc.Name,
			sdk.NewCoins(sdk.NewCoin("utxo", sdk.NewInt(100000)))),
	)
	// send coins to address
	addr1 := sdk.AccAddress("addr1_______________")
	suite.Require().NoError(
		suite.bankKeeper.SendCoinsFromModuleToAccount(
			suite.ctx,
			multiPermAcc.Name,
			addr1,
			sdk.NewCoins(sdk.NewCoin("utxo", sdk.NewInt(50000))),
		),
	)

	// burn coins from module account
	suite.Require().NoError(
		suite.bankKeeper.BurnCoins(
			suite.ctx,
			multiPermAcc.Name,
			sdk.NewCoins(sdk.NewInt64Coin("utxo", 1000)),
		),
	)

	// process balances and supply from events
	supply := sdk.NewCoins()

	balances := make(map[string]sdk.Coins)

	for _, e := range suite.ctx.EventManager().ABCIEvents() {
		switch e.Type {
		case types.EventTypeCoinBurn:
			burnedCoins, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			suite.Require().NoError(err)
			supply = supply.Sub(burnedCoins...)

		case types.EventTypeCoinMint:
			mintedCoins, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			suite.Require().NoError(err)
			supply = supply.Add(mintedCoins...)

		case types.EventTypeCoinSpent:
			coinsSpent, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			suite.Require().NoError(err)
			spender, err := sdk.AccAddressFromBech32((string)(e.Attributes[0].Value))
			suite.Require().NoError(err)
			balances[spender.String()] = balances[spender.String()].Sub(coinsSpent...)

		case types.EventTypeCoinReceived:
			coinsRecv, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			suite.Require().NoError(err)
			receiver, err := sdk.AccAddressFromBech32((string)(e.Attributes[0].Value))
			suite.Require().NoError(err)
			balances[receiver.String()] = balances[receiver.String()].Add(coinsRecv...)
		}
	}

	// check balance and supply tracking
	suite.Require().True(suite.bankKeeper.HasSupply(suite.ctx, "utxo"))
	savedSupply := suite.bankKeeper.GetSupply(suite.ctx, "utxo")
	utxoSupply := savedSupply
	suite.Require().Equal(utxoSupply.Amount, supply.AmountOf("utxo"))
	// iterate accounts and check balances
	suite.bankKeeper.IterateAllBalances(suite.ctx, func(address sdk.AccAddress, coin sdk.Coin) (stop bool) {
		// if it's not utxo coin then skip
		if coin.Denom != "utxo" {
			return false
		}

		balance, exists := balances[address.String()]
		suite.Require().True(exists)

		expectedUtxo := sdk.NewCoin("utxo", balance.AmountOf(coin.Denom))
		suite.Require().Equal(expectedUtxo.String(), coin.String())
		return false
	})
}

func (suite *IntegrationTestSuite) getTestMetadata() []types.Metadata {
	return []types.Metadata{
		{
			Name:        "Cosmos Hub Atom",
			Symbol:      "ATOM",
			Description: "The native staking token of the Cosmos Hub.",
			DenomUnits: []*types.DenomUnit{
				{Denom: "uatom", Exponent: uint32(0), Aliases: []string{"microatom"}},
				{Denom: "matom", Exponent: uint32(3), Aliases: []string{"milliatom"}},
				{Denom: "atom", Exponent: uint32(6), Aliases: nil},
			},
			Base:    "uatom",
			Display: "atom",
		},
		{
			Name:        "Token",
			Symbol:      "TOKEN",
			Description: "The native staking token of the Token Hub.",
			DenomUnits: []*types.DenomUnit{
				{Denom: "1token", Exponent: uint32(5), Aliases: []string{"decitoken"}},
				{Denom: "2token", Exponent: uint32(4), Aliases: []string{"centitoken"}},
				{Denom: "3token", Exponent: uint32(7), Aliases: []string{"dekatoken"}},
			},
			Base:    "utoken",
			Display: "token",
		},
	}
}

func (suite *IntegrationTestSuite) TestMintCoinRestrictions() {
	type BankMintingRestrictionFn func(ctx sdk.Context, coins sdk.Coins) error

	maccPerms := simapp.GetMaccPerms()
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}

	suite.accountKeeper = authkeeper.NewAccountKeeper(
		suite.appCodec, suite.fetchStoreKey(authtypes.StoreKey),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	suite.accountKeeper.SetModuleAccount(suite.ctx, multiPermAcc)

	type testCase struct {
		coinsToTry sdk.Coin
		expectPass bool
	}

	tests := []struct {
		name          string
		restrictionFn BankMintingRestrictionFn
		testCases     []testCase
	}{
		{
			"restriction",
			func(_ sdk.Context, coins sdk.Coins) error {
				for _, coin := range coins {
					if coin.Denom != fooDenom {
						return fmt.Errorf("Module %s only has perms for minting %s coins, tried minting %s coins", types.ModuleName, fooDenom, coin.Denom)
					}
				}
				return nil
			},
			[]testCase{
				{
					coinsToTry: newFooCoin(100),
					expectPass: true,
				},
				{
					coinsToTry: newBarCoin(100),
					expectPass: false,
				},
			},
		},
	}

	for _, test := range tests {
		suite.bankKeeper = keeper.NewBaseKeeper(suite.appCodec, suite.fetchStoreKey(types.StoreKey),
			suite.accountKeeper, nil, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		).WithMintCoinsRestriction(keeper.MintingRestrictionFn(test.restrictionFn))
		for _, testCase := range test.testCases {
			if testCase.expectPass {
				suite.Require().NoError(
					suite.bankKeeper.MintCoins(
						suite.ctx,
						multiPermAcc.Name,
						sdk.NewCoins(testCase.coinsToTry),
					),
				)
			} else {
				suite.Require().Error(
					suite.bankKeeper.MintCoins(
						suite.ctx,
						multiPermAcc.Name,
						sdk.NewCoins(testCase.coinsToTry),
					),
				)
			}
		}
	}
}

func (suite *IntegrationTestSuite) TestIsSendEnabledDenom() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper

	defaultCoin := "defaultCoin"
	enabledCoin := "enabledCoin"
	disabledCoin := "disabledCoin"
	bankKeeper.DeleteSendEnabled(ctx, defaultCoin)
	bankKeeper.SetSendEnabled(ctx, enabledCoin, true)
	bankKeeper.SetSendEnabled(ctx, disabledCoin, false)

	tests := []struct {
		denom  string
		exp    bool
		expDef bool
	}{
		{
			denom:  "defaultCoin",
			expDef: true,
		},
		{
			denom: enabledCoin,
			exp:   true,
		},
		{
			denom: disabledCoin,
			exp:   false,
		},
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		suite.Require().NoError(bankKeeper.SetParams(ctx, params))
		for _, tc := range tests {
			suite.T().Run(fmt.Sprintf("%s default %t", tc.denom, def), func(t *testing.T) {
				actual := suite.bankKeeper.IsSendEnabledDenom(suite.ctx, tc.denom)
				exp := tc.exp
				if tc.expDef {
					exp = def
				}
				assert.Equal(t, exp, actual)
			})
		}
	}
}

func (suite *IntegrationTestSuite) TestGetSendEnabledEntry() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper

	bankKeeper.SetAllSendEnabled(ctx, []*types.SendEnabled{
		{Denom: "gettruecoin", Enabled: true},
		{Denom: "getfalsecoin", Enabled: false},
	})

	tests := []struct {
		denom string
		expSE types.SendEnabled
		expF  bool
	}{
		{
			denom: "missing",
			expSE: types.SendEnabled{},
			expF:  false,
		},
		{
			denom: "gettruecoin",
			expSE: types.SendEnabled{Denom: "gettruecoin", Enabled: true},
			expF:  true,
		},
		{
			denom: "getfalsecoin",
			expSE: types.SendEnabled{Denom: "getfalsecoin", Enabled: false},
			expF:  true,
		},
	}

	for _, tc := range tests {
		suite.T().Run(tc.denom, func(t *testing.T) {
			actualSE, actualF := bankKeeper.GetSendEnabledEntry(ctx, tc.denom)
			assert.Equal(t, tc.expF, actualF, "found")
			assert.Equal(t, tc.expSE, actualSE, "SendEnabled")
		})
	}
}

func (suite *IntegrationTestSuite) TestSetSendEnabled() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper

	tests := []struct {
		name  string
		denom string
		value bool
	}{
		{
			name:  "very short denom true",
			denom: "f",
			value: true,
		},
		{
			name:  "very short denom false",
			denom: "f",
			value: true,
		},
		{
			name:  "falseFirstCoin false",
			denom: "falseFirstCoin",
			value: false,
		},
		{
			name:  "falseFirstCoin true",
			denom: "falseFirstCoin",
			value: true,
		},
		{
			name:  "falseFirstCoin true again",
			denom: "falseFirstCoin",
			value: true,
		},
		{
			name:  "trueFirstCoin true",
			denom: "falseFirstCoin",
			value: false,
		},
		{
			name:  "trueFirstCoin false",
			denom: "falseFirstCoin",
			value: false,
		},
		{
			name:  "trueFirstCoin false again",
			denom: "falseFirstCoin",
			value: false,
		},
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		suite.Require().NoError(bankKeeper.SetParams(ctx, params))
		for _, tc := range tests {
			suite.T().Run(fmt.Sprintf("%s default %t", tc.name, def), func(t *testing.T) {
				bankKeeper.SetSendEnabled(ctx, tc.denom, tc.value)
				actual := bankKeeper.IsSendEnabledDenom(ctx, tc.denom)
				assert.Equal(t, tc.value, actual)
			})
		}
	}
}

func (suite *IntegrationTestSuite) TestSetAllSendEnabled() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper

	tests := []struct {
		name         string
		sendEnableds []*types.SendEnabled
	}{
		{
			name:         "nil",
			sendEnableds: nil,
		},
		{
			name:         "empty",
			sendEnableds: []*types.SendEnabled{},
		},
		{
			name: "one true",
			sendEnableds: []*types.SendEnabled{
				{Denom: "aonecoin", Enabled: true},
			},
		},
		{
			name: "one false",
			sendEnableds: []*types.SendEnabled{
				{Denom: "bonecoin", Enabled: false},
			},
		},
		{
			name: "two true",
			sendEnableds: []*types.SendEnabled{
				{Denom: "conecoin", Enabled: true},
				{Denom: "ctwocoin", Enabled: true},
			},
		},
		{
			name: "two true false",
			sendEnableds: []*types.SendEnabled{
				{Denom: "donecoin", Enabled: true},
				{Denom: "dtwocoin", Enabled: false},
			},
		},
		{
			name: "two false true",
			sendEnableds: []*types.SendEnabled{
				{Denom: "eonecoin", Enabled: false},
				{Denom: "etwocoin", Enabled: true},
			},
		},
		{
			name: "two false",
			sendEnableds: []*types.SendEnabled{
				{Denom: "fonecoin", Enabled: false},
				{Denom: "ftwocoin", Enabled: false},
			},
		},
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		suite.Require().NoError(bankKeeper.SetParams(ctx, params))
		for _, tc := range tests {
			suite.T().Run(fmt.Sprintf("%s default %t", tc.name, def), func(t *testing.T) {
				bankKeeper.SetAllSendEnabled(ctx, tc.sendEnableds)
				for _, se := range tc.sendEnableds {
					actual := bankKeeper.IsSendEnabledDenom(ctx, se.Denom)
					assert.Equal(t, se.Enabled, actual, se.Denom)
				}
			})
		}
	}
}

func (suite *IntegrationTestSuite) TestDeleteSendEnabled() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		suite.Require().NoError(bankKeeper.SetParams(ctx, params))
		suite.T().Run(fmt.Sprintf("default %t", def), func(t *testing.T) {
			denom := fmt.Sprintf("somerand%tcoin", !def)
			bankKeeper.SetSendEnabled(ctx, denom, !def)
			require.Equal(t, !def, bankKeeper.IsSendEnabledDenom(ctx, denom))
			bankKeeper.DeleteSendEnabled(ctx, denom)
			require.Equal(t, def, bankKeeper.IsSendEnabledDenom(ctx, denom))
		})
	}
}

func (suite *IntegrationTestSuite) TestIterateSendEnabledEntries() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper

	suite.T().Run("no entries to iterate", func(t *testing.T) {
		count := 0
		bankKeeper.IterateSendEnabledEntries(ctx, func(_ string, _ bool) (stop bool) {
			count++
			return false
		})
		assert.Equal(t, 0, count)
	})

	alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	denoms := make([]string, len(alpha)*2)
	for i, l := range alpha {
		denoms[i*2] = fmt.Sprintf("%sitercointrue", l)
		denoms[i*2+1] = fmt.Sprintf("%sitercoinfalse", l)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2], true)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2+1], false)
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		suite.Require().NoError(bankKeeper.SetParams(ctx, params))
		var seen []string
		suite.T().Run(fmt.Sprintf("all denoms have expected values default %t", def), func(t *testing.T) {
			bankKeeper.IterateSendEnabledEntries(ctx, func(denom string, sendEnabled bool) (stop bool) {
				seen = append(seen, denom)
				exp := true
				if strings.HasSuffix(denom, "false") {
					exp = false
				}
				assert.Equal(t, exp, sendEnabled, denom)
				return false
			})
		})
		suite.T().Run(fmt.Sprintf("all denoms were seen default %t", def), func(t *testing.T) {
			assert.ElementsMatch(t, denoms, seen)
		})
	}

	for _, denom := range denoms {
		bankKeeper.DeleteSendEnabled(ctx, denom)
	}

	suite.T().Run("no entries to iterate again after deleting all of them", func(t *testing.T) {
		count := 0
		bankKeeper.IterateSendEnabledEntries(ctx, func(_ string, _ bool) (stop bool) {
			count++
			return false
		})
		assert.Equal(t, 0, count)
	})
}

func (suite *IntegrationTestSuite) TestGetAllSendEnabledEntries() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper

	suite.T().Run("no entries", func(t *testing.T) {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		assert.Len(t, actual, 0)
	})

	alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	denoms := make([]string, len(alpha)*2)
	for i, l := range alpha {
		denoms[i*2] = fmt.Sprintf("%sitercointrue", l)
		denoms[i*2+1] = fmt.Sprintf("%sitercoinfalse", l)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2], true)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2+1], false)
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		suite.Require().NoError(bankKeeper.SetParams(ctx, params))
		var seen []string
		suite.T().Run(fmt.Sprintf("all denoms have expected values default %t", def), func(t *testing.T) {
			actual := bankKeeper.GetAllSendEnabledEntries(ctx)
			for _, se := range actual {
				seen = append(seen, se.Denom)
				exp := true
				if strings.HasSuffix(se.Denom, "false") {
					exp = false
				}
				assert.Equal(t, exp, se.Enabled, se.Denom)
			}
		})
		suite.T().Run(fmt.Sprintf("all denoms were seen default %t", def), func(t *testing.T) {
			assert.ElementsMatch(t, denoms, seen)
		})
	}

	for _, denom := range denoms {
		bankKeeper.DeleteSendEnabled(ctx, denom)
	}

	suite.T().Run("no entries again after deleting all of them", func(t *testing.T) {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		assert.Len(t, actual, 0)
	})
}

type mockSubspace struct {
	ps types.Params
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps exported.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

func (ms mockSubspace) Get(ctx sdk.Context, key []byte, ptr interface{}) {}

func (suite *IntegrationTestSuite) TestMigrator_Migrate3to4() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		suite.Require().NoError(bankKeeper.SetParams(ctx, params))
		suite.T().Run(fmt.Sprintf("default %t does not change", def), func(t *testing.T) {
			legacySubspace := func(ps types.Params) mockSubspace {
				return mockSubspace{ps: ps}
			}(types.NewParams(def))
			migrator := keeper.NewMigrator(bankKeeper, legacySubspace)
			require.NoError(t, migrator.Migrate3to4(ctx))
			actual := bankKeeper.GetParams(ctx)
			assert.Equal(t, params.DefaultSendEnabled, actual.DefaultSendEnabled)
		})
	}

	for _, def := range []bool{true, false} {
		params := types.Params{
			SendEnabled: []*types.SendEnabled{
				{Denom: fmt.Sprintf("truecoin%t", def), Enabled: true},
				{Denom: fmt.Sprintf("falsecoin%t", def), Enabled: false},
			},
		}
		suite.Require().NoError(bankKeeper.SetParams(ctx, params))
		suite.T().Run(fmt.Sprintf("default %t send enabled info moved to store", def), func(t *testing.T) {
			legacySubspace := func(ps types.Params) mockSubspace {
				return mockSubspace{ps: ps}
			}(types.NewParams(def))
			migrator := keeper.NewMigrator(bankKeeper, legacySubspace)
			require.NoError(t, migrator.Migrate3to4(ctx))
			newParams := bankKeeper.GetParams(ctx)
			assert.Len(t, newParams.SendEnabled, 0)
			for _, se := range params.SendEnabled {
				actual := bankKeeper.IsSendEnabledDenom(ctx, se.Denom)
				assert.Equal(t, se.Enabled, actual, se.Denom)
			}
		})
	}
}

func (suite *IntegrationTestSuite) TestSetParams() {
	ctx, bankKeeper := suite.ctx, suite.bankKeeper
	params := types.NewParams(true)
	params.SendEnabled = []*types.SendEnabled{
		{Denom: "paramscointrue", Enabled: true},
		{Denom: "paramscoinfalse", Enabled: false},
	}
	suite.Require().NoError(bankKeeper.SetParams(ctx, params))

	suite.Run("stored params are as expected", func() {
		actual := bankKeeper.GetParams(ctx)
		suite.Assert().True(actual.DefaultSendEnabled, "DefaultSendEnabled")
		suite.Assert().Len(actual.SendEnabled, 0, "SendEnabled")
	})

	suite.Run("send enabled params converted to store", func() {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		if suite.Assert().Len(actual, 2) {
			suite.Equal("paramscoinfalse", actual[0].Denom, "actual[0].Denom")
			suite.False(actual[0].Enabled, "actual[0].Enabled")
			suite.Equal("paramscointrue", actual[1].Denom, "actual[1].Denom")
			suite.True(actual[1].Enabled, "actual[1].Enabled")
		}
	})
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
