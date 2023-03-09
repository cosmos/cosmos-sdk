package keeper_test

import (
	"fmt"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil/configurator"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ types.BankHooks = &MockBankHooksReceiver{}

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
)

type testingSuite struct {
	BankKeeper    bankkeeper.Keeper
	AccountKeeper types.AccountKeeper
	StakingKeeper stakingkeeper.Keeper
	App           *runtime.App
}

func createTestSuite(t *testing.T, genesisAccounts []authtypes.GenesisAccount) testingSuite {
	res := testingSuite{}

	var genAccounts []simtestutil.GenesisAccount
	for _, acc := range genesisAccounts {
		genAccounts = append(genAccounts, simtestutil.GenesisAccount{GenesisAccount: acc})
	}

	startupCfg := simtestutil.DefaultStartUpConfig()
	startupCfg.GenesisAccounts = genAccounts

	app, err := simtestutil.SetupWithConfiguration(configurator.NewAppConfig(
		configurator.ParamsModule(),
		configurator.AuthModule(),
		configurator.StakingModule(),
		configurator.TxModule(),
		configurator.ConsensusModule(),
		configurator.BankModule(),
		configurator.GovModule(),
	),
		startupCfg, &res.BankKeeper, &res.AccountKeeper, &res.StakingKeeper)

	res.App = app

	require.NoError(t, err)
	return res
}

// BankHooks event hooks for bank (noalias)
type MockBankHooksReceiver struct{}

// Mock BlockBeforeSend bank hook that doesn't allow the sending of exactly 100 coins of any denom.
func (h *MockBankHooksReceiver) BlockBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	for _, coin := range amount {
		if coin.Amount.Equal(sdk.NewInt(100)) {
			return fmt.Errorf("not allowed; expected %v, got: %v", 100, coin.Amount)
		}
	}
	return nil
}

// variable for counting `TrackBeforeSend`
var (
	countTrackBeforeSend = 0
	expNextCount         = 1
)

// Mock TrackBeforeSend bank hook that simply tracks the sending of exactly 50 coins of any denom.
func (h *MockBankHooksReceiver) TrackBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) {
	for _, coin := range amount {
		if coin.Amount.Equal(sdk.NewInt(50)) {
			countTrackBeforeSend += 1
		}
	}
}

func TestHooks(t *testing.T) {
	acc := &authtypes.BaseAccount{
		Address: addr1.String(),
	}

	genAccs := []authtypes.GenesisAccount{acc}
	app := createTestSuite(t, genAccs)
	baseApp := app.App.BaseApp
	ctx := baseApp.NewContext(false, tmproto.Header{})

	addrs := simtestutil.AddTestAddrs(app.BankKeeper, app.StakingKeeper, ctx, 2, sdk.NewInt(1000))
	banktestutil.FundModuleAccount(app.BankKeeper, ctx, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1000))))

	// create a valid send amount which is 1 coin, and an invalidSendAmount which is 100 coins
	validSendAmount := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1)))
	triggerTrackSendAmount := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(50)))
	invalidBlockSendAmount := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(100)))

	// setup our mock bank hooks receiver that prevents the send of 100 coins
	bankHooksReceiver := MockBankHooksReceiver{}
	baseBankKeeper, ok := app.BankKeeper.(keeper.BaseKeeper)
	require.True(t, ok)
	keeper.UnsafeSetHooks(
		&baseBankKeeper, types.NewMultiBankHooks(&bankHooksReceiver),
	)
	app.BankKeeper = baseBankKeeper

	// try sending a validSendAmount and it should work
	err := app.BankKeeper.SendCoins(ctx, addrs[0], addrs[1], validSendAmount)
	require.NoError(t, err)

	// try sending an trigger track send amount and it should work
	err = app.BankKeeper.SendCoins(ctx, addrs[0], addrs[1], triggerTrackSendAmount)
	require.NoError(t, err)

	require.Equal(t, countTrackBeforeSend, expNextCount)
	expNextCount++

	// try sending an invalidSendAmount and it should not work
	err = app.BankKeeper.SendCoins(ctx, addrs[0], addrs[1], invalidBlockSendAmount)
	require.Error(t, err)

	// try doing SendManyCoins and make sure if even a single subsend is invalid, the entire function fails
	err = app.BankKeeper.SendManyCoins(ctx, addrs[0], []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{invalidBlockSendAmount, validSendAmount})
	require.Error(t, err)

	err = app.BankKeeper.SendManyCoins(ctx, addrs[0], []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{triggerTrackSendAmount, validSendAmount})
	require.Equal(t, countTrackBeforeSend, expNextCount)
	expNextCount++

	// make sure that account to module doesn't bypass hook
	err = app.BankKeeper.SendCoinsFromAccountToModule(ctx, addrs[0], stakingtypes.BondedPoolName, validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromAccountToModule(ctx, addrs[0], stakingtypes.BondedPoolName, invalidBlockSendAmount)
	require.Error(t, err)
	err = app.BankKeeper.SendCoinsFromAccountToModule(ctx, addrs[0], stakingtypes.BondedPoolName, triggerTrackSendAmount)
	require.Equal(t, countTrackBeforeSend, expNextCount)
	expNextCount++

	// make sure that module to account doesn't bypass hook
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, stakingtypes.BondedPoolName, addrs[0], validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, stakingtypes.BondedPoolName, addrs[0], invalidBlockSendAmount)
	require.Error(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, stakingtypes.BondedPoolName, addrs[0], triggerTrackSendAmount)
	require.Equal(t, countTrackBeforeSend, expNextCount)
	expNextCount++

	// make sure that module to module doesn't bypass hook
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, invalidBlockSendAmount)
	// there should be no error since module to module does not call block before send hooks
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, triggerTrackSendAmount)
	require.Equal(t, countTrackBeforeSend, expNextCount)
	expNextCount++

	// make sure that module to many accounts doesn't bypass hook
	err = app.BankKeeper.SendCoinsFromModuleToManyAccounts(ctx, stakingtypes.BondedPoolName, []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{validSendAmount, validSendAmount})
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToManyAccounts(ctx, stakingtypes.BondedPoolName, []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{validSendAmount, invalidBlockSendAmount})
	require.Error(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToManyAccounts(ctx, stakingtypes.BondedPoolName, []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{validSendAmount, triggerTrackSendAmount})
	require.Equal(t, countTrackBeforeSend, expNextCount)
	expNextCount++

	// make sure that DelegateCoins doesn't bypass the hook
	err = app.BankKeeper.DelegateCoins(ctx, addrs[0], app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.DelegateCoins(ctx, addrs[0], app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), invalidBlockSendAmount)
	require.Error(t, err)
	err = app.BankKeeper.DelegateCoins(ctx, addrs[0], app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), triggerTrackSendAmount)
	require.Equal(t, countTrackBeforeSend, expNextCount)
	expNextCount++

	// make sure that UndelegateCoins doesn't bypass the hook
	err = app.BankKeeper.UndelegateCoins(ctx, app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), addrs[0], validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.UndelegateCoins(ctx, app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), addrs[0], invalidBlockSendAmount)
	require.Error(t, err)

	err = app.BankKeeper.UndelegateCoins(ctx, app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), addrs[0], triggerTrackSendAmount)
	require.Equal(t, countTrackBeforeSend, expNextCount)
	expNextCount++
}
