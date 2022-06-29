package keeper_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func newMonikerValidator(t testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey, moniker string) types.Validator {
	v, err := types.NewValidator(operator, pubKey, types.Description{Moniker: moniker})
	require.NoError(t, err)
	return v
}

func bootstrapValidatorTest(t testing.TB, power int64, numAddrs int) (bankkeeper.Keeper, *keeper.Keeper, authkeeper.AccountKeeper, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	var (
		bankKeeper    bankkeeper.Keeper
		stakingKeeper *keeper.Keeper
		accountKeeper authkeeper.AccountKeeper
	)

	app, err := simtestutil.Setup(testutil.AppConfig,
		&bankKeeper,
		&stakingKeeper,
		&accountKeeper,
	)
	require.NoError(t, err)

	ctx := app.BaseApp.NewContext(false, tmproto.Header{})

	addrDels := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, numAddrs, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	amt := stakingKeeper.TokensFromConsensusPower(ctx, power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)
	// set bonded pool supply
	accountKeeper.SetModuleAccount(ctx, notBondedPool)

	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), totalSupply))

	// unbond genesis validator delegations
	delegations := stakingKeeper.GetAllDelegations(ctx)
	require.Len(t, delegations, 1)
	delegation := delegations[0]

	_, err = stakingKeeper.Undelegate(ctx, delegation.GetDelegatorAddr(), delegation.GetValidatorAddr(), delegation.Shares)
	require.NoError(t, err)

	// end block to unbond genesis validator
	staking.EndBlocker(ctx, stakingKeeper)

	return bankKeeper, stakingKeeper, accountKeeper, ctx, addrDels, addrVals
}

func initValidators(t testing.TB, power int64, numAddrs int, powers []int64) (bankkeeper.Keeper, *keeper.Keeper, authkeeper.AccountKeeper, sdk.Context, []sdk.AccAddress, []sdk.ValAddress, []types.Validator) {
	bankKeeper, stakingKeeper, accountKeeper, ctx, addrs, valAddrs := bootstrapValidatorTest(t, power, numAddrs)
	pks := simtestutil.CreateTestPubKeys(numAddrs)

	vs := make([]types.Validator, len(powers))
	for i, power := range powers {
		vs[i] = teststaking.NewValidator(t, sdk.ValAddress(addrs[i]), pks[i])
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		vs[i], _ = vs[i].AddTokensFromDel(tokens)
	}
	return bankKeeper, stakingKeeper, accountKeeper, ctx, addrs, valAddrs, vs
}

func TestSetValidator(t *testing.T) {
	_, stakingKeeper, _, ctx, _, _ := bootstrapValidatorTest(t, 10, 1000)

	valPubKey := PKs[0]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	valTokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)

	// test how the validator is set from a purely unbonbed pool
	validator := teststaking.NewValidator(t, valAddr, valPubKey)
	validator, _ = validator.AddTokensFromDel(valTokens)
	require.Equal(t, types.Unbonded, validator.Status)
	require.Equal(t, valTokens, validator.Tokens)
	require.Equal(t, valTokens, validator.DelegatorShares.RoundInt())
	stakingKeeper.SetValidator(ctx, validator)
	stakingKeeper.SetValidatorByPowerIndex(ctx, validator)

	// ensure update
	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, 1)
	validator, found := stakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, validator.ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])

	// after the save the validator should be bonded
	require.Equal(t, types.Bonded, validator.Status)
	require.Equal(t, valTokens, validator.Tokens)
	require.Equal(t, valTokens, validator.DelegatorShares.RoundInt())

	// Check each store for being saved
	resVal, found := stakingKeeper.GetValidator(ctx, valAddr)
	require.True(ValEq(t, validator, resVal))
	require.True(t, found)

	resVals := stakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validator, resVals[0]))

	resVals = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validator, resVals[0]))

	resVals = stakingKeeper.GetValidators(ctx, 1)
	require.Equal(t, 1, len(resVals))

	resVals = stakingKeeper.GetValidators(ctx, 10)
	require.Equal(t, 2, len(resVals))

	allVals := stakingKeeper.GetAllValidators(ctx)
	require.Equal(t, 2, len(allVals))
}

func TestUpdateValidatorByPowerIndex(t *testing.T) {
	bankKeeper, stakingKeeper, accountKeeper, ctx, _, _ := bootstrapValidatorTest(t, 0, 1000)

	addrDels := simtestutil.AddTestAddrsIncremental(bankKeeper, stakingKeeper, ctx, 1, sdk.NewInt(10000))
	addrVals := simtestutil.ConvertAddrsToValAddrs(addrDels)

	bondedPool := stakingKeeper.GetBondedPool(ctx)
	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), stakingKeeper.TokensFromConsensusPower(ctx, 1234)))))
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), stakingKeeper.TokensFromConsensusPower(ctx, 10000)))))

	accountKeeper.SetModuleAccount(ctx, bondedPool)
	accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// add a validator
	validator := teststaking.NewValidator(t, addrVals[0], PKs[0])
	validator, delSharesCreated := validator.AddTokensFromDel(stakingKeeper.TokensFromConsensusPower(ctx, 100))
	require.Equal(t, types.Unbonded, validator.Status)
	require.Equal(t, stakingKeeper.TokensFromConsensusPower(ctx, 100), validator.Tokens)
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validator, true)
	validator, found := stakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.Equal(t, stakingKeeper.TokensFromConsensusPower(ctx, 100), validator.Tokens)

	power := types.GetValidatorsByPowerIndexKey(validator, stakingKeeper.PowerReduction(ctx))
	require.True(t, keeper.ValidatorByPowerIndexExists(ctx, stakingKeeper, power))

	// burn half the delegator shares
	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validator)
	validator, burned := validator.RemoveDelShares(delSharesCreated.Quo(sdk.NewDec(2)))
	require.Equal(t, stakingKeeper.TokensFromConsensusPower(ctx, 50), burned)
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validator, true) // update the validator, possibly kicking it out
	require.False(t, keeper.ValidatorByPowerIndexExists(ctx, stakingKeeper, power))

	validator, found = stakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)

	power = types.GetValidatorsByPowerIndexKey(validator, stakingKeeper.PowerReduction(ctx))
	require.True(t, keeper.ValidatorByPowerIndexExists(ctx, stakingKeeper, power))
}

func TestUpdateBondedValidatorsDecreaseCliff(t *testing.T) {
	numVals := 10
	maxVals := 5

	// create context, keeper, and pool for tests
	bankKeeper, stakingKeeper, accountKeeper, ctx, _, valAddrs := bootstrapValidatorTest(t, 10, 100)

	bondedPool := stakingKeeper.GetBondedPool(ctx)
	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)

	// create keeper parameters
	params := stakingKeeper.GetParams(ctx)
	params.MaxValidators = uint32(maxVals)
	stakingKeeper.SetParams(ctx, params)

	// create a random pool
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), stakingKeeper.TokensFromConsensusPower(ctx, 1234)))))
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), stakingKeeper.TokensFromConsensusPower(ctx, 10000)))))

	accountKeeper.SetModuleAccount(ctx, bondedPool)
	accountKeeper.SetModuleAccount(ctx, notBondedPool)

	validators := make([]types.Validator, numVals)
	for i := 0; i < len(validators); i++ {
		moniker := fmt.Sprintf("val#%d", int64(i))
		val := newMonikerValidator(t, valAddrs[i], PKs[i], moniker)
		delTokens := stakingKeeper.TokensFromConsensusPower(ctx, int64((i+1)*10))
		val, _ = val.AddTokensFromDel(delTokens)

		val = keeper.TestingUpdateValidator(stakingKeeper, ctx, val, true)
		validators[i] = val
	}

	nextCliffVal := validators[numVals-maxVals+1]

	// remove enough tokens to kick out the validator below the current cliff
	// validator and next in line cliff validator
	stakingKeeper.DeleteValidatorByPowerIndex(ctx, nextCliffVal)
	shares := stakingKeeper.TokensFromConsensusPower(ctx, 21)
	nextCliffVal, _ = nextCliffVal.RemoveDelShares(sdk.NewDecFromInt(shares))
	nextCliffVal = keeper.TestingUpdateValidator(stakingKeeper, ctx, nextCliffVal, true)

	expectedValStatus := map[int]types.BondStatus{
		9: types.Bonded, 8: types.Bonded, 7: types.Bonded, 5: types.Bonded, 4: types.Bonded,
		0: types.Unbonding, 1: types.Unbonding, 2: types.Unbonding, 3: types.Unbonding, 6: types.Unbonding,
	}

	// require all the validators have their respective statuses
	for valIdx, status := range expectedValStatus {
		valAddr := validators[valIdx].OperatorAddress
		addr, err := sdk.ValAddressFromBech32(valAddr)
		require.NoError(t, err)
		val, _ := stakingKeeper.GetValidator(ctx, addr)

		require.Equal(
			t, status, val.GetStatus(),
			fmt.Sprintf("expected validator at index %v to have status: %s", valIdx, status),
		)
	}
}

func TestSlashToZeroPowerRemoved(t *testing.T) {
	// initialize setup
	bankKeeper, stakingKeeper, accountKeeper, ctx, _, addrVals := bootstrapValidatorTest(t, 100, 20)

	// add a validator
	validator := teststaking.NewValidator(t, addrVals[0], PKs[0])
	valTokens := stakingKeeper.TokensFromConsensusPower(ctx, 100)

	bondedPool := stakingKeeper.GetBondedPool(ctx)

	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), valTokens))))

	accountKeeper.SetModuleAccount(ctx, bondedPool)

	validator, _ = validator.AddTokensFromDel(valTokens)
	require.Equal(t, types.Unbonded, validator.Status)
	require.Equal(t, valTokens, validator.Tokens)
	stakingKeeper.SetValidatorByConsAddr(ctx, validator)
	validator = keeper.TestingUpdateValidator(stakingKeeper, ctx, validator, true)
	require.Equal(t, valTokens, validator.Tokens, "\nvalidator %v\npool %v", validator, valTokens)

	// slash the validator by 100%
	stakingKeeper.Slash(ctx, sdk.ConsAddress(PKs[0].Address()), 0, 100, sdk.OneDec())
	// apply TM updates
	applyValidatorSetUpdates(t, ctx, stakingKeeper, -1)
	// validator should be unbonding
	validator, _ = stakingKeeper.GetValidator(ctx, addrVals[0])
	require.Equal(t, validator.GetStatus(), types.Unbonding)
}

// This function tests UpdateValidator, GetValidator, GetLastValidators, RemoveValidator
func TestValidatorBasics(t *testing.T) {
	_, stakingKeeper, _, ctx, _, addrVals := bootstrapValidatorTest(t, 1000, 20)

	// construct the validators
	var validators [3]types.Validator
	powers := []int64{9, 8, 7}
	for i, power := range powers {
		validators[i] = teststaking.NewValidator(t, addrVals[i], PKs[i])
		validators[i].Status = types.Unbonded
		validators[i].Tokens = sdk.ZeroInt()
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)

		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}
	require.Equal(t, stakingKeeper.TokensFromConsensusPower(ctx, 9), validators[0].Tokens)
	require.Equal(t, stakingKeeper.TokensFromConsensusPower(ctx, 8), validators[1].Tokens)
	require.Equal(t, stakingKeeper.TokensFromConsensusPower(ctx, 7), validators[2].Tokens)

	// check the empty keeper first
	_, found := stakingKeeper.GetValidator(ctx, addrVals[0])
	require.False(t, found)
	resVals := stakingKeeper.GetLastValidators(ctx)
	require.Zero(t, len(resVals))

	resVals = stakingKeeper.GetValidators(ctx, 2)
	require.Len(t, resVals, 1)

	// set and retrieve a record
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], true)
	stakingKeeper.SetValidatorByConsAddr(ctx, validators[0])
	resVal, found := stakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.True(ValEq(t, validators[0], resVal))

	// retrieve from consensus
	resVal, found = stakingKeeper.GetValidatorByConsAddr(ctx, sdk.ConsAddress(PKs[0].Address()))
	require.True(t, found)
	require.True(ValEq(t, validators[0], resVal))
	resVal, found = stakingKeeper.GetValidatorByConsAddr(ctx, sdk.GetConsAddress(PKs[0]))
	require.True(t, found)
	require.True(ValEq(t, validators[0], resVal))

	resVals = stakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validators[0], resVals[0]))
	require.Equal(t, types.Bonded, validators[0].Status)
	require.True(sdk.IntEq(t, stakingKeeper.TokensFromConsensusPower(ctx, 9), validators[0].BondedTokens()))

	// modify a records, save, and retrieve
	validators[0].Status = types.Bonded
	validators[0].Tokens = stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validators[0].DelegatorShares = sdk.NewDecFromInt(validators[0].Tokens)
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], true)
	resVal, found = stakingKeeper.GetValidator(ctx, addrVals[0])
	require.True(t, found)
	require.True(ValEq(t, validators[0], resVal))

	resVals = stakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 1, len(resVals))
	require.True(ValEq(t, validators[0], resVals[0]))

	// add other validators
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[2], true)
	resVal, found = stakingKeeper.GetValidator(ctx, addrVals[1])
	require.True(t, found)
	require.True(ValEq(t, validators[1], resVal))
	resVal, found = stakingKeeper.GetValidator(ctx, addrVals[2])
	require.True(t, found)
	require.True(ValEq(t, validators[2], resVal))

	resVals = stakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 3, len(resVals))
	require.True(ValEq(t, validators[0], resVals[0])) // order doesn't matter here
	require.True(ValEq(t, validators[1], resVals[1]))
	require.True(ValEq(t, validators[2], resVals[2]))

	// remove a record

	// shouldn't be able to remove if status is not unbonded
	require.PanicsWithValue(t,
		"cannot call RemoveValidator on bonded or unbonding validators",
		func() { stakingKeeper.RemoveValidator(ctx, validators[1].GetOperator()) })

	// shouldn't be able to remove if there are still tokens left
	validators[1].Status = types.Unbonded
	stakingKeeper.SetValidator(ctx, validators[1])
	require.PanicsWithValue(t,
		"attempting to remove a validator which still contains tokens",
		func() { stakingKeeper.RemoveValidator(ctx, validators[1].GetOperator()) })

	validators[1].Tokens = sdk.ZeroInt()                            // ...remove all tokens
	stakingKeeper.SetValidator(ctx, validators[1])                  // ...set the validator
	stakingKeeper.RemoveValidator(ctx, validators[1].GetOperator()) // Now it can be removed.
	_, found = stakingKeeper.GetValidator(ctx, addrVals[1])
	require.False(t, found)
}

// test how the validators are sorted, tests GetBondedValidatorsByPower
func TestGetValidatorSortingUnmixed(t *testing.T) {
	_, stakingKeeper, _, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// initialize some validators into the state
	amts := []sdk.Int{
		sdk.NewIntFromUint64(0),
		stakingKeeper.PowerReduction(ctx).MulRaw(100),
		stakingKeeper.PowerReduction(ctx),
		stakingKeeper.PowerReduction(ctx).MulRaw(400),
		stakingKeeper.PowerReduction(ctx).MulRaw(200),
	}
	n := len(amts)
	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		validators[i].Status = types.Bonded
		validators[i].Tokens = amt
		validators[i].DelegatorShares = sdk.NewDecFromInt(amt)
		keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[i], true)
	}

	// first make sure everything made it in to the gotValidator group
	resValidators := stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, n, len(resValidators))
	require.Equal(t, sdk.NewInt(400).Mul(stakingKeeper.PowerReduction(ctx)), resValidators[0].BondedTokens(), "%v", resValidators)
	require.Equal(t, sdk.NewInt(200).Mul(stakingKeeper.PowerReduction(ctx)), resValidators[1].BondedTokens(), "%v", resValidators)
	require.Equal(t, sdk.NewInt(100).Mul(stakingKeeper.PowerReduction(ctx)), resValidators[2].BondedTokens(), "%v", resValidators)
	require.Equal(t, sdk.NewInt(1).Mul(stakingKeeper.PowerReduction(ctx)), resValidators[3].BondedTokens(), "%v", resValidators)
	require.Equal(t, sdk.NewInt(0), resValidators[4].BondedTokens(), "%v", resValidators)
	require.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	require.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
	require.Equal(t, validators[1].OperatorAddress, resValidators[2].OperatorAddress, "%v", resValidators)
	require.Equal(t, validators[2].OperatorAddress, resValidators[3].OperatorAddress, "%v", resValidators)
	require.Equal(t, validators[0].OperatorAddress, resValidators[4].OperatorAddress, "%v", resValidators)

	// test a basic increase in voting power
	validators[3].Tokens = sdk.NewInt(500).Mul(stakingKeeper.PowerReduction(ctx))
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[3], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	require.True(ValEq(t, validators[3], resValidators[0]))

	// test a decrease in voting power
	validators[3].Tokens = sdk.NewInt(300).Mul(stakingKeeper.PowerReduction(ctx))
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[3], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	require.True(ValEq(t, validators[3], resValidators[0]))
	require.True(ValEq(t, validators[4], resValidators[1]))

	// test equal voting power, different age
	validators[3].Tokens = sdk.NewInt(200).Mul(stakingKeeper.PowerReduction(ctx))
	ctx = ctx.WithBlockHeight(10)
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[3], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	require.True(ValEq(t, validators[3], resValidators[0]))
	require.True(ValEq(t, validators[4], resValidators[1]))

	// no change in voting power - no change in sort
	ctx = ctx.WithBlockHeight(20)
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[4], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	require.True(ValEq(t, validators[3], resValidators[0]))
	require.True(ValEq(t, validators[4], resValidators[1]))

	// change in voting power of both validators, both still in v-set, no age change
	validators[3].Tokens = sdk.NewInt(300).Mul(stakingKeeper.PowerReduction(ctx))
	validators[4].Tokens = sdk.NewInt(300).Mul(stakingKeeper.PowerReduction(ctx))
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[3], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n)
	ctx = ctx.WithBlockHeight(30)
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[4], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, len(resValidators), n, "%v", resValidators)
	require.True(ValEq(t, validators[3], resValidators[0]))
	require.True(ValEq(t, validators[4], resValidators[1]))
}

func TestGetValidatorSortingMixed(t *testing.T) {
	bankKeeper, stakingKeeper, accountKeeper, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	bondedPool := stakingKeeper.GetBondedPool(ctx)
	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)

	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), stakingKeeper.TokensFromConsensusPower(ctx, 501)))))
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(stakingKeeper.BondDenom(ctx), stakingKeeper.TokensFromConsensusPower(ctx, 0)))))

	accountKeeper.SetModuleAccount(ctx, notBondedPool)
	accountKeeper.SetModuleAccount(ctx, bondedPool)

	// now 2 max resValidators
	params := stakingKeeper.GetParams(ctx)
	params.MaxValidators = 2
	stakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	amts := []sdk.Int{
		sdk.NewIntFromUint64(0),
		stakingKeeper.PowerReduction(ctx).MulRaw(100),
		stakingKeeper.PowerReduction(ctx),
		stakingKeeper.PowerReduction(ctx).MulRaw(400),
		stakingKeeper.PowerReduction(ctx).MulRaw(200),
	}

	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = teststaking.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		validators[i].DelegatorShares = sdk.NewDecFromInt(amt)
		validators[i].Status = types.Bonded
		validators[i].Tokens = amt
		keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[i], true)
	}

	val0, found := stakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[0]))
	require.True(t, found)
	val1, found := stakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[1]))
	require.True(t, found)
	val2, found := stakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[2]))
	require.True(t, found)
	val3, found := stakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[3]))
	require.True(t, found)
	val4, found := stakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[4]))
	require.True(t, found)
	require.Equal(t, types.Bonded, val0.Status)
	require.Equal(t, types.Unbonding, val1.Status)
	require.Equal(t, types.Unbonding, val2.Status)
	require.Equal(t, types.Bonded, val3.Status)
	require.Equal(t, types.Bonded, val4.Status)

	// first make sure everything made it in to the gotValidator group
	resValidators := stakingKeeper.GetBondedValidatorsByPower(ctx)
	// The validators returned should match the max validators
	require.Equal(t, 2, len(resValidators))
	require.Equal(t, sdk.NewInt(400).Mul(stakingKeeper.PowerReduction(ctx)), resValidators[0].BondedTokens(), "%v", resValidators)
	require.Equal(t, sdk.NewInt(200).Mul(stakingKeeper.PowerReduction(ctx)), resValidators[1].BondedTokens(), "%v", resValidators)
	require.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	require.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
}

// TODO separate out into multiple tests
func TestGetValidatorsEdgeCases(t *testing.T) {
	bankKeeper, stakingKeeper, accountKeeper, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// set max validators to 2
	params := stakingKeeper.GetParams(ctx)
	nMax := uint32(2)
	params.MaxValidators = nMax
	stakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400}
	var validators [4]types.Validator
	for i, power := range powers {
		moniker := fmt.Sprintf("val#%d", int64(i))
		validators[i] = newMonikerValidator(t, sdk.ValAddress(addrs[i]), PKs[i], moniker)

		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		notBondedPool := stakingKeeper.GetNotBondedPool(ctx)
		require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, tokens))))
		accountKeeper.SetModuleAccount(ctx, notBondedPool)
		validators[i] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[i], true)
	}

	// ensure that the first two bonded validators are the largest validators
	resValidators := stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	require.True(ValEq(t, validators[2], resValidators[0]))
	require.True(ValEq(t, validators[3], resValidators[1]))

	// delegate 500 tokens to validator 0
	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	delTokens := stakingKeeper.TokensFromConsensusPower(ctx, 500)
	validators[0], _ = validators[0].AddTokensFromDel(delTokens)
	notBondedPool := stakingKeeper.GetNotBondedPool(ctx)

	newTokens := sdk.NewCoins()

	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), newTokens))
	accountKeeper.SetModuleAccount(ctx, notBondedPool)

	// test that the two largest validators are
	//   a) validator 0 with 500 tokens
	//   b) validator 2 with 400 tokens (delegated before validator 3)
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	require.True(ValEq(t, validators[0], resValidators[0]))
	require.True(ValEq(t, validators[2], resValidators[1]))

	// A validator which leaves the bonded validator set due to a decrease in voting power,
	// then increases to the original voting power, does not get its spot back in the
	// case of a tie.
	//
	// Order of operations for this test:
	//  - validator 3 enter validator set with 1 new token
	//  - validator 3 removed validator set by removing 201 tokens (validator 2 enters)
	//  - validator 3 adds 200 tokens (equal to validator 2 now) and does not get its spot back

	// validator 3 enters bonded validator set
	ctx = ctx.WithBlockHeight(40)

	var found bool
	validators[3], found = stakingKeeper.GetValidator(ctx, validators[3].GetOperator())
	require.True(t, found)
	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = validators[3].AddTokensFromDel(stakingKeeper.TokensFromConsensusPower(ctx, 1))

	notBondedPool = stakingKeeper.GetNotBondedPool(ctx)
	newTokens = sdk.NewCoins(sdk.NewCoin(params.BondDenom, stakingKeeper.TokensFromConsensusPower(ctx, 1)))
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), newTokens))
	accountKeeper.SetModuleAccount(ctx, notBondedPool)

	validators[3] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[3], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	require.True(ValEq(t, validators[0], resValidators[0]))
	require.True(ValEq(t, validators[3], resValidators[1]))

	// validator 3 kicked out temporarily
	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	rmTokens := validators[3].TokensFromShares(sdk.NewDec(201)).TruncateInt()
	validators[3], _ = validators[3].RemoveDelShares(sdk.NewDec(201))

	bondedPool := stakingKeeper.GetBondedPool(ctx)
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, rmTokens))))
	accountKeeper.SetModuleAccount(ctx, bondedPool)

	validators[3] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[3], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	require.True(ValEq(t, validators[0], resValidators[0]))
	require.True(ValEq(t, validators[2], resValidators[1]))

	// validator 3 does not get spot back
	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = validators[3].AddTokensFromDel(sdk.NewInt(200))

	notBondedPool = stakingKeeper.GetNotBondedPool(ctx)
	require.NoError(t, banktestutil.FundModuleAccount(bankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, sdk.NewInt(200)))))
	accountKeeper.SetModuleAccount(ctx, notBondedPool)

	validators[3] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[3], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, nMax, uint32(len(resValidators)))
	require.True(ValEq(t, validators[0], resValidators[0]))
	require.True(ValEq(t, validators[2], resValidators[1]))
	_, exists := stakingKeeper.GetValidator(ctx, validators[3].GetOperator())
	require.True(t, exists)
}

func TestValidatorBondHeight(t *testing.T) {
	_, stakingKeeper, _, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// now 2 max resValidators
	params := stakingKeeper.GetParams(ctx)
	params.MaxValidators = 2
	stakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	var validators [3]types.Validator
	validators[0] = teststaking.NewValidator(t, sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0])
	validators[1] = teststaking.NewValidator(t, sdk.ValAddress(addrs[1]), PKs[1])
	validators[2] = teststaking.NewValidator(t, sdk.ValAddress(addrs[2]), PKs[2])

	tokens0 := stakingKeeper.TokensFromConsensusPower(ctx, 200)
	tokens1 := stakingKeeper.TokensFromConsensusPower(ctx, 100)
	tokens2 := stakingKeeper.TokensFromConsensusPower(ctx, 100)
	validators[0], _ = validators[0].AddTokensFromDel(tokens0)
	validators[1], _ = validators[1].AddTokensFromDel(tokens1)
	validators[2], _ = validators[2].AddTokensFromDel(tokens2)

	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], true)

	////////////////////////////////////////
	// If two validators both increase to the same voting power in the same block,
	// the one with the first transaction should become bonded
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[2], true)

	resValidators := stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, uint32(len(resValidators)), params.MaxValidators)

	require.True(ValEq(t, validators[0], resValidators[0]))
	require.True(ValEq(t, validators[1], resValidators[1]))
	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[2])
	delTokens := stakingKeeper.TokensFromConsensusPower(ctx, 50)
	validators[1], _ = validators[1].AddTokensFromDel(delTokens)
	validators[2], _ = validators[2].AddTokensFromDel(delTokens)
	validators[2] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[2], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, params.MaxValidators, uint32(len(resValidators)))
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], true)
	require.True(ValEq(t, validators[0], resValidators[0]))
	require.True(ValEq(t, validators[2], resValidators[1]))
}

func TestFullValidatorSetPowerChange(t *testing.T) {
	_, stakingKeeper, _, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	params := stakingKeeper.GetParams(ctx)
	max := 2
	params.MaxValidators = uint32(2)
	stakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400, 200}
	var validators [5]types.Validator
	for i, power := range powers {
		validators[i] = teststaking.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
		keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[i], true)
	}
	for i := range powers {
		var found bool
		validators[i], found = stakingKeeper.GetValidator(ctx, validators[i].GetOperator())
		require.True(t, found)
	}
	require.Equal(t, types.Unbonded, validators[0].Status)
	require.Equal(t, types.Unbonding, validators[1].Status)
	require.Equal(t, types.Bonded, validators[2].Status)
	require.Equal(t, types.Bonded, validators[3].Status)
	require.Equal(t, types.Unbonded, validators[4].Status)
	resValidators := stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, max, len(resValidators))
	require.True(ValEq(t, validators[2], resValidators[0])) // in the order of txs
	require.True(ValEq(t, validators[3], resValidators[1]))

	// test a swap in voting power

	tokens := stakingKeeper.TokensFromConsensusPower(ctx, 600)
	validators[0], _ = validators[0].AddTokensFromDel(tokens)
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], true)
	resValidators = stakingKeeper.GetBondedValidatorsByPower(ctx)
	require.Equal(t, max, len(resValidators))
	require.True(ValEq(t, validators[0], resValidators[0]))
	require.True(ValEq(t, validators[2], resValidators[1]))
}

func TestApplyAndReturnValidatorSetUpdatesAllNone(t *testing.T) {
	_, stakingKeeper, _, ctx, _, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = teststaking.NewValidator(t, valAddr, valPubKey)
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}

	// test from nothing to something
	//  tendermintUpdate set: {} -> {c1, c3}
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)
	stakingKeeper.SetValidator(ctx, validators[0])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[0])
	stakingKeeper.SetValidator(ctx, validators[1])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[1])

	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)
	validators[0], _ = stakingKeeper.GetValidator(ctx, validators[0].GetOperator())
	validators[1], _ = stakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	require.Equal(t, validators[0].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[1])
	require.Equal(t, validators[1].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesIdentical(t *testing.T) {
	_, stakingKeeper, _, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = teststaking.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])

		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)

	// test identical,
	//  tendermintUpdate set: {} -> {}
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)
}

func TestApplyAndReturnValidatorSetUpdatesSingleValueChange(t *testing.T) {
	_, stakingKeeper, _, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = teststaking.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])

		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)

	// test single value change
	//  tendermintUpdate set: {} -> {c1'}
	validators[0].Status = types.Bonded
	validators[0].Tokens = stakingKeeper.TokensFromConsensusPower(ctx, 600)
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)

	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, 1)
	require.Equal(t, validators[0].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesMultipleValueChange(t *testing.T) {
	powers := []int64{10, 20}
	// TODO: use it in other places
	_, stakingKeeper, _, ctx, _, _, validators := initValidators(t, 1000, 20, powers)

	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := stakingKeeper.TokensFromConsensusPower(ctx, 190)
	delTokens2 := stakingKeeper.TokensFromConsensusPower(ctx, 80)
	validators[0], _ = validators[0].AddTokensFromDel(delTokens1)
	validators[1], _ = validators[1].AddTokensFromDel(delTokens2)
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)

	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)
	require.Equal(t, validators[0].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesInserted(t *testing.T) {
	powers := []int64{10, 20, 5, 15, 25}
	_, stakingKeeper, _, ctx, _, _, validators := initValidators(t, 1000, 20, powers)

	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	stakingKeeper.SetValidator(ctx, validators[2])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, 1)
	validators[2], _ = stakingKeeper.GetValidator(ctx, validators[2].GetOperator())
	require.Equal(t, validators[2].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	stakingKeeper.SetValidator(ctx, validators[3])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[3])
	updates = applyValidatorSetUpdates(t, ctx, stakingKeeper, 1)
	validators[3], _ = stakingKeeper.GetValidator(ctx, validators[3].GetOperator())
	require.Equal(t, validators[3].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])

	// test validtor added at the end
	//  tendermintUpdate set: {} -> {c0}
	stakingKeeper.SetValidator(ctx, validators[4])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[4])
	updates = applyValidatorSetUpdates(t, ctx, stakingKeeper, 1)
	validators[4], _ = stakingKeeper.GetValidator(ctx, validators[4].GetOperator())
	require.Equal(t, validators[4].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesWithCliffValidator(t *testing.T) {
	_, stakingKeeper, _, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	params := types.DefaultParams()
	params.MaxValidators = 2
	stakingKeeper.SetParams(ctx, params)

	powers := []int64{10, 20, 5}
	var validators [5]types.Validator
	for i, power := range powers {
		validators[i] = teststaking.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)

	// test validator added at the end but not inserted in the valset
	//  tendermintUpdate set: {} -> {}
	keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[2], false)
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)

	// test validator change its power and become a gotValidator (pushing out an existing)
	//  tendermintUpdate set: {}     -> {c0, c4}
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)

	tokens := stakingKeeper.TokensFromConsensusPower(ctx, 10)
	validators[2], _ = validators[2].AddTokensFromDel(tokens)
	stakingKeeper.SetValidator(ctx, validators[2])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)
	validators[2], _ = stakingKeeper.GetValidator(ctx, validators[2].GetOperator())
	require.Equal(t, validators[0].ABCIValidatorUpdateZero(), updates[1])
	require.Equal(t, validators[2].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesPowerDecrease(t *testing.T) {
	_, stakingKeeper, _, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{100, 100}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = teststaking.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)

	// check initial power
	require.Equal(t, int64(100), validators[0].GetConsensusPower(stakingKeeper.PowerReduction(ctx)))
	require.Equal(t, int64(100), validators[1].GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := stakingKeeper.TokensFromConsensusPower(ctx, 20)
	delTokens2 := stakingKeeper.TokensFromConsensusPower(ctx, 30)
	validators[0], _ = validators[0].RemoveDelShares(sdk.NewDecFromInt(delTokens1))
	validators[1], _ = validators[1].RemoveDelShares(sdk.NewDecFromInt(delTokens2))
	validators[0] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(stakingKeeper, ctx, validators[1], false)

	// power has changed
	require.Equal(t, int64(80), validators[0].GetConsensusPower(stakingKeeper.PowerReduction(ctx)))
	require.Equal(t, int64(70), validators[1].GetConsensusPower(stakingKeeper.PowerReduction(ctx)))

	// Tendermint updates should reflect power change
	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)
	require.Equal(t, validators[0].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesNewValidator(t *testing.T) {
	_, stakingKeeper, _, ctx, _, _ := bootstrapValidatorTest(t, 1000, 20)
	params := stakingKeeper.GetParams(ctx)
	params.MaxValidators = uint32(3)

	stakingKeeper.SetParams(ctx, params)

	powers := []int64{100, 100}
	var validators [2]types.Validator

	// initialize some validators into the state
	for i, power := range powers {
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = teststaking.NewValidator(t, valAddr, valPubKey)
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		stakingKeeper.SetValidator(ctx, validators[i])
		stakingKeeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, len(validators))
	validators[0], _ = stakingKeeper.GetValidator(ctx, validators[0].GetOperator())
	validators[1], _ = stakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	require.Equal(t, validators[0].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[1])

	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)

	// update initial validator set
	for i, power := range powers {

		stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[i])
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		stakingKeeper.SetValidator(ctx, validators[i])
		stakingKeeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// add a new validator that goes from zero power, to non-zero power, back to
	// zero power
	valPubKey := PKs[len(validators)+1]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	amt := sdk.NewInt(100)

	validator := teststaking.NewValidator(t, valAddr, valPubKey)
	validator, _ = validator.AddTokensFromDel(amt)

	stakingKeeper.SetValidator(ctx, validator)

	validator, _ = validator.RemoveDelShares(sdk.NewDecFromInt(amt))
	stakingKeeper.SetValidator(ctx, validator)
	stakingKeeper.SetValidatorByPowerIndex(ctx, validator)

	// add a new validator that increases in power
	valPubKey = PKs[len(validators)+2]
	valAddr = sdk.ValAddress(valPubKey.Address().Bytes())

	validator = teststaking.NewValidator(t, valAddr, valPubKey)
	tokens := stakingKeeper.TokensFromConsensusPower(ctx, 500)
	validator, _ = validator.AddTokensFromDel(tokens)
	stakingKeeper.SetValidator(ctx, validator)
	stakingKeeper.SetValidatorByPowerIndex(ctx, validator)

	// verify initial Tendermint updates are correct
	updates = applyValidatorSetUpdates(t, ctx, stakingKeeper, len(validators)+1)
	validator, _ = stakingKeeper.GetValidator(ctx, validator.GetOperator())
	validators[0], _ = stakingKeeper.GetValidator(ctx, validators[0].GetOperator())
	validators[1], _ = stakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	require.Equal(t, validator.ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
	require.Equal(t, validators[0].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[1])
	require.Equal(t, validators[1].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[2])
}

func TestApplyAndReturnValidatorSetUpdatesBondTransition(t *testing.T) {
	_, stakingKeeper, _, ctx, _, _ := bootstrapValidatorTest(t, 1000, 20)
	params := stakingKeeper.GetParams(ctx)
	params.MaxValidators = uint32(2)

	stakingKeeper.SetParams(ctx, params)

	powers := []int64{100, 200, 300}
	var validators [3]types.Validator

	// initialize some validators into the state
	for i, power := range powers {
		moniker := fmt.Sprintf("%d", i)
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = newMonikerValidator(t, valAddr, valPubKey, moniker)
		tokens := stakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
		stakingKeeper.SetValidator(ctx, validators[i])
		stakingKeeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := applyValidatorSetUpdates(t, ctx, stakingKeeper, 2)
	validators[2], _ = stakingKeeper.GetValidator(ctx, validators[2].GetOperator())
	validators[1], _ = stakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	require.Equal(t, validators[2].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])
	require.Equal(t, validators[1].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[1])

	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)

	// delegate to validator with lowest power but not enough to bond
	ctx = ctx.WithBlockHeight(1)

	var found bool
	validators[0], found = stakingKeeper.GetValidator(ctx, validators[0].GetOperator())
	require.True(t, found)

	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	tokens := stakingKeeper.TokensFromConsensusPower(ctx, 1)
	validators[0], _ = validators[0].AddTokensFromDel(tokens)
	stakingKeeper.SetValidator(ctx, validators[0])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[0])

	// verify initial Tendermint updates are correct
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)

	// create a series of events that will bond and unbond the validator with
	// lowest power in a single block context (height)
	ctx = ctx.WithBlockHeight(2)

	validators[1], found = stakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	require.True(t, found)

	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	validators[0], _ = validators[0].RemoveDelShares(validators[0].DelegatorShares)
	stakingKeeper.SetValidator(ctx, validators[0])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[0])
	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)

	stakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	tokens = stakingKeeper.TokensFromConsensusPower(ctx, 250)
	validators[1], _ = validators[1].AddTokensFromDel(tokens)
	stakingKeeper.SetValidator(ctx, validators[1])
	stakingKeeper.SetValidatorByPowerIndex(ctx, validators[1])

	// verify initial Tendermint updates are correct
	updates = applyValidatorSetUpdates(t, ctx, stakingKeeper, 1)
	require.Equal(t, validators[1].ABCIValidatorUpdate(stakingKeeper.PowerReduction(ctx)), updates[0])

	applyValidatorSetUpdates(t, ctx, stakingKeeper, 0)
}

func TestUpdateValidatorCommission(t *testing.T) {
	_, stakingKeeper, _, ctx, _, addrVals := bootstrapValidatorTest(t, 1000, 20)
	ctx = ctx.WithBlockHeader(tmproto.Header{Time: time.Now().UTC()})

	// Set MinCommissionRate to 0.05
	params := stakingKeeper.GetParams(ctx)
	params.MinCommissionRate = sdk.NewDecWithPrec(5, 2)
	stakingKeeper.SetParams(ctx, params)

	commission1 := types.NewCommissionWithTime(
		sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(3, 1),
		sdk.NewDecWithPrec(1, 1), time.Now().UTC().Add(time.Duration(-1)*time.Hour),
	)
	commission2 := types.NewCommission(sdk.NewDecWithPrec(1, 1), sdk.NewDecWithPrec(3, 1), sdk.NewDecWithPrec(1, 1))

	val1 := teststaking.NewValidator(t, addrVals[0], PKs[0])
	val2 := teststaking.NewValidator(t, addrVals[1], PKs[1])

	val1, _ = val1.SetInitialCommission(commission1)
	val2, _ = val2.SetInitialCommission(commission2)

	stakingKeeper.SetValidator(ctx, val1)
	stakingKeeper.SetValidator(ctx, val2)

	testCases := []struct {
		validator   types.Validator
		newRate     sdk.Dec
		expectedErr bool
	}{
		{val1, sdk.ZeroDec(), true},
		{val2, sdk.NewDecWithPrec(-1, 1), true},
		{val2, sdk.NewDecWithPrec(4, 1), true},
		{val2, sdk.NewDecWithPrec(3, 1), true},
		{val2, sdk.NewDecWithPrec(1, 2), true},
		{val2, sdk.NewDecWithPrec(2, 1), false},
	}

	for i, tc := range testCases {
		commission, err := stakingKeeper.UpdateValidatorCommission(ctx, tc.validator, tc.newRate)

		if tc.expectedErr {
			require.Error(t, err, "expected error for test case #%d with rate: %s", i, tc.newRate)
		} else {
			tc.validator.Commission = commission
			stakingKeeper.SetValidator(ctx, tc.validator)
			val, found := stakingKeeper.GetValidator(ctx, tc.validator.GetOperator())

			require.True(t, found,
				"expected to find validator for test case #%d with rate: %s", i, tc.newRate,
			)
			require.NoError(t, err,
				"unexpected error for test case #%d with rate: %s", i, tc.newRate,
			)
			require.Equal(t, tc.newRate, val.Commission.Rate,
				"expected new validator commission rate for test case #%d with rate: %s", i, tc.newRate,
			)
			require.Equal(t, ctx.BlockHeader().Time, val.Commission.UpdateTime,
				"expected new validator commission update time for test case #%d with rate: %s", i, tc.newRate,
			)
		}
	}
}

func applyValidatorSetUpdates(t *testing.T, ctx sdk.Context, k *keeper.Keeper, expectedUpdatesLen int) []abci.ValidatorUpdate {
	updates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	if expectedUpdatesLen >= 0 {
		require.Equal(t, expectedUpdatesLen, len(updates), "%v", updates)
	}
	return updates
}
