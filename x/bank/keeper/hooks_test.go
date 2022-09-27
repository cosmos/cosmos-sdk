package keeper_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

var _ types.BankHooks = &MockBankHooksReceiver{}

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

// Mock TrackBeforeSend bank hook that doesn't allow the sending of exactly 50 coins of any denom.
func (h *MockBankHooksReceiver) TrackBeforeSend(ctx sdk.Context, from, to sdk.AccAddress, amount sdk.Coins) error {
	for _, coin := range amount {
		if coin.Amount.Equal(sdk.NewInt(50)) {
			return fmt.Errorf("not allowed; expected %v, got: %v", 100, coin.Amount)
		}
	}

	return nil
}

func TestHooks(t *testing.T) {
	app := simapp.Setup(false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrs := simapp.AddTestAddrs(app, ctx, 2, sdk.NewInt(1000))
	simapp.FundModuleAccount(app.BankKeeper, ctx, stakingtypes.BondedPoolName, sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1000))))

	// create a valid send amount which is 1 coin, and an invalidSendAmount which is 100 coins
	validSendAmount := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(1)))
	invalidTrackSendAmount := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), sdk.NewInt(50)))
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

	// try sending an invalidSendAmount and it should not work
	err = app.BankKeeper.SendCoins(ctx, addrs[0], addrs[1], invalidTrackSendAmount)
	require.Error(t, err)

	// try sending an invalidSendAmount and it should not work
	err = app.BankKeeper.SendCoins(ctx, addrs[0], addrs[1], invalidBlockSendAmount)
	require.Error(t, err)

	// try doing SendManyCoins and make sure if even a single subsend is invalid, the entire function fails
	err = app.BankKeeper.SendManyCoins(ctx, addrs[0], []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{invalidBlockSendAmount, validSendAmount})
	require.Error(t, err)

	err = app.BankKeeper.SendManyCoins(ctx, addrs[0], []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{invalidTrackSendAmount, validSendAmount})
	require.Error(t, err)

	// make sure that account to module doesn't bypass hook
	err = app.BankKeeper.SendCoinsFromAccountToModule(ctx, addrs[0], stakingtypes.BondedPoolName, validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromAccountToModule(ctx, addrs[0], stakingtypes.BondedPoolName, invalidBlockSendAmount)
	require.Error(t, err)
	err = app.BankKeeper.SendCoinsFromAccountToModule(ctx, addrs[0], stakingtypes.BondedPoolName, invalidTrackSendAmount)
	require.Error(t, err)

	// make sure that module to account doesn't bypass hook
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, stakingtypes.BondedPoolName, addrs[0], validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, stakingtypes.BondedPoolName, addrs[0], invalidBlockSendAmount)
	require.Error(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToAccount(ctx, stakingtypes.BondedPoolName, addrs[0], invalidTrackSendAmount)
	require.Error(t, err)

	// make sure that module to module doesn't bypass hook
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, invalidBlockSendAmount)
	// there should be no error since module to module does not call block before send hooks
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToModule(ctx, stakingtypes.BondedPoolName, stakingtypes.NotBondedPoolName, invalidTrackSendAmount)
	require.Error(t, err)

	// make sure that module to many accounts doesn't bypass hook
	err = app.BankKeeper.SendCoinsFromModuleToManyAccounts(ctx, stakingtypes.BondedPoolName, []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{validSendAmount, validSendAmount})
	require.NoError(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToManyAccounts(ctx, stakingtypes.BondedPoolName, []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{validSendAmount, invalidBlockSendAmount})
	require.Error(t, err)
	err = app.BankKeeper.SendCoinsFromModuleToManyAccounts(ctx, stakingtypes.BondedPoolName, []sdk.AccAddress{addrs[0], addrs[1]}, []sdk.Coins{validSendAmount, invalidTrackSendAmount})
	require.Error(t, err)

	// make sure that DelegateCoins doesn't bypass the hook
	err = app.BankKeeper.DelegateCoins(ctx, addrs[0], app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.DelegateCoins(ctx, addrs[0], app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), invalidBlockSendAmount)
	require.Error(t, err)
	err = app.BankKeeper.DelegateCoins(ctx, addrs[0], app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), invalidTrackSendAmount)
	require.Error(t, err)

	// make sure that UndelegateCoins doesn't bypass the hook
	err = app.BankKeeper.UndelegateCoins(ctx, app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), addrs[0], validSendAmount)
	require.NoError(t, err)
	err = app.BankKeeper.UndelegateCoins(ctx, app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), addrs[0], invalidBlockSendAmount)
	require.Error(t, err)
	err = app.BankKeeper.UndelegateCoins(ctx, app.AccountKeeper.GetModuleAddress(stakingtypes.BondedPoolName), addrs[0], invalidTrackSendAmount)
	require.Error(t, err)
}
