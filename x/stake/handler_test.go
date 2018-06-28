package stake

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
	keep "github.com/cosmos/cosmos-sdk/x/stake/keeper"
	"github.com/cosmos/cosmos-sdk/x/stake/types"
)

//______________________________________________________________________

func newTestMsgCreateValidator(address sdk.Address, pubKey crypto.PubKey, amt int64) MsgCreateValidator {
	return MsgCreateValidator{
		Description:    Description{},
		ValidatorAddr:  address,
		PubKey:         pubKey,
		SelfDelegation: sdk.Coin{"steak", sdk.NewInt(amt)},
	}
}

func newTestMsgDelegate(delegatorAddr, validatorAddr sdk.Address, amt int64) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
		Bond:          sdk.Coin{"steak", sdk.NewInt(amt)},
	}
}

// retrieve params which are instant
func setInstantUnbondPeriod(keeper keep.Keeper, ctx sdk.Context) types.Params {
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	keeper.SetParams(ctx, params)
	return params
}

//______________________________________________________________________

func TestValidatorByPowerIndex(t *testing.T) {
	validatorAddr, validatorAddr3 := keep.Addrs[0], keep.Addrs[1]

	initBond := int64(1000000)
	ctx, _, keeper := keep.CreateTestInput(t, false, initBond)
	_ = setInstantUnbondPeriod(keeper, ctx)

	// create validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[0], initBond)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// verify the self-delegation exists
	bond, found := keeper.GetDelegation(ctx, validatorAddr, validatorAddr)
	require.True(t, found)
	gotBond := bond.Shares.Evaluate()
	require.Equal(t, initBond, gotBond,
		"initBond: %v\ngotBond: %v\nbond: %v\n",
		initBond, gotBond, bond)

	// verify that the by power index exists
	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	pool := keeper.GetPool(ctx)
	power := keep.GetValidatorsByPowerIndexKey(validator, pool)
	require.True(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power))

	// create a second validator keep it bonded
	msgCreateValidator = newTestMsgCreateValidator(validatorAddr3, keep.PKs[2], int64(1000000))
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// slash and revoke the first validator
	keeper.Slash(ctx, keep.PKs[0], 0, sdk.NewRat(1, 2))
	keeper.Revoke(ctx, keep.PKs[0])
	validator, found = keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, sdk.Unbonded, validator.PoolShares.Status)             // ensure is unbonded
	require.Equal(t, int64(500000), validator.PoolShares.Amount.Evaluate()) // ensure is unbonded

	// the old power record should have been deleted as the power changed
	assert.False(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power))

	// but the new power record should have been created
	validator, found = keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	pool = keeper.GetPool(ctx)
	power2 := GetValidatorsByPowerIndexKey(validator, pool)
	require.True(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power2))

	// inflate a bunch
	for i := 0; i < 20000; i++ {
		pool = keeper.ProcessProvisions(ctx)
		keeper.SetPool(ctx, pool)
	}

	// now the new record power index should be the same as the original record
	power3 := GetValidatorsByPowerIndexKey(validator, pool)
	assert.Equal(t, power2, power3)

	// unbond self-delegation
	msgBeginUnbonding := NewMsgBeginUnbonding(validatorAddr, validatorAddr, sdk.NewRat(1000000))
	msgCompleteUnbonding := NewMsgCompleteUnbonding(validatorAddr, validatorAddr)
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)
	got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbonding, keeper)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	// verify that by power key nolonger exists
	_, found = keeper.GetValidator(ctx, validatorAddr)
	require.False(t, found)
	assert.False(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power3))
}

func TestDuplicatesMsgCreateValidator(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)

	validatorAddr := keep.Addrs[0]
	pk := keep.PKs[0]
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, pk, 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.True(t, got.IsOK(), "%v", got)
	validator, found := keeper.GetValidator(ctx, validatorAddr)

	require.True(t, found)
	assert.Equal(t, sdk.Bonded, validator.Status())
	assert.Equal(t, validatorAddr, validator.Owner)
	assert.Equal(t, pk, validator.PubKey)
	assert.Equal(t, sdk.NewRat(10), validator.PoolShares.Bonded())
	assert.Equal(t, sdk.NewRat(10), validator.DelegatorShares)
	assert.Equal(t, Description{}, validator.Description)

	// one validator cannot bond twice
	msgCreateValidator.PubKey = keep.PKs[1]
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.False(t, got.IsOK(), "%v", got)
}

func TestIncrementsMsgDelegate(t *testing.T) {
	initBond := int64(1000)
	ctx, accMapper, keeper := keep.CreateTestInput(t, false, initBond)
	params := keeper.GetParams(ctx)

	bondAmount := int64(10)
	validatorAddr, delegatorAddr := keep.Addrs[0], keep.Addrs[1]

	// first create validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[0], bondAmount)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.True(t, got.IsOK(), "expected create validator msg to be ok, got %v", got)

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, sdk.Bonded, validator.Status())
	assert.Equal(t, bondAmount, validator.DelegatorShares.Evaluate())
	assert.Equal(t, bondAmount, validator.PoolShares.Bonded().Evaluate(), "validator: %v", validator)

	_, found = keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
	require.False(t, found)

	bond, found := keeper.GetDelegation(ctx, validatorAddr, validatorAddr)
	require.True(t, found)
	assert.Equal(t, bondAmount, bond.Shares.Evaluate())

	pool := keeper.GetPool(ctx)
	exRate := validator.DelegatorShareExRate(pool)
	require.True(t, exRate.Equal(sdk.OneRat()), "expected exRate 1 got %v", exRate)
	assert.Equal(t, bondAmount, pool.BondedShares.Evaluate())
	assert.Equal(t, bondAmount, pool.BondedTokens)

	// just send the same msgbond multiple times
	msgDelegate := newTestMsgDelegate(delegatorAddr, validatorAddr, bondAmount)

	for i := 0; i < 5; i++ {
		ctx = ctx.WithBlockHeight(int64(i))

		got := handleMsgDelegate(ctx, msgDelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		validator, found := keeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		bond, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)

		pool := keeper.GetPool(ctx)
		exRate := validator.DelegatorShareExRate(pool)
		require.True(t, exRate.Equal(sdk.OneRat()), "expected exRate 1 got %v, i = %v", exRate, i)

		expBond := int64(i+1) * bondAmount
		expDelegatorShares := int64(i+2) * bondAmount // (1 self delegation)
		expDelegatorAcc := sdk.NewInt(initBond - expBond)

		require.Equal(t, bond.Height, int64(i), "Incorrect bond height")

		gotBond := bond.Shares.Evaluate()
		gotDelegatorShares := validator.DelegatorShares.Evaluate()
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

func TestIncrementsMsgUnbond(t *testing.T) {
	initBond := int64(1000)
	ctx, accMapper, keeper := keep.CreateTestInput(t, false, initBond)
	params := setInstantUnbondPeriod(keeper, ctx)

	// create validator, delegate
	validatorAddr, delegatorAddr := keep.Addrs[0], keep.Addrs[1]

	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[0], initBond)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	msgDelegate := newTestMsgDelegate(delegatorAddr, validatorAddr, initBond)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	assert.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	assert.Equal(t, initBond*2, validator.DelegatorShares.Evaluate())
	assert.Equal(t, initBond*2, validator.PoolShares.Bonded().Evaluate())

	// just send the same msgUnbond multiple times
	// TODO use decimals here
	unbondShares := sdk.NewRat(10)
	msgBeginUnbonding := NewMsgBeginUnbonding(delegatorAddr, validatorAddr, unbondShares)
	msgCompleteUnbonding := NewMsgCompleteUnbonding(delegatorAddr, validatorAddr)
	numUnbonds := 5
	for i := 0; i < numUnbonds; i++ {
		got := handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)
		got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbonding, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		validator, found = keeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		bond, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)

		expBond := initBond - int64(i+1)*unbondShares.Evaluate()
		expDelegatorShares := 2*initBond - int64(i+1)*unbondShares.Evaluate()
		expDelegatorAcc := sdk.NewInt(initBond - expBond)

		gotBond := bond.Shares.Evaluate()
		gotDelegatorShares := validator.DelegatorShares.Evaluate()
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

	// these are more than we have bonded now
	errorCases := []int64{
		//1<<64 - 1, // more than int64
		//1<<63 + 1, // more than int64
		1<<63 - 1,
		1 << 31,
		initBond,
	}
	for _, c := range errorCases {
		unbondShares := sdk.NewRat(int64(c))
		msgBeginUnbonding := NewMsgBeginUnbonding(delegatorAddr, validatorAddr, unbondShares)
		got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
		require.False(t, got.IsOK(), "expected unbond msg to fail")
	}

	leftBonded := initBond - int64(numUnbonds)*unbondShares.Evaluate()

	// should be unable to unbond one more than we have
	unbondShares = sdk.NewRat(leftBonded + 1)
	msgBeginUnbonding = NewMsgBeginUnbonding(delegatorAddr, validatorAddr, unbondShares)
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	assert.False(t, got.IsOK(),
		"got: %v\nmsgUnbond: %v\nshares: %v\nleftBonded: %v\n", got, msgBeginUnbonding, unbondShares.String(), leftBonded)

	// should be able to unbond just what we have
	unbondShares = sdk.NewRat(leftBonded)
	msgBeginUnbonding = NewMsgBeginUnbonding(delegatorAddr, validatorAddr, unbondShares)
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	assert.True(t, got.IsOK(),
		"got: %v\nmsgUnbond: %v\nshares: %v\nleftBonded: %v\n", got, msgBeginUnbonding, unbondShares, leftBonded)
}

func TestMultipleMsgCreateValidator(t *testing.T) {
	initBond := int64(1000)
	ctx, accMapper, keeper := keep.CreateTestInput(t, false, initBond)
	params := setInstantUnbondPeriod(keeper, ctx)

	validatorAddrs := []sdk.Address{keep.Addrs[0], keep.Addrs[1], keep.Addrs[2]}

	// bond them all
	for i, validatorAddr := range validatorAddrs {
		msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[i], 10)
		got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		validators := keeper.GetValidators(ctx, 100)
		require.Equal(t, (i + 1), len(validators))
		val := validators[i]
		balanceExpd := sdk.NewInt(initBond - 10)
		balanceGot := accMapper.GetAccount(ctx, val.Owner).GetCoins().AmountOf(params.BondDenom)
		require.Equal(t, i+1, len(validators), "expected %d validators got %d, validators: %v", i+1, len(validators), validators)
		require.Equal(t, 10, int(val.DelegatorShares.Evaluate()), "expected %d shares, got %d", 10, val.DelegatorShares)
		require.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, validatorAddr := range validatorAddrs {
		validatorPre, found := keeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		msgBeginUnbonding := NewMsgBeginUnbonding(validatorAddr, validatorAddr, sdk.NewRat(10)) // self-delegation
		msgCompleteUnbonding := NewMsgCompleteUnbonding(validatorAddr, validatorAddr)
		got := handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)
		got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbonding, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		validators := keeper.GetValidators(ctx, 100)
		require.Equal(t, len(validatorAddrs)-(i+1), len(validators),
			"expected %d validators got %d", len(validatorAddrs)-(i+1), len(validators))

		_, found = keeper.GetValidator(ctx, validatorAddr)
		require.False(t, found)

		expBalance := sdk.NewInt(initBond)
		gotBalance := accMapper.GetAccount(ctx, validatorPre.Owner).GetCoins().AmountOf(params.BondDenom)
		require.Equal(t, expBalance, gotBalance, "expected account to have %d, got %d", expBalance, gotBalance)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr, delegatorAddrs := keep.Addrs[0], keep.Addrs[1:]
	_ = setInstantUnbondPeriod(keeper, ctx)

	//first make a validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[0], 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected msg to be ok, got %v", got)

	// delegate multiple parties
	for i, delegatorAddr := range delegatorAddrs {
		msgDelegate := newTestMsgDelegate(delegatorAddr, validatorAddr, 10)
		got := handleMsgDelegate(ctx, msgDelegate, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		bond, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)
		require.NotNil(t, bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for i, delegatorAddr := range delegatorAddrs {
		msgBeginUnbonding := NewMsgBeginUnbonding(delegatorAddr, validatorAddr, sdk.NewRat(10))
		msgCompleteUnbonding := NewMsgCompleteUnbonding(delegatorAddr, validatorAddr)
		got := handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)
		got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbonding, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		_, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.False(t, found)
	}
}

func TestRevokeValidator(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr, delegatorAddr := keep.Addrs[0], keep.Addrs[1]
	_ = setInstantUnbondPeriod(keeper, ctx)

	// create the validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[0], 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// bond a delegator
	msgDelegate := newTestMsgDelegate(delegatorAddr, validatorAddr, 10)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)

	validator, _ := keeper.GetValidator(ctx, validatorAddr)

	// unbond the validators bond portion
	msgBeginUnbondingValidator := NewMsgBeginUnbonding(validatorAddr, validatorAddr, sdk.NewRat(10))
	msgCompleteUnbondingValidator := NewMsgCompleteUnbonding(validatorAddr, validatorAddr)
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbondingValidator, keeper)
	require.True(t, got.IsOK(), "expected no error")
	got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbondingValidator, keeper)
	require.True(t, got.IsOK(), "expected no error")

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.Revoked, "%v", validator)

	// test that this address cannot yet be bonded too because is revoked
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	assert.False(t, got.IsOK(), "expected error, got %v", got)

	// test that the delegator can still withdraw their bonds
	msgBeginUnbondingDelegator := NewMsgBeginUnbonding(delegatorAddr, validatorAddr, sdk.NewRat(10))
	msgCompleteUnbondingDelegator := NewMsgCompleteUnbonding(delegatorAddr, validatorAddr)
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbondingDelegator, keeper)
	require.True(t, got.IsOK(), "expected no error")
	got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbondingDelegator, keeper)
	require.True(t, got.IsOK(), "expected no error")

	// verify that the pubkey can now be reused
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.True(t, got.IsOK(), "expected ok, got %v", got)
}

func TestUnbondingPeriod(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr := keep.Addrs[0]

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 7
	keeper.SetParams(ctx, params)

	// create the validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[0], 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// begin unbonding
	msgBeginUnbonding := NewMsgBeginUnbonding(validatorAddr, validatorAddr, sdk.NewRat(10))
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	require.True(t, got.IsOK(), "expected no error")

	// cannot complete unbonding at same time
	msgCompleteUnbonding := NewMsgCompleteUnbonding(validatorAddr, validatorAddr)
	got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbonding, keeper)
	require.True(t, !got.IsOK(), "expected no error")

	// cannot complete unbonding at time 6 seconds later
	origHeader := ctx.BlockHeader()
	headerTime6 := origHeader
	headerTime6.Time += 6
	ctx = ctx.WithBlockHeader(headerTime6)
	got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbonding, keeper)
	require.True(t, !got.IsOK(), "expected no error")

	// can complete unbonding at time 7 seconds later
	headerTime7 := origHeader
	headerTime7.Time += 7
	ctx = ctx.WithBlockHeader(headerTime7)
	got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbonding, keeper)
	require.True(t, got.IsOK(), "expected no error")
}

func TestRedelegationPeriod(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr, validatorAddr2 := keep.Addrs[0], keep.Addrs[1]

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 7
	keeper.SetParams(ctx, params)

	// create the validators
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[0], 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = newTestMsgCreateValidator(validatorAddr2, keep.PKs[1], 10)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// begin redelegate
	msgBeginRedelegate := NewMsgBeginRedelegate(validatorAddr, validatorAddr, validatorAddr2, sdk.NewRat(10))
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, %v", got)

	// cannot complete redelegation at same time
	msgCompleteRedelegate := NewMsgCompleteRedelegate(validatorAddr, validatorAddr, validatorAddr2)
	got = handleMsgCompleteRedelegate(ctx, msgCompleteRedelegate, keeper)
	require.True(t, !got.IsOK(), "expected an error")

	// cannot complete redelegation at time 6 seconds later
	origHeader := ctx.BlockHeader()
	headerTime6 := origHeader
	headerTime6.Time += 6
	ctx = ctx.WithBlockHeader(headerTime6)
	got = handleMsgCompleteRedelegate(ctx, msgCompleteRedelegate, keeper)
	require.True(t, !got.IsOK(), "expected an error")

	// can complete redelegation at time 7 seconds later
	headerTime7 := origHeader
	headerTime7.Time += 7
	ctx = ctx.WithBlockHeader(headerTime7)
	got = handleMsgCompleteRedelegate(ctx, msgCompleteRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error")
}

func TestTransitiveRedelegation(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr, validatorAddr2, validatorAddr3 := keep.Addrs[0], keep.Addrs[1], keep.Addrs[2]

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	keeper.SetParams(ctx, params)

	// create the validators
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, keep.PKs[0], 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = newTestMsgCreateValidator(validatorAddr2, keep.PKs[1], 10)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = newTestMsgCreateValidator(validatorAddr3, keep.PKs[2], 10)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// begin redelegate
	msgBeginRedelegate := NewMsgBeginRedelegate(validatorAddr, validatorAddr, validatorAddr2, sdk.NewRat(10))
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error, %v", got)

	// cannot redelegation to next validator while first delegation exists
	msgBeginRedelegate = NewMsgBeginRedelegate(validatorAddr, validatorAddr2, validatorAddr3, sdk.NewRat(10))
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, !got.IsOK(), "expected an error, msg: %v", msgBeginRedelegate)

	// complete first redelegation
	msgCompleteRedelegate := NewMsgCompleteRedelegate(validatorAddr, validatorAddr, validatorAddr2)
	got = handleMsgCompleteRedelegate(ctx, msgCompleteRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error")

	// now should be able to redelegate from the second validator to the third
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error")
}
