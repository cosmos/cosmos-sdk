package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/simapp"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestUnbondingDelegationsMaxEntries(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 1, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	startTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)

	bondDenom := app.StakingKeeper.BondDenom(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, startTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// create a validator and a delegator to that validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])

	validator, issuedShares := validator.AddTokensFromDel(startTokens)
	require.Equal(t, startTokens, issuedShares.RoundInt())

	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	require.True(math.IntEq(t, startTokens, validator.BondedTokens()))
	require.True(t, validator.IsBonded())

	delegation := types.NewDelegation(addrDels[0], addrVals[0], issuedShares)
	app.StakingKeeper.SetDelegation(ctx, delegation)

	maxEntries := app.StakingKeeper.MaxEntries(ctx)

	oldBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// should all pass
	var completionTime time.Time
	for i := int64(0); i < int64(maxEntries); i++ {
		var err error
		ctx = ctx.WithBlockHeight(i)
		completionTime, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
		require.NoError(t, err)
	}

	newBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded := app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded.SubRaw(int64(maxEntries))))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(int64(maxEntries))))

	oldBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// an additional unbond should fail due to max entries
	_, err := app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
	require.Error(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	require.True(math.IntEq(t, newBonded, oldBonded))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded))

	// mature unbonding delegations
	ctx = ctx.WithBlockTime(completionTime)
	_, err = app.StakingKeeper.CompleteUnbonding(ctx, addrDels[0], addrVals[0])
	require.NoError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.SubRaw(int64(maxEntries))))

	oldNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount

	// unbonding  should work again
	_, err = app.StakingKeeper.Undelegate(ctx, addrDels[0], addrVals[0], math.LegacyNewDec(1))
	require.NoError(t, err)

	newBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetBondedPool(ctx).GetAddress(), bondDenom).Amount
	newNotBonded = app.BankKeeper.GetBalance(ctx, app.StakingKeeper.GetNotBondedPool(ctx).GetAddress(), bondDenom).Amount
	require.True(math.IntEq(t, newBonded, oldBonded.SubRaw(1)))
	require.True(math.IntEq(t, newNotBonded, oldNotBonded.AddRaw(1)))
}

func TestTransferDelegation(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 3, math.NewInt(10000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)

	// construct the validators
	amts := []math.Int{math.NewInt(9), math.NewInt(8), math.NewInt(7)}
	var validators [3]types.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(t, valAddrs[i], PKs[i])
		validators[i], _ = validators[i].AddTokensFromDel(amt)
	}
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[2], true)

	// try a transfer when there's nothing
	transferred := app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], math.LegacyNewDec(1000))
	require.Equal(t, math.LegacyZeroDec(), transferred)

	// stake some tokens
	bond1to1 := types.NewDelegation(addrDels[0], valAddrs[0], sdk.NewDec(99))
	app.StakingKeeper.SetDelegation(ctx, bond1to1)
	// stake to an unrelated validator so implementation has to skip it
	bond1to3 := types.NewDelegation(addrDels[0], valAddrs[2], sdk.NewDec(9))
	app.StakingKeeper.SetDelegation(ctx, bond1to3)

	// transfer nothing
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.ZeroDec())
	require.Equal(t, math.LegacyZeroDec(), transferred)

	// partial transfer, empty recipient
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.NewDec(10))
	require.Equal(t, sdk.NewDec(10), transferred)
	resBond, found := app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(89), resBond.Shares)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(10), resBond.Shares)

	// partial transfer, existing recipient
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.NewDec(11))
	require.Equal(t, transferred, sdk.NewDec(11))
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, math.LegacyNewDec(78), resBond.Shares)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)
	require.Equal(t, math.LegacyNewDec(21), resBond.Shares)

	// full transfer
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[0], sdk.NewDec(9999))
	require.Equal(t, transferred, math.LegacyNewDec(78))
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[0])
	require.False(t, found)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(99), resBond.Shares)

	// simulate redelegate to another validator
	bond1to2 := types.NewDelegation(addrDels[0], valAddrs[1], sdk.NewDec(20))
	app.StakingKeeper.SetDelegation(ctx, bond1to2)
	rd := types.NewRedelegation(
		addrDels[0],
		valAddrs[0],
		valAddrs[1],
		0,
		time.Unix(0, 0).UTC(),
		sdk.NewInt(20),
		sdk.NewDec(20),
		uint64(0),
	)
	app.StakingKeeper.SetRedelegation(ctx, rd)

	// partial transfer from redelegation
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[1], sdk.NewDec(7))
	require.Equal(t, sdk.NewDec(7), transferred)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[1])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(13), resBond.Shares)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[1])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(7), resBond.Shares)

	// stake more alongside redelegation
	bond1to2, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[1])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(13), bond1to2.Shares)
	bond1to2.Shares = sdk.NewDec(47) // add 34 shares
	app.StakingKeeper.SetDelegation(ctx, bond1to2)

	// full transfer from partial redelegation
	transferred = app.StakingKeeper.TransferDelegation(ctx, addrDels[0], addrDels[1], valAddrs[1], sdk.NewDec(9999))
	require.Equal(t, sdk.NewDec(47), transferred)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[0], valAddrs[1])
	require.False(t, found)
	resBond, found = app.StakingKeeper.GetDelegation(ctx, addrDels[1], valAddrs[1])
	require.True(t, found)
	require.Equal(t, sdk.NewDec(54), resBond.Shares)
}

func TestTransferUnbonding(t *testing.T) {
	_, app, ctx := createTestInput(t)

	addrDels := simapp.AddTestAddrsIncremental(app, ctx, 2, math.NewInt(10000))
	valAddrs := simtestutil.ConvertAddrsToValAddrs(addrDels)

	// try to transfer when there's nothing
	transferred := app.StakingKeeper.TransferUnbonding(ctx, addrDels[0], addrDels[1], valAddrs[0], math.NewInt(30))
	require.Equal(t, math.ZeroInt(), transferred)
	_, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[1], valAddrs[0])
	require.False(t, found)

	// set an UnbondingDelegation with one entry
	ubd := types.NewUnbondingDelegation(
		addrDels[0],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		math.NewInt(5),
		uint64(0),
	)
	app.StakingKeeper.SetUnbondingDelegation(ctx, ubd)

	// transfer nothing
	transferred = app.StakingKeeper.TransferUnbonding(ctx, addrDels[0], addrDels[1], valAddrs[0], math.ZeroInt())
	require.Equal(t, math.ZeroInt(), transferred)

	// partial transfer
	transferred = app.StakingKeeper.TransferUnbonding(ctx, addrDels[0], addrDels[1], valAddrs[0], math.NewInt(3))
	require.Equal(t, math.NewInt(3), transferred)
	ubd.Entries[0].Balance = math.NewInt(2)
	resUnbond, found := app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], valAddrs[0])
	require.True(t, found)
	require.Equal(t, ubd, resUnbond)
	resUnbond, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)
	wantDestUnbond := types.NewUnbondingDelegation(
		addrDels[1],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		math.NewInt(3),
		uint64(1),
	)

	require.Equal(t, wantDestUnbond, resUnbond)

	// add another entry
	completionTime := time.Unix(3600, 0).UTC()
	ubdTo := app.StakingKeeper.SetUnbondingDelegationEntry(ctx, addrDels[0], valAddrs[0], 1, completionTime, math.NewInt(57))
	app.StakingKeeper.InsertUBDQueue(ctx, ubdTo, completionTime)

	// full transfer
	resUnbond, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)

	transferred = app.StakingKeeper.TransferUnbonding(ctx, addrDels[0], addrDels[1], valAddrs[0], math.NewInt(999))
	require.Equal(t, math.NewInt(59), transferred)
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[0], valAddrs[0])
	require.False(t, found)
	resUnbond, found = app.StakingKeeper.GetUnbondingDelegation(ctx, addrDels[1], valAddrs[0])
	require.True(t, found)
	require.Equal(t, 2, len(resUnbond.Entries))
	require.Equal(t, math.NewInt(5), resUnbond.Entries[0].Balance)
	require.Equal(t, math.NewInt(57), resUnbond.Entries[1].Balance)
}
