package staking

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/secp256k1"
	tmtypes "github.com/tendermint/tendermint/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

//______________________________________________________________________

// retrieve params which are instant
func setInstantUnbondPeriod(keeper keep.Keeper, ctx sdk.Context) types.Params {
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	keeper.SetParams(ctx, params)
	return params
}

//______________________________________________________________________

func TestValidatorByPowerIndex(t *testing.T) {
	validatorAddr, validatorAddr3 := sdk.ValAddress(keep.Addrs[0]), sdk.ValAddress(keep.Addrs[1])

	initPower := int64(1000000)
	initBond := sdk.TokensFromConsensusPower(initPower)
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, initPower)
	_ = setInstantUnbondPeriod(keeper, ctx)

	// create validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], initBond)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// must end-block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// verify the self-delegation exists
	bond, found := keeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	gotBond := bond.Shares.RoundInt()
	require.Equal(t, initBond, gotBond)

	// verify that the by power index exists
	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	power := GetValidatorsByPowerIndexKey(validator)
	require.True(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power))

	// create a second validator keep it bonded
	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr3, keep.PKs[2], initBond)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// must end-block
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// slash and jail the first validator
	consAddr0 := sdk.ConsAddress(keep.PKs[0].Address())
	keeper.Slash(ctx, consAddr0, 0, initPower, sdk.NewDecWithPrec(5, 1))
	keeper.Jail(ctx, consAddr0)
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator, found = keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, sdk.Unbonding, validator.Status)      // ensure is unbonding
	require.Equal(t, initBond.QuoRaw(2), validator.Tokens) // ensure tokens slashed
	keeper.Unjail(ctx, consAddr0)

	// the old power record should have been deleted as the power changed
	require.False(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power))

	// but the new power record should have been created
	validator, found = keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	power2 := GetValidatorsByPowerIndexKey(validator)
	require.True(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power2))

	// now the new record power index should be the same as the original record
	power3 := GetValidatorsByPowerIndexKey(validator)
	require.Equal(t, power2, power3)

	// unbond self-delegation
	totalBond := validator.TokensFromShares(bond.GetShares()).TruncateInt()
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, totalBond)
	msgUndelegate := NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)

	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	var finishTime time.Time
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)

	ctx = ctx.WithBlockTime(finishTime)
	EndBlocker(ctx, keeper)
	EndBlocker(ctx, keeper)

	// verify that by power key nolonger exists
	_, found = keeper.GetValidator(ctx, validatorAddr)
	require.False(t, found)
	require.False(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power3))
}

func TestDuplicatesMsgCreateValidator(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)

	addr1, addr2 := sdk.ValAddress(keep.Addrs[0]), sdk.ValAddress(keep.Addrs[1])
	pk1, pk2 := keep.PKs[0], keep.PKs[1]

	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator1 := NewTestMsgCreateValidator(addr1, pk1, valTokens)
	got := handleMsgCreateValidator(ctx, msgCreateValidator1, keeper)
	require.True(t, got.IsOK(), "%v", got)

	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator, found := keeper.GetValidator(ctx, addr1)
	require.True(t, found)
	assert.Equal(t, sdk.Bonded, validator.Status)
	assert.Equal(t, addr1, validator.OperatorAddress)
	assert.Equal(t, pk1, validator.ConsPubKey)
	assert.Equal(t, valTokens, validator.BondedTokens())
	assert.Equal(t, valTokens.ToDec(), validator.DelegatorShares)
	assert.Equal(t, Description{}, validator.Description)

	// two validators can't have the same operator address
	msgCreateValidator2 := NewTestMsgCreateValidator(addr1, pk2, valTokens)
	got = handleMsgCreateValidator(ctx, msgCreateValidator2, keeper)
	require.False(t, got.IsOK(), "%v", got)

	// two validators can't have the same pubkey
	msgCreateValidator3 := NewTestMsgCreateValidator(addr2, pk1, valTokens)
	got = handleMsgCreateValidator(ctx, msgCreateValidator3, keeper)
	require.False(t, got.IsOK(), "%v", got)

	// must have different pubkey and operator
	msgCreateValidator4 := NewTestMsgCreateValidator(addr2, pk2, valTokens)
	got = handleMsgCreateValidator(ctx, msgCreateValidator4, keeper)
	require.True(t, got.IsOK(), "%v", got)

	// must end-block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found = keeper.GetValidator(ctx, addr2)

	require.True(t, found)
	assert.Equal(t, sdk.Bonded, validator.Status)
	assert.Equal(t, addr2, validator.OperatorAddress)
	assert.Equal(t, pk2, validator.ConsPubKey)
	assert.True(sdk.IntEq(t, valTokens, validator.Tokens))
	assert.True(sdk.DecEq(t, valTokens.ToDec(), validator.DelegatorShares))
	assert.Equal(t, Description{}, validator.Description)
}

func TestInvalidPubKeyTypeMsgCreateValidator(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)

	addr := sdk.ValAddress(keep.Addrs[0])
	invalidPk := secp256k1.GenPrivKey().PubKey()

	// invalid pukKey type should not be allowed
	msgCreateValidator := NewTestMsgCreateValidator(addr, invalidPk, sdk.NewInt(10))
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.False(t, got.IsOK(), "%v", got)

	ctx = ctx.WithConsensusParams(&abci.ConsensusParams{
		Validator: &abci.ValidatorParams{PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeSecp256k1}},
	})

	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "%v", got)
}

func TestLegacyValidatorDelegations(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, int64(1000))
	setInstantUnbondPeriod(keeper, ctx)

	bondAmount := sdk.TokensFromConsensusPower(10)
	valAddr := sdk.ValAddress(keep.Addrs[0])
	valConsPubKey, valConsAddr := keep.PKs[0], sdk.ConsAddress(keep.PKs[0].Address())
	delAddr := keep.Addrs[1]

	// create validator
	msgCreateVal := NewTestMsgCreateValidator(valAddr, valConsPubKey, bondAmount)
	got := handleMsgCreateValidator(ctx, msgCreateVal, keeper)
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got %v", got)

	// must end-block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// verify the validator exists and has the correct attributes
	validator, found := keeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount, validator.BondedTokens())

	// delegate tokens to the validator
	msgDelegate := NewTestMsgDelegate(delAddr, valAddr, bondAmount)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	// verify validator bonded shares
	validator, found = keeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(2), validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(2), validator.BondedTokens())

	// unbond validator total self-delegations (which should jail the validator)
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, bondAmount)
	msgUndelegate := NewMsgUndelegate(sdk.AccAddress(valAddr), valAddr, unbondAmt)

	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected begin unbonding validator msg to be ok, got %v", got)

	var finishTime time.Time
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)
	ctx = ctx.WithBlockTime(finishTime)
	EndBlocker(ctx, keeper)

	// verify the validator record still exists, is jailed, and has correct tokens
	validator, found = keeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.True(t, validator.Jailed)
	require.Equal(t, bondAmount, validator.Tokens)

	// verify delegation still exists
	bond, found := keeper.GetDelegation(ctx, delAddr, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount, bond.Shares.RoundInt())
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())

	// verify the validator can still self-delegate
	msgSelfDelegate := NewTestMsgDelegate(sdk.AccAddress(valAddr), valAddr, bondAmount)
	got = handleMsgDelegate(ctx, msgSelfDelegate, keeper)
	require.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	// verify validator bonded shares
	validator, found = keeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(2), validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(2), validator.Tokens)

	// unjail the validator now that is has non-zero self-delegated shares
	keeper.Unjail(ctx, valConsAddr)

	// verify the validator can now accept delegations
	msgDelegate = NewTestMsgDelegate(delAddr, valAddr, bondAmount)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	// verify validator bonded shares
	validator, found = keeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(3), validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(3), validator.Tokens)

	// verify new delegation
	bond, found = keeper.GetDelegation(ctx, delAddr, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(2), bond.Shares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(3), validator.DelegatorShares.RoundInt())
}

func TestIncrementsMsgDelegate(t *testing.T) {
	initPower := int64(1000)
	initBond := sdk.TokensFromConsensusPower(initPower)
	ctx, accMapper, keeper, _ := keep.CreateTestInput(t, false, initPower)
	params := keeper.GetParams(ctx)

	bondAmount := sdk.TokensFromConsensusPower(10)
	validatorAddr, delegatorAddr := sdk.ValAddress(keep.Addrs[0]), keep.Addrs[1]

	// first create validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], bondAmount)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got %v", got)

	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, sdk.Bonded, validator.Status)
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount, validator.BondedTokens(), "validator: %v", validator)

	_, found = keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
	require.False(t, found)

	bond, found := keeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	require.Equal(t, bondAmount, bond.Shares.RoundInt())

	bondedTokens := keeper.TotalBondedTokens(ctx)
	require.Equal(t, bondAmount.Int64(), bondedTokens.Int64())

	// just send the same msgbond multiple times
	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, bondAmount)

	for i := int64(0); i < 5; i++ {
		ctx = ctx.WithBlockHeight(int64(i))

		got := handleMsgDelegate(ctx, msgDelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		validator, found := keeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		bond, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)

		expBond := bondAmount.MulRaw(i + 1)
		expDelegatorShares := bondAmount.MulRaw(i + 2) // (1 self delegation)
		expDelegatorAcc := initBond.Sub(expBond)

		gotBond := bond.Shares.RoundInt()
		gotDelegatorShares := validator.DelegatorShares.RoundInt()
		gotDelegatorAcc := accMapper.GetAccount(ctx, delegatorAddr).GetCoins().AmountOf(params.BondDenom)

		require.Equal(t, expBond, gotBond,
			"i: %v\nexpBond: %v\ngotBond: %v\nvalidator: %v\nbond: %v\n",
			i, expBond, gotBond, validator, bond)
		require.Equal(t, expDelegatorShares, gotDelegatorShares,
			"i: %v\nexpDelegatorShares: %v\ngotDelegatorShares: %v\nvalidator: %v\nbond: %v\n",
			i, expDelegatorShares, gotDelegatorShares, validator, bond)
		require.Equal(t, expDelegatorAcc, gotDelegatorAcc,
			"i: %v\nexpDelegatorAcc: %v\ngotDelegatorAcc: %v\nvalidator: %v\nbond: %v\n",
			i, expDelegatorAcc, gotDelegatorAcc, validator, bond)
	}
}

func TestEditValidatorDecreaseMinSelfDelegation(t *testing.T) {
	validatorAddr := sdk.ValAddress(keep.Addrs[0])

	initPower := int64(100)
	initBond := sdk.TokensFromConsensusPower(100)
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, initPower)
	_ = setInstantUnbondPeriod(keeper, ctx)

	// create validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], initBond)
	msgCreateValidator.MinSelfDelegation = sdk.NewInt(2)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// must end-block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// verify the self-delegation exists
	bond, found := keeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	gotBond := bond.Shares.RoundInt()
	require.Equal(t, initBond, gotBond,
		"initBond: %v\ngotBond: %v\nbond: %v\n",
		initBond, gotBond, bond)

	newMinSelfDelegation := sdk.OneInt()
	msgEditValidator := NewMsgEditValidator(validatorAddr, Description{}, nil, &newMinSelfDelegation)
	got = handleMsgEditValidator(ctx, msgEditValidator, keeper)
	require.False(t, got.IsOK(), "should not be able to decrease minSelfDelegation")
}

func TestEditValidatorIncreaseMinSelfDelegationBeyondCurrentBond(t *testing.T) {
	validatorAddr := sdk.ValAddress(keep.Addrs[0])

	initPower := int64(100)
	initBond := sdk.TokensFromConsensusPower(100)
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, initPower)
	_ = setInstantUnbondPeriod(keeper, ctx)

	// create validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], initBond)
	msgCreateValidator.MinSelfDelegation = sdk.NewInt(2)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// must end-block
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// verify the self-delegation exists
	bond, found := keeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	gotBond := bond.Shares.RoundInt()
	require.Equal(t, initBond, gotBond,
		"initBond: %v\ngotBond: %v\nbond: %v\n",
		initBond, gotBond, bond)

	newMinSelfDelegation := initBond.Add(sdk.OneInt())
	msgEditValidator := NewMsgEditValidator(validatorAddr, Description{}, nil, &newMinSelfDelegation)
	got = handleMsgEditValidator(ctx, msgEditValidator, keeper)
	require.False(t, got.IsOK(), "should not be able to increase minSelfDelegation above current self delegation")
}

func TestIncrementsMsgUnbond(t *testing.T) {
	initPower := int64(1000)
	initBond := sdk.TokensFromConsensusPower(initPower)
	ctx, accMapper, keeper, _ := keep.CreateTestInput(t, false, initPower)
	params := setInstantUnbondPeriod(keeper, ctx)
	denom := params.BondDenom

	// create validator, delegate
	validatorAddr, delegatorAddr := sdk.ValAddress(keep.Addrs[0]), keep.Addrs[1]

	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], initBond)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// initial balance
	amt1 := accMapper.GetAccount(ctx, delegatorAddr).GetCoins().AmountOf(denom)

	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, initBond)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	// balance should have been subtracted after delegation
	amt2 := accMapper.GetAccount(ctx, delegatorAddr).GetCoins().AmountOf(denom)
	require.True(sdk.IntEq(t, amt1.Sub(initBond), amt2))

	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, initBond.MulRaw(2), validator.DelegatorShares.RoundInt())
	require.Equal(t, initBond.MulRaw(2), validator.BondedTokens())

	// just send the same msgUnbond multiple times
	// TODO use decimals here
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgUndelegate := NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)
	numUnbonds := int64(5)
	for i := int64(0); i < numUnbonds; i++ {

		got := handleMsgUndelegate(ctx, msgUndelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)
		var finishTime time.Time
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)
		ctx = ctx.WithBlockTime(finishTime)
		EndBlocker(ctx, keeper)

		// check that the accounts and the bond account have the appropriate values
		validator, found = keeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		bond, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)

		expBond := initBond.Sub(unbondAmt.Amount.Mul(sdk.NewInt(i + 1)))
		expDelegatorShares := initBond.MulRaw(2).Sub(unbondAmt.Amount.Mul(sdk.NewInt(i + 1)))
		expDelegatorAcc := initBond.Sub(expBond)

		gotBond := bond.Shares.RoundInt()
		gotDelegatorShares := validator.DelegatorShares.RoundInt()
		gotDelegatorAcc := accMapper.GetAccount(ctx, delegatorAddr).GetCoins().AmountOf(params.BondDenom)

		require.Equal(t, expBond.Int64(), gotBond.Int64(),
			"i: %v\nexpBond: %v\ngotBond: %v\nvalidator: %v\nbond: %v\n",
			i, expBond, gotBond, validator, bond)
		require.Equal(t, expDelegatorShares.Int64(), gotDelegatorShares.Int64(),
			"i: %v\nexpDelegatorShares: %v\ngotDelegatorShares: %v\nvalidator: %v\nbond: %v\n",
			i, expDelegatorShares, gotDelegatorShares, validator, bond)
		require.Equal(t, expDelegatorAcc.Int64(), gotDelegatorAcc.Int64(),
			"i: %v\nexpDelegatorAcc: %v\ngotDelegatorAcc: %v\nvalidator: %v\nbond: %v\n",
			i, expDelegatorAcc, gotDelegatorAcc, validator, bond)
	}

	// these are more than we have bonded now
	errorCases := []sdk.Int{
		//1<<64 - 1, // more than int64 power
		//1<<63 + 1, // more than int64 power
		sdk.TokensFromConsensusPower(1<<63 - 1),
		sdk.TokensFromConsensusPower(1 << 31),
		initBond,
	}

	for i, c := range errorCases {
		unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, c)
		msgUndelegate := NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)
		got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
		require.False(t, got.IsOK(), "expected unbond msg to fail, index: %v", i)
	}

	leftBonded := initBond.Sub(unbondAmt.Amount.Mul(sdk.NewInt(numUnbonds)))

	// should be able to unbond remaining
	unbondAmt = sdk.NewCoin(sdk.DefaultBondDenom, leftBonded)
	msgUndelegate = NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(),
		"got: %v\nmsgUnbond: %v\nshares: %s\nleftBonded: %s\n", got.Log, msgUndelegate, unbondAmt, leftBonded)
}

func TestMultipleMsgCreateValidator(t *testing.T) {
	initPower := int64(1000)
	initTokens := sdk.TokensFromConsensusPower(initPower)
	ctx, accMapper, keeper, _ := keep.CreateTestInput(t, false, initPower)
	params := setInstantUnbondPeriod(keeper, ctx)

	validatorAddrs := []sdk.ValAddress{
		sdk.ValAddress(keep.Addrs[0]),
		sdk.ValAddress(keep.Addrs[1]),
		sdk.ValAddress(keep.Addrs[2]),
	}
	delegatorAddrs := []sdk.AccAddress{
		keep.Addrs[0],
		keep.Addrs[1],
		keep.Addrs[2],
	}

	// bond them all
	for i, validatorAddr := range validatorAddrs {
		valTokens := sdk.TokensFromConsensusPower(10)
		msgCreateValidatorOnBehalfOf := NewTestMsgCreateValidator(validatorAddr, keep.PKs[i], valTokens)

		got := handleMsgCreateValidator(ctx, msgCreateValidatorOnBehalfOf, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		// verify that the account is bonded
		validators := keeper.GetValidators(ctx, 100)
		require.Equal(t, (i + 1), len(validators))

		val := validators[i]
		balanceExpd := initTokens.Sub(valTokens)
		balanceGot := accMapper.GetAccount(ctx, delegatorAddrs[i]).GetCoins().AmountOf(params.BondDenom)

		require.Equal(t, i+1, len(validators), "expected %d validators got %d, validators: %v", i+1, len(validators), validators)
		require.Equal(t, valTokens, val.DelegatorShares.RoundInt(), "expected %d shares, got %d", 10, val.DelegatorShares)
		require.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all by removing delegation
	for i, validatorAddr := range validatorAddrs {
		_, found := keeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)

		unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))
		msgUndelegate := NewMsgUndelegate(delegatorAddrs[i], validatorAddr, unbondAmt) // remove delegation
		got := handleMsgUndelegate(ctx, msgUndelegate, keeper)

		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)
		var finishTime time.Time

		// Jump to finishTime for unbonding period and remove from unbonding queue
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)
		ctx = ctx.WithBlockTime(finishTime)

		EndBlocker(ctx, keeper)

		// Check that the validator is deleted from state
		validators := keeper.GetValidators(ctx, 100)
		require.Equal(t, len(validatorAddrs)-(i+1), len(validators),
			"expected %d validators got %d", len(validatorAddrs)-(i+1), len(validators))

		_, found = keeper.GetValidator(ctx, validatorAddr)
		require.False(t, found)

		gotBalance := accMapper.GetAccount(ctx, delegatorAddrs[i]).GetCoins().AmountOf(params.BondDenom)
		require.Equal(t, initTokens, gotBalance, "expected account to have %d, got %d", initTokens, gotBalance)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	validatorAddr, delegatorAddrs := sdk.ValAddress(keep.Addrs[0]), keep.Addrs[1:]
	_ = setInstantUnbondPeriod(keeper, ctx)

	// first make a validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], sdk.NewInt(10))
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	// delegate multiple parties
	for i, delegatorAddr := range delegatorAddrs {
		msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, sdk.NewInt(10))
		got := handleMsgDelegate(ctx, msgDelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		// check that the account is bonded
		bond, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)
		require.NotNil(t, bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for i, delegatorAddr := range delegatorAddrs {
		unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
		msgUndelegate := NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)

		got := handleMsgUndelegate(ctx, msgUndelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		var finishTime time.Time
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)

		ctx = ctx.WithBlockTime(finishTime)
		EndBlocker(ctx, keeper)

		// check that the account is unbonded
		_, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.False(t, found)
	}
}

func TestJailValidator(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	validatorAddr, delegatorAddr := sdk.ValAddress(keep.Addrs[0]), keep.Addrs[1]
	_ = setInstantUnbondPeriod(keeper, ctx)

	// create the validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], sdk.NewInt(10))
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// bond a delegator
	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, sdk.NewInt(10))
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)

	// unbond the validators bond portion
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgUndelegateValidator := NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error: %v", got)

	var finishTime time.Time
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)

	ctx = ctx.WithBlockTime(finishTime)
	EndBlocker(ctx, keeper)

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.Jailed, "%v", validator)

	// test that the delegator can still withdraw their bonds
	msgUndelegateDelegator := NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)

	got = handleMsgUndelegate(ctx, msgUndelegateDelegator, keeper)
	require.True(t, got.IsOK(), "expected no error")
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)

	ctx = ctx.WithBlockTime(finishTime)
	EndBlocker(ctx, keeper)

	// verify that the pubkey can now be reused
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)
}

func TestValidatorQueue(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	validatorAddr, delegatorAddr := sdk.ValAddress(keep.Addrs[0]), keep.Addrs[1]

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	keeper.SetParams(ctx, params)

	// create the validator
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], valTokens)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// bond a delegator
	delTokens := sdk.TokensFromConsensusPower(10)
	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, delTokens)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)

	EndBlocker(ctx, keeper)

	// unbond the all self-delegation to put validator in unbonding state
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, delTokens)
	msgUndelegateValidator := NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error: %v", got)

	var finishTime time.Time
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)

	ctx = ctx.WithBlockTime(finishTime)
	EndBlocker(ctx, keeper)

	origHeader := ctx.BlockHeader()

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonding(), "%v", validator)

	// should still be unbonding at time 6 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 6))
	EndBlocker(ctx, keeper)

	validator, found = keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonding(), "%v", validator)

	// should be in unbonded state at time 7 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 7))
	EndBlocker(ctx, keeper)

	validator, found = keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonded(), "%v", validator)
}

func TestUnbondingPeriod(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	validatorAddr := sdk.ValAddress(keep.Addrs[0])

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	keeper.SetParams(ctx, params)

	// create the validator
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], valTokens)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	EndBlocker(ctx, keeper)

	// begin unbonding
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))
	msgUndelegate := NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected no error")

	origHeader := ctx.BlockHeader()

	_, found := keeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// cannot complete unbonding at same time
	EndBlocker(ctx, keeper)
	_, found = keeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// cannot complete unbonding at time 6 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 6))
	EndBlocker(ctx, keeper)
	_, found = keeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// can complete unbonding at time 7 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 7))
	EndBlocker(ctx, keeper)
	_, found = keeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.False(t, found, "should have unbonded")
}

func TestUnbondingFromUnbondingValidator(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	validatorAddr, delegatorAddr := sdk.ValAddress(keep.Addrs[0]), keep.Addrs[1]

	// create the validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], sdk.NewInt(10))
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// bond a delegator
	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, sdk.NewInt(10))
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)

	// unbond the validators bond portion
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgUndelegateValidator := NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error")

	// change the ctx to Block Time one second before the validator would have unbonded
	var finishTime time.Time
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(got.Data, &finishTime)
	ctx = ctx.WithBlockTime(finishTime.Add(time.Second * -1))

	// unbond the delegator from the validator
	msgUndelegateDelegator := NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegateDelegator, keeper)
	require.True(t, got.IsOK(), "expected no error")

	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(keeper.UnbondingTime(ctx)))

	// Run the EndBlocker
	EndBlocker(ctx, keeper)

	// Check to make sure that the unbonding delegation is no longer in state
	// (meaning it was deleted in the above EndBlocker)
	_, found := keeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	require.False(t, found, "should be removed from state")
}

func TestRedelegationPeriod(t *testing.T) {
	ctx, AccMapper, keeper, _ := keep.CreateTestInput(t, false, 1000)
	validatorAddr, validatorAddr2 := sdk.ValAddress(keep.Addrs[0]), sdk.ValAddress(keep.Addrs[1])
	denom := keeper.GetParams(ctx).BondDenom

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	keeper.SetParams(ctx, params)

	// create the validators
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], sdk.NewInt(10))

	// initial balance
	amt1 := AccMapper.GetAccount(ctx, sdk.AccAddress(validatorAddr)).GetCoins().AmountOf(denom)

	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// balance should have been subtracted after creation
	amt2 := AccMapper.GetAccount(ctx, sdk.AccAddress(validatorAddr)).GetCoins().AmountOf(denom)
	require.Equal(t, amt1.Sub(sdk.NewInt(10)).Int64(), amt2.Int64(), "expected coins to be subtracted")

	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr2, keep.PKs[1], sdk.NewInt(10))
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	bal1 := AccMapper.GetAccount(ctx, sdk.AccAddress(validatorAddr)).GetCoins()

	// begin redelegate
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgBeginRedelegate := NewMsgBeginRedelegate(sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2, redAmt)
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, %v", got)

	// origin account should not lose tokens as with a regular delegation
	bal2 := AccMapper.GetAccount(ctx, sdk.AccAddress(validatorAddr)).GetCoins()
	require.Equal(t, bal1, bal2)

	origHeader := ctx.BlockHeader()

	// cannot complete redelegation at same time
	EndBlocker(ctx, keeper)
	_, found := keeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.True(t, found, "should not have unbonded")

	// cannot complete redelegation at time 6 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 6))
	EndBlocker(ctx, keeper)
	_, found = keeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.True(t, found, "should not have unbonded")

	// can complete redelegation at time 7 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 7))
	EndBlocker(ctx, keeper)
	_, found = keeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.False(t, found, "should have unbonded")
}

func TestTransitiveRedelegation(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	validatorAddr := sdk.ValAddress(keep.Addrs[0])
	validatorAddr2 := sdk.ValAddress(keep.Addrs[1])
	validatorAddr3 := sdk.ValAddress(keep.Addrs[2])

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	keeper.SetParams(ctx, params)

	// create the validators
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, keep.PKs[0], sdk.NewInt(10))
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr2, keep.PKs[1], sdk.NewInt(10))
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr3, keep.PKs[2], sdk.NewInt(10))
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// begin redelegate
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgBeginRedelegate := NewMsgBeginRedelegate(sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2, redAmt)
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, %v", got)

	// cannot redelegation to next validator while first delegation exists
	msgBeginRedelegate = NewMsgBeginRedelegate(sdk.AccAddress(validatorAddr), validatorAddr2, validatorAddr3, redAmt)
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, !got.IsOK(), "expected an error, msg: %v", msgBeginRedelegate)

	// complete first redelegation
	EndBlocker(ctx, keeper)

	// now should be able to redelegate from the second validator to the third
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error")
}

func TestMultipleRedelegationAtSameTime(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	valAddr := sdk.ValAddress(keep.Addrs[0])
	valAddr2 := sdk.ValAddress(keep.Addrs[1])

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 1 * time.Second
	keeper.SetParams(ctx, params)

	// create the validators
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valAddr, keep.PKs[0], valTokens)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = NewTestMsgCreateValidator(valAddr2, keep.PKs[1], valTokens)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// end block to bond them
	EndBlocker(ctx, keeper)

	// begin a redelegate
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgBeginRedelegate := NewMsgBeginRedelegate(selfDelAddr, valAddr, valAddr2, redAmt)
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, %v", got)

	// there should only be one entry in the redelegation object
	rd, found := keeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 1)

	// start a second redelegation at this same time as the first
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, msg: %v", msgBeginRedelegate)

	// now there should be two entries
	rd, found = keeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 2)

	// move forward in time, should complete both redelegations
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(1 * time.Second))
	EndBlocker(ctx, keeper)

	rd, found = keeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.False(t, found)
}

func TestMultipleRedelegationAtUniqueTimes(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	valAddr := sdk.ValAddress(keep.Addrs[0])
	valAddr2 := sdk.ValAddress(keep.Addrs[1])

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 10 * time.Second
	keeper.SetParams(ctx, params)

	// create the validators
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valAddr, keep.PKs[0], valTokens)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = NewTestMsgCreateValidator(valAddr2, keep.PKs[1], valTokens)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// end block to bond them
	EndBlocker(ctx, keeper)

	// begin a redelegate
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgBeginRedelegate := NewMsgBeginRedelegate(selfDelAddr, valAddr, valAddr2, redAmt)
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, %v", got)

	// move forward in time and start a second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, msg: %v", msgBeginRedelegate)

	// now there should be two entries
	rd, found := keeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 2)

	// move forward in time, should complete the first redelegation, but not the second
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	EndBlocker(ctx, keeper)
	rd, found = keeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 1)

	// move forward in time, should complete the second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	EndBlocker(ctx, keeper)
	rd, found = keeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.False(t, found)
}

func TestMultipleUnbondingDelegationAtSameTime(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	valAddr := sdk.ValAddress(keep.Addrs[0])

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 1 * time.Second
	keeper.SetParams(ctx, params)

	// create the validator
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valAddr, keep.PKs[0], valTokens)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// end block to bond
	EndBlocker(ctx, keeper)

	// begin an unbonding delegation
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgUndelegate := NewMsgUndelegate(selfDelAddr, valAddr, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, %v", got)

	// there should only be one entry in the ubd object
	ubd, found := keeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// start a second ubd at this same time as the first
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, msg: %v", msgUndelegate)

	// now there should be two entries
	ubd, found = keeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 2)

	// move forwaubd in time, should complete both ubds
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(1 * time.Second))
	EndBlocker(ctx, keeper)

	ubd, found = keeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.False(t, found)
}

func TestMultipleUnbondingDelegationAtUniqueTimes(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	valAddr := sdk.ValAddress(keep.Addrs[0])

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 10 * time.Second
	keeper.SetParams(ctx, params)

	// create the validator
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valAddr, keep.PKs[0], valTokens)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// end block to bond
	EndBlocker(ctx, keeper)

	// begin an unbonding delegation
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgUndelegate := NewMsgUndelegate(selfDelAddr, valAddr, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, %v", got)

	// there should only be one entry in the ubd object
	ubd, found := keeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// move forwaubd in time and start a second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, msg: %v", msgUndelegate)

	// now there should be two entries
	ubd, found = keeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 2)

	// move forwaubd in time, should complete the first redelegation, but not the second
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	EndBlocker(ctx, keeper)
	ubd, found = keeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// move forwaubd in time, should complete the second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	EndBlocker(ctx, keeper)
	ubd, found = keeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.False(t, found)
}

func TestUnbondingWhenExcessValidators(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	validatorAddr1 := sdk.ValAddress(keep.Addrs[0])
	validatorAddr2 := sdk.ValAddress(keep.Addrs[1])
	validatorAddr3 := sdk.ValAddress(keep.Addrs[2])

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	// add three validators
	valTokens1 := sdk.TokensFromConsensusPower(50)
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr1, keep.PKs[0], valTokens1)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")
	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(keeper.GetLastValidators(ctx)))

	valTokens2 := sdk.TokensFromConsensusPower(30)
	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr2, keep.PKs[1], valTokens2)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")
	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(keeper.GetLastValidators(ctx)))

	valTokens3 := sdk.TokensFromConsensusPower(10)
	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr3, keep.PKs[2], valTokens3)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")
	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(keeper.GetLastValidators(ctx)))

	// unbond the validator-2
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens2)
	msgUndelegate := NewMsgUndelegate(sdk.AccAddress(validatorAddr2), validatorAddr2, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgUndelegate")

	// apply TM updates
	keeper.ApplyAndReturnValidatorSetUpdates(ctx)

	// because there are extra validators waiting to get in, the queued
	// validator (aka. validator-1) should make it into the bonded group, thus
	// the total number of validators should stay the same
	vals := keeper.GetLastValidators(ctx)
	require.Equal(t, 2, len(vals), "vals %v", vals)
	val1, found := keeper.GetValidator(ctx, validatorAddr1)
	require.True(t, found)
	require.Equal(t, sdk.Bonded, val1.Status, "%v", val1)
}

func TestBondUnbondRedelegateSlashTwice(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	valA, valB, del := sdk.ValAddress(keep.Addrs[0]), sdk.ValAddress(keep.Addrs[1]), keep.Addrs[2]
	consAddr0 := sdk.ConsAddress(keep.PKs[0].Address())

	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valA, keep.PKs[0], valTokens)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = NewTestMsgCreateValidator(valB, keep.PKs[1], valTokens)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// delegate 10 stake
	msgDelegate := NewTestMsgDelegate(del, valA, valTokens)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgDelegate")

	// apply Tendermint updates
	updates := keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))

	// a block passes
	ctx = ctx.WithBlockHeight(1)

	// begin unbonding 4 stake
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(4))
	msgUndelegate := NewMsgUndelegate(del, valA, unbondAmt)
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgUndelegate")

	// begin redelegate 6 stake
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(6))
	msgBeginRedelegate := NewMsgBeginRedelegate(del, valA, valB, redAmt)
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgBeginRedelegate")

	// destination delegation should have 6 shares
	delegation, found := keeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewDecFromInt(redAmt.Amount), delegation.Shares)

	// must apply validator updates
	updates = keeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))

	// slash the validator by half
	keeper.Slash(ctx, consAddr0, 0, 20, sdk.NewDecWithPrec(5, 1))

	// unbonding delegation should have been slashed by half
	ubd, found := keeper.GetUnbondingDelegation(ctx, del, valA)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	require.Equal(t, unbondAmt.Amount.QuoRaw(2), ubd.Entries[0].Balance)

	// redelegation should have been slashed by half
	redelegation, found := keeper.GetRedelegation(ctx, del, valA, valB)
	require.True(t, found)
	require.Len(t, redelegation.Entries, 1)

	// destination delegation should have been slashed by half
	delegation, found = keeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewDecFromInt(redAmt.Amount.QuoRaw(2)), delegation.Shares)

	// validator power should have been reduced by half
	validator, found := keeper.GetValidator(ctx, valA)
	require.True(t, found)
	require.Equal(t, valTokens.QuoRaw(2), validator.GetBondedTokens())

	// slash the validator for an infraction committed after the unbonding and redelegation begin
	ctx = ctx.WithBlockHeight(3)
	keeper.Slash(ctx, consAddr0, 2, 10, sdk.NewDecWithPrec(5, 1))

	// unbonding delegation should be unchanged
	ubd, found = keeper.GetUnbondingDelegation(ctx, del, valA)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	require.Equal(t, unbondAmt.Amount.QuoRaw(2), ubd.Entries[0].Balance)

	// redelegation should be unchanged
	redelegation, found = keeper.GetRedelegation(ctx, del, valA, valB)
	require.True(t, found)
	require.Len(t, redelegation.Entries, 1)

	// destination delegation should be unchanged
	delegation, found = keeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewDecFromInt(redAmt.Amount.QuoRaw(2)), delegation.Shares)

	// end blocker
	EndBlocker(ctx, keeper)

	// validator power should have been reduced to zero
	// validator should be in unbonding state
	validator, _ = keeper.GetValidator(ctx, valA)
	require.Equal(t, validator.GetStatus(), sdk.Unbonding)
}

func TestInvalidMsg(t *testing.T) {
	k := keep.Keeper{}
	h := NewHandler(k)

	res := h(sdk.NewContext(nil, abci.Header{}, false, nil), sdk.NewTestMsg())
	require.False(t, res.IsOK())
	require.True(t, strings.Contains(res.Log, "unrecognized staking message type"))
}

func TestInvalidCoinDenom(t *testing.T) {
	ctx, _, keeper, _ := keep.CreateTestInput(t, false, 1000)
	valA, valB, delAddr := sdk.ValAddress(keep.Addrs[0]), sdk.ValAddress(keep.Addrs[1]), keep.Addrs[2]

	valTokens := sdk.TokensFromConsensusPower(100)
	invalidCoin := sdk.NewCoin("churros", valTokens)
	validCoin := sdk.NewCoin(sdk.DefaultBondDenom, valTokens)
	oneCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())

	commission := types.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.ZeroDec())

	msgCreate := types.NewMsgCreateValidator(valA, keep.PKs[0], invalidCoin, Description{}, commission, sdk.OneInt())
	got := handleMsgCreateValidator(ctx, msgCreate, keeper)
	require.False(t, got.IsOK())
	msgCreate = types.NewMsgCreateValidator(valA, keep.PKs[0], validCoin, Description{}, commission, sdk.OneInt())
	got = handleMsgCreateValidator(ctx, msgCreate, keeper)
	require.True(t, got.IsOK())
	msgCreate = types.NewMsgCreateValidator(valB, keep.PKs[1], validCoin, Description{}, commission, sdk.OneInt())
	got = handleMsgCreateValidator(ctx, msgCreate, keeper)
	require.True(t, got.IsOK())

	msgDelegate := types.NewMsgDelegate(delAddr, valA, invalidCoin)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.False(t, got.IsOK())
	msgDelegate = types.NewMsgDelegate(delAddr, valA, validCoin)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK())

	msgUndelegate := types.NewMsgUndelegate(delAddr, valA, invalidCoin)
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.False(t, got.IsOK())
	msgUndelegate = types.NewMsgUndelegate(delAddr, valA, oneCoin)
	got = handleMsgUndelegate(ctx, msgUndelegate, keeper)
	require.True(t, got.IsOK())

	msgRedelegate := types.NewMsgBeginRedelegate(delAddr, valA, valB, invalidCoin)
	got = handleMsgBeginRedelegate(ctx, msgRedelegate, keeper)
	require.False(t, got.IsOK())
	msgRedelegate = types.NewMsgBeginRedelegate(delAddr, valA, valB, oneCoin)
	got = handleMsgBeginRedelegate(ctx, msgRedelegate, keeper)
	require.True(t, got.IsOK())
}
