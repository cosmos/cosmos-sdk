package staking_test

import (
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankKeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	stakingKeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

var (
	priv1 = secp256k1.GenPrivKey()
	addr1 = sdk.AccAddress(priv1.PubKey().Address())
	priv2 = secp256k1.GenPrivKey()
	addr2 = sdk.AccAddress(priv2.PubKey().Address())
	priv3 = secp256k1.GenPrivKey()
	addr3 = sdk.AccAddress(priv3.PubKey().Address())

	valKey          = ed25519.GenPrivKey()
	valKey2         = ed25519.GenPrivKey()
	commissionRates = types.NewCommissionRates(math.LegacyZeroDec(), math.LegacyZeroDec(), math.LegacyZeroDec())
)

func TestStakingMsgs(t *testing.T) {
	genTokens := sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	acc1 := &authtypes.BaseAccount{Address: addr1.String()}
	acc2 := &authtypes.BaseAccount{Address: addr2.String()}
	accs := []simtestutil.GenesisAccount{
		{GenesisAccount: acc1, Coins: sdk.Coins{genCoin}},
		{GenesisAccount: acc2, Coins: sdk.Coins{genCoin}},
	}

	var (
		bankKeeper    bankKeeper.Keeper
		stakingKeeper *stakingKeeper.Keeper
	)

	startupCfg := simtestutil.DefaultStartUpConfig()
	startupCfg.GenesisAccounts = accs

	app, err := simtestutil.SetupWithConfiguration(
		simtestutil.AppConfigWithNopLogger(testutil.AppConfig),
		startupCfg, &bankKeeper, &stakingKeeper)
	require.NoError(t, err)
	ctxCheck := app.NewContext(true)

	require.True(t, sdk.Coins{genCoin}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr1)))
	require.True(t, sdk.Coins{genCoin}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr2)))

	// create validator
	description := types.NewDescription("foo_moniker", "", "", "", "")
	createValidatorMsg, err := types.NewMsgCreateValidator(
		sdk.ValAddress(addr1).String(), valKey.PubKey(), bondCoin, description, commissionRates, math.OneInt(),
	)
	require.NoError(t, err)

	header := cmtproto.Header{Height: app.LastBlockHeight() + 1}
	txConfig := moduletestutil.MakeTestTxConfig()
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, app.BaseApp, header, []sdk.Msg{createValidatorMsg}, "", []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)
	require.True(t, sdk.Coins{genCoin.Sub(bondCoin)}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr1)))

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	require.NoError(t, err)
	ctxCheck = app.NewContext(true)
	validator, err := stakingKeeper.GetValidator(ctxCheck, sdk.ValAddress(addr1))
	require.NoError(t, err)

	require.Equal(t, sdk.ValAddress(addr1).String(), validator.OperatorAddress)
	require.Equal(t, types.Bonded, validator.Status)
	require.True(math.IntEq(t, bondTokens, validator.BondedTokens()))

	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: app.LastBlockHeight() + 1})
	require.NoError(t, err)

	// edit the validator
	description = types.NewDescription("bar_moniker", "", "", "", "")
	editValidatorMsg := types.NewMsgEditValidator(sdk.ValAddress(addr1).String(), description, nil, nil)

	header = cmtproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, app.BaseApp, header, []sdk.Msg{editValidatorMsg}, "", []uint64{0}, []uint64{1}, true, true, priv1)
	require.NoError(t, err)

	ctxCheck = app.NewContext(true)
	validator, err = stakingKeeper.GetValidator(ctxCheck, sdk.ValAddress(addr1))
	require.NoError(t, err)
	require.Equal(t, description, validator.Description)

	// delegate
	require.True(t, sdk.Coins{genCoin}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr2)))
	delegateMsg := types.NewMsgDelegate(addr2.String(), sdk.ValAddress(addr1).String(), bondCoin)

	header = cmtproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, app.BaseApp, header, []sdk.Msg{delegateMsg}, "", []uint64{1}, []uint64{0}, true, true, priv2)
	require.NoError(t, err)

	ctxCheck = app.NewContext(true)
	require.True(t, sdk.Coins{genCoin.Sub(bondCoin)}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr2)))
	_, err = stakingKeeper.GetDelegation(ctxCheck, addr2, sdk.ValAddress(addr1))
	require.NoError(t, err)

	// begin unbonding
	beginUnbondingMsg := types.NewMsgUndelegate(addr2.String(), sdk.ValAddress(addr1).String(), bondCoin)
	header = cmtproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(t, txConfig, app.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, "", []uint64{1}, []uint64{1}, true, true, priv2)
	require.NoError(t, err)

	// delegation should not exist anymore
	ctxCheck = app.NewContext(true)
	_, err = stakingKeeper.GetDelegation(ctxCheck, addr2, sdk.ValAddress(addr1))
	require.ErrorIs(t, err, types.ErrNoDelegation)

	// balance should be the same because bonding not yet complete
	require.True(t, sdk.Coins{genCoin.Sub(bondCoin)}.Equal(bankKeeper.GetAllBalances(ctxCheck, addr2)))
}

func TestBeginRedelegateAllSharesFromUnbondedSource(t *testing.T) {
	genTokens := sdk.TokensFromConsensusPower(100, sdk.DefaultPowerReduction)
	valTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	bobTokens := sdk.TokensFromConsensusPower(5, sdk.DefaultPowerReduction)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)

	acc1 := &authtypes.BaseAccount{Address: addr1.String()}
	acc2 := &authtypes.BaseAccount{Address: addr2.String()}
	acc3 := &authtypes.BaseAccount{Address: addr3.String()}
	accs := []simtestutil.GenesisAccount{
		{GenesisAccount: acc1, Coins: sdk.Coins{genCoin}},
		{GenesisAccount: acc2, Coins: sdk.Coins{genCoin}},
		{GenesisAccount: acc3, Coins: sdk.Coins{genCoin}},
	}

	var (
		bankKeeper    bankKeeper.Keeper
		stakingKeeper *stakingKeeper.Keeper
	)

	startupCfg := simtestutil.DefaultStartUpConfig()
	startupCfg.GenesisAccounts = accs

	app, err := simtestutil.SetupWithConfiguration(
		simtestutil.AppConfigWithNopLogger(testutil.AppConfig),
		startupCfg, &bankKeeper, &stakingKeeper)
	require.NoError(t, err)

	txConfig := moduletestutil.MakeTestTxConfig()
	nextHeight := func() int64 { return app.LastBlockHeight() + 1 }

	// Create destination validator (addr1).
	createDstMsg, err := types.NewMsgCreateValidator(
		sdk.ValAddress(addr1).String(), valKey.PubKey(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
		types.NewDescription("dst", "", "", "", ""), commissionRates, math.OneInt(),
	)
	require.NoError(t, err)
	_, _, err = simtestutil.SignCheckDeliver(
		t, txConfig, app.BaseApp, cmtproto.Header{Height: nextHeight()}, []sdk.Msg{createDstMsg},
		"", []uint64{0}, []uint64{0}, true, true, priv1,
	)
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: nextHeight()})
	require.NoError(t, err)

	// Create source validator (addr2).
	createSrcMsg, err := types.NewMsgCreateValidator(
		sdk.ValAddress(addr2).String(), valKey2.PubKey(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens),
		types.NewDescription("src", "", "", "", ""), commissionRates, math.OneInt(),
	)
	require.NoError(t, err)
	_, _, err = simtestutil.SignCheckDeliver(
		t, txConfig, app.BaseApp, cmtproto.Header{Height: nextHeight()}, []sdk.Msg{createSrcMsg},
		"", []uint64{1}, []uint64{0}, true, true, priv2,
	)
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: nextHeight()})
	require.NoError(t, err)

	// Bob delegates to source validator.
	delegateMsg := types.NewMsgDelegate(addr3.String(), sdk.ValAddress(addr2).String(), sdk.NewCoin(sdk.DefaultBondDenom, bobTokens))
	_, _, err = simtestutil.SignCheckDeliver(
		t, txConfig, app.BaseApp, cmtproto.Header{Height: nextHeight()}, []sdk.Msg{delegateMsg},
		"", []uint64{2}, []uint64{0}, true, true, priv3,
	)
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: nextHeight()})
	require.NoError(t, err)

	// Source operator undelegates all self-delegation so Bob becomes sole delegator.
	undelegateSelfMsg := types.NewMsgUndelegate(addr2.String(), sdk.ValAddress(addr2).String(), sdk.NewCoin(sdk.DefaultBondDenom, valTokens))
	_, _, err = simtestutil.SignCheckDeliver(
		t, txConfig, app.BaseApp, cmtproto.Header{Height: nextHeight()}, []sdk.Msg{undelegateSelfMsg},
		"", []uint64{1}, []uint64{1}, true, true, priv2,
	)
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{Height: nextHeight()})
	require.NoError(t, err)

	// Advance block time past unbonding period to trigger unbonding->unbonded via normal block flow.
	ctx := app.NewContext(true)
	unbondingTime, err := stakingKeeper.UnbondingTime(ctx)
	require.NoError(t, err)
	_, err = app.FinalizeBlock(&abci.RequestFinalizeBlock{
		Height: nextHeight(),
		Time:   ctx.BlockTime().Add(unbondingTime).Add(time.Second),
	})
	require.NoError(t, err)
	_, err = app.Commit()
	require.NoError(t, err)

	// Bob redelegates 100% of remaining shares from source to destination.
	beginRedelegateMsg := types.NewMsgBeginRedelegate(
		addr3.String(), sdk.ValAddress(addr2).String(), sdk.ValAddress(addr1).String(), sdk.NewCoin(sdk.DefaultBondDenom, bobTokens),
	)
	_, _, err = simtestutil.SignCheckDeliver(
		t, txConfig, app.BaseApp, cmtproto.Header{Height: nextHeight()}, []sdk.Msg{beginRedelegateMsg},
		"", []uint64{2}, []uint64{1}, true, true, priv3,
	)
	require.NoError(t, err)

	// This path should be complete-now: no redelegation entry, source validator removed.
	ctx = app.NewContext(true)
	_, err = stakingKeeper.GetValidator(ctx, sdk.ValAddress(addr2))
	require.ErrorIs(t, err, types.ErrNoValidatorFound)
	_, err = stakingKeeper.GetDelegation(ctx, addr3, sdk.ValAddress(addr2))
	require.ErrorIs(t, err, types.ErrNoDelegation)
	dstDel, err := stakingKeeper.GetDelegation(ctx, addr3, sdk.ValAddress(addr1))
	require.NoError(t, err)
	require.Equal(t, bobTokens, dstDel.Shares.RoundInt())
	_, err = stakingKeeper.GetRedelegation(ctx, addr3, sdk.ValAddress(addr2), sdk.ValAddress(addr1))
	require.ErrorIs(t, err, types.ErrNoRedelegation)
}
