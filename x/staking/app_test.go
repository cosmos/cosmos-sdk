package staking_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func checkValidator(t *testing.T, ba *baseapp.BaseApp, stakingKeeper *keeper.Keeper, addr sdk.ValAddress, expFound bool) types.Validator {
	ctxCheck := ba.NewContext(true, tmproto.Header{})
	validator, found := stakingKeeper.GetValidator(ctxCheck, addr)

	require.Equal(t, expFound, found)
	return validator
}

func checkDelegation(
	t *testing.T, ba *baseapp.BaseApp, stakingKeeper *keeper.Keeper, delegatorAddr sdk.AccAddress,
	validatorAddr sdk.ValAddress, expFound bool, expShares sdk.Dec,
) {
	ctxCheck := ba.NewContext(true, tmproto.Header{})
	delegation, found := stakingKeeper.GetDelegation(ctxCheck, delegatorAddr, validatorAddr)
	if expFound {
		require.True(t, found)
		require.True(sdk.DecEq(t, expShares, delegation.Shares))

		return
	}

	require.False(t, found)
}

func TestStakingMsgs(t *testing.T) {
	genTokens := sdk.TokensFromConsensusPower(42, sdk.DefaultPowerReduction)
	bondTokens := sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	genCoin := sdk.NewCoin(sdk.DefaultBondDenom, genTokens)
	bondCoin := sdk.NewCoin(sdk.DefaultBondDenom, bondTokens)

	var (
		txConfig      client.TxConfig
		bankKeeper    bankkeeper.Keeper
		stakingKeeper *keeper.Keeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&txConfig,
		&bankKeeper,
		&stakingKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})
	banktestutil.FundAccount(bankKeeper, ctx, addr1, sdk.Coins{genCoin})
	banktestutil.FundAccount(bankKeeper, ctx, addr2, sdk.Coins{genCoin})
	app.Commit()

	require.True(t, simtestutil.CheckBalance(app.BaseApp, bankKeeper, addr1, sdk.Coins{genCoin}))
	require.True(t, simtestutil.CheckBalance(app.BaseApp, bankKeeper, addr2, sdk.Coins{genCoin}))

	// create validator
	description := types.NewDescription("foo_moniker", "", "", "", "")
	createValidatorMsg, err := types.NewMsgCreateValidator(
		sdk.ValAddress(addr1), valKey.PubKey(), bondCoin, description, commissionRates, sdk.OneInt(),
	)
	require.NoError(t, err)

	header := tmproto.Header{Height: app.LastBlockHeight() + 1}

	_, _, err = simtestutil.SignCheckDeliver(txConfig, app.BaseApp, header, []sdk.Msg{createValidatorMsg}, "", []uint64{0}, []uint64{0}, true, true, priv1)
	require.NoError(t, err)
	require.True(t, simtestutil.CheckBalance(app.BaseApp, bankKeeper, addr1, sdk.Coins{genCoin.Sub(bondCoin)}))

	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	validator := checkValidator(t, app.BaseApp, stakingKeeper, sdk.ValAddress(addr1), true)
	require.Equal(t, sdk.ValAddress(addr1).String(), validator.OperatorAddress)
	require.Equal(t, types.Bonded, validator.Status)
	require.True(sdk.IntEq(t, bondTokens, validator.BondedTokens()))

	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	app.BeginBlock(abci.RequestBeginBlock{Header: header})

	// edit the validator
	description = types.NewDescription("bar_moniker", "", "", "", "")
	editValidatorMsg := types.NewMsgEditValidator(sdk.ValAddress(addr1), description, nil, nil)

	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(txConfig, app.BaseApp, header, []sdk.Msg{editValidatorMsg}, "", []uint64{0}, []uint64{1}, true, true, priv1)
	require.NoError(t, err)

	validator = checkValidator(t, app.BaseApp, stakingKeeper, sdk.ValAddress(addr1), true)
	require.Equal(t, description, validator.Description)

	// delegate
	require.True(t, simtestutil.CheckBalance(app.BaseApp, bankKeeper, addr2, sdk.Coins{genCoin}))
	delegateMsg := types.NewMsgDelegate(addr2, sdk.ValAddress(addr1), bondCoin)

	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(txConfig, app.BaseApp, header, []sdk.Msg{delegateMsg}, "", []uint64{1}, []uint64{0}, true, true, priv2)
	require.NoError(t, err)

	require.True(t, simtestutil.CheckBalance(app.BaseApp, bankKeeper, addr2, sdk.Coins{genCoin.Sub(bondCoin)}))
	checkDelegation(t, app.BaseApp, stakingKeeper, addr2, sdk.ValAddress(addr1), true, sdk.NewDecFromInt(bondTokens))

	// begin unbonding
	beginUnbondingMsg := types.NewMsgUndelegate(addr2, sdk.ValAddress(addr1), bondCoin)
	header = tmproto.Header{Height: app.LastBlockHeight() + 1}
	_, _, err = simtestutil.SignCheckDeliver(txConfig, app.BaseApp, header, []sdk.Msg{beginUnbondingMsg}, "", []uint64{1}, []uint64{1}, true, true, priv2)
	require.NoError(t, err)

	// delegation should exist anymore
	checkDelegation(t, app.BaseApp, stakingKeeper, addr2, sdk.ValAddress(addr1), false, sdk.Dec{})

	// balance should be the same because bonding not yet complete
	require.True(t, simtestutil.CheckBalance(app.BaseApp, bankKeeper, addr2, sdk.Coins{genCoin.Sub(bondCoin)}))
}
