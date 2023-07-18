package keeper_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtime "github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
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

	app         *simapp.SimApp
	ctx         sdk.Context
	queryClient types.QueryClient
}

func (suite *IntegrationTestSuite) initKeepersWithmAccPerms(blockedAddrs map[string]bool) (authkeeper.AccountKeeper, *keeper.BaseKeeper) {
	app := suite.app
	maccPerms := simapp.GetMaccPerms()
	appCodec := simapp.MakeTestEncodingConfig().Codec

	maccPerms[holder] = nil
	maccPerms[authtypes.Burner] = []string{authtypes.Burner}
	maccPerms[authtypes.Minter] = []string{authtypes.Minter}
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}
	maccPerms[randomPerm] = []string{"random"}
	authKeeper := authkeeper.NewAccountKeeper(
		appCodec, app.GetKey(types.StoreKey), app.GetSubspace(types.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
	)
	keeper := keeper.NewBaseKeeper(
		appCodec, app.GetKey(types.StoreKey), authKeeper,
		app.GetSubspace(types.ModuleName), blockedAddrs,
	)

	return authKeeper, keeper
}

func (suite *IntegrationTestSuite) SetupTest() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	app.AccountKeeper.SetParams(ctx, authtypes.DefaultParams())
	app.BankKeeper.SetParams(ctx, types.DefaultParams())

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.BankKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	suite.app = app
	suite.ctx = ctx
	suite.queryClient = queryClient
}

func (suite *IntegrationTestSuite) TestSupply() {
	ctx := suite.ctx

	require := suite.Require()

	// add module accounts to supply keeper
	authKeeper, keeper := suite.initKeepersWithmAccPerms(make(map[string]bool))

	genesisSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	require.NoError(err)

	initialPower := int64(100)
	initTokens := suite.app.StakingKeeper.TokensFromConsensusPower(ctx, initialPower)
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
	supplyAfterBurn, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	suite.Require().NoError(err)
	suite.Require().NoError(err)
	suite.Require().Equal(sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, multiPermAcc.GetName()).String())
	suite.Require().Equal(supplyAfterInflation.Sub(initCoins...), supplyAfterBurn)
}

func (suite *IntegrationTestSuite) TestSendCoinsNewAccount() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Nil(app.AccountKeeper.GetAccount(ctx, addr2))
	app.BankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Empty(app.BankKeeper.GetAllBalances(ctx, addr2))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(50))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	acc1Balances = app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(sendAmt, acc2Balances)
	updatedAcc1Bal := balances.Sub(sendAmt...)
	suite.Require().Len(acc1Balances, len(updatedAcc1Bal))
	suite.Require().Equal(acc1Balances, updatedAcc1Bal)
	suite.Require().NotNil(app.AccountKeeper.GetAccount(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestInputOutputNewAccount() {
	app, ctx := suite.app, suite.ctx

	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Nil(app.AccountKeeper.GetAccount(ctx, addr2))
	suite.Require().Empty(app.BankKeeper.GetAllBalances(ctx, addr2))

	inputs := types.Input{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))}
	outputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	suite.Require().NoError(app.BankKeeper.InputOutputCoins(ctx, inputs, outputs))

	expected := sdk.NewCoins(newFooCoin(30), newBarCoin(10))
	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Equal(expected, acc2Balances)
	suite.Require().NotNil(app.AccountKeeper.GetAccount(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestInputOutputCoins() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(90), newBarCoin(30))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, acc2)

	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	acc3 := app.AccountKeeper.NewAccountWithAddress(ctx, addr3)
	app.AccountKeeper.SetAccount(ctx, acc3)

	input := types.Input{
		Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(60), newBarCoin(20)),
	}
	outputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
		{Address: addr3.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, input, []types.Output{}))
	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, input, outputs))

	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))

	insufficientInput := types.Input{
		Address: addr1.String(),
		Coins:   sdk.NewCoins(newFooCoin(300), newBarCoin(100)),
	}
	insufficientOutputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
		{Address: addr3.String(), Coins: sdk.NewCoins(newFooCoin(300), newBarCoin(100))},
	}
	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, insufficientInput, insufficientOutputs))
	suite.Require().NoError(app.BankKeeper.InputOutputCoins(ctx, input, outputs))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(30), newBarCoin(10))
	suite.Require().Equal(expected, acc1Balances)

	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	suite.Require().Equal(expected, acc2Balances)

	acc3Balances := app.BankKeeper.GetAllBalances(ctx, addr3)
	suite.Require().Equal(expected, acc3Balances)
}

func (suite *IntegrationTestSuite) TestInputOutputCoinsWithRestrictions() {
	type restrictionArgs struct {
		ctx      sdk.Context
		fromAddr sdk.AccAddress
		toAddr   sdk.AccAddress
		amt      sdk.Coins
	}
	var actualRestrictionArgs []*restrictionArgs
	restrictionError := func(messages ...string) types.SendRestrictionFn {
		i := -1
		return func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
			actualRestrictionArgs = append(actualRestrictionArgs, &restrictionArgs{
				ctx:      ctx,
				fromAddr: fromAddr,
				toAddr:   toAddr,
				amt:      amt,
			})
			i++
			if i < len(messages) {
				if len(messages[i]) > 0 {
					return nil, errors.New(messages[i])
				}
			}
			return toAddr, nil
		}
	}
	restrictionPassthrough := func() types.SendRestrictionFn {
		return func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
			actualRestrictionArgs = append(actualRestrictionArgs, &restrictionArgs{
				ctx:      ctx,
				fromAddr: fromAddr,
				toAddr:   toAddr,
				amt:      amt,
			})
			return toAddr, nil
		}
	}
	restrictionNewTo := func(newToAddrs ...sdk.AccAddress) types.SendRestrictionFn {
		i := -1
		return func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
			actualRestrictionArgs = append(actualRestrictionArgs, &restrictionArgs{
				ctx:      ctx,
				fromAddr: fromAddr,
				toAddr:   toAddr,
				amt:      amt,
			})
			i++
			if i < len(newToAddrs) {
				if len(newToAddrs[i]) > 0 {
					return newToAddrs[i], nil
				}
			}
			return toAddr, nil
		}
	}
	type expBals struct {
		from sdk.Coins
		to1  sdk.Coins
		to2  sdk.Coins
	}

	app, setupCtx := suite.app, suite.ctx

	addr1 := sdk.AccAddress("addr1_iocoinsr______")
	addr2 := sdk.AccAddress("addr2_iocoinsr______")
	addr3 := sdk.AccAddress("addr3_iocoinsr______")

	balances := sdk.NewCoins(newFooCoin(1000), newBarCoin(500))
	fromAddr := addr1
	toAddr1 := addr2
	toAddr2 := addr3

	acc1 := app.AccountKeeper.NewAccountWithAddress(setupCtx, addr1)
	app.AccountKeeper.SetAccount(setupCtx, acc1)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, setupCtx, addr1, balances))
	acc2 := app.AccountKeeper.NewAccountWithAddress(setupCtx, addr2)
	app.AccountKeeper.SetAccount(setupCtx, acc2)
	acc3 := app.AccountKeeper.NewAccountWithAddress(setupCtx, addr3)
	app.AccountKeeper.SetAccount(setupCtx, acc3)

	tests := []struct {
		name        string
		fn          types.SendRestrictionFn
		inputCoins  sdk.Coins
		outputs     []types.Output
		outputAddrs []sdk.AccAddress
		expArgs     []*restrictionArgs
		expErr      string
		expBals     expBals
	}{
		{
			name:        "nil restriction",
			fn:          nil,
			inputCoins:  sdk.NewCoins(newFooCoin(5)),
			outputs:     []types.Output{{Address: toAddr1.String(), Coins: sdk.NewCoins(newFooCoin(5))}},
			outputAddrs: []sdk.AccAddress{toAddr1},
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(995), newBarCoin(500)),
				to1:  sdk.NewCoins(newFooCoin(5)),
				to2:  sdk.Coins{},
			},
		},
		{
			name:        "passthrough restriction single output",
			fn:          restrictionPassthrough(),
			inputCoins:  sdk.NewCoins(newFooCoin(10)),
			outputs:     []types.Output{{Address: toAddr1.String(), Coins: sdk.NewCoins(newFooCoin(10))}},
			outputAddrs: []sdk.AccAddress{toAddr1},
			expArgs:     []*restrictionArgs{{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newFooCoin(10))}},
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(985), newBarCoin(500)),
				to1:  sdk.NewCoins(newFooCoin(15)),
				to2:  sdk.Coins{},
			},
		},
		{
			name:        "new to restriction single output",
			fn:          restrictionNewTo(toAddr2),
			inputCoins:  sdk.NewCoins(newFooCoin(26)),
			outputs:     []types.Output{{Address: toAddr1.String(), Coins: sdk.NewCoins(newFooCoin(26))}},
			outputAddrs: []sdk.AccAddress{toAddr2},
			expArgs:     []*restrictionArgs{{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newFooCoin(26))}},
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(959), newBarCoin(500)),
				to1:  sdk.NewCoins(newFooCoin(15)),
				to2:  sdk.NewCoins(newFooCoin(26)),
			},
		},
		{
			name:        "error restriction single output",
			fn:          restrictionError("restriction test error"),
			inputCoins:  sdk.NewCoins(newBarCoin(88)),
			outputs:     []types.Output{{Address: toAddr1.String(), Coins: sdk.NewCoins(newBarCoin(88))}},
			outputAddrs: []sdk.AccAddress{},
			expArgs:     []*restrictionArgs{{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newBarCoin(88))}},
			expErr:      "restriction test error",
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(959), newBarCoin(500)),
				to1:  sdk.NewCoins(newFooCoin(15)),
				to2:  sdk.NewCoins(newFooCoin(26)),
			},
		},
		{
			name:       "passthrough restriction two outputs",
			fn:         restrictionPassthrough(),
			inputCoins: sdk.NewCoins(newFooCoin(11), newBarCoin(12)),
			outputs: []types.Output{
				{Address: toAddr1.String(), Coins: sdk.NewCoins(newFooCoin(11))},
				{Address: toAddr2.String(), Coins: sdk.NewCoins(newBarCoin(12))},
			},
			outputAddrs: []sdk.AccAddress{toAddr1, toAddr2},
			expArgs: []*restrictionArgs{
				{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newFooCoin(11))},
				{fromAddr: fromAddr, toAddr: toAddr2, amt: sdk.NewCoins(newBarCoin(12))},
			},
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(948), newBarCoin(488)),
				to1:  sdk.NewCoins(newFooCoin(26)),
				to2:  sdk.NewCoins(newFooCoin(26), newBarCoin(12)),
			},
		},
		{
			name:       "error restriction two outputs error on second",
			fn:         restrictionError("", "second restriction error"),
			inputCoins: sdk.NewCoins(newFooCoin(44)),
			outputs: []types.Output{
				{Address: toAddr1.String(), Coins: sdk.NewCoins(newFooCoin(12))},
				{Address: toAddr2.String(), Coins: sdk.NewCoins(newFooCoin(32))},
			},
			outputAddrs: []sdk.AccAddress{toAddr1},
			expArgs: []*restrictionArgs{
				{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newFooCoin(12))},
				{fromAddr: fromAddr, toAddr: toAddr2, amt: sdk.NewCoins(newFooCoin(32))},
			},
			expErr: "second restriction error",
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(948), newBarCoin(488)),
				to1:  sdk.NewCoins(newFooCoin(26)),
				to2:  sdk.NewCoins(newFooCoin(26), newBarCoin(12)),
			},
		},
		{
			name:       "new to restriction two outputs",
			fn:         restrictionNewTo(toAddr2, toAddr1),
			inputCoins: sdk.NewCoins(newBarCoin(35)),
			outputs: []types.Output{
				{Address: toAddr1.String(), Coins: sdk.NewCoins(newBarCoin(10))},
				{Address: toAddr2.String(), Coins: sdk.NewCoins(newBarCoin(25))},
			},
			outputAddrs: []sdk.AccAddress{toAddr1, toAddr2},
			expArgs: []*restrictionArgs{
				{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newBarCoin(10))},
				{fromAddr: fromAddr, toAddr: toAddr2, amt: sdk.NewCoins(newBarCoin(25))},
			},
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(948), newBarCoin(453)),
				to1:  sdk.NewCoins(newFooCoin(26), newBarCoin(25)),
				to2:  sdk.NewCoins(newFooCoin(26), newBarCoin(22)),
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			baseKeeper, ok := app.BankKeeper.(*keeper.BaseKeeper)
			suite.Require().True(ok, "app.BankKeeper.(*keeper.BaseKeeper")
			existingSendRestrictionFn := baseKeeper.GetSendRestrictionFn()
			defer baseKeeper.SetSendRestriction(existingSendRestrictionFn)
			actualRestrictionArgs = nil
			baseKeeper.SetSendRestriction(tc.fn)

			ctx, writeCache := suite.ctx.CacheContext()
			input := types.Input{
				Address: fromAddr.String(),
				Coins:   tc.inputCoins,
			}

			var err error
			testFunc := func() {
				err = baseKeeper.InputOutputCoins(ctx, input, tc.outputs)
			}
			suite.Require().NotPanics(testFunc, "InputOutputCoins")
			if err == nil {
				writeCache()
			}
			if len(tc.expErr) > 0 {
				suite.Assert().EqualError(err, tc.expErr, "InputOutputCoins error")
			} else {
				suite.Assert().NoError(err, "InputOutputCoins error")
			}
			if len(tc.expArgs) > 0 {
				for i, expArgs := range tc.expArgs {
					suite.Assert().Equal(ctx, actualRestrictionArgs[i].ctx, "[%d] ctx provided to restriction", i)
					suite.Assert().Equal(expArgs.fromAddr, actualRestrictionArgs[i].fromAddr, "[%d] fromAddr provided to restriction", i)
					suite.Assert().Equal(expArgs.toAddr, actualRestrictionArgs[i].toAddr, "[%d] toAddr provided to restriction", i)
					suite.Assert().Equal(expArgs.amt.String(), actualRestrictionArgs[i].amt.String(), "[%d] amt provided to restriction", i)
				}
			} else {
				suite.Assert().Nil(actualRestrictionArgs, "args provided to a restriction")
			}
			fromBal := baseKeeper.GetAllBalances(suite.ctx, fromAddr)
			suite.Assert().Equal(tc.expBals.from.String(), fromBal.String(), "fromAddr balance")
			to1Bal := baseKeeper.GetAllBalances(suite.ctx, toAddr1)
			suite.Assert().Equal(tc.expBals.to1.String(), to1Bal.String(), "toAddr1 balance")
			to2Bal := baseKeeper.GetAllBalances(suite.ctx, toAddr2)
			suite.Assert().Equal(tc.expBals.to2.String(), to2Bal.String(), "toAddr2 balance")
		})
	}
}

func (suite *IntegrationTestSuite) TestSendCoins() {
	app, ctx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress("addr1_______________")
	acc1 := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress("addr2_______________")
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	app.AccountKeeper.SetAccount(ctx, acc2)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, balances))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc1Balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	suite.Require().Equal(expected, acc1Balances)

	acc2Balances := app.BankKeeper.GetAllBalances(ctx, addr2)
	expected = sdk.NewCoins(newFooCoin(150), newBarCoin(75))
	suite.Require().Equal(expected, acc2Balances)

	// we sent all foo coins to acc2, so foo balance should be deleted for acc1 and bar should be still there
	var coins []sdk.Coin
	app.BankKeeper.IterateAccountBalances(ctx, addr1, func(c sdk.Coin) (stop bool) {
		coins = append(coins, c)
		return true
	})
	suite.Require().Len(coins, 1)
	suite.Require().Equal(newBarCoin(25), coins[0], "expected only bar coins in the account balance, got: %v", coins)
}

func (suite *IntegrationTestSuite) TestSendCoinsWithRestrictions() {
	type restrictionArgs struct {
		ctx      sdk.Context
		fromAddr sdk.AccAddress
		toAddr   sdk.AccAddress
		amt      sdk.Coins
	}
	var actualRestrictionArgs *restrictionArgs
	restrictionError := func(message string) types.SendRestrictionFn {
		return func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
			actualRestrictionArgs = &restrictionArgs{
				ctx:      ctx,
				fromAddr: fromAddr,
				toAddr:   toAddr,
				amt:      amt,
			}
			return nil, errors.New(message)
		}
	}
	restrictionPassthrough := func() types.SendRestrictionFn {
		return func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
			actualRestrictionArgs = &restrictionArgs{
				ctx:      ctx,
				fromAddr: fromAddr,
				toAddr:   toAddr,
				amt:      amt,
			}
			return toAddr, nil
		}
	}
	restrictionNewTo := func(newToAddr sdk.AccAddress) types.SendRestrictionFn {
		return func(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amt sdk.Coins) (sdk.AccAddress, error) {
			actualRestrictionArgs = &restrictionArgs{
				ctx:      ctx,
				fromAddr: fromAddr,
				toAddr:   toAddr,
				amt:      amt,
			}
			return newToAddr, nil
		}
	}
	type expBals struct {
		from sdk.Coins
		to1  sdk.Coins
		to2  sdk.Coins
	}

	addr1 := sdk.AccAddress("addr1_sendcoinsr____")
	addr2 := sdk.AccAddress("addr2_sendcoinsr____")
	addr3 := sdk.AccAddress("addr3_sendcoinsr____")

	app, setupCtx := suite.app, suite.ctx
	balances := sdk.NewCoins(newFooCoin(1000), newBarCoin(500))
	fromAddr := addr1
	toAddr1 := addr2
	toAddr2 := addr3

	acc1 := app.AccountKeeper.NewAccountWithAddress(setupCtx, addr1)
	app.AccountKeeper.SetAccount(setupCtx, acc1)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, setupCtx, addr1, balances))
	acc2 := app.AccountKeeper.NewAccountWithAddress(setupCtx, addr2)
	app.AccountKeeper.SetAccount(setupCtx, acc2)
	acc3 := app.AccountKeeper.NewAccountWithAddress(setupCtx, addr3)
	app.AccountKeeper.SetAccount(setupCtx, acc3)

	tests := []struct {
		name      string
		fn        types.SendRestrictionFn
		toAddr    sdk.AccAddress
		finalAddr sdk.AccAddress
		amt       sdk.Coins
		expArgs   *restrictionArgs
		expErr    string
		expBals   expBals
	}{
		{
			name:      "nil restriction",
			fn:        nil,
			toAddr:    toAddr1,
			finalAddr: toAddr1,
			amt:       sdk.NewCoins(newFooCoin(5)),
			expArgs:   nil,
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(995), newBarCoin(500)),
				to1:  sdk.NewCoins(newFooCoin(5)),
				to2:  sdk.Coins{},
			},
		},
		{
			name:      "passthrough restriction",
			fn:        restrictionPassthrough(),
			toAddr:    toAddr1,
			finalAddr: toAddr1,
			amt:       sdk.NewCoins(newFooCoin(10)),
			expArgs:   &restrictionArgs{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newFooCoin(10))},
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(985), newBarCoin(500)),
				to1:  sdk.NewCoins(newFooCoin(15)),
				to2:  sdk.Coins{},
			},
		},
		{
			name:      "new to addr restriction",
			fn:        restrictionNewTo(toAddr2),
			toAddr:    toAddr1,
			finalAddr: toAddr2,
			amt:       sdk.NewCoins(newBarCoin(27)),
			expArgs:   &restrictionArgs{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newBarCoin(27))},
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(985), newBarCoin(473)),
				to1:  sdk.NewCoins(newFooCoin(15)),
				to2:  sdk.NewCoins(newBarCoin(27)),
			},
		},
		{
			name:      "restriction returns error",
			fn:        restrictionError("test restriction error"),
			toAddr:    toAddr1,
			finalAddr: toAddr1,
			amt:       sdk.NewCoins(newFooCoin(100), newBarCoin(200)),
			expArgs:   &restrictionArgs{fromAddr: fromAddr, toAddr: toAddr1, amt: sdk.NewCoins(newFooCoin(100), newBarCoin(200))},
			expErr:    "test restriction error",
			expBals: expBals{
				from: sdk.NewCoins(newFooCoin(985), newBarCoin(473)),
				to1:  sdk.NewCoins(newFooCoin(15)),
				to2:  sdk.NewCoins(newBarCoin(27)),
			},
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			baseKeeper, ok := app.BankKeeper.(*keeper.BaseKeeper)
			suite.Require().True(ok, "app.BankKeeper.(*keeper.BaseKeeper")
			existingSendRestrictionFn := baseKeeper.GetSendRestrictionFn()
			defer baseKeeper.SetSendRestriction(existingSendRestrictionFn)
			actualRestrictionArgs = nil
			baseKeeper.SetSendRestriction(tc.fn)
			ctx, writeCache := suite.ctx.CacheContext()

			var err error
			testFunc := func() {
				err = baseKeeper.SendCoins(ctx, fromAddr, tc.toAddr, tc.amt)
			}
			suite.Require().NotPanics(testFunc, "SendCoins")
			if err == nil {
				writeCache()
			}
			if len(tc.expErr) > 0 {
				suite.Assert().EqualError(err, tc.expErr, "SendCoins error")
			} else {
				suite.Assert().NoError(err, "SendCoins error")
			}
			if tc.expArgs != nil {
				suite.Assert().Equal(ctx, actualRestrictionArgs.ctx, "ctx provided to restriction")
				suite.Assert().Equal(tc.expArgs.fromAddr, actualRestrictionArgs.fromAddr, "fromAddr provided to restriction")
				suite.Assert().Equal(tc.expArgs.toAddr, actualRestrictionArgs.toAddr, "toAddr provided to restriction")
				suite.Assert().Equal(tc.expArgs.amt.String(), actualRestrictionArgs.amt.String(), "amt provided to restriction")
			} else {
				suite.Assert().Nil(actualRestrictionArgs, "args provided to a restriction")
			}
			fromBal := baseKeeper.GetAllBalances(suite.ctx, fromAddr)
			suite.Assert().Equal(tc.expBals.from.String(), fromBal.String(), "fromAddr balance")
			to1Bal := baseKeeper.GetAllBalances(suite.ctx, toAddr1)
			suite.Assert().Equal(tc.expBals.to1.String(), to1Bal.String(), "toAddr1 balance")
			to2Bal := baseKeeper.GetAllBalances(suite.ctx, toAddr2)
			suite.Assert().Equal(tc.expBals.to2.String(), to2Bal.String(), "toAddr2 balance")
		})
	}
}

func (suite *IntegrationTestSuite) TestValidateBalance() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	suite.Require().Error(app.BankKeeper.ValidateBalance(ctx, addr1))

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	app.AccountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, balances))
	suite.Require().NoError(app.BankKeeper.ValidateBalance(ctx, addr1))

	bacc := authtypes.NewBaseAccountWithAddress(addr2)
	vacc := vesting.NewContinuousVestingAccount(bacc, balances.Add(balances...), now.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, balances))
	suite.Require().Error(app.BankKeeper.ValidateBalance(ctx, addr2))
}

func (suite *IntegrationTestSuite) TestSendEnabled() {
	app, ctx := suite.app, suite.ctx
	enabled := true
	params := types.DefaultParams()
	suite.Require().Equal(enabled, params.DefaultSendEnabled)

	app.BankKeeper.SetParams(ctx, params)

	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())
	fooCoin := sdk.NewCoin("foocoin", sdk.OneInt())
	barCoin := sdk.NewCoin("barcoin", sdk.OneInt())

	// assert with default (all denom) send enabled both Bar and Bond Denom are enabled
	suite.Require().Equal(enabled, app.BankKeeper.IsSendEnabledCoin(ctx, barCoin))
	suite.Require().Equal(enabled, app.BankKeeper.IsSendEnabledCoin(ctx, bondCoin))

	// Both coins should be send enabled.
	err := app.BankKeeper.IsSendEnabledCoins(ctx, fooCoin, bondCoin)
	suite.Require().NoError(err)

	// Set default send_enabled to !enabled, add a foodenom that overrides default as enabled
	params.DefaultSendEnabled = !enabled
	app.BankKeeper.SetParams(ctx, params)
	app.BankKeeper.SetSendEnabled(ctx, fooCoin.Denom, enabled)

	// Expect our specific override to be enabled, others to be !enabled.
	suite.Require().Equal(enabled, app.BankKeeper.IsSendEnabledCoin(ctx, fooCoin))
	suite.Require().Equal(!enabled, app.BankKeeper.IsSendEnabledCoin(ctx, barCoin))
	suite.Require().Equal(!enabled, app.BankKeeper.IsSendEnabledCoin(ctx, bondCoin))

	// Foo coin should be send enabled.
	err = app.BankKeeper.IsSendEnabledCoins(ctx, fooCoin)
	suite.Require().NoError(err)

	// Expect an error when one coin is not send enabled.
	err = app.BankKeeper.IsSendEnabledCoins(ctx, fooCoin, bondCoin)
	suite.Require().Error(err)

	// Expect an error when all coins are not send enabled.
	err = app.BankKeeper.IsSendEnabledCoins(ctx, bondCoin, barCoin)
	suite.Require().Error(err)
}

func (suite *IntegrationTestSuite) TestHasBalance() {
	app, ctx := suite.app, suite.ctx
	addr := sdk.AccAddress([]byte("addr1_______________"))

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	app.AccountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	suite.Require().False(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(99)))

	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, balances))
	suite.Require().False(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(101)))
	suite.Require().True(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(100)))
	suite.Require().True(app.BankKeeper.HasBalance(ctx, addr, newFooCoin(1)))
}

func (suite *IntegrationTestSuite) TestMsgSendEvents() {
	app, ctx := suite.app, suite.ctx
	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)

	app.AccountKeeper.SetAccount(ctx, acc)
	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, newCoins))

	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr, addr2, newCoins))
	event1 := sdk.Event{
		Type:       types.EventTypeTransfer,
		Attributes: []abci.EventAttribute{},
	}
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeyRecipient), Value: []byte(addr2.String())},
	)
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())},
	)
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: []byte(sdk.AttributeKeyAmount), Value: []byte(newCoins.String())},
	)

	event2 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event2.Attributes = append(
		event2.Attributes,
		abci.EventAttribute{Key: []byte(types.AttributeKeySender), Value: []byte(addr.String())},
	)

	// events are shifted due to the funding account events
	events := ctx.EventManager().ABCIEvents()
	suite.Require().Equal(10, len(events))
	suite.Require().Equal(abci.Event(event1), events[8])
	suite.Require().Equal(abci.Event(event2), events[9])
}

func (suite *IntegrationTestSuite) TestMsgMultiSendEvents() {
	app, ctx := suite.app, suite.ctx

	app.BankKeeper.SetParams(ctx, types.DefaultParams())

	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	addr4 := sdk.AccAddress([]byte("addr4_______________"))
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr)
	acc2 := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, acc2)

	coins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50), sdk.NewInt64Coin(barDenom, 100))
	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))
	newCoins2 := sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))
	input := types.Input{
		Address: addr.String(), Coins: coins,
	}
	outputs := []types.Output{
		{Address: addr3.String(), Coins: newCoins},
		{Address: addr4.String(), Coins: newCoins2},
	}

	abciEventsStrings := func(events []abci.Event) []string {
		rv := []string(nil)
		for i, event := range events {
			if len(event.Type) == 0 {
				rv = append(rv, fmt.Sprintf("[%d]{ignored}", i))
			}
			for j, attr := range event.Attributes {
				rv = append(rv, fmt.Sprintf("[%d]%s[%d]: %q = %q", i, event.Type, j, string(attr.Key), string(attr.Value)))
			}
		}
		return rv
	}

	abciEvent := func(typeName string, attrs ...abci.EventAttribute) abci.Event {
		return abci.Event{
			Type:       typeName,
			Attributes: attrs,
		}
	}
	abciAttr := func(key, value string) abci.EventAttribute {
		return abci.EventAttribute{
			Key:   []byte(key),
			Value: []byte(value),
		}
	}

	coinReceived := types.EventTypeCoinReceived
	transfer := types.EventTypeTransfer
	receiver := types.AttributeKeyReceiver
	recipient := types.AttributeKeyRecipient
	amount := sdk.AttributeKeyAmount

	expFailEvents := []abci.Event{}

	emF := sdk.NewEventManager()
	ctx = ctx.WithEventManager(emF)
	suite.Require().Error(app.BankKeeper.InputOutputCoins(ctx, input, outputs))
	events := emF.ABCIEvents()
	if !suite.Assert().Equal(expFailEvents, events, "events after InputOutputCoins failure") {
		suite.T().Logf("Expected Events:\n%s", strings.Join(abciEventsStrings(expFailEvents), "\n"))
		suite.T().Logf("Actual Events:\n%s", strings.Join(abciEventsStrings(events), "\n"))
	}

	// Set addr's coins
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr, coins))
	newCoins = sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))

	coinSpent := types.EventTypeCoinSpent
	spender := types.AttributeKeySpender
	message := sdk.EventTypeMessage
	sender := types.AttributeKeySender

	expPassEvents := []abci.Event{
		abciEvent(coinSpent, abciAttr(spender, addr.String()), abciAttr(amount, coins.String())),
		abciEvent(message, abciAttr(sender, addr.String())),
		abciEvent(coinReceived, abciAttr(receiver, addr3.String()), abciAttr(amount, newCoins.String())),
		abciEvent(transfer, abciAttr(recipient, addr3.String()), abciAttr(amount, newCoins.String())),
		abciEvent(coinReceived, abciAttr(receiver, addr4.String()), abciAttr(amount, newCoins2.String())),
		abciEvent(transfer, abciAttr(recipient, addr4.String()), abciAttr(amount, newCoins2.String())),
	}

	emS := sdk.NewEventManager()
	ctx = ctx.WithEventManager(emS)
	suite.Require().NoError(app.BankKeeper.InputOutputCoins(ctx, input, outputs))
	events = emS.ABCIEvents()
	if !suite.Assert().Equal(expPassEvents, events, "events after InputOutputCoins success") {
		suite.T().Logf("Expected Events:\n%s", strings.Join(abciEventsStrings(expPassEvents), "\n"))
		suite.T().Logf("Actual Events:\n%s", strings.Join(abciEventsStrings(events), "\n"))
	}
}

func (suite *IntegrationTestSuite) TestGetLockedCoinsFnWrapper() {
	// Not using the coin/coins constructors since we want to be able to create bad ones.
	c := func(amt int64, denom string) sdk.Coin {
		return sdk.Coin{Amount: sdk.NewInt(amt), Denom: denom}
	}
	cz := func(coins ...sdk.Coin) sdk.Coins {
		return coins
	}

	tests := []struct {
		name string
		rv   sdk.Coins
		exp  sdk.Coins
	}{
		{
			name: "one positive coin",
			rv:   cz(c(1, "okcoin")),
			exp:  cz(c(1, "okcoin")),
		},
		{
			name: "one zero coin",
			rv:   cz(c(0, "zerocoin")),
			exp:  sdk.Coins{},
		},
		{
			name: "one negative coin",
			rv:   cz(c(-1, "negcoin")),
			exp:  sdk.Coins{},
		},
		{
			name: "one positive one zero and one negative",
			rv:   cz(c(1, "okcoin"), c(0, "zerocoin"), c(-1, "negcoin")),
			exp:  cz(c(1, "okcoin")),
		},
		{
			name: "two of same denom both positive",
			rv:   cz(c(1, "twocoin"), c(4, "twocoin")),
			exp:  cz(c(5, "twocoin")),
		},
		{
			name: "two of same denom but one is negative",
			rv:   cz(c(1, "badcoin"), c(-1, "badcoin")),
			exp:  cz(c(1, "badcoin")),
		},
		{
			name: "a bit of everything",
			rv:   cz(c(-1, "weird"), c(1, "weird"), c(0, "weird"), c(8, "okay"), c(4, "weird"), c(0, "zero")),
			exp:  cz(c(8, "okay"), c(5, "weird")),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			testAddr := sdk.AccAddress("testAddr____________")
			var addr sdk.AccAddress
			getter := func(_ sdk.Context, getterAddr sdk.AccAddress) sdk.Coins {
				addr = getterAddr
				return tc.rv
			}
			wrapped := keeper.GetLockedCoinsFnWrapper(getter)
			coins := wrapped(sdk.Context{}, testAddr)
			suite.Assert().Equal(tc.exp.String(), coins.String(), "wrapped result")
			suite.Assert().Equal(testAddr, addr, "address provided to wrapped getter")
		})
	}
}

func (suite *IntegrationTestSuite) TestSpendableCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule)
	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, macc)
	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.SpendableCoins(ctx, addr2))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	suite.Require().Equal(origCoins.Sub(delCoins...), app.BankKeeper.SpendableCoins(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestSpendableCoinsWithInjection() {
	app, ctx := suite.app, suite.ctx
	now := time.Unix(1713586800, 0).UTC()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(12 * time.Hour)
	bk := app.BankKeeper.(*keeper.BaseKeeper)
	origGetter := bk.GetLockedCoinsGetter()
	defer bk.SetLockedCoinsGetter(origGetter)

	baseStake := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	baseOther := sdk.NewCoins(sdk.NewInt64Coin("other", 50))
	baseCoins := baseStake.Add(baseOther...)
	vestingCoins := sdk.NewCoins(sdk.NewInt64Coin("vest", 88))
	expVestingBalances := baseCoins.Add(vestingCoins...)

	addr1 := sdk.AccAddress("addr1_______________")
	addr2 := sdk.AccAddress("addr2_______________")
	addrV := sdk.AccAddress("addrV_______________")

	bacc := authtypes.NewBaseAccountWithAddress(addrV)
	vacc := vesting.NewDelayedVestingAccount(bacc, vestingCoins, endTime.Unix())
	app.AccountKeeper.SetAccount(ctx, vacc)

	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, baseCoins), "FundAccount(addr1)")
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, baseCoins), "FundAccount(addr2)")
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addrV, expVestingBalances), "FundAccount(addrV)")

	ctx = ctx.WithBlockTime(now.Add(1 * time.Hour))

	bal1 := app.BankKeeper.GetAllBalances(ctx, addr1)
	bal2 := app.BankKeeper.GetAllBalances(ctx, addr2)
	balV := app.BankKeeper.GetAllBalances(ctx, addrV)
	suite.Require().Equal(baseCoins.String(), bal1.String(), "GetAllBalances(addr1)")
	suite.Require().Equal(baseCoins.String(), bal2.String(), "GetAllBalances(addr2)")
	suite.Require().Equal(expVestingBalances.String(), balV.String(), "GetAllBalances(addrV)")

	tests := []struct {
		name     string
		setup    func()
		expAddr1 sdk.Coins
		expAddr2 sdk.Coins
		expAddrV sdk.Coins
	}{
		{
			name:     "no locked coins getter",
			expAddr1: baseCoins,
			expAddr2: baseCoins,
			expAddrV: expVestingBalances,
		},
		{
			name:     "with only unvested coins getter",
			setup:    func() { app.BankKeeper.AppendLockedCoinsGetter(app.BankKeeper.UnvestedCoins) },
			expAddr1: baseCoins,
			expAddr2: baseCoins,
			expAddrV: baseCoins,
		},
		{
			name: "with extra locked coins getters too",
			setup: func() {
				app.BankKeeper.AppendLockedCoinsGetter(app.BankKeeper.UnvestedCoins)
				app.BankKeeper.AppendLockedCoinsGetter(func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
					if addr1.Equals(addr) {
						return sdk.NewCoins(sdk.NewInt64Coin("stake", 15), sdk.NewInt64Coin("other", 32))
					}
					return sdk.NewCoins()
				})
				app.BankKeeper.AppendLockedCoinsGetter(func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
					if addr1.Equals(addr) {
						return sdk.NewCoins(sdk.NewInt64Coin("stake", 10), sdk.NewInt64Coin("other", 18))
					}
					return sdk.NewCoins()
				})
			},
			expAddr1: sdk.NewCoins(sdk.NewInt64Coin("stake", 75)),
			expAddr2: baseCoins,
			expAddrV: baseCoins,
		},
		{
			name: "only appended getter that returns a negative amount",
			setup: func() {
				app.BankKeeper.AppendLockedCoinsGetter(func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
					return sdk.Coins{sdk.Coin{Denom: "stake", Amount: sdk.NewInt(-1)}}
				})
			},
			expAddr1: baseCoins,
			expAddr2: baseCoins,
			expAddrV: expVestingBalances,
		},
		{
			name: "only prepended getter that returns a negative amount",
			setup: func() {
				app.BankKeeper.PrependLockedCoinsGetter(func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
					return sdk.Coins{sdk.Coin{Denom: "stake", Amount: sdk.NewInt(-1)}}
				})
			},
			expAddr1: baseCoins,
			expAddr2: baseCoins,
			expAddrV: expVestingBalances,
		},
		{
			name: "getters return more than account has for a denom",
			setup: func() {
				app.BankKeeper.AppendLockedCoinsGetter(func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
					return sdk.NewCoins(sdk.NewInt64Coin("stake", 1))
				})
				app.BankKeeper.AppendLockedCoinsGetter(func(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
					return baseStake
				})
			},
			expAddr1: baseOther,
			expAddr2: baseOther,
			expAddrV: baseOther.Add(vestingCoins...),
		},
	}

	for _, tc := range tests {
		suite.Run(tc.name, func() {
			app.BankKeeper.ClearLockedCoinsGetter()
			if tc.setup != nil {
				tc.setup()
			}

			spend1 := app.BankKeeper.SpendableCoins(ctx, addr1)
			suite.Assert().Equal(tc.expAddr1.String(), spend1.String(), "SpendableCoins(addr1)")
			spend2 := app.BankKeeper.SpendableCoins(ctx, addr2)
			suite.Assert().Equal(tc.expAddr2.String(), spend2.String(), "SpendableCoins(addr2)")
			spendV := app.BankKeeper.SpendableCoins(ctx, addrV)
			suite.Assert().Equal(tc.expAddrV.String(), spendV.String(), "SpendableCoins(addrV)")
		})
	}
}

func (suite *IntegrationTestSuite) TestVestingAccountSend() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))

	// receive some coins
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, sendCoins))
	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestPeriodicVestingAccountSend() {
	app, ctx := suite.app, suite.ctx
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

	app.AccountKeeper.SetAccount(ctx, vacc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	suite.Require().Error(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))

	// receive some coins
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, sendCoins))

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
}

func (suite *IntegrationTestSuite) TestVestingAccountReceive() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	// send some coins to the vesting account
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(origCoins.Add(sendCoins...), balances)
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now)...), sendCoins)

	// require coins are spendable plus any that have vested
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))...), origCoins)
}

func (suite *IntegrationTestSuite) TestPeriodicVestingAccountReceive() {
	app, ctx := suite.app, suite.ctx
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
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	// send some coins to the vesting account
	suite.Require().NoError(app.BankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = app.AccountKeeper.GetAccount(ctx, addr1).(*vesting.PeriodicVestingAccount)
	balances := app.BankKeeper.GetAllBalances(ctx, addr1)
	suite.Require().Equal(origCoins.Add(sendCoins...), balances)
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now)...), sendCoins)

	// require coins are spendable plus any that have vested
	suite.Require().Equal(balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))...), origCoins)
}

func (suite *IntegrationTestSuite) TestDelegateCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	suite.Require().Equal(origCoins.Sub(delCoins...), app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addr1))

	// require that delegated vesting amount is equal to what was delegated with DelegateCoins
	acc = app.AccountKeeper.GetAccount(ctx, addr1)
	vestingAcc, ok := acc.(types.VestingAccount)
	suite.Require().True(ok)
	suite.Require().Equal(delCoins, vestingAcc.GetDelegatedVesting())
}

func (suite *IntegrationTestSuite) TestDelegateCoinsWithRestriction() {
	app, ctx := suite.app, suite.ctx

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress("addr1_______________")
	addrModule := sdk.AccAddress("moduleAcc___________")

	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing

	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))

	// Now that the accounts are set up and funded, add a send restriction that just blocks everything.
	expErr := "this is a test restriction: restriction of no"
	restrictionOfNo := func(_ sdk.Context, _, _ sdk.AccAddress, _ sdk.Coins) (sdk.AccAddress, error) {
		return nil, errors.New(expErr)
	}
	app.BankKeeper.AppendSendRestriction(restrictionOfNo)

	err := app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins)
	suite.Require().EqualError(err, expErr, "DelegateCoins")
}

func (suite *IntegrationTestSuite) TestDelegateCoins_Invalid() {
	app, ctx := suite.app, suite.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)

	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	invalidCoins := sdk.Coins{sdk.Coin{Denom: "fooDenom", Amount: sdk.NewInt(-50)}}
	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, invalidCoins))

	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	app.AccountKeeper.SetAccount(ctx, acc)
	suite.Require().Error(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, origCoins.Add(origCoins...)))
}

func (suite *IntegrationTestSuite) TestUndelegateCoins() {
	app, ctx := suite.app, suite.ctx
	now := tmtime.Now()
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing

	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr2)

	app.AccountKeeper.SetAccount(ctx, vacc)
	app.AccountKeeper.SetAccount(ctx, acc)
	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := app.BankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins)
	suite.Require().NoError(err)

	suite.Require().Equal(origCoins.Sub(delCoins...), app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a non-vesting account to undelegate
	suite.Require().NoError(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr2, delCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr2))
	suite.Require().True(app.BankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require the ability for a vesting account to delegate
	suite.Require().NoError(app.BankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))

	suite.Require().Equal(origCoins.Sub(delCoins...), app.BankKeeper.GetAllBalances(ctx, addr1))
	suite.Require().Equal(delCoins, app.BankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to undelegate
	suite.Require().NoError(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	suite.Require().Equal(origCoins, app.BankKeeper.GetAllBalances(ctx, addr1))
	suite.Require().True(app.BankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require that delegated vesting amount is completely empty, since they were completely undelegated
	acc = app.AccountKeeper.GetAccount(ctx, addr1)
	vestingAcc, ok := acc.(types.VestingAccount)
	suite.Require().True(ok)
	suite.Require().Empty(vestingAcc.GetDelegatedVesting())
}

func (suite *IntegrationTestSuite) TestUndelegateCoins_Invalid() {
	app, ctx := suite.app, suite.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))
	macc := app.AccountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := app.AccountKeeper.NewAccountWithAddress(ctx, addr1)

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	app.AccountKeeper.SetAccount(ctx, macc)
	suite.Require().NoError(testutil.FundAccount(app.BankKeeper, ctx, addr1, origCoins))

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))
	app.AccountKeeper.SetAccount(ctx, acc)

	suite.Require().Error(app.BankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))
}

func (suite *IntegrationTestSuite) TestSetDenomMetaData() {
	app, ctx := suite.app, suite.ctx

	metadata := suite.getTestMetadata()

	for i := range []int{1, 2} {
		app.BankKeeper.SetDenomMetaData(ctx, metadata[i])
	}

	actualMetadata, found := app.BankKeeper.GetDenomMetaData(ctx, metadata[1].Base)
	suite.Require().True(found)
	found = app.BankKeeper.HasDenomMetaData(ctx, metadata[1].Base)
	suite.Require().True(found)
	suite.Require().Equal(metadata[1].GetBase(), actualMetadata.GetBase())
	suite.Require().Equal(metadata[1].GetDisplay(), actualMetadata.GetDisplay())
	suite.Require().Equal(metadata[1].GetDescription(), actualMetadata.GetDescription())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetDenom(), actualMetadata.GetDenomUnits()[1].GetDenom())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetExponent(), actualMetadata.GetDenomUnits()[1].GetExponent())
	suite.Require().Equal(metadata[1].GetDenomUnits()[1].GetAliases(), actualMetadata.GetDenomUnits()[1].GetAliases())
}

func (suite *IntegrationTestSuite) TestIterateAllDenomMetaData() {
	app, ctx := suite.app, suite.ctx

	expectedMetadata := suite.getTestMetadata()
	// set metadata
	for i := range []int{1, 2} {
		app.BankKeeper.SetDenomMetaData(ctx, expectedMetadata[i])
	}
	// retrieve metadata
	actualMetadata := make([]types.Metadata, 0)
	app.BankKeeper.IterateAllDenomMetaData(ctx, func(metadata types.Metadata) bool {
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

	suite.app.AccountKeeper = authkeeper.NewAccountKeeper(
		suite.app.AppCodec(), suite.app.GetKey(authtypes.StoreKey), suite.app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
	)

	suite.app.BankKeeper = keeper.NewBaseKeeper(suite.app.AppCodec(), suite.app.GetKey(types.StoreKey),
		suite.app.AccountKeeper, suite.app.GetSubspace(types.ModuleName), nil,
	)

	// set account with multiple permissions
	suite.app.AccountKeeper.SetModuleAccount(suite.ctx, multiPermAcc)
	// mint coins
	suite.Require().NoError(
		suite.app.BankKeeper.MintCoins(
			suite.ctx,
			multiPermAcc.Name,
			sdk.NewCoins(sdk.NewCoin("utxo", sdk.NewInt(100000)))),
	)
	// send coins to address
	addr1 := sdk.AccAddress("addr1_______________")
	suite.Require().NoError(
		suite.app.BankKeeper.SendCoinsFromModuleToAccount(
			suite.ctx,
			multiPermAcc.Name,
			addr1,
			sdk.NewCoins(sdk.NewCoin("utxo", sdk.NewInt(50000))),
		),
	)

	// burn coins from module account
	suite.Require().NoError(
		suite.app.BankKeeper.BurnCoins(
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
	suite.Require().True(suite.app.BankKeeper.HasSupply(suite.ctx, "utxo"))
	savedSupply := suite.app.BankKeeper.GetSupply(suite.ctx, "utxo")
	utxoSupply := savedSupply
	suite.Require().Equal(utxoSupply.Amount, supply.AmountOf("utxo"))
	// iterate accounts and check balances
	suite.app.BankKeeper.IterateAllBalances(suite.ctx, func(address sdk.AccAddress, coin sdk.Coin) (stop bool) {
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
				{"uatom", uint32(0), []string{"microatom"}},
				{"matom", uint32(3), []string{"milliatom"}},
				{"atom", uint32(6), nil},
			},
			Base:    "uatom",
			Display: "atom",
		},
		{
			Name:        "Token",
			Symbol:      "TOKEN",
			Description: "The native staking token of the Token Hub.",
			DenomUnits: []*types.DenomUnit{
				{"1token", uint32(5), []string{"decitoken"}},
				{"2token", uint32(4), []string{"centitoken"}},
				{"3token", uint32(7), []string{"dekatoken"}},
			},
			Base:    "utoken",
			Display: "token",
		},
	}
}

func (suite *IntegrationTestSuite) TestMintCoinRestrictions() {
	maccPerms := simapp.GetMaccPerms()
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}

	suite.app.AccountKeeper = authkeeper.NewAccountKeeper(
		suite.app.AppCodec(), suite.app.GetKey(authtypes.StoreKey), suite.app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
	)
	suite.app.AccountKeeper.SetModuleAccount(suite.ctx, multiPermAcc)

	type testCase struct {
		coinsToTry sdk.Coin
		expectPass bool
	}

	tests := []struct {
		name          string
		restrictionFn types.MintingRestrictionFn
		testCases     []testCase
	}{
		{
			"restriction",
			func(ctx sdk.Context, coins sdk.Coins) error {
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
		suite.app.BankKeeper = keeper.NewBaseKeeper(suite.app.AppCodec(), suite.app.GetKey(types.StoreKey),
			suite.app.AccountKeeper, suite.app.GetSubspace(types.ModuleName), nil,
		).WithMintCoinsRestriction(test.restrictionFn)
		for _, tc := range test.testCases {
			if tc.expectPass {
				suite.Require().NoError(
					suite.app.BankKeeper.MintCoins(
						suite.ctx,
						multiPermAcc.Name,
						sdk.NewCoins(tc.coinsToTry),
					),
				)
			} else {
				suite.Require().Error(
					suite.app.BankKeeper.MintCoins(
						suite.ctx,
						multiPermAcc.Name,
						sdk.NewCoins(tc.coinsToTry),
					),
				)
			}
		}
	}

	suite.Run("WithMintCoinsRestriction does not update original", func() {
		mintCoinsOrig := func(ctx sdk.Context, coins sdk.Coins) error {
			return fmt.Errorf("this is the original")
		}
		mintCoinsSecond := func(ctx sdk.Context, coins sdk.Coins) error {
			return fmt.Errorf("no can do: second one")
		}
		origKeeper := suite.app.BankKeeper.WithMintCoinsRestriction(mintCoinsOrig)
		secondKeeper := origKeeper.WithMintCoinsRestriction(mintCoinsSecond)

		amt := sdk.NewCoins(newFooCoin(100))
		// Make sure the original keeper still uses the original minting restriction.
		err := origKeeper.MintCoins(suite.ctx, multiPermAcc.Name, amt)
		suite.Assert().EqualError(err, "this is the original", "origKeeper.MintCoins")

		// Make sure the second keeper has the expected minting restriction.
		err = secondKeeper.MintCoins(suite.ctx, multiPermAcc.Name, amt)
		suite.Assert().EqualError(err, "no can do: second one", "secondKeeper.MintCoins")
	})
}

func (suite *IntegrationTestSuite) TestIsSendEnabledDenom() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

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
		bankKeeper.SetParams(ctx, params)
		for _, tc := range tests {
			suite.T().Run(fmt.Sprintf("%s default %t", tc.denom, def), func(t *testing.T) {
				actual := suite.app.BankKeeper.IsSendEnabledDenom(suite.ctx, tc.denom)
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
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

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
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

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
		bankKeeper.SetParams(ctx, params)
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
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

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
				{"aonecoin", true},
			},
		},
		{
			name: "one false",
			sendEnableds: []*types.SendEnabled{
				{"bonecoin", false},
			},
		},
		{
			name: "two true",
			sendEnableds: []*types.SendEnabled{
				{"conecoin", true},
				{"ctwocoin", true},
			},
		},
		{
			name: "two true false",
			sendEnableds: []*types.SendEnabled{
				{"donecoin", true},
				{"dtwocoin", false},
			},
		},
		{
			name: "two false true",
			sendEnableds: []*types.SendEnabled{
				{"eonecoin", false},
				{"etwocoin", true},
			},
		},
		{
			name: "two false",
			sendEnableds: []*types.SendEnabled{
				{"fonecoin", false},
				{"ftwocoin", false},
			},
		},
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
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
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
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
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	suite.T().Run("no entries to iterate", func(t *testing.T) {
		count := 0
		bankKeeper.IterateSendEnabledEntries(ctx, func(denom string, sendEnabled bool) (stop bool) {
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
		bankKeeper.SetParams(ctx, params)
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
		bankKeeper.IterateSendEnabledEntries(ctx, func(denom string, sendEnabled bool) (stop bool) {
			count++
			return false
		})
		assert.Equal(t, 0, count)
	})
}

func (suite *IntegrationTestSuite) TestGetAllSendEnabledEntries() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

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
		bankKeeper.SetParams(ctx, params)
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

func (suite *IntegrationTestSuite) TestMigrator_Migrate3to4() {
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		bankKeeper.SetParams(ctx, params)
		suite.T().Run(fmt.Sprintf("default %t does not change", def), func(t *testing.T) {
			kp := bankKeeper.(*keeper.BaseKeeper)
			migrator := keeper.NewMigrator(*kp)
			require.NoError(t, migrator.Migrate3to4(ctx))
			actual := bankKeeper.GetParams(ctx)
			assert.Equal(t, params.DefaultSendEnabled, actual.DefaultSendEnabled)
		})
	}

	for _, def := range []bool{true, false} {
		params := types.Params{
			SendEnabled: []*types.SendEnabled{
				{fmt.Sprintf("truecoin%t", def), true},
				{fmt.Sprintf("falsecoin%t", def), false},
			},
		}
		bankKeeper.SetParams(ctx, params)
		suite.T().Run(fmt.Sprintf("default %t send enabled info moved to store", def), func(t *testing.T) {
			kp := bankKeeper.(*keeper.BaseKeeper)
			migrator := keeper.NewMigrator(*kp)
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
	ctx, bankKeeper := suite.ctx, suite.app.BankKeeper
	params := types.NewParams(true)
	params.SendEnabled = []*types.SendEnabled{
		{"paramscointrue", true},
		{"paramscoinfalse", false},
	}
	bankKeeper.SetParams(ctx, params)

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
