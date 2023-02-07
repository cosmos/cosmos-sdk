package keeper_test

import (
	"fmt"
	"testing"

	"cosmossdk.io/math"
	"cosmossdk.io/simapp"
	abci "github.com/cometbft/cometbft/abci/types"
	"gotest.tools/v3/assert"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func newMonikerValidator(t testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey, moniker string) types.Validator {
	v, err := types.NewValidator(operator, pubKey, types.Description{Moniker: moniker})
	assert.NilError(t, err)
	return v
}

func bootstrapValidatorTest(t testing.TB, power int64, numAddrs int) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	_, app, ctx := createTestInput(&testing.T{})

	addrDels, addrVals := generateAddresses(app, ctx, numAddrs)

	amt := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	// set bonded pool supply
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), totalSupply))

	// unbond genesis validator delegations
	delegations := app.StakingKeeper.GetAllDelegations(ctx)
	assert.Assert(t, len(delegations) == 1)
	delegation := delegations[0]

	_, _, err := app.StakingKeeper.Undelegate(ctx, delegation.GetDelegatorAddr(), delegation.GetValidatorAddr(), delegation.Shares)
	assert.NilError(t, err)

	// end block to unbond genesis validator
	staking.EndBlocker(ctx, app.StakingKeeper)

	return app, ctx, addrDels, addrVals
}

func initValidators(t testing.TB, power int64, numAddrs int, powers []int64) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress, []types.Validator) {
	app, ctx, addrs, valAddrs := bootstrapValidatorTest(t, power, numAddrs)
	pks := simtestutil.CreateTestPubKeys(numAddrs)

	vs := make([]types.Validator, len(powers))
	for i, power := range powers {
		vs[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), pks[i])
		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		vs[i], _ = vs[i].AddTokensFromDel(tokens)
	}
	return app, ctx, addrs, valAddrs, vs
}

func TestUpdateBondedValidatorsDecreaseCliff(t *testing.T) {
	numVals := 10
	maxVals := 5

	// create context, keeper, and pool for tests
	app, ctx, _, valAddrs := bootstrapValidatorTest(t, 0, 100)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	// create keeper parameters
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = uint32(maxVals)
	app.StakingKeeper.SetParams(ctx, params)

	// create a random pool
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), app.StakingKeeper.TokensFromConsensusPower(ctx, 1234)))))
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), app.StakingKeeper.TokensFromConsensusPower(ctx, 10000)))))

	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	validators := make([]types.Validator, numVals)
	for i := 0; i < len(validators); i++ {
		moniker := fmt.Sprintf("val#%d", int64(i))
		val := newMonikerValidator(t, valAddrs[i], PKs[i], moniker)
		delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, int64((i+1)*10))
		val, _ = val.AddTokensFromDel(delTokens)

		val = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, val, true)
		validators[i] = val
	}

	nextCliffVal := validators[numVals-maxVals+1]

	// remove enough tokens to kick out the validator below the current cliff
	// validator and next in line cliff validator
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, nextCliffVal)
	shares := app.StakingKeeper.TokensFromConsensusPower(ctx, 21)
	nextCliffVal, _ = nextCliffVal.RemoveDelShares(sdk.NewDecFromInt(shares))
	_ = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, nextCliffVal, true)

	expectedValStatus := map[int]types.BondStatus{
		9: types.Bonded, 8: types.Bonded, 7: types.Bonded, 5: types.Bonded, 4: types.Bonded,
		0: types.Unbonding, 1: types.Unbonding, 2: types.Unbonding, 3: types.Unbonding, 6: types.Unbonding,
	}

	// require all the validators have their respective statuses
	for valIdx, status := range expectedValStatus {
		valAddr := validators[valIdx].OperatorAddress
		addr, err := sdk.ValAddressFromBech32(valAddr)
		assert.NilError(t, err)
		val, _ := app.StakingKeeper.GetValidator(ctx, addr)

		assert.Equal(
			t, status, val.GetStatus(),
			fmt.Sprintf("expected validator at index %v to have status: %s", valIdx, status),
		)
	}
}

func TestSlashToZeroPowerRemoved(t *testing.T) {
	// initialize setup
	app, ctx, _, addrVals := bootstrapValidatorTest(t, 100, 20)

	// add a validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])
	valTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 100)

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)

	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), valTokens))))

	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)

	validator, _ = validator.AddTokensFromDel(valTokens)
	assert.Equal(t, types.Unbonded, validator.Status)
	assert.DeepEqual(t, valTokens, validator.Tokens)
	app.StakingKeeper.SetValidatorByConsAddr(ctx, validator)
	validator = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validator, true)
	assert.DeepEqual(t, valTokens, validator.Tokens)

	// slash the validator by 100%
	app.StakingKeeper.Slash(ctx, sdk.ConsAddress(PKs[0].Address()), 0, 100, math.LegacyOneDec())
	// apply TM updates
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, -1)
	// validator should be unbonding
	validator, _ = app.StakingKeeper.GetValidator(ctx, addrVals[0])
	assert.Equal(t, validator.GetStatus(), types.Unbonding)
}

// test how the validators are sorted, tests GetBondedValidatorsByPower
func TestGetValidatorSortingUnmixed(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// initialize some validators into the state
	amts := []math.Int{
		sdk.NewIntFromUint64(0),
		app.StakingKeeper.PowerReduction(ctx).MulRaw(100),
		app.StakingKeeper.PowerReduction(ctx),
		app.StakingKeeper.PowerReduction(ctx).MulRaw(400),
		app.StakingKeeper.PowerReduction(ctx).MulRaw(200),
	}
	n := len(amts)
	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		validators[i].Status = types.Bonded
		validators[i].Tokens = amt
		validators[i].DelegatorShares = sdk.NewDecFromInt(amt)
		keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[i], true)
	}

	// first make sure everything made it in to the gotValidator group
	resValidators := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, n, len(resValidators))
	assert.DeepEqual(t, sdk.NewInt(400).Mul(app.StakingKeeper.PowerReduction(ctx)), resValidators[0].BondedTokens())
	assert.DeepEqual(t, sdk.NewInt(200).Mul(app.StakingKeeper.PowerReduction(ctx)), resValidators[1].BondedTokens())
	assert.DeepEqual(t, sdk.NewInt(100).Mul(app.StakingKeeper.PowerReduction(ctx)), resValidators[2].BondedTokens())
	assert.DeepEqual(t, sdk.NewInt(1).Mul(app.StakingKeeper.PowerReduction(ctx)), resValidators[3].BondedTokens())
	assert.DeepEqual(t, sdk.NewInt(0), resValidators[4].BondedTokens())
	assert.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[1].OperatorAddress, resValidators[2].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[2].OperatorAddress, resValidators[3].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[0].OperatorAddress, resValidators[4].OperatorAddress, "%v", resValidators)

	// test a basic increase in voting power
	validators[3].Tokens = sdk.NewInt(500).Mul(app.StakingKeeper.PowerReduction(ctx))
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, len(resValidators), n)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))

	// test a decrease in voting power
	validators[3].Tokens = sdk.NewInt(300).Mul(app.StakingKeeper.PowerReduction(ctx))
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, len(resValidators), n)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))
	assert.Assert(ValEq(t, validators[4], resValidators[1]))

	// test equal voting power, different age
	validators[3].Tokens = sdk.NewInt(200).Mul(app.StakingKeeper.PowerReduction(ctx))
	ctx = ctx.WithBlockHeight(10)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, len(resValidators), n)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))
	assert.Assert(ValEq(t, validators[4], resValidators[1]))

	// no change in voting power - no change in sort
	ctx = ctx.WithBlockHeight(20)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[4], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, len(resValidators), n)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))
	assert.Assert(ValEq(t, validators[4], resValidators[1]))

	// change in voting power of both validators, both still in v-set, no age change
	validators[3].Tokens = sdk.NewInt(300).Mul(app.StakingKeeper.PowerReduction(ctx))
	validators[4].Tokens = sdk.NewInt(300).Mul(app.StakingKeeper.PowerReduction(ctx))
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, len(resValidators), n)
	ctx = ctx.WithBlockHeight(30)
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[4], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, len(resValidators), n, "%v", resValidators)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))
	assert.Assert(ValEq(t, validators[4], resValidators[1]))
}

func TestGetValidatorSortingMixed(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), app.StakingKeeper.TokensFromConsensusPower(ctx, 501)))))
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), app.StakingKeeper.TokensFromConsensusPower(ctx, 0)))))

	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)
	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)

	// now 2 max resValidators
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = 2
	app.StakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	amts := []math.Int{
		sdk.NewIntFromUint64(0),
		app.StakingKeeper.PowerReduction(ctx).MulRaw(100),
		app.StakingKeeper.PowerReduction(ctx),
		app.StakingKeeper.PowerReduction(ctx).MulRaw(400),
		app.StakingKeeper.PowerReduction(ctx).MulRaw(200),
	}

	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		validators[i].DelegatorShares = sdk.NewDecFromInt(amt)
		validators[i].Status = types.Bonded
		validators[i].Tokens = amt
		keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[i], true)
	}

	val0, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[0]))
	assert.Assert(t, found)
	val1, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[1]))
	assert.Assert(t, found)
	val2, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[2]))
	assert.Assert(t, found)
	val3, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[3]))
	assert.Assert(t, found)
	val4, found := app.StakingKeeper.GetValidator(ctx, sdk.ValAddress(addrs[4]))
	assert.Assert(t, found)
	assert.Equal(t, types.Bonded, val0.Status)
	assert.Equal(t, types.Unbonding, val1.Status)
	assert.Equal(t, types.Unbonding, val2.Status)
	assert.Equal(t, types.Bonded, val3.Status)
	assert.Equal(t, types.Bonded, val4.Status)

	// first make sure everything made it in to the gotValidator group
	resValidators := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	// The validators returned should match the max validators
	assert.Equal(t, 2, len(resValidators))
	assert.DeepEqual(t, sdk.NewInt(400).Mul(app.StakingKeeper.PowerReduction(ctx)), resValidators[0].BondedTokens())
	assert.DeepEqual(t, sdk.NewInt(200).Mul(app.StakingKeeper.PowerReduction(ctx)), resValidators[1].BondedTokens())
	assert.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
}

// TODO separate out into multiple tests
func TestGetValidatorsEdgeCases(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// set max validators to 2
	params := app.StakingKeeper.GetParams(ctx)
	nMax := uint32(2)
	params.MaxValidators = nMax
	app.StakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400}
	var validators [4]types.Validator
	for i, power := range powers {
		moniker := fmt.Sprintf("val#%d", int64(i))
		validators[i] = newMonikerValidator(t, sdk.ValAddress(addrs[i]), PKs[i], moniker)

		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
		assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, tokens))))
		app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)
		validators[i] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[i], true)
	}

	// ensure that the first two bonded validators are the largest validators
	resValidators := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[2], resValidators[0]))
	assert.Assert(ValEq(t, validators[3], resValidators[1]))

	// delegate 500 tokens to validator 0
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 500)
	validators[0], _ = validators[0].AddTokensFromDel(delTokens)
	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)

	newTokens := sdk.NewCoins()

	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), newTokens))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	// test that the two largest validators are
	//   a) validator 0 with 500 tokens
	//   b) validator 2 with 400 tokens (delegated before validator 3)
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))

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
	validators[3], found = app.StakingKeeper.GetValidator(ctx, validators[3].GetOperator())
	assert.Assert(t, found)
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = validators[3].AddTokensFromDel(app.StakingKeeper.TokensFromConsensusPower(ctx, 1))

	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)
	newTokens = sdk.NewCoins(sdk.NewCoin(params.BondDenom, app.StakingKeeper.TokensFromConsensusPower(ctx, 1)))
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), newTokens))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	validators[3] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[3], resValidators[1]))

	// validator 3 kicked out temporarily
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	rmTokens := validators[3].TokensFromShares(math.LegacyNewDec(201)).TruncateInt()
	validators[3], _ = validators[3].RemoveDelShares(math.LegacyNewDec(201))

	bondedPool := app.StakingKeeper.GetBondedPool(ctx)
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, rmTokens))))
	app.AccountKeeper.SetModuleAccount(ctx, bondedPool)

	validators[3] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))

	// validator 3 does not get spot back
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[3])
	validators[3], _ = validators[3].AddTokensFromDel(sdk.NewInt(200))

	notBondedPool = app.StakingKeeper.GetNotBondedPool(ctx)
	assert.NilError(t, banktestutil.FundModuleAccount(app.BankKeeper, ctx, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, sdk.NewInt(200)))))
	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)

	validators[3] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[3], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))
	_, exists := app.StakingKeeper.GetValidator(ctx, validators[3].GetOperator())
	assert.Assert(t, exists)
}

func TestValidatorBondHeight(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// now 2 max resValidators
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = 2
	app.StakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	var validators [3]types.Validator
	validators[0] = testutil.NewValidator(t, sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0])
	validators[1] = testutil.NewValidator(t, sdk.ValAddress(addrs[1]), PKs[1])
	validators[2] = testutil.NewValidator(t, sdk.ValAddress(addrs[2]), PKs[2])

	tokens0 := app.StakingKeeper.TokensFromConsensusPower(ctx, 200)
	tokens1 := app.StakingKeeper.TokensFromConsensusPower(ctx, 100)
	tokens2 := app.StakingKeeper.TokensFromConsensusPower(ctx, 100)
	validators[0], _ = validators[0].AddTokensFromDel(tokens0)
	validators[1], _ = validators[1].AddTokensFromDel(tokens1)
	validators[2], _ = validators[2].AddTokensFromDel(tokens2)

	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)

	////////////////////////////////////////
	// If two validators both increase to the same voting power in the same block,
	// the one with the first transaction should become bonded
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[2], true)

	resValidators := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, uint32(len(resValidators)), params.MaxValidators)

	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[1], resValidators[1]))
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[2])
	delTokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 50)
	validators[1], _ = validators[1].AddTokensFromDel(delTokens)
	validators[2], _ = validators[2].AddTokensFromDel(delTokens)
	validators[2] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[2], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, params.MaxValidators, uint32(len(resValidators)))
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], true)
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))
}

func TestFullValidatorSetPowerChange(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	params := app.StakingKeeper.GetParams(ctx)
	max := 2
	params.MaxValidators = uint32(2)
	app.StakingKeeper.SetParams(ctx, params)

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400, 200}
	var validators [5]types.Validator
	for i, power := range powers {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
		keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[i], true)
	}
	for i := range powers {
		var found bool
		validators[i], found = app.StakingKeeper.GetValidator(ctx, validators[i].GetOperator())
		assert.Assert(t, found)
	}
	assert.Equal(t, types.Unbonded, validators[0].Status)
	assert.Equal(t, types.Unbonding, validators[1].Status)
	assert.Equal(t, types.Bonded, validators[2].Status)
	assert.Equal(t, types.Bonded, validators[3].Status)
	assert.Equal(t, types.Unbonded, validators[4].Status)
	resValidators := app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, max, len(resValidators))
	assert.Assert(ValEq(t, validators[2], resValidators[0])) // in the order of txs
	assert.Assert(ValEq(t, validators[3], resValidators[1]))

	// test a swap in voting power

	tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 600)
	validators[0], _ = validators[0].AddTokensFromDel(tokens)
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], true)
	resValidators = app.StakingKeeper.GetBondedValidatorsByPower(ctx)
	assert.Equal(t, max, len(resValidators))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))
}

func TestApplyAndReturnValidatorSetUpdatesAllNone(t *testing.T) {
	app, ctx, _, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = testutil.NewValidator(t, valAddr, valPubKey)
		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}

	// test from nothing to something
	//  tendermintUpdate set: {} -> {c1, c3}
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)
	app.StakingKeeper.SetValidator(ctx, validators[0])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[0])
	app.StakingKeeper.SetValidator(ctx, validators[1])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[1])

	updates := applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)
	validators[0], _ = app.StakingKeeper.GetValidator(ctx, validators[0].GetOperator())
	validators[1], _ = app.StakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	assert.DeepEqual(t, validators[0].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[1])
	assert.DeepEqual(t, validators[1].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesIdentical(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])

		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)

	// test identical,
	//  tendermintUpdate set: {} -> {}
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)
}

func TestApplyAndReturnValidatorSetUpdatesSingleValueChange(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])

		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)

	// test single value change
	//  tendermintUpdate set: {} -> {c1'}
	validators[0].Status = types.Bonded
	validators[0].Tokens = app.StakingKeeper.TokensFromConsensusPower(ctx, 600)
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], false)

	updates := applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)
	assert.DeepEqual(t, validators[0].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesMultipleValueChange(t *testing.T) {
	powers := []int64{10, 20}
	// TODO: use it in other places
	app, ctx, _, _, validators := initValidators(t, 1000, 20, powers)

	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := app.StakingKeeper.TokensFromConsensusPower(ctx, 190)
	delTokens2 := app.StakingKeeper.TokensFromConsensusPower(ctx, 80)
	validators[0], _ = validators[0].AddTokensFromDel(delTokens1)
	validators[1], _ = validators[1].AddTokensFromDel(delTokens2)
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], false)

	updates := applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)
	assert.DeepEqual(t, validators[0].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])
	assert.DeepEqual(t, validators[1].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesInserted(t *testing.T) {
	powers := []int64{10, 20, 5, 15, 25}
	app, ctx, _, _, validators := initValidators(t, 1000, 20, powers)

	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	app.StakingKeeper.SetValidator(ctx, validators[2])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates := applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)
	validators[2], _ = app.StakingKeeper.GetValidator(ctx, validators[2].GetOperator())
	assert.DeepEqual(t, validators[2].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])

	// test validtor added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	app.StakingKeeper.SetValidator(ctx, validators[3])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[3])
	updates = applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)
	validators[3], _ = app.StakingKeeper.GetValidator(ctx, validators[3].GetOperator())
	assert.DeepEqual(t, validators[3].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])

	// test validtor added at the end
	//  tendermintUpdate set: {} -> {c0}
	app.StakingKeeper.SetValidator(ctx, validators[4])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[4])
	updates = applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)
	validators[4], _ = app.StakingKeeper.GetValidator(ctx, validators[4].GetOperator())
	assert.DeepEqual(t, validators[4].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesWithCliffValidator(t *testing.T) {
	app, ctx, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	params := types.DefaultParams()
	params.MaxValidators = 2
	app.StakingKeeper.SetParams(ctx, params)

	powers := []int64{10, 20, 5}
	var validators [5]types.Validator
	for i, power := range powers {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}
	validators[0] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[1], false)
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)

	// test validator added at the end but not inserted in the valset
	//  tendermintUpdate set: {} -> {}
	keeper.TestingUpdateValidator(app.StakingKeeper, ctx, validators[2], false)
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)

	// test validator change its power and become a gotValidator (pushing out an existing)
	//  tendermintUpdate set: {}     -> {c0, c4}
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)

	tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 10)
	validators[2], _ = validators[2].AddTokensFromDel(tokens)
	app.StakingKeeper.SetValidator(ctx, validators[2])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[2])
	updates := applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)
	validators[2], _ = app.StakingKeeper.GetValidator(ctx, validators[2].GetOperator())
	assert.DeepEqual(t, validators[0].ABCIValidatorUpdateZero(), updates[1])
	assert.DeepEqual(t, validators[2].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesNewValidator(t *testing.T) {
	app, ctx, _, _ := bootstrapValidatorTest(t, 1000, 20)
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = uint32(3)

	app.StakingKeeper.SetParams(ctx, params)

	powers := []int64{100, 100}
	var validators [2]types.Validator

	// initialize some validators into the state
	for i, power := range powers {
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = testutil.NewValidator(t, valAddr, valPubKey)
		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		app.StakingKeeper.SetValidator(ctx, validators[i])
		app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := applyValidatorSetUpdates(t, ctx, app.StakingKeeper, len(validators))
	validators[0], _ = app.StakingKeeper.GetValidator(ctx, validators[0].GetOperator())
	validators[1], _ = app.StakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	assert.DeepEqual(t, validators[0].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])
	assert.DeepEqual(t, validators[1].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[1])

	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)

	// update initial validator set
	for i, power := range powers {

		app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[i])
		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		app.StakingKeeper.SetValidator(ctx, validators[i])
		app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// add a new validator that goes from zero power, to non-zero power, back to
	// zero power
	valPubKey := PKs[len(validators)+1]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	amt := sdk.NewInt(100)

	validator := testutil.NewValidator(t, valAddr, valPubKey)
	validator, _ = validator.AddTokensFromDel(amt)

	app.StakingKeeper.SetValidator(ctx, validator)

	validator, _ = validator.RemoveDelShares(sdk.NewDecFromInt(amt))
	app.StakingKeeper.SetValidator(ctx, validator)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validator)

	// add a new validator that increases in power
	valPubKey = PKs[len(validators)+2]
	valAddr = sdk.ValAddress(valPubKey.Address().Bytes())

	validator = testutil.NewValidator(t, valAddr, valPubKey)
	tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 500)
	validator, _ = validator.AddTokensFromDel(tokens)
	app.StakingKeeper.SetValidator(ctx, validator)
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validator)

	// verify initial Tendermint updates are correct
	updates = applyValidatorSetUpdates(t, ctx, app.StakingKeeper, len(validators)+1)
	validator, _ = app.StakingKeeper.GetValidator(ctx, validator.GetOperator())
	validators[0], _ = app.StakingKeeper.GetValidator(ctx, validators[0].GetOperator())
	validators[1], _ = app.StakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	assert.DeepEqual(t, validator.ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])
	assert.DeepEqual(t, validators[0].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[1])
	assert.DeepEqual(t, validators[1].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[2])
}

func TestApplyAndReturnValidatorSetUpdatesBondTransition(t *testing.T) {
	app, ctx, _, _ := bootstrapValidatorTest(t, 1000, 20)
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = uint32(2)

	app.StakingKeeper.SetParams(ctx, params)

	powers := []int64{100, 200, 300}
	var validators [3]types.Validator

	// initialize some validators into the state
	for i, power := range powers {
		moniker := fmt.Sprintf("%d", i)
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = newMonikerValidator(t, valAddr, valPubKey, moniker)
		tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
		app.StakingKeeper.SetValidator(ctx, validators[i])
		app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[i])
	}

	// verify initial Tendermint updates are correct
	updates := applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 2)
	validators[2], _ = app.StakingKeeper.GetValidator(ctx, validators[2].GetOperator())
	validators[1], _ = app.StakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	assert.DeepEqual(t, validators[2].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])
	assert.DeepEqual(t, validators[1].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[1])

	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)

	// delegate to validator with lowest power but not enough to bond
	ctx = ctx.WithBlockHeight(1)

	var found bool
	validators[0], found = app.StakingKeeper.GetValidator(ctx, validators[0].GetOperator())
	assert.Assert(t, found)

	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	tokens := app.StakingKeeper.TokensFromConsensusPower(ctx, 1)
	validators[0], _ = validators[0].AddTokensFromDel(tokens)
	app.StakingKeeper.SetValidator(ctx, validators[0])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[0])

	// verify initial Tendermint updates are correct
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)

	// create a series of events that will bond and unbond the validator with
	// lowest power in a single block context (height)
	ctx = ctx.WithBlockHeight(2)

	validators[1], found = app.StakingKeeper.GetValidator(ctx, validators[1].GetOperator())
	assert.Assert(t, found)

	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[0])
	validators[0], _ = validators[0].RemoveDelShares(validators[0].DelegatorShares)
	app.StakingKeeper.SetValidator(ctx, validators[0])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[0])
	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)

	app.StakingKeeper.DeleteValidatorByPowerIndex(ctx, validators[1])
	tokens = app.StakingKeeper.TokensFromConsensusPower(ctx, 250)
	validators[1], _ = validators[1].AddTokensFromDel(tokens)
	app.StakingKeeper.SetValidator(ctx, validators[1])
	app.StakingKeeper.SetValidatorByPowerIndex(ctx, validators[1])

	// verify initial Tendermint updates are correct
	updates = applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 1)
	assert.DeepEqual(t, validators[1].ABCIValidatorUpdate(app.StakingKeeper.PowerReduction(ctx)), updates[0])

	applyValidatorSetUpdates(t, ctx, app.StakingKeeper, 0)
}

func applyValidatorSetUpdates(t *testing.T, ctx sdk.Context, k *keeper.Keeper, expectedUpdatesLen int) []abci.ValidatorUpdate {
	updates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	assert.NilError(t, err)
	if expectedUpdatesLen >= 0 {
		assert.Equal(t, expectedUpdatesLen, len(updates), "%v", updates)
	}
	return updates
}
