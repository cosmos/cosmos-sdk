package stake

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	crypto "github.com/tendermint/go-crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//______________________________________________________________________

func newTestMsgCreateValidator(address sdk.Address, pubKey crypto.PubKey, amt int64) MsgCreateValidator {
	return MsgCreateValidator{
		Description:   Description{},
		ValidatorAddr: address,
		PubKey:        pubKey,
		Bond:          sdk.Coin{"steak", amt},
	}
}

func newTestMsgDelegate(delegatorAddr, validatorAddr sdk.Address, amt int64) MsgDelegate {
	return MsgDelegate{
		DelegatorAddr: delegatorAddr,
		ValidatorAddr: validatorAddr,
		Bond:          sdk.Coin{"steak", amt},
	}
}

//______________________________________________________________________

func TestDuplicatesMsgCreateValidator(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 1000)

	validatorAddr := addrs[0]
	pk := pks[0]
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
	msgCreateValidator.PubKey = pks[1]
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.False(t, got.IsOK(), "%v", got)
}

func TestIncrementsMsgDelegate(t *testing.T) {
	initBond := int64(1000)
	ctx, accMapper, keeper := createTestInput(t, false, initBond)
	params := keeper.GetParams(ctx)

	bondAmount := int64(10)
	validatorAddr, delegatorAddr := addrs[0], addrs[1]

	// first create validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, pks[0], bondAmount)
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
		expDelegatorAcc := initBond - expBond

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
	ctx, accMapper, keeper := createTestInput(t, false, initBond)
	params := keeper.GetParams(ctx)

	// create validator, delegate
	validatorAddr, delegatorAddr := addrs[0], addrs[1]

	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, pks[0], initBond)
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
	unbondShares, unbondSharesStr := int64(10), "10"
	msgUnbond := NewMsgUnbond(delegatorAddr, validatorAddr, unbondSharesStr)
	numUnbonds := 5
	for i := 0; i < numUnbonds; i++ {
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the accounts and the bond account have the appropriate values
		validator, found = keeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		bond, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)

		expBond := initBond - int64(i+1)*unbondShares
		expDelegatorShares := 2*initBond - int64(i+1)*unbondShares
		expDelegatorAcc := initBond - expBond

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
		unbondShares := strconv.Itoa(int(c))
		msgUnbond := NewMsgUnbond(delegatorAddr, validatorAddr, unbondShares)
		got = handleMsgUnbond(ctx, msgUnbond, keeper)
		require.False(t, got.IsOK(), "expected unbond msg to fail")
	}

	leftBonded := initBond - unbondShares*int64(numUnbonds)

	// should be unable to unbond one more than we have
	unbondSharesStr = strconv.Itoa(int(leftBonded) + 1)
	msgUnbond = NewMsgUnbond(delegatorAddr, validatorAddr, unbondSharesStr)
	got = handleMsgUnbond(ctx, msgUnbond, keeper)
	assert.False(t, got.IsOK(),
		"got: %v\nmsgUnbond: %v\nshares: %v\nleftBonded: %v\n", got, msgUnbond, unbondSharesStr, leftBonded)

	// should be able to unbond just what we have
	unbondSharesStr = strconv.Itoa(int(leftBonded))
	msgUnbond = NewMsgUnbond(delegatorAddr, validatorAddr, unbondSharesStr)
	got = handleMsgUnbond(ctx, msgUnbond, keeper)
	assert.True(t, got.IsOK(),
		"got: %v\nmsgUnbond: %v\nshares: %v\nleftBonded: %v\n", got, msgUnbond, unbondSharesStr, leftBonded)
}

func TestMultipleMsgCreateValidator(t *testing.T) {
	initBond := int64(1000)
	ctx, accMapper, keeper := createTestInput(t, false, initBond)
	params := keeper.GetParams(ctx)
	validatorAddrs := []sdk.Address{addrs[0], addrs[1], addrs[2]}

	// bond them all
	for i, validatorAddr := range validatorAddrs {
		msgCreateValidator := newTestMsgCreateValidator(validatorAddr, pks[i], 10)
		got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is bonded
		validators := keeper.GetValidators(ctx, 100)
		require.Equal(t, (i + 1), len(validators))
		val := validators[i]
		balanceExpd := initBond - 10
		balanceGot := accMapper.GetAccount(ctx, val.Owner).GetCoins().AmountOf(params.BondDenom)
		require.Equal(t, i+1, len(validators), "expected %d validators got %d, validators: %v", i+1, len(validators), validators)
		require.Equal(t, 10, int(val.DelegatorShares.Evaluate()), "expected %d shares, got %d", 10, val.DelegatorShares)
		require.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	// unbond them all
	for i, validatorAddr := range validatorAddrs {
		validatorPre, found := keeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		msgUnbond := NewMsgUnbond(validatorAddr, validatorAddr, "10") // self-delegation
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		validators := keeper.GetValidators(ctx, 100)
		require.Equal(t, len(validatorAddrs)-(i+1), len(validators),
			"expected %d validators got %d", len(validatorAddrs)-(i+1), len(validators))

		_, found = keeper.GetValidator(ctx, validatorAddr)
		require.False(t, found)

		expBalance := initBond
		gotBalance := accMapper.GetAccount(ctx, validatorPre.Owner).GetCoins().AmountOf(params.BondDenom)
		require.Equal(t, expBalance, gotBalance, "expected account to have %d, got %d", expBalance, gotBalance)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 1000)
	validatorAddr, delegatorAddrs := addrs[0], addrs[1:]

	//first make a validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, pks[0], 10)
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
		msgUnbond := NewMsgUnbond(delegatorAddr, validatorAddr, "10")
		got := handleMsgUnbond(ctx, msgUnbond, keeper)
		require.True(t, got.IsOK(), "expected msg %d to be ok, got %v", i, got)

		//Check that the account is unbonded
		_, found := keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.False(t, found)
	}
}

func TestRevokeValidator(t *testing.T) {
	ctx, _, keeper := createTestInput(t, false, 1000)
	validatorAddr, delegatorAddr := addrs[0], addrs[1]

	// create the validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, pks[0], 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// bond a delegator
	msgDelegate := newTestMsgDelegate(delegatorAddr, validatorAddr, 10)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)

	// unbond the validators bond portion
	msgUnbondValidator := NewMsgUnbond(validatorAddr, validatorAddr, "10")
	got = handleMsgUnbond(ctx, msgUnbondValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")
	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.Revoked)

	// test that this address cannot yet be bonded too because is revoked
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	assert.False(t, got.IsOK(), "expected error, got %v", got)

	// test that the delegator can still withdraw their bonds
	msgUnbondDelegator := NewMsgUnbond(delegatorAddr, validatorAddr, "10")
	got = handleMsgUnbond(ctx, msgUnbondDelegator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// verify that the pubkey can now be reused
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	assert.True(t, got.IsOK(), "expected ok, got %v", got)
}
