package keeper_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	authmodulev1 "cosmossdk.io/api/cosmos/auth/module/v1"
	"cosmossdk.io/math"
	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	cmttime "github.com/cometbft/cometbft/types/time"
	"gotest.tools/v3/assert"

	storetypes "cosmossdk.io/store/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	"github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	_ "github.com/cosmos/cosmos-sdk/x/auth/vesting"
	_ "github.com/cosmos/cosmos-sdk/x/consensus"
	_ "github.com/cosmos/cosmos-sdk/x/params"
	_ "github.com/cosmos/cosmos-sdk/x/staking"
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

func assertElementsMatch(listA, listB []string) bool {
	// Create maps for each slice
	mapA, mapB := make(map[string]int), make(map[string]int)

	// Populate the maps with the elements of each slice
	for _, value := range listA {
		mapA[value]++
	}
	for _, value := range listB {
		mapB[value]++
	}

	return reflect.DeepEqual(mapA, mapB)
}

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

type fixture struct {
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

func initFixture(t assert.TestingT) *fixture {
	f := &fixture{}

	var interfaceRegistry codectypes.InterfaceRegistry

	app, err := sims.Setup(
		configurator.NewAppConfig(
			configurator.AuthModule(),
			configurator.BankModule(),
			configurator.StakingModule(),
			configurator.ParamsModule(),
			configurator.ConsensusModule(),
			configurator.VestingModule()),
		&f.accountKeeper, &f.bankKeeper, &f.stakingKeeper,
		&f.appCodec, &f.authConfig, &interfaceRegistry,
	)
	assert.NilError(t, err)

	ctx := app.BaseApp.NewContext(false, cmtproto.Header{})
	f.ctx = ctx
	f.fetchStoreKey = app.UnsafeFindStoreKey

	queryHelper := baseapp.NewQueryServerTestHelper(f.ctx, interfaceRegistry)
	types.RegisterQueryServer(queryHelper, f.bankKeeper)
	queryClient := types.NewQueryClient(queryHelper)
	types.RegisterInterfaces(interfaceRegistry)

	f.queryClient = queryClient
	f.msgServer = keeper.NewMsgServerImpl(f.bankKeeper)

	return f
}

func initKeepersWithmAccPerms(f *fixture, blockedAddrs map[string]bool) (authkeeper.AccountKeeper, keeper.BaseKeeper) {
	maccPerms := map[string][]string{}
	for _, permission := range f.authConfig.ModuleAccountPermissions {
		maccPerms[permission.Account] = permission.Permissions
	}

	appCodec := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{}, bank.AppModuleBasic{}).Codec

	maccPerms[holder] = nil
	maccPerms[authtypes.Burner] = []string{authtypes.Burner}
	maccPerms[authtypes.Minter] = []string{authtypes.Minter}
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}
	maccPerms[randomPerm] = []string{"random"}
	authKeeper := authkeeper.NewAccountKeeper(
		appCodec, f.fetchStoreKey(types.StoreKey), authtypes.ProtoBaseAccount,
		maccPerms, sdk.Bech32MainPrefix, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	bankKeeper := keeper.NewBaseKeeper(
		appCodec, f.fetchStoreKey(types.StoreKey), authKeeper, blockedAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	return authKeeper, bankKeeper
}

func TestSupply(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	// add module accounts to supply keeper
	authKeeper, keeper := initKeepersWithmAccPerms(f, make(map[string]bool))

	genesisSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)

	initialPower := int64(100)
	initTokens := f.stakingKeeper.TokensFromConsensusPower(ctx, initialPower)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initTokens))

	// set burnerAcc balance
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	assert.NilError(t, keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	assert.NilError(t, keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Minter, burnerAcc.GetAddress(), initCoins))

	total, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)

	expTotalSupply := initCoins.Add(genesisSupply...)
	assert.DeepEqual(t, expTotalSupply, total)

	// burning all supplied tokens
	err = keeper.BurnCoins(ctx, authtypes.Burner, initCoins)
	assert.NilError(t, err)

	total, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)
	assert.DeepEqual(t, total, genesisSupply)
}

func TestSendCoinsFromModuleToAccount_Blocklist(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	// add module accounts to supply keeper
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	_, keeper := initKeepersWithmAccPerms(f, map[string]bool{addr1.String(): true})

	assert.NilError(t, keeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	assert.Error(t, keeper.SendCoinsFromModuleToAccount(
		ctx, minttypes.ModuleName, addr1, initCoins), fmt.Sprintf("%s is not allowed to receive funds: unauthorized", addr1))
}

func TestSupply_SendCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	// add module accounts to supply keeper
	authKeeper, keeper := initKeepersWithmAccPerms(f, make(map[string]bool))

	baseAcc := authKeeper.NewAccountWithAddress(ctx, authtypes.NewModuleAddress("baseAcc"))

	// set initial balances
	assert.NilError(t, keeper.MintCoins(ctx, minttypes.ModuleName, initCoins))

	assert.NilError(t, keeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, holderAcc.GetAddress(), initCoins))

	authKeeper.SetModuleAccount(ctx, holderAcc)
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	authKeeper.SetAccount(ctx, baseAcc)

	testutil.AssertPanics(t, func() {
		_ = keeper.SendCoinsFromModuleToModule(ctx, "", holderAcc.GetName(), initCoins) // nolint:errcheck
	})

	testutil.AssertPanics(t, func() {
		_ = keeper.SendCoinsFromModuleToModule(ctx, authtypes.Burner, "", initCoins) // nolint:errcheck
	})

	testutil.AssertPanics(t, func() {
		_ = keeper.SendCoinsFromModuleToAccount(ctx, "", baseAcc.GetAddress(), initCoins) // nolint:errcheck
	})

	assert.Error(t,
		keeper.SendCoinsFromModuleToAccount(ctx, holderAcc.GetName(), baseAcc.GetAddress(), initCoins.Add(initCoins...)),
		fmt.Sprintf("spendable balance %s is smaller than %s: insufficient funds", initCoins, initCoins.Add(initCoins...)),
	)

	assert.NilError(t,
		keeper.SendCoinsFromModuleToModule(ctx, holderAcc.GetName(), authtypes.Burner, initCoins),
	)
	assert.Equal(t, sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, holderAcc.GetName()).String())
	assert.DeepEqual(t, initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner))

	assert.NilError(t,
		keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Burner, baseAcc.GetAddress(), initCoins),
	)
	assert.Equal(t, sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner).String())
	assert.DeepEqual(t, initCoins, keeper.GetAllBalances(ctx, baseAcc.GetAddress()))

	assert.NilError(t, keeper.SendCoinsFromAccountToModule(ctx, baseAcc.GetAddress(), authtypes.Burner, initCoins))
	assert.Equal(t, sdk.NewCoins().String(), keeper.GetAllBalances(ctx, baseAcc.GetAddress()).String())
	assert.DeepEqual(t, initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner))
}

func TestSupply_MintCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	// add module accounts to supply keeper
	authKeeper, keeper := initKeepersWithmAccPerms(f, make(map[string]bool))

	authKeeper.SetModuleAccount(ctx, burnerAcc)
	authKeeper.SetModuleAccount(ctx, minterAcc)
	authKeeper.SetModuleAccount(ctx, multiPermAcc)
	authKeeper.SetModuleAccount(ctx, randomPermAcc)

	initialSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)

	// no module account
	testutil.AssertPanics(t, func() { keeper.MintCoins(ctx, "", initCoins) }) // nolint:errcheck
	// invalid permission
	testutil.AssertPanics(t, func() { keeper.MintCoins(ctx, authtypes.Burner, initCoins) }) // nolint:errcheck

	err = keeper.MintCoins(ctx, authtypes.Minter, sdk.Coins{sdk.Coin{Denom: "denom", Amount: sdk.NewInt(-10)}})
	assert.Error(t, err, fmt.Sprintf("%sdenom: invalid coins", sdk.NewInt(-10)))

	testutil.AssertPanics(t, func() { keeper.MintCoins(ctx, randomPerm, initCoins) }) // nolint:errcheck

	err = keeper.MintCoins(ctx, authtypes.Minter, initCoins)
	assert.NilError(t, err)

	assert.DeepEqual(t, initCoins, getCoinsByName(ctx, keeper, authKeeper, authtypes.Minter))
	totalSupply, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)

	assert.DeepEqual(t, initialSupply.Add(initCoins...), totalSupply)

	// test same functionality on module account with multiple permissions
	initialSupply, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)

	err = keeper.MintCoins(ctx, multiPermAcc.GetName(), initCoins)
	assert.NilError(t, err)

	totalSupply, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)
	assert.DeepEqual(t, initCoins, getCoinsByName(ctx, keeper, authKeeper, multiPermAcc.GetName()))
	assert.DeepEqual(t, initialSupply.Add(initCoins...), totalSupply)
	testutil.AssertPanics(t, func() { keeper.MintCoins(ctx, authtypes.Burner, initCoins) }) // nolint:errcheck
}

func TestSupply_BurnCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	// add module accounts to supply keeper
	authKeeper, keeper := initKeepersWithmAccPerms(f, make(map[string]bool))

	// set burnerAcc balance
	authKeeper.SetModuleAccount(ctx, burnerAcc)
	assert.NilError(t, keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	assert.NilError(t, keeper.SendCoinsFromModuleToAccount(ctx, authtypes.Minter, burnerAcc.GetAddress(), initCoins))

	// inflate supply
	assert.NilError(t, keeper.MintCoins(ctx, authtypes.Minter, initCoins))
	supplyAfterInflation, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)
	// no module account
	testutil.AssertPanics(t, func() { keeper.BurnCoins(ctx, "", initCoins) }) // nolint:errcheck
	// invalid permission
	testutil.AssertPanics(t, func() { keeper.BurnCoins(ctx, authtypes.Minter, initCoins) }) // nolint:errcheck
	// random permission
	testutil.AssertPanics(t, func() { keeper.BurnCoins(ctx, randomPerm, supplyAfterInflation) }) // nolint:errcheck
	err = keeper.BurnCoins(ctx, authtypes.Burner, supplyAfterInflation)
	assert.Error(t, err, fmt.Sprintf("spendable balance %s is smaller than %s: insufficient funds", initCoins, supplyAfterInflation))

	err = keeper.BurnCoins(ctx, authtypes.Burner, initCoins)
	assert.NilError(t, err)
	supplyAfterBurn, _, err := keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)
	assert.Equal(t, sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, authtypes.Burner).String())
	assert.DeepEqual(t, supplyAfterInflation.Sub(initCoins...), supplyAfterBurn)

	// test same functionality on module account with multiple permissions
	assert.NilError(t, keeper.MintCoins(ctx, authtypes.Minter, initCoins))

	supplyAfterInflation, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)
	assert.NilError(t, keeper.SendCoins(ctx, authtypes.NewModuleAddress(authtypes.Minter), multiPermAcc.GetAddress(), initCoins))
	authKeeper.SetModuleAccount(ctx, multiPermAcc)

	err = keeper.BurnCoins(ctx, multiPermAcc.GetName(), initCoins)
	assert.NilError(t, err)
	supplyAfterBurn, _, err = keeper.GetPaginatedTotalSupply(ctx, &query.PageRequest{})
	assert.NilError(t, err)
	assert.Equal(t, sdk.NewCoins().String(), getCoinsByName(ctx, keeper, authKeeper, multiPermAcc.GetName()).String())
	assert.DeepEqual(t, supplyAfterInflation.Sub(initCoins...), supplyAfterBurn)
}

func TestSendCoinsNewAccount(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := f.accountKeeper.NewAccountWithAddress(ctx, addr1)
	f.accountKeeper.SetAccount(ctx, acc1)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, balances))

	acc1Balances := f.bankKeeper.GetAllBalances(ctx, addr1)
	assert.DeepEqual(t, balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	assert.Assert(t, f.accountKeeper.GetAccount(ctx, addr2) == nil)
	f.bankKeeper.GetAllBalances(ctx, addr2)
	assert.Assert(t, f.bankKeeper.GetAllBalances(ctx, addr2).Empty())

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(50))
	assert.NilError(t, f.bankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc2Balances := f.bankKeeper.GetAllBalances(ctx, addr2)
	acc1Balances = f.bankKeeper.GetAllBalances(ctx, addr1)
	assert.DeepEqual(t, sendAmt, acc2Balances)
	updatedAcc1Bal := balances.Sub(sendAmt...)
	assert.Assert(t, len(acc1Balances) == len(updatedAcc1Bal))
	assert.DeepEqual(t, acc1Balances, updatedAcc1Bal)
	assert.Assert(t, f.accountKeeper.GetAccount(ctx, addr2) != nil)
}

func TestInputOutputNewAccount(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))
	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := f.accountKeeper.NewAccountWithAddress(ctx, addr1)
	f.accountKeeper.SetAccount(ctx, acc1)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, balances))

	acc1Balances := f.bankKeeper.GetAllBalances(ctx, addr1)
	assert.DeepEqual(t, balances, acc1Balances)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	assert.Assert(t, f.accountKeeper.GetAccount(ctx, addr2) == nil)
	assert.Assert(t, f.bankKeeper.GetAllBalances(ctx, addr2).Empty())

	inputs := []types.Input{
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}
	outputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	assert.NilError(t, f.bankKeeper.InputOutputCoins(ctx, inputs, outputs))

	expected := sdk.NewCoins(newFooCoin(30), newBarCoin(10))
	acc2Balances := f.bankKeeper.GetAllBalances(ctx, addr2)
	assert.DeepEqual(t, expected, acc2Balances)
	assert.Assert(t, f.accountKeeper.GetAccount(ctx, addr2) != nil)
}

func TestInputOutputCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	balances := sdk.NewCoins(newFooCoin(90), newBarCoin(30))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	acc1 := f.accountKeeper.NewAccountWithAddress(ctx, addr1)
	f.accountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	acc2 := f.accountKeeper.NewAccountWithAddress(ctx, addr2)
	f.accountKeeper.SetAccount(ctx, acc2)

	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	acc3 := f.accountKeeper.NewAccountWithAddress(ctx, addr3)
	f.accountKeeper.SetAccount(ctx, acc3)

	input := []types.Input{
		{Address: addr1.String(), Coins: sdk.NewCoins(newFooCoin(60), newBarCoin(20))},
	}
	outputs := []types.Output{
		{Address: addr2.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
		{Address: addr3.String(), Coins: sdk.NewCoins(newFooCoin(30), newBarCoin(10))},
	}

	assert.Error(t, f.bankKeeper.InputOutputCoins(ctx, input, []types.Output{}), "sum inputs != sum outputs")
	assert.Error(t, f.bankKeeper.InputOutputCoins(ctx, input, outputs), "spendable balance  is smaller than 20bar: insufficient funds")

	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, balances))

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
	assert.Error(t, f.bankKeeper.InputOutputCoins(ctx, insufficientInput, insufficientOutputs), "sum inputs != sum outputs")
	assert.NilError(t, f.bankKeeper.InputOutputCoins(ctx, input, outputs))

	acc1Balances := f.bankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(30), newBarCoin(10))
	assert.DeepEqual(t, expected, acc1Balances)

	acc2Balances := f.bankKeeper.GetAllBalances(ctx, addr2)
	assert.DeepEqual(t, expected, acc2Balances)

	acc3Balances := f.bankKeeper.GetAllBalances(ctx, addr3)
	assert.DeepEqual(t, expected, acc3Balances)
}

func TestSendCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	balances := sdk.NewCoins(newFooCoin(100), newBarCoin(50))

	addr1 := sdk.AccAddress("addr1_______________")
	acc1 := f.accountKeeper.NewAccountWithAddress(ctx, addr1)
	f.accountKeeper.SetAccount(ctx, acc1)

	addr2 := sdk.AccAddress("addr2_______________")
	acc2 := f.accountKeeper.NewAccountWithAddress(ctx, addr2)
	f.accountKeeper.SetAccount(ctx, acc2)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr2, balances))

	sendAmt := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	assert.Error(t, f.bankKeeper.SendCoins(ctx, addr1, addr2, sendAmt), "spendable balance  is smaller than 25bar: insufficient funds")

	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, balances))
	assert.NilError(t, f.bankKeeper.SendCoins(ctx, addr1, addr2, sendAmt))

	acc1Balances := f.bankKeeper.GetAllBalances(ctx, addr1)
	expected := sdk.NewCoins(newFooCoin(50), newBarCoin(25))
	assert.DeepEqual(t, expected, acc1Balances)

	acc2Balances := f.bankKeeper.GetAllBalances(ctx, addr2)
	expected = sdk.NewCoins(newFooCoin(150), newBarCoin(75))
	assert.DeepEqual(t, expected, acc2Balances)

	// we sent all foo coins to acc2, so foo balance should be deleted for acc1 and bar should be still there
	var coins []sdk.Coin
	f.bankKeeper.IterateAccountBalances(ctx, addr1, func(c sdk.Coin) (stop bool) {
		coins = append(coins, c)
		return true
	})
	assert.Assert(t, len(coins) == 1)
	assert.DeepEqual(t, newBarCoin(25), coins[0])
}

func TestValidateBalance(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	assert.Error(t, f.bankKeeper.ValidateBalance(ctx, addr1), fmt.Sprintf("account %s does not exist: unknown address", addr1))

	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr1)
	f.accountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, balances))
	assert.NilError(t, f.bankKeeper.ValidateBalance(ctx, addr1))

	bacc := authtypes.NewBaseAccountWithAddress(addr2)
	vacc := vesting.NewContinuousVestingAccount(bacc, balances.Add(balances...), now.Unix(), endTime.Unix())

	f.accountKeeper.SetAccount(ctx, vacc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr2, balances))
	assert.Error(t, f.bankKeeper.ValidateBalance(ctx, addr2), "vesting amount 200foo cannot be greater than total amount 100foo")
}

func TestSendCoins_Invalid_SendLockedCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	balances := sdk.NewCoins(newFooCoin(50))
	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	now := cmttime.Now()
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	acc0 := authtypes.NewBaseAccountWithAddress(addr)
	vacc := vesting.NewContinuousVestingAccount(acc0, origCoins, now.Unix(), endTime.Unix())
	f.accountKeeper.SetAccount(ctx, vacc)

	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, f.ctx, addr2, balances))
	assert.Error(t, f.bankKeeper.SendCoins(ctx, addr, addr2, sendCoins), fmt.Sprintf("locked amount exceeds account balance funds: %s > 0stake: insufficient funds", origCoins))
}

func TestSendEnabled(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	enabled := true
	params := types.DefaultParams()
	assert.Equal(t, enabled, params.DefaultSendEnabled)

	assert.NilError(t, f.bankKeeper.SetParams(ctx, params))

	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, math.OneInt())
	fooCoin := sdk.NewCoin("foocoin", math.OneInt())
	barCoin := sdk.NewCoin("barcoin", math.OneInt())

	// assert with default (all denom) send enabled both Bar and Bond Denom are enabled
	assert.Equal(t, enabled, f.bankKeeper.IsSendEnabledCoin(ctx, barCoin))
	assert.Equal(t, enabled, f.bankKeeper.IsSendEnabledCoin(ctx, bondCoin))

	// Both coins should be send enabled.
	err := f.bankKeeper.IsSendEnabledCoins(ctx, fooCoin, bondCoin)
	assert.NilError(t, err)

	// Set default send_enabled to !enabled, add a foodenom that overrides default as enabled
	params.DefaultSendEnabled = !enabled
	assert.NilError(t, f.bankKeeper.SetParams(ctx, params))
	f.bankKeeper.SetSendEnabled(ctx, fooCoin.Denom, enabled)

	// Expect our specific override to be enabled, others to be !enabled.
	assert.Equal(t, enabled, f.bankKeeper.IsSendEnabledCoin(ctx, fooCoin))
	assert.Equal(t, !enabled, f.bankKeeper.IsSendEnabledCoin(ctx, barCoin))
	assert.Equal(t, !enabled, f.bankKeeper.IsSendEnabledCoin(ctx, bondCoin))

	// Foo coin should be send enabled.
	err = f.bankKeeper.IsSendEnabledCoins(ctx, fooCoin)
	assert.NilError(t, err)

	// Expect an error when one coin is not send enabled.
	err = f.bankKeeper.IsSendEnabledCoins(ctx, fooCoin, bondCoin)
	assert.Error(t, err, "stake transfers are currently disabled: send transactions are disabled")

	// Expect an error when all coins are not send enabled.
	err = f.bankKeeper.IsSendEnabledCoins(ctx, bondCoin, barCoin)
	assert.Error(t, err, "stake transfers are currently disabled: send transactions are disabled")
}

func TestHasBalance(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	addr := sdk.AccAddress([]byte("addr1_______________"))

	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr)
	f.accountKeeper.SetAccount(ctx, acc)

	balances := sdk.NewCoins(newFooCoin(100))
	assert.Assert(t, f.bankKeeper.HasBalance(ctx, addr, newFooCoin(99)) == false)

	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr, balances))
	assert.Assert(t, f.bankKeeper.HasBalance(ctx, addr, newFooCoin(101)) == false)
	assert.Assert(t, f.bankKeeper.HasBalance(ctx, addr, newFooCoin(100)))
	assert.Assert(t, f.bankKeeper.HasBalance(ctx, addr, newFooCoin(1)))
}

func TestMsgSendEvents(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr)

	f.accountKeeper.SetAccount(ctx, acc)
	newCoins := sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr, newCoins))

	assert.NilError(t, f.bankKeeper.SendCoins(ctx, addr, addr2, newCoins))
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
	assert.Equal(t, 10, len(events))
	assert.DeepEqual(t, abci.Event(event1), events[8])
	assert.DeepEqual(t, abci.Event(event2), events[9])
}

func TestMsgMultiSendEvents(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	assert.NilError(t, f.bankKeeper.SetParams(ctx, types.DefaultParams()))

	addr := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addr3 := sdk.AccAddress([]byte("addr3_______________"))
	addr4 := sdk.AccAddress([]byte("addr4_______________"))
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr)
	acc2 := f.accountKeeper.NewAccountWithAddress(ctx, addr2)

	f.accountKeeper.SetAccount(ctx, acc)
	f.accountKeeper.SetAccount(ctx, acc2)

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

	assert.Error(t, f.bankKeeper.InputOutputCoins(ctx, input, outputs), fmt.Sprintf("spendable balance  is smaller than %s: insufficient funds", newCoins2))

	events := ctx.EventManager().ABCIEvents()
	assert.Equal(t, 0, len(events))

	// Set addr's coins but not addr2's coins
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50), sdk.NewInt64Coin(barDenom, 100))))
	assert.NilError(t, f.bankKeeper.InputOutputCoins(ctx, input, outputs))

	events = ctx.EventManager().ABCIEvents()
	assert.Equal(t, 12, len(events)) // 12 events because account funding causes extra minting + coin_spent + coin_recv events

	event1 := sdk.Event{
		Type:       sdk.EventTypeMessage,
		Attributes: []abci.EventAttribute{},
	}
	event1.Attributes = append(
		event1.Attributes,
		abci.EventAttribute{Key: types.AttributeKeySender, Value: addr.String()},
	)
	assert.DeepEqual(t, abci.Event(event1), events[7])

	// Set addr's coins and addr2's coins
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))))
	newCoins = sdk.NewCoins(sdk.NewInt64Coin(fooDenom, 50))

	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr, sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))))
	newCoins2 = sdk.NewCoins(sdk.NewInt64Coin(barDenom, 100))

	assert.NilError(t, f.bankKeeper.InputOutputCoins(ctx, input, outputs))

	events = ctx.EventManager().ABCIEvents()
	assert.Equal(t, 30, len(events)) // 27 due to account funding + coin_spent + coin_recv events

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
	assert.DeepEqual(t, abci.Event(event1), events[25])
	assert.DeepEqual(t, abci.Event(event2), events[27])
	assert.DeepEqual(t, abci.Event(event3), events[29])
}

func TestSpendableCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	macc := f.accountKeeper.NewAccountWithAddress(ctx, addrModule)
	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr2)

	f.accountKeeper.SetAccount(ctx, macc)
	f.accountKeeper.SetAccount(ctx, vacc)
	f.accountKeeper.SetAccount(ctx, acc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, origCoins))
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr2, origCoins))

	assert.DeepEqual(t, origCoins, f.bankKeeper.SpendableCoins(ctx, addr2))
	assert.DeepEqual(t, origCoins[0], f.bankKeeper.SpendableCoin(ctx, addr2, "stake"))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	assert.NilError(t, f.bankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	assert.DeepEqual(t, origCoins.Sub(delCoins...), f.bankKeeper.SpendableCoins(ctx, addr1))
	assert.DeepEqual(t, origCoins.Sub(delCoins...)[0], f.bankKeeper.SpendableCoin(ctx, addr1, "stake"))
}

func TestVestingAccountSend(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, now.Unix(), endTime.Unix())

	f.accountKeeper.SetAccount(ctx, vacc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	assert.Error(t, f.bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins), fmt.Sprintf("spendable balance  is smaller than %s: insufficient funds", sendCoins))

	// receive some coins
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, sendCoins))
	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	assert.NilError(t, f.bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	assert.DeepEqual(t, origCoins, f.bankKeeper.GetAllBalances(ctx, addr1))
}

func TestPeriodicVestingAccountSend(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})
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

	f.accountKeeper.SetAccount(ctx, vacc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, origCoins))

	// require that no coins be sendable at the beginning of the vesting schedule
	assert.Error(t, f.bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins), fmt.Sprintf("spendable balance  is smaller than %s: insufficient funds", sendCoins))

	// receive some coins
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, sendCoins))

	// require that all vested coins are spendable plus any received
	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))
	assert.NilError(t, f.bankKeeper.SendCoins(ctx, addr1, addr2, sendCoins))
	assert.DeepEqual(t, origCoins, f.bankKeeper.GetAllBalances(ctx, addr1))
}

func TestVestingAccountReceive(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	sendCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr2)

	f.accountKeeper.SetAccount(ctx, vacc)
	f.accountKeeper.SetAccount(ctx, acc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, origCoins))
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr2, origCoins))

	// send some coins to the vesting account
	assert.NilError(t, f.bankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = f.accountKeeper.GetAccount(ctx, addr1).(*vesting.ContinuousVestingAccount)
	balances := f.bankKeeper.GetAllBalances(ctx, addr1)
	assert.DeepEqual(t, origCoins.Add(sendCoins...), balances)
	assert.DeepEqual(t, balances.Sub(vacc.LockedCoins(now)...), sendCoins)

	// require coins are spendable plus any that have vested
	assert.DeepEqual(t, balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))...), origCoins)
}

func TestPeriodicVestingAccountReceive(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})

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
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr2)

	f.accountKeeper.SetAccount(ctx, vacc)
	f.accountKeeper.SetAccount(ctx, acc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, origCoins))
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr2, origCoins))

	// send some coins to the vesting account
	assert.NilError(t, f.bankKeeper.SendCoins(ctx, addr2, addr1, sendCoins))

	// require the coins are spendable
	vacc = f.accountKeeper.GetAccount(ctx, addr1).(*vesting.PeriodicVestingAccount)
	balances := f.bankKeeper.GetAllBalances(ctx, addr1)
	assert.DeepEqual(t, origCoins.Add(sendCoins...), balances)
	assert.DeepEqual(t, balances.Sub(vacc.LockedCoins(now)...), sendCoins)

	// require coins are spendable plus any that have vested
	assert.DeepEqual(t, balances.Sub(vacc.LockedCoins(now.Add(12*time.Hour))...), origCoins)
}

func TestDelegateCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	macc := f.accountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr2)
	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())

	f.accountKeeper.SetAccount(ctx, vacc)
	f.accountKeeper.SetAccount(ctx, acc)
	f.accountKeeper.SetAccount(ctx, macc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, origCoins))
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	assert.NilError(t, f.bankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins))
	assert.DeepEqual(t, origCoins.Sub(delCoins...), f.bankKeeper.GetAllBalances(ctx, addr2))
	assert.DeepEqual(t, delCoins, f.bankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to delegate
	assert.NilError(t, f.bankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))
	assert.DeepEqual(t, delCoins, f.bankKeeper.GetAllBalances(ctx, addr1))

	// require that delegated vesting amount is equal to what was delegated with DelegateCoins
	acc = f.accountKeeper.GetAccount(ctx, addr1)
	vestingAcc, ok := acc.(types.VestingAccount)
	assert.Assert(t, ok)
	assert.DeepEqual(t, delCoins, vestingAcc.GetDelegatedVesting())
}

func TestDelegateCoins_Invalid(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))
	macc := f.accountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr1)

	assert.Error(t, f.bankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins), fmt.Sprintf("module account %s does not exist: unknown address", addrModule.String()))
	invalidCoins := sdk.Coins{sdk.Coin{Denom: "fooDenom", Amount: sdk.NewInt(-50)}}
	assert.Error(t, f.bankKeeper.DelegateCoins(ctx, addr1, addrModule, invalidCoins), fmt.Sprintf("module account %s does not exist: unknown address", addrModule.String()))

	f.accountKeeper.SetAccount(ctx, macc)
	assert.Error(t, f.bankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins), fmt.Sprintf("failed to delegate; 0foo is smaller than %s: insufficient funds", delCoins))
	f.accountKeeper.SetAccount(ctx, acc)
	assert.Error(t, f.bankKeeper.DelegateCoins(ctx, addr1, addrModule, origCoins.Add(origCoins...)), fmt.Sprintf("failed to delegate; 0foo is smaller than %s: insufficient funds", origCoins.Add(origCoins...)))
}

func TestUndelegateCoins(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx
	now := cmttime.Now()
	ctx = ctx.WithBlockHeader(cmtproto.Header{Time: now})
	endTime := now.Add(24 * time.Hour)

	origCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 100))
	delCoins := sdk.NewCoins(sdk.NewInt64Coin("stake", 50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addr2 := sdk.AccAddress([]byte("addr2_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))

	bacc := authtypes.NewBaseAccountWithAddress(addr1)
	macc := f.accountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing

	vacc := vesting.NewContinuousVestingAccount(bacc, origCoins, ctx.BlockHeader().Time.Unix(), endTime.Unix())
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr2)

	f.accountKeeper.SetAccount(ctx, vacc)
	f.accountKeeper.SetAccount(ctx, acc)
	f.accountKeeper.SetAccount(ctx, macc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, origCoins))
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr2, origCoins))

	ctx = ctx.WithBlockTime(now.Add(12 * time.Hour))

	// require the ability for a non-vesting account to delegate
	err := f.bankKeeper.DelegateCoins(ctx, addr2, addrModule, delCoins)
	assert.NilError(t, err)

	assert.DeepEqual(t, origCoins.Sub(delCoins...), f.bankKeeper.GetAllBalances(ctx, addr2))
	assert.DeepEqual(t, delCoins, f.bankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a non-vesting account to undelegate
	assert.NilError(t, f.bankKeeper.UndelegateCoins(ctx, addrModule, addr2, delCoins))

	assert.DeepEqual(t, origCoins, f.bankKeeper.GetAllBalances(ctx, addr2))
	assert.Assert(t, f.bankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require the ability for a vesting account to delegate
	assert.NilError(t, f.bankKeeper.DelegateCoins(ctx, addr1, addrModule, delCoins))

	assert.DeepEqual(t, origCoins.Sub(delCoins...), f.bankKeeper.GetAllBalances(ctx, addr1))
	assert.DeepEqual(t, delCoins, f.bankKeeper.GetAllBalances(ctx, addrModule))

	// require the ability for a vesting account to undelegate
	assert.NilError(t, f.bankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins))

	assert.DeepEqual(t, origCoins, f.bankKeeper.GetAllBalances(ctx, addr1))
	assert.Assert(t, f.bankKeeper.GetAllBalances(ctx, addrModule).Empty())

	// require that delegated vesting amount is completely empty, since they were completely undelegated
	acc = f.accountKeeper.GetAccount(ctx, addr1)
	vestingAcc, ok := acc.(types.VestingAccount)
	assert.Assert(t, ok)
	assert.Assert(t, vestingAcc.GetDelegatedVesting().Empty())
}

func TestUndelegateCoins_Invalid(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	origCoins := sdk.NewCoins(newFooCoin(100))
	delCoins := sdk.NewCoins(newFooCoin(50))

	addr1 := sdk.AccAddress([]byte("addr1_______________"))
	addrModule := sdk.AccAddress([]byte("moduleAcc___________"))
	macc := f.accountKeeper.NewAccountWithAddress(ctx, addrModule) // we don't need to define an actual module account bc we just need the address for testing
	acc := f.accountKeeper.NewAccountWithAddress(ctx, addr1)

	assert.Error(t, f.bankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins), fmt.Sprintf("module account %s does not exist: unknown address", addrModule.String()))

	f.accountKeeper.SetAccount(ctx, macc)
	assert.NilError(t, banktestutil.FundAccount(f.bankKeeper, ctx, addr1, origCoins))

	assert.Error(t, f.bankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins), fmt.Sprintf("spendable balance  is smaller than %s: insufficient funds", delCoins))
	f.accountKeeper.SetAccount(ctx, acc)

	assert.Error(t, f.bankKeeper.UndelegateCoins(ctx, addrModule, addr1, delCoins), fmt.Sprintf("spendable balance  is smaller than %s: insufficient funds", delCoins))
}

func TestSetDenomMetaData(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	metadata := getTestMetadata()

	for i := range []int{1, 2} {
		f.bankKeeper.SetDenomMetaData(ctx, metadata[i])
	}

	actualMetadata, found := f.bankKeeper.GetDenomMetaData(ctx, metadata[1].Base)
	assert.Assert(t, found)
	found = f.bankKeeper.HasDenomMetaData(ctx, metadata[1].Base)
	assert.Assert(t, found)
	assert.Equal(t, metadata[1].GetBase(), actualMetadata.GetBase())
	assert.Equal(t, metadata[1].GetDisplay(), actualMetadata.GetDisplay())
	assert.Equal(t, metadata[1].GetDescription(), actualMetadata.GetDescription())
	assert.Equal(t, metadata[1].GetDenomUnits()[1].GetDenom(), actualMetadata.GetDenomUnits()[1].GetDenom())
	assert.Equal(t, metadata[1].GetDenomUnits()[1].GetExponent(), actualMetadata.GetDenomUnits()[1].GetExponent())
	assert.DeepEqual(t, metadata[1].GetDenomUnits()[1].GetAliases(), actualMetadata.GetDenomUnits()[1].GetAliases())
}

func TestIterateAllDenomMetaData(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx := f.ctx

	expectedMetadata := getTestMetadata()
	// set metadata
	for i := range []int{1, 2} {
		f.bankKeeper.SetDenomMetaData(ctx, expectedMetadata[i])
	}
	// retrieve metadata
	actualMetadata := make([]types.Metadata, 0)
	f.bankKeeper.IterateAllDenomMetaData(ctx, func(metadata types.Metadata) bool {
		actualMetadata = append(actualMetadata, metadata)
		return false
	})
	// execute checks
	for i := range []int{1, 2} {
		assert.Equal(t, expectedMetadata[i].GetBase(), actualMetadata[i].GetBase())
		assert.Equal(t, expectedMetadata[i].GetDisplay(), actualMetadata[i].GetDisplay())
		assert.Equal(t, expectedMetadata[i].GetDescription(), actualMetadata[i].GetDescription())
		assert.Equal(t, expectedMetadata[i].GetDenomUnits()[1].GetDenom(), actualMetadata[i].GetDenomUnits()[1].GetDenom())
		assert.Equal(t, expectedMetadata[i].GetDenomUnits()[1].GetExponent(), actualMetadata[i].GetDenomUnits()[1].GetExponent())
		assert.DeepEqual(t, expectedMetadata[i].GetDenomUnits()[1].GetAliases(), actualMetadata[i].GetDenomUnits()[1].GetAliases())
	}
}

func TestBalanceTrackingEvents(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	// replace account keeper and bank keeper otherwise the account keeper won't be aware of the
	// existence of the new module account because GetModuleAccount checks for the existence via
	// permissions map and not via state... weird
	maccPerms := make(map[string][]string)

	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}

	f.accountKeeper = authkeeper.NewAccountKeeper(
		f.appCodec, f.fetchStoreKey(authtypes.StoreKey),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	f.bankKeeper = keeper.NewBaseKeeper(f.appCodec, f.fetchStoreKey(types.StoreKey),
		f.accountKeeper, nil, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// set account with multiple permissions
	f.accountKeeper.SetModuleAccount(f.ctx, multiPermAcc)
	// mint coins
	assert.NilError(t,
		f.bankKeeper.MintCoins(
			f.ctx,
			multiPermAcc.Name,
			sdk.NewCoins(sdk.NewCoin("utxo", sdk.NewInt(100000)))),
	)
	// send coins to address
	addr1 := sdk.AccAddress("addr1_______________")
	assert.NilError(t,
		f.bankKeeper.SendCoinsFromModuleToAccount(
			f.ctx,
			multiPermAcc.Name,
			addr1,
			sdk.NewCoins(sdk.NewCoin("utxo", sdk.NewInt(50000))),
		),
	)

	// burn coins from module account
	assert.NilError(t,
		f.bankKeeper.BurnCoins(
			f.ctx,
			multiPermAcc.Name,
			sdk.NewCoins(sdk.NewInt64Coin("utxo", 1000)),
		),
	)

	// process balances and supply from events
	supply := sdk.NewCoins()

	balances := make(map[string]sdk.Coins)

	for _, e := range f.ctx.EventManager().ABCIEvents() {
		switch e.Type {
		case types.EventTypeCoinBurn:
			burnedCoins, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			assert.NilError(t, err)
			supply = supply.Sub(burnedCoins...)

		case types.EventTypeCoinMint:
			mintedCoins, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			assert.NilError(t, err)
			supply = supply.Add(mintedCoins...)

		case types.EventTypeCoinSpent:
			coinsSpent, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			assert.NilError(t, err)
			spender, err := sdk.AccAddressFromBech32((string)(e.Attributes[0].Value))
			assert.NilError(t, err)
			balances[spender.String()] = balances[spender.String()].Sub(coinsSpent...)

		case types.EventTypeCoinReceived:
			coinsRecv, err := sdk.ParseCoinsNormalized((string)(e.Attributes[1].Value))
			assert.NilError(t, err)
			receiver, err := sdk.AccAddressFromBech32((string)(e.Attributes[0].Value))
			assert.NilError(t, err)
			balances[receiver.String()] = balances[receiver.String()].Add(coinsRecv...)
		}
	}

	// check balance and supply tracking
	assert.Assert(t, f.bankKeeper.HasSupply(f.ctx, "utxo"))
	savedSupply := f.bankKeeper.GetSupply(f.ctx, "utxo")
	utxoSupply := savedSupply
	assert.DeepEqual(t, utxoSupply.Amount, supply.AmountOf("utxo"))
	// iterate accounts and check balances
	f.bankKeeper.IterateAllBalances(f.ctx, func(address sdk.AccAddress, coin sdk.Coin) (stop bool) {
		// if it's not utxo coin then skip
		if coin.Denom != "utxo" {
			return false
		}

		balance, exists := balances[address.String()]
		assert.Assert(t, exists)

		expectedUtxo := sdk.NewCoin("utxo", balance.AmountOf(coin.Denom))
		assert.Equal(t, expectedUtxo.String(), coin.String())
		return false
	})
}

func getTestMetadata() []types.Metadata {
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

func TestMintCoinRestrictions(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	type BankMintingRestrictionFn func(ctx sdk.Context, coins sdk.Coins) error

	maccPerms := make(map[string][]string)
	maccPerms[multiPerm] = []string{authtypes.Burner, authtypes.Minter, authtypes.Staking}

	f.accountKeeper = authkeeper.NewAccountKeeper(
		f.appCodec, f.fetchStoreKey(authtypes.StoreKey),
		authtypes.ProtoBaseAccount, maccPerms, sdk.Bech32MainPrefix,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	f.accountKeeper.SetModuleAccount(f.ctx, multiPermAcc)

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
		f.bankKeeper = keeper.NewBaseKeeper(f.appCodec, f.fetchStoreKey(types.StoreKey),
			f.accountKeeper, nil, authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		).WithMintCoinsRestriction(keeper.MintingRestrictionFn(test.restrictionFn))
		for _, testCase := range test.testCases {
			if testCase.expectPass {
				assert.NilError(t,
					f.bankKeeper.MintCoins(
						f.ctx,
						multiPermAcc.Name,
						sdk.NewCoins(testCase.coinsToTry),
					),
				)
			} else {
				assert.Error(t,
					f.bankKeeper.MintCoins(
						f.ctx,
						multiPermAcc.Name,
						sdk.NewCoins(testCase.coinsToTry),
					),
					"Module bank only has perms for minting foo coins, tried minting bar coins",
				)
			}
		}
	}
}

func TestIsSendEnabledDenom(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper

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
		assert.NilError(t, bankKeeper.SetParams(ctx, params))
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s default %t", tc.denom, def), func(t *testing.T) {
				actual := f.bankKeeper.IsSendEnabledDenom(f.ctx, tc.denom)
				exp := tc.exp
				if tc.expDef {
					exp = def
				}
				assert.Equal(t, exp, actual)
			})
		}
	}
}

func TestGetSendEnabledEntry(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper

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
		t.Run(tc.denom, func(t *testing.T) {
			actualSE, actualF := bankKeeper.GetSendEnabledEntry(ctx, tc.denom)
			assert.Equal(t, tc.expF, actualF, "found")
			assert.Equal(t, tc.expSE, actualSE, "SendEnabled")
		})
	}
}

func TestSetSendEnabled(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper

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
		assert.NilError(t, bankKeeper.SetParams(ctx, params))
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s default %t", tc.name, def), func(t *testing.T) {
				bankKeeper.SetSendEnabled(ctx, tc.denom, tc.value)
				actual := bankKeeper.IsSendEnabledDenom(ctx, tc.denom)
				assert.Equal(t, tc.value, actual)
			})
		}
	}
}

func TestSetAllSendEnabled(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper

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
		assert.NilError(t, bankKeeper.SetParams(ctx, params))
		for _, tc := range tests {
			t.Run(fmt.Sprintf("%s default %t", tc.name, def), func(t *testing.T) {
				bankKeeper.SetAllSendEnabled(ctx, tc.sendEnableds)
				for _, se := range tc.sendEnableds {
					actual := bankKeeper.IsSendEnabledDenom(ctx, se.Denom)
					assert.Equal(t, se.Enabled, actual, se.Denom)
				}
			})
		}
	}
}

func TestDeleteSendEnabled(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		assert.NilError(t, bankKeeper.SetParams(ctx, params))
		t.Run(fmt.Sprintf("default %t", def), func(t *testing.T) {
			denom := fmt.Sprintf("somerand%tcoin", !def)
			bankKeeper.SetSendEnabled(ctx, denom, !def)
			assert.Equal(t, !def, bankKeeper.IsSendEnabledDenom(ctx, denom))
			bankKeeper.DeleteSendEnabled(ctx, denom)
			assert.Equal(t, def, bankKeeper.IsSendEnabledDenom(ctx, denom))
		})
	}
}

func TestIterateSendEnabledEntries(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper

	t.Run("no entries to iterate", func(t *testing.T) {
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
		assert.NilError(t, bankKeeper.SetParams(ctx, params))
		var seen []string
		t.Run(fmt.Sprintf("all denoms have expected values default %t", def), func(t *testing.T) {
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
		t.Run(fmt.Sprintf("all denoms were seen default %t", def), func(t *testing.T) {
			match := assertElementsMatch(denoms, seen)
			assert.Assert(t, match)
		})
	}

	for _, denom := range denoms {
		bankKeeper.DeleteSendEnabled(ctx, denom)
	}

	t.Run("no entries to iterate again after deleting all of them", func(t *testing.T) {
		count := 0
		bankKeeper.IterateSendEnabledEntries(ctx, func(_ string, _ bool) (stop bool) {
			count++
			return false
		})
		assert.Equal(t, 0, count)
	})
}

func TestGetAllSendEnabledEntries(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper

	t.Run("no entries", func(t *testing.T) {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		assert.Assert(t, len(actual) == 0)
	})

	alpha := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	denoms := make([]string, len(alpha)*2)
	for i, l := range alpha {
		denoms[i*2] = fmt.Sprintf("%sitercoinfalse", l)
		denoms[i*2+1] = fmt.Sprintf("%sitercointrue", l)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2], false)
		bankKeeper.SetSendEnabled(ctx, denoms[i*2+1], true)
	}

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		assert.NilError(t, bankKeeper.SetParams(ctx, params))
		var seen []string
		t.Run(fmt.Sprintf("all denoms have expected values default %t", def), func(t *testing.T) {
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
		t.Run(fmt.Sprintf("all denoms were seen default %t", def), func(t *testing.T) {
			assert.DeepEqual(t, denoms, seen)
		})
	}

	for _, denom := range denoms {
		bankKeeper.DeleteSendEnabled(ctx, denom)
	}

	t.Run("no entries again after deleting all of them", func(t *testing.T) {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		assert.Assert(t, len(actual) == 0)
	})
}

type mockSubspace struct {
	ps types.Params
}

func (ms mockSubspace) GetParamSet(ctx sdk.Context, ps exported.ParamSet) {
	*ps.(*types.Params) = ms.ps
}

func TestMigrator_Migrate3to4(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper

	for _, def := range []bool{true, false} {
		params := types.Params{DefaultSendEnabled: def}
		assert.NilError(t, bankKeeper.SetParams(ctx, params))
		t.Run(fmt.Sprintf("default %t does not change", def), func(t *testing.T) {
			legacySubspace := func(ps types.Params) mockSubspace {
				return mockSubspace{ps: ps}
			}(types.NewParams(def))
			migrator := keeper.NewMigrator(bankKeeper, legacySubspace)
			assert.NilError(t, migrator.Migrate3to4(ctx))
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
		assert.NilError(t, bankKeeper.SetParams(ctx, params))
		t.Run(fmt.Sprintf("default %t send enabled info moved to store", def), func(t *testing.T) {
			legacySubspace := func(ps types.Params) mockSubspace {
				return mockSubspace{ps: ps}
			}(types.NewParams(def))
			migrator := keeper.NewMigrator(bankKeeper, legacySubspace)
			assert.NilError(t, migrator.Migrate3to4(ctx))
			newParams := bankKeeper.GetParams(ctx)
			assert.Assert(t, len(newParams.SendEnabled) == 0)
			for _, se := range params.SendEnabled {
				actual := bankKeeper.IsSendEnabledDenom(ctx, se.Denom)
				assert.Equal(t, se.Enabled, actual, se.Denom)
			}
		})
	}
}

func TestSetParams(t *testing.T) {
	f := initFixture(t)
	t.Parallel()

	ctx, bankKeeper := f.ctx, f.bankKeeper
	params := types.NewParams(true)
	params.SendEnabled = []*types.SendEnabled{
		{Denom: "paramscointrue", Enabled: true},
		{Denom: "paramscoinfalse", Enabled: false},
	}
	assert.NilError(t, bankKeeper.SetParams(ctx, params))

	t.Run("stored params are as expected", func(t *testing.T) {
		actual := bankKeeper.GetParams(ctx)
		assert.Assert(t, actual.DefaultSendEnabled, "DefaultSendEnabled")
		assert.Assert(t, len(actual.SendEnabled) == 0, "SendEnabled")
	})

	t.Run("send enabled params converted to store", func(t *testing.T) {
		actual := bankKeeper.GetAllSendEnabledEntries(ctx)
		if assert.Check(t, len(actual) == 2) {
			assert.Equal(t, "paramscoinfalse", actual[0].Denom, "actual[0].Denom")
			assert.Assert(t, actual[0].Enabled == false, "actual[0].Enabled")
			assert.Equal(t, "paramscointrue", actual[1].Denom, "actual[1].Denom")
			assert.Assert(t, actual[1].Enabled, "actual[1].Enabled")
		}
	})
}
