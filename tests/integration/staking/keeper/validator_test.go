package keeper_test

import (
	"fmt"
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	banktestutil "cosmossdk.io/x/bank/testutil"
	"cosmossdk.io/x/staking/keeper"
	"cosmossdk.io/x/staking/testutil"
	"cosmossdk.io/x/staking/types"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

func newMonikerValidator(tb testing.TB, operator sdk.ValAddress, pubKey cryptotypes.PubKey, moniker string) types.Validator {
	tb.Helper()
	v, err := types.NewValidator(operator.String(), pubKey, types.Description{Moniker: moniker})
	assert.NilError(tb, err)
	return v
}

func bootstrapValidatorTest(tb testing.TB, power int64, numAddrs int) (*fixture, []sdk.AccAddress, []sdk.ValAddress) {
	tb.Helper()
	f := initFixture(tb)

	addrDels, addrVals := generateAddresses(f, numAddrs)

	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(tb, err)

	amt := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(bondDenom, amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	// set bonded pool supply
	f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)

	assert.NilError(tb, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), totalSupply))

	return f, addrDels, addrVals
}

func initValidators(tb testing.TB, power int64, numAddrs int, powers []int64) (*fixture, []sdk.AccAddress, []sdk.ValAddress, []types.Validator) {
	tb.Helper()
	f, addrs, valAddrs := bootstrapValidatorTest(tb, power, numAddrs)
	pks := simtestutil.CreateTestPubKeys(numAddrs)

	vs := make([]types.Validator, len(powers))
	for i, power := range powers {
		vs[i] = testutil.NewValidator(tb, sdk.ValAddress(addrs[i]), pks[i])
		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		vs[i], _ = vs[i].AddTokensFromDel(tokens)
	}
	return f, addrs, valAddrs, vs
}

func TestUpdateBondedValidatorsDecreaseCliff(t *testing.T) {
	numVals := 10
	maxVals := 5

	// create context, keeper, and pool for tests
	f, _, valAddrs := bootstrapValidatorTest(t, 1, 100)

	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	// create keeper parameters
	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)
	params.MaxValidators = uint32(maxVals)
	assert.NilError(t, f.stakingKeeper.Params.Set(f.sdkCtx, params))

	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	// create a random pool
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1234)))))
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 10000)))))

	f.accountKeeper.SetModuleAccount(f.sdkCtx, bondedPool)
	f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)

	validators := make([]types.Validator, numVals)
	for i := 0; i < len(validators); i++ {
		moniker := fmt.Sprintf("val#%d", int64(i))
		val := newMonikerValidator(t, valAddrs[i], PKs[i], moniker)
		delTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, int64((i+1)*10))
		val, _ = val.AddTokensFromDel(delTokens)

		val = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, val, true)
		validators[i] = val
	}

	nextCliffVal := validators[numVals-maxVals+1]

	// remove enough tokens to kick out the validator below the current cliff
	// validator and next in line cliff validator
	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, nextCliffVal))
	shares := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 21)
	nextCliffVal, _ = nextCliffVal.RemoveDelShares(math.LegacyNewDecFromInt(shares))
	_ = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, nextCliffVal, true)

	expectedValStatus := map[int]sdk.BondStatus{
		9: sdk.Bonded, 8: sdk.Bonded, 7: sdk.Bonded, 5: sdk.Bonded, 4: sdk.Bonded,
		0: sdk.Unbonding, 1: sdk.Unbonding, 2: sdk.Unbonding, 3: sdk.Unbonding, 6: sdk.Unbonding,
	}

	// require all the validators have their respective statuses
	for valIdx, status := range expectedValStatus {
		valAddr := validators[valIdx].OperatorAddress
		addr, err := sdk.ValAddressFromBech32(valAddr)
		assert.NilError(t, err)
		val, _ := f.stakingKeeper.GetValidator(f.sdkCtx, addr)

		assert.Equal(
			t, status, val.GetStatus(),
			fmt.Sprintf("expected validator at index %v to have status: %x", valIdx, status),
		)
	}
}

func TestSlashToZeroPowerRemoved(t *testing.T) {
	// initialize setup
	f, _, addrVals := bootstrapValidatorTest(t, 100, 20)

	// add a validator
	validator := testutil.NewValidator(t, addrVals[0], PKs[0])
	valTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 100)

	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, valTokens))))

	f.accountKeeper.SetModuleAccount(f.sdkCtx, bondedPool)

	validator, _ = validator.AddTokensFromDel(valTokens)
	assert.Equal(t, types.Unbonded, validator.Status)
	assert.DeepEqual(t, valTokens, validator.Tokens)
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.sdkCtx, validator))
	validator = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validator, true)
	assert.DeepEqual(t, valTokens, validator.Tokens)

	// slash the validator by 100%
	_, err = f.stakingKeeper.Slash(f.sdkCtx, sdk.ConsAddress(PKs[0].Address()), 0, 100, math.LegacyOneDec())
	assert.NilError(t, err)
	// apply TM updates
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, -1)
	// validator should be unbonding
	validator, _ = f.stakingKeeper.GetValidator(f.sdkCtx, addrVals[0])
	assert.Equal(t, validator.GetStatus(), sdk.Unbonding)
}

// test how the validators are sorted, tests GetBondedValidatorsByPower
func TestGetValidatorSortingUnmixed(t *testing.T) {
	f, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// initialize some validators into the state
	amts := []math.Int{
		math.NewIntFromUint64(0),
		f.stakingKeeper.PowerReduction(f.sdkCtx).MulRaw(100),
		f.stakingKeeper.PowerReduction(f.sdkCtx),
		f.stakingKeeper.PowerReduction(f.sdkCtx).MulRaw(400),
		f.stakingKeeper.PowerReduction(f.sdkCtx).MulRaw(200),
	}
	n := len(amts)
	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		validators[i].Status = types.Bonded
		validators[i].Tokens = amt
		validators[i].DelegatorShares = math.LegacyNewDecFromInt(amt)
		keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[i], true)
	}

	// first make sure everything made it in to the gotValidator group
	resValidators, err := f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, n, len(resValidators))
	assert.DeepEqual(t, math.NewInt(400).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx)), resValidators[0].BondedTokens())
	assert.DeepEqual(t, math.NewInt(200).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx)), resValidators[1].BondedTokens())
	assert.DeepEqual(t, math.NewInt(100).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx)), resValidators[2].BondedTokens())
	assert.DeepEqual(t, math.NewInt(1).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx)), resValidators[3].BondedTokens())
	assert.DeepEqual(t, math.NewInt(0), resValidators[4].BondedTokens())
	assert.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[1].OperatorAddress, resValidators[2].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[2].OperatorAddress, resValidators[3].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[0].OperatorAddress, resValidators[4].OperatorAddress, "%v", resValidators)

	// test a basic increase in voting power
	validators[3].Tokens = math.NewInt(500).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx))
	keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[3], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, len(resValidators), n)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))

	// test a decrease in voting power
	validators[3].Tokens = math.NewInt(300).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx))
	keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[3], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, len(resValidators), n)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))
	assert.Assert(ValEq(t, validators[4], resValidators[1]))

	// test equal voting power, different age
	validators[3].Tokens = math.NewInt(200).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx))
	f.sdkCtx = f.sdkCtx.WithBlockHeight(10)
	keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[3], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, len(resValidators), n)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))
	assert.Assert(ValEq(t, validators[4], resValidators[1]))

	// no change in voting power - no change in sort
	f.sdkCtx = f.sdkCtx.WithBlockHeight(20)
	keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[4], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, len(resValidators), n)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))
	assert.Assert(ValEq(t, validators[4], resValidators[1]))

	// change in voting power of both validators, both still in v-set, no age change
	validators[3].Tokens = math.NewInt(300).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx))
	validators[4].Tokens = math.NewInt(300).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx))
	keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[3], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, len(resValidators), n)
	f.sdkCtx = f.sdkCtx.WithBlockHeight(30)
	keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[4], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, len(resValidators), n, "%v", resValidators)
	assert.Assert(ValEq(t, validators[3], resValidators[0]))
	assert.Assert(ValEq(t, validators[4], resValidators[1]))
}

func TestGetValidatorSortingMixed(t *testing.T) {
	f, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 501)))))
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(bondDenom, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 0)))))

	f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)
	f.accountKeeper.SetModuleAccount(f.sdkCtx, bondedPool)

	// now 2 max resValidators
	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)
	params.MaxValidators = 2
	assert.NilError(t, f.stakingKeeper.Params.Set(f.sdkCtx, params))

	// initialize some validators into the state
	amts := []math.Int{
		math.NewIntFromUint64(0),
		f.stakingKeeper.PowerReduction(f.sdkCtx).MulRaw(100),
		f.stakingKeeper.PowerReduction(f.sdkCtx),
		f.stakingKeeper.PowerReduction(f.sdkCtx).MulRaw(400),
		f.stakingKeeper.PowerReduction(f.sdkCtx).MulRaw(200),
	}

	var validators [5]types.Validator
	for i, amt := range amts {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		validators[i].DelegatorShares = math.LegacyNewDecFromInt(amt)
		validators[i].Status = types.Bonded
		validators[i].Tokens = amt
		keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[i], true)
	}

	val0, found := f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[0]))
	assert.Assert(t, found)
	val1, found := f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[1]))
	assert.Assert(t, found)
	val2, found := f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[2]))
	assert.Assert(t, found)
	val3, found := f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[3]))
	assert.Assert(t, found)
	val4, found := f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[4]))
	assert.Assert(t, found)
	assert.Equal(t, types.Bonded, val0.Status)
	assert.Equal(t, types.Unbonding, val1.Status)
	assert.Equal(t, types.Unbonding, val2.Status)
	assert.Equal(t, types.Bonded, val3.Status)
	assert.Equal(t, types.Bonded, val4.Status)

	// first make sure everything made it in to the gotValidator group
	resValidators, err := f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	// The validators returned should match the max validators
	assert.Equal(t, 2, len(resValidators))
	assert.DeepEqual(t, math.NewInt(400).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx)), resValidators[0].BondedTokens())
	assert.DeepEqual(t, math.NewInt(200).Mul(f.stakingKeeper.PowerReduction(f.sdkCtx)), resValidators[1].BondedTokens())
	assert.Equal(t, validators[3].OperatorAddress, resValidators[0].OperatorAddress, "%v", resValidators)
	assert.Equal(t, validators[4].OperatorAddress, resValidators[1].OperatorAddress, "%v", resValidators)
}

// TODO separate out into multiple tests
func TestGetValidatorsEdgeCases(t *testing.T) {
	f, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// set max validators to 2
	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)
	nMax := uint32(2)
	params.MaxValidators = nMax
	assert.NilError(t, f.stakingKeeper.Params.Set(f.sdkCtx, params))
	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400}
	var validators [4]types.Validator
	for i, power := range powers {
		moniker := fmt.Sprintf("val#%d", int64(i))
		validators[i] = newMonikerValidator(t, sdk.ValAddress(addrs[i]), PKs[i], moniker)

		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)
		assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, tokens))))
		f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)
		validators[i] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[i], true)
	}

	// ensure that the first two bonded validators are the largest validators
	resValidators, err := f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[2], resValidators[0]))
	assert.Assert(ValEq(t, validators[3], resValidators[1]))

	// delegate 500 tokens to validator 0
	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[0]))
	delTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 500)
	validators[0], _ = validators[0].AddTokensFromDel(delTokens)
	notBondedPool := f.stakingKeeper.GetNotBondedPool(f.sdkCtx)

	newTokens := sdk.NewCoins()

	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), newTokens))
	f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)

	// test that the two largest validators are
	//   a) validator 0 with 500 tokens
	//   b) validator 2 with 400 tokens (delegated before validator 3)
	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
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
	f.sdkCtx = f.sdkCtx.WithBlockHeight(40)

	valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[3].GetOperator())
	assert.NilError(t, err)

	validators[3], err = f.stakingKeeper.GetValidator(f.sdkCtx, valbz)
	assert.NilError(t, err)
	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[3]))
	validators[3], _ = validators[3].AddTokensFromDel(f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1))

	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.sdkCtx)
	newTokens = sdk.NewCoins(sdk.NewCoin(params.BondDenom, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1)))
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), newTokens))
	f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)

	validators[3] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[3], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[3], resValidators[1]))

	// validator 3 kicked out temporarily
	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[3]))
	rmTokens := validators[3].TokensFromShares(math.LegacyNewDec(201)).TruncateInt()
	validators[3], _ = validators[3].RemoveDelShares(math.LegacyNewDec(201))

	bondedPool := f.stakingKeeper.GetBondedPool(f.sdkCtx)
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, bondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, rmTokens))))
	f.accountKeeper.SetModuleAccount(f.sdkCtx, bondedPool)

	validators[3] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[3], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))

	// validator 3 does not get spot back
	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[3]))
	validators[3], _ = validators[3].AddTokensFromDel(math.NewInt(200))

	notBondedPool = f.stakingKeeper.GetNotBondedPool(f.sdkCtx)
	assert.NilError(t, banktestutil.FundModuleAccount(f.sdkCtx, f.bankKeeper, notBondedPool.GetName(), sdk.NewCoins(sdk.NewCoin(params.BondDenom, math.NewInt(200)))))
	f.accountKeeper.SetModuleAccount(f.sdkCtx, notBondedPool)

	validators[3] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[3], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, nMax, uint32(len(resValidators)))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))
	_, exists := f.stakingKeeper.GetValidator(f.sdkCtx, valbz)
	assert.Assert(t, exists)
}

func TestValidatorBondHeight(t *testing.T) {
	f, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	// now 2 max resValidators
	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)
	params.MaxValidators = 2
	assert.NilError(t, f.stakingKeeper.Params.Set(f.sdkCtx, params))
	// initialize some validators into the state
	var validators [3]types.Validator
	validators[0] = testutil.NewValidator(t, sdk.ValAddress(PKs[0].Address().Bytes()), PKs[0])
	validators[1] = testutil.NewValidator(t, sdk.ValAddress(addrs[1]), PKs[1])
	validators[2] = testutil.NewValidator(t, sdk.ValAddress(addrs[2]), PKs[2])

	tokens0 := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 200)
	tokens1 := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 100)
	tokens2 := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 100)
	validators[0], _ = validators[0].AddTokensFromDel(tokens0)
	validators[1], _ = validators[1].AddTokensFromDel(tokens1)
	validators[2], _ = validators[2].AddTokensFromDel(tokens2)

	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], true)

	////////////////////////////////////////
	// If two validators both increase to the same voting power in the same block,
	// the one with the first transaction should become bonded
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], true)
	validators[2] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[2], true)

	resValidators, err := f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, uint32(len(resValidators)), params.MaxValidators)

	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[1], resValidators[1]))
	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[1]))
	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[2]))
	delTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 50)
	validators[1], _ = validators[1].AddTokensFromDel(delTokens)
	validators[2], _ = validators[2].AddTokensFromDel(delTokens)
	validators[2] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[2], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, params.MaxValidators, uint32(len(resValidators)))
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], true)
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))
}

func TestFullValidatorSetPowerChange(t *testing.T) {
	f, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)
	max := 2
	params.MaxValidators = uint32(2)
	assert.NilError(t, f.stakingKeeper.Params.Set(f.sdkCtx, params))

	// initialize some validators into the state
	powers := []int64{0, 100, 400, 400, 200}
	var validators [5]types.Validator
	for i, power := range powers {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
		keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[i], true)
	}
	for i := range powers {
		valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[i].GetOperator())
		assert.NilError(t, err)

		validators[i], err = f.stakingKeeper.GetValidator(f.sdkCtx, valbz)
		assert.NilError(t, err)
	}
	assert.Equal(t, types.Unbonded, validators[0].Status)
	assert.Equal(t, types.Unbonding, validators[1].Status)
	assert.Equal(t, types.Bonded, validators[2].Status)
	assert.Equal(t, types.Bonded, validators[3].Status)
	assert.Equal(t, types.Unbonded, validators[4].Status)
	resValidators, err := f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, max, len(resValidators))
	assert.Assert(ValEq(t, validators[2], resValidators[0])) // in the order of txs
	assert.Assert(ValEq(t, validators[3], resValidators[1]))

	// test a swap in voting power

	tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 600)
	validators[0], _ = validators[0].AddTokensFromDel(tokens)
	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], true)
	resValidators, err = f.stakingKeeper.GetBondedValidatorsByPower(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, max, len(resValidators))
	assert.Assert(ValEq(t, validators[0], resValidators[0]))
	assert.Assert(ValEq(t, validators[2], resValidators[1]))
}

func TestApplyAndReturnValidatorSetUpdatesAllNone(t *testing.T) {
	f, _, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = testutil.NewValidator(t, valAddr, valPubKey)
		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}

	// test from nothing to something
	//  tendermintUpdate set: {} -> {c1, c3}
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[0]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[0]))
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[1]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[1]))

	updates := applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)
	val0bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	assert.NilError(t, err)
	val1bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[1].GetOperator())
	assert.NilError(t, err)
	validators[0], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val0bz)
	validators[1], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val1bz)
	assert.DeepEqual(t, validators[0].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[1])
	assert.DeepEqual(t, validators[1].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesIdentical(t *testing.T) {
	f, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])

		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], false)
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)

	// test identical,
	//  tendermintUpdate set: {} -> {}
	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], false)
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)
}

func TestApplyAndReturnValidatorSetUpdatesSingleValueChange(t *testing.T) {
	f, addrs, _ := bootstrapValidatorTest(t, 1000, 20)

	powers := []int64{10, 20}
	var validators [2]types.Validator
	for i, power := range powers {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])

		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

	}
	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], false)
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)

	// test single value change
	//  tendermintUpdate set: {} -> {c1'}
	validators[0].Status = types.Bonded
	validators[0].Tokens = f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 600)
	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], false)

	updates := applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)
	assert.DeepEqual(t, validators[0].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesMultipleValueChange(t *testing.T) {
	powers := []int64{10, 20}
	// TODO: use it in other places
	f, _, _, validators := initValidators(t, 1000, 20, powers)

	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], false)
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)

	// test multiple value change
	//  tendermintUpdate set: {c1, c3} -> {c1', c3'}
	delTokens1 := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 190)
	delTokens2 := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 80)
	validators[0], _ = validators[0].AddTokensFromDel(delTokens1)
	validators[1], _ = validators[1].AddTokensFromDel(delTokens2)
	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], false)

	updates := applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)
	assert.DeepEqual(t, validators[0].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])
	assert.DeepEqual(t, validators[1].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[1])
}

func TestApplyAndReturnValidatorSetUpdatesInserted(t *testing.T) {
	powers := []int64{10, 20, 5, 15, 25}
	f, _, _, validators := initValidators(t, 1000, 20, powers)

	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], false)
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)

	// test validator added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[2]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[2]))
	updates := applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)
	val2bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[2].GetOperator())
	assert.NilError(t, err)
	validators[2], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val2bz)
	assert.DeepEqual(t, validators[2].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])

	// test validator added at the beginning
	//  tendermintUpdate set: {} -> {c0}
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[3]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[3]))
	updates = applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)
	val3bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[3].GetOperator())
	assert.NilError(t, err)
	validators[3], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val3bz)
	assert.DeepEqual(t, validators[3].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])

	// test validator added at the end
	//  tendermintUpdate set: {} -> {c0}
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[4]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[4]))
	updates = applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)
	val4bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[4].GetOperator())
	assert.NilError(t, err)
	validators[4], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val4bz)
	assert.DeepEqual(t, validators[4].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesWithCliffValidator(t *testing.T) {
	f, addrs, _ := bootstrapValidatorTest(t, 1000, 20)
	params := types.DefaultParams()
	params.MaxValidators = 2
	err := f.stakingKeeper.Params.Set(f.sdkCtx, params)
	assert.NilError(t, err)
	powers := []int64{10, 20, 5}
	var validators [5]types.Validator
	for i, power := range powers {
		validators[i] = testutil.NewValidator(t, sdk.ValAddress(addrs[i]), PKs[i])
		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
	}
	validators[0] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[0], false)
	validators[1] = keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[1], false)
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)

	// test validator added at the end but not inserted in the valset
	//  tendermintUpdate set: {} -> {}
	keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validators[2], false)
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)

	// test validator change its power and become a gotValidator (pushing out an existing)
	//  tendermintUpdate set: {}     -> {c0, c4}
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)

	tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 10)
	validators[2], _ = validators[2].AddTokensFromDel(tokens)
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[2]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[2]))
	updates := applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)
	val2bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[2].GetOperator())
	assert.NilError(t, err)
	validators[2], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val2bz)
	assert.DeepEqual(t, validators[0].ModuleValidatorUpdateZero(), updates[1])
	assert.DeepEqual(t, validators[2].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])
}

func TestApplyAndReturnValidatorSetUpdatesNewValidator(t *testing.T) {
	f, _, _ := bootstrapValidatorTest(t, 1000, 20)
	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)
	params.MaxValidators = uint32(3)

	assert.NilError(t, f.stakingKeeper.Params.Set(f.sdkCtx, params))

	powers := []int64{100, 100}
	var validators [2]types.Validator

	// initialize some validators into the state
	for i, power := range powers {
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = testutil.NewValidator(t, valAddr, valPubKey)
		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[i]))
		assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[i]))
	}

	// verify initial CometBFT updates are correct
	updates := applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, len(validators))

	val0bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	assert.NilError(t, err)
	val1bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[1].GetOperator())
	assert.NilError(t, err)
	validators[0], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val0bz)
	validators[1], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val1bz)
	assert.DeepEqual(t, validators[0].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])
	assert.DeepEqual(t, validators[1].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[1])

	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)

	// update initial validator set
	for i, power := range powers {

		assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[i]))
		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)

		assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[i]))
		assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[i]))
	}

	// add a new validator that goes from zero power, to non-zero power, back to
	// zero power
	valPubKey := PKs[len(validators)+1]
	valAddr := sdk.ValAddress(valPubKey.Address().Bytes())
	amt := math.NewInt(100)

	validator := testutil.NewValidator(t, valAddr, valPubKey)
	validator, _ = validator.AddTokensFromDel(amt)

	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validator))

	validator, _ = validator.RemoveDelShares(math.LegacyNewDecFromInt(amt))
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validator))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validator))

	// add a new validator that increases in power
	valPubKey = PKs[len(validators)+2]
	valAddr = sdk.ValAddress(valPubKey.Address().Bytes())

	validator = testutil.NewValidator(t, valAddr, valPubKey)
	tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 500)
	validator, _ = validator.AddTokensFromDel(tokens)
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validator))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validator))

	// verify initial CometBFT updates are correct
	updates = applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, len(validators)+1)
	valbz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validator.GetOperator())
	assert.NilError(t, err)
	validator, _ = f.stakingKeeper.GetValidator(f.sdkCtx, valbz)
	validators[0], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val0bz)
	validators[1], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val1bz)
	assert.DeepEqual(t, validator.ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])
	assert.DeepEqual(t, validators[0].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[1])
	assert.DeepEqual(t, validators[1].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[2])
}

func TestApplyAndReturnValidatorSetUpdatesBondTransition(t *testing.T) {
	f, _, _ := bootstrapValidatorTest(t, 1000, 20)
	params, err := f.stakingKeeper.Params.Get(f.sdkCtx)
	assert.NilError(t, err)
	params.MaxValidators = uint32(2)

	assert.NilError(t, f.stakingKeeper.Params.Set(f.sdkCtx, params))

	powers := []int64{100, 200, 300}
	var validators [3]types.Validator

	// initialize some validators into the state
	for i, power := range powers {
		moniker := fmt.Sprintf("%d", i)
		valPubKey := PKs[i+1]
		valAddr := sdk.ValAddress(valPubKey.Address().Bytes())

		validators[i] = newMonikerValidator(t, valAddr, valPubKey, moniker)
		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, power)
		validators[i], _ = validators[i].AddTokensFromDel(tokens)
		assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[i]))
		assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[i]))
	}

	// verify initial CometBFT updates are correct
	updates := applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 2)
	val1bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[1].GetOperator())
	assert.NilError(t, err)
	val2bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[2].GetOperator())
	assert.NilError(t, err)
	validators[2], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val2bz)
	validators[1], _ = f.stakingKeeper.GetValidator(f.sdkCtx, val1bz)
	assert.DeepEqual(t, validators[2].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])
	assert.DeepEqual(t, validators[1].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[1])

	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)

	// delegate to validator with lowest power but not enough to bond
	f.sdkCtx = f.sdkCtx.WithBlockHeight(1)

	val0bz, err := f.stakingKeeper.ValidatorAddressCodec().StringToBytes(validators[0].GetOperator())
	assert.NilError(t, err)
	validators[0], err = f.stakingKeeper.GetValidator(f.sdkCtx, val0bz)
	assert.NilError(t, err)

	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[0]))
	tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1)
	validators[0], _ = validators[0].AddTokensFromDel(tokens)
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[0]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[0]))

	// verify initial CometBFT updates are correct
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)

	// create a series of events that will bond and unbond the validator with
	// lowest power in a single block context (height)
	f.sdkCtx = f.sdkCtx.WithBlockHeight(2)

	validators[1], err = f.stakingKeeper.GetValidator(f.sdkCtx, val1bz)
	assert.NilError(t, err)

	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[0]))
	validators[0], _ = validators[0].RemoveDelShares(validators[0].DelegatorShares)
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[0]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[0]))
	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)

	assert.NilError(t, f.stakingKeeper.DeleteValidatorByPowerIndex(f.sdkCtx, validators[1]))
	tokens = f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 250)
	validators[1], _ = validators[1].AddTokensFromDel(tokens)
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, validators[1]))
	assert.NilError(t, f.stakingKeeper.SetValidatorByPowerIndex(f.sdkCtx, validators[1]))

	// verify initial CometBFT updates are correct
	updates = applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)
	assert.DeepEqual(t, validators[1].ModuleValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)), updates[0])

	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 0)
}

func applyValidatorSetUpdates(t *testing.T, ctx sdk.Context, k *keeper.Keeper, expectedUpdatesLen int) []module.ValidatorUpdate {
	t.Helper()
	updates, err := k.ApplyAndReturnValidatorSetUpdates(ctx)
	assert.NilError(t, err)
	if expectedUpdatesLen >= 0 {
		assert.Equal(t, expectedUpdatesLen, len(updates), "%v", updates)
	}
	return updates
}
