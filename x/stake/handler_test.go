package stake

import (
	"testing"

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
	require.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// verify the self-delegation exists
	bond, found := keeper.GetDelegation(ctx, validatorAddr, validatorAddr)
	require.True(t, found)
	gotBond := bond.Shares.RoundInt64()
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
	require.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	// slash and revoke the first validator
	keeper.Slash(ctx, keep.PKs[0], 0, initBond, sdk.NewRat(1, 2))
	keeper.Revoke(ctx, keep.PKs[0])
	validator, found = keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, sdk.Unbonded, validator.PoolShares.Status)               // ensure is unbonded
	require.Equal(t, int64(500000), validator.PoolShares.Amount.RoundInt64()) // ensure is unbonded

	// the old power record should have been deleted as the power changed
	require.False(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power))

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
	require.Equal(t, power2, power3)

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
	require.False(t, keep.ValidatorByPowerIndexExists(ctx, keeper, power3))
}

func TestDuplicatesMsgCreateValidator(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)

	validatorAddr := keep.Addrs[0]
	pk := keep.PKs[0]
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr, pk, 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "%v", got)
	validator, found := keeper.GetValidator(ctx, validatorAddr)

	require.True(t, found)
	require.Equal(t, sdk.Bonded, validator.Status())
	require.Equal(t, validatorAddr, validator.Owner)
	require.Equal(t, pk, validator.PubKey)
	require.Equal(t, sdk.NewRat(10), validator.PoolShares.Bonded())
	require.Equal(t, sdk.NewRat(10), validator.DelegatorShares)
	require.Equal(t, Description{}, validator.Description)

	// one validator cannot bond twice
	msgCreateValidator.PubKey = keep.PKs[1]
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.False(t, got.IsOK(), "%v", got)
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
	require.True(t, got.IsOK(), "expected create validator msg to be ok, got %v", got)

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, sdk.Bonded, validator.Status())
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt64())
	require.Equal(t, bondAmount, validator.PoolShares.Bonded().RoundInt64(), "validator: %v", validator)

	_, found = keeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
	require.False(t, found)

	bond, found := keeper.GetDelegation(ctx, validatorAddr, validatorAddr)
	require.True(t, found)
	require.Equal(t, bondAmount, bond.Shares.RoundInt64())

	pool := keeper.GetPool(ctx)
	exRate := validator.DelegatorShareExRate(pool)
	require.True(t, exRate.Equal(sdk.OneRat()), "expected exRate 1 got %v", exRate)
	require.Equal(t, bondAmount, pool.BondedShares.RoundInt64())
	require.Equal(t, bondAmount, pool.BondedTokens)

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

		gotBond := bond.Shares.RoundInt64()
		gotDelegatorShares := validator.DelegatorShares.RoundInt64()
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
	require.True(t, got.IsOK(), "expected create-validator to be ok, got %v", got)

	msgDelegate := newTestMsgDelegate(delegatorAddr, validatorAddr, initBond)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected delegation to be ok, got %v", got)

	validator, found := keeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, initBond*2, validator.DelegatorShares.RoundInt64())
	require.Equal(t, initBond*2, validator.PoolShares.Bonded().RoundInt64())

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

		expBond := initBond - int64(i+1)*unbondShares.RoundInt64()
		expDelegatorShares := 2*initBond - int64(i+1)*unbondShares.RoundInt64()
		expDelegatorAcc := sdk.NewInt(initBond - expBond)

		gotBond := bond.Shares.RoundInt64()
		gotDelegatorShares := validator.DelegatorShares.RoundInt64()
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

	leftBonded := initBond - int64(numUnbonds)*unbondShares.RoundInt64()

	// should be unable to unbond one more than we have
	unbondShares = sdk.NewRat(leftBonded + 1)
	msgBeginUnbonding = NewMsgBeginUnbonding(delegatorAddr, validatorAddr, unbondShares)
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	require.False(t, got.IsOK(),
		"got: %v\nmsgUnbond: %v\nshares: %v\nleftBonded: %v\n", got, msgBeginUnbonding, unbondShares.String(), leftBonded)

	// should be able to unbond just what we have
	unbondShares = sdk.NewRat(leftBonded)
	msgBeginUnbonding = NewMsgBeginUnbonding(delegatorAddr, validatorAddr, unbondShares)
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	require.True(t, got.IsOK(),
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
		require.Equal(t, 10, int(val.DelegatorShares.RoundInt64()), "expected %d shares, got %d", 10, val.DelegatorShares)
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
	require.False(t, got.IsOK(), "expected error, got %v", got)

	// test that the delegator can still withdraw their bonds
	msgBeginUnbondingDelegator := NewMsgBeginUnbonding(delegatorAddr, validatorAddr, sdk.NewRat(10))
	msgCompleteUnbondingDelegator := NewMsgCompleteUnbonding(delegatorAddr, validatorAddr)
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbondingDelegator, keeper)
	require.True(t, got.IsOK(), "expected no error")
	got = handleMsgCompleteUnbonding(ctx, msgCompleteUnbondingDelegator, keeper)
	require.True(t, got.IsOK(), "expected no error")

	// verify that the pubkey can now be reused
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected ok, got %v", got)
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

func TestUnbondingWhenExcessValidators(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr1, validatorAddr2, validatorAddr3 := keep.Addrs[0], keep.Addrs[1], keep.Addrs[2]

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	// add three validators
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr1, keep.PKs[0], 50)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")
	require.Equal(t, 1, len(keeper.GetValidatorsBonded(ctx)))

	msgCreateValidator = newTestMsgCreateValidator(validatorAddr2, keep.PKs[1], 30)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")
	require.Equal(t, 2, len(keeper.GetValidatorsBonded(ctx)))

	msgCreateValidator = newTestMsgCreateValidator(validatorAddr3, keep.PKs[2], 10)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")
	require.Equal(t, 2, len(keeper.GetValidatorsBonded(ctx)))

	// unbond the valdator-2
	msgBeginUnbonding := NewMsgBeginUnbonding(validatorAddr2, validatorAddr2, sdk.NewRat(30))
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgBeginUnbonding")

	// because there are extra validators waiting to get in, the queued
	// validator (aka. validator-1) should make it into the bonded group, thus
	// the total number of validators should stay the same
	vals := keeper.GetValidatorsBonded(ctx)
	require.Equal(t, 2, len(vals), "vals %v", vals)
	val1, found := keeper.GetValidator(ctx, validatorAddr1)
	require.True(t, found)
	require.Equal(t, sdk.Bonded, val1.Status(), "%v", val1)
}

func TestJoiningAsCliffValidator(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr1, validatorAddr2 := keep.Addrs[0], keep.Addrs[1]

	// make sure that the cliff validator is nil to begin with
	cliffVal := keeper.GetCliffValidator(ctx)
	require.Equal(t, []byte(nil), cliffVal)

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	// add the first validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr1, keep.PKs[0], 50)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// cliff validator should still be nil
	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, []byte(nil), cliffVal)

	// Add the second validator
	msgCreateValidator = newTestMsgCreateValidator(validatorAddr2, keep.PKs[1], 30)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// now that we've reached maximum validators, the val-2 should be added to the cliff (top)
	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, validatorAddr2.Bytes(), cliffVal)
}

func TestJoiningToCreateFirstCliffValidator(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr1, validatorAddr2 := keep.Addrs[0], keep.Addrs[1]

	// make sure that the cliff validator is nil to begin with
	cliffVal := keeper.GetCliffValidator(ctx)
	require.Equal(t, []byte(nil), cliffVal)

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	// add the first validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr1, keep.PKs[0], 50)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// cliff validator should still be nil
	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, []byte(nil), cliffVal)

	// Add the second validator
	msgCreateValidator = newTestMsgCreateValidator(validatorAddr2, keep.PKs[1], 60)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// now that we've reached maximum validators, validator-1 should be added to the cliff (top)
	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, validatorAddr1.Bytes(), cliffVal)
}

func TestCliffValidator(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	validatorAddr1, validatorAddr2, validatorAddr3 := keep.Addrs[0], keep.Addrs[1], keep.Addrs[2]

	// make sure that the cliff validator is nil to begin with
	cliffVal := keeper.GetCliffValidator(ctx)
	require.Equal(t, []byte(nil), cliffVal)

	// set the unbonding time
	params := keeper.GetParams(ctx)
	params.UnbondingTime = 0
	params.MaxValidators = 2
	keeper.SetParams(ctx, params)

	// add the first validator
	msgCreateValidator := newTestMsgCreateValidator(validatorAddr1, keep.PKs[0], 50)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// cliff validator should still be nil
	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, []byte(nil), cliffVal)

	// Add the second validator
	msgCreateValidator = newTestMsgCreateValidator(validatorAddr2, keep.PKs[1], 30)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// now that we've reached maximum validators, validator-2 should be added to the cliff (top)
	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, validatorAddr2.Bytes(), cliffVal)

	// add the third validator, which should not make it to being bonded,
	// so the cliff validator should not change because nobody has been kicked out
	msgCreateValidator = newTestMsgCreateValidator(validatorAddr3, keep.PKs[2], 10)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, validatorAddr2.Bytes(), cliffVal)

	// unbond valdator-2
	msgBeginUnbonding := NewMsgBeginUnbonding(validatorAddr2, validatorAddr2, sdk.NewRat(30))
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgBeginUnbonding")

	vals := keeper.GetValidatorsBonded(ctx)
	require.Equal(t, 2, len(vals))

	// now the validator set should be updated,
	// where val-3 enters the validator set on the cliff
	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, validatorAddr3.Bytes(), cliffVal)

	// unbond valdator-1
	msgBeginUnbonding = NewMsgBeginUnbonding(validatorAddr1, validatorAddr1, sdk.NewRat(50))
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgBeginUnbonding")

	// get bonded validators - should just be one
	vals = keeper.GetValidatorsBonded(ctx)
	require.Equal(t, 1, len(vals))

	// cliff now should be empty
	cliffVal = keeper.GetCliffValidator(ctx)
	require.Equal(t, []byte(nil), cliffVal)
}

func TestBondUnbondRedelegateSlashTwice(t *testing.T) {
	ctx, _, keeper := keep.CreateTestInput(t, false, 1000)
	valA, valB, del := keep.Addrs[0], keep.Addrs[1], keep.Addrs[2]

	msgCreateValidator := newTestMsgCreateValidator(valA, keep.PKs[0], 10)
	got := handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	msgCreateValidator = newTestMsgCreateValidator(valB, keep.PKs[1], 10)
	got = handleMsgCreateValidator(ctx, msgCreateValidator, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgCreateValidator")

	// delegate 10 stake
	msgDelegate := newTestMsgDelegate(del, valA, 10)
	got = handleMsgDelegate(ctx, msgDelegate, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgDelegate")

	// a block passes
	ctx = ctx.WithBlockHeight(1)

	// begin unbonding 4 stake
	msgBeginUnbonding := NewMsgBeginUnbonding(del, valA, sdk.NewRat(4))
	got = handleMsgBeginUnbonding(ctx, msgBeginUnbonding, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgBeginUnbonding")

	// begin redelegate 6 stake
	msgBeginRedelegate := NewMsgBeginRedelegate(del, valA, valB, sdk.NewRat(6))
	got = handleMsgBeginRedelegate(ctx, msgBeginRedelegate, keeper)
	require.True(t, got.IsOK(), "expected no error on runMsgBeginRedelegate")

	// destination delegation should have 6 shares
	delegation, found := keeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewRat(6), delegation.Shares)

	// slash the validator by half
	keeper.Slash(ctx, keep.PKs[0], 0, 20, sdk.NewRat(1, 2))

	// unbonding delegation should have been slashed by half
	unbonding, found := keeper.GetUnbondingDelegation(ctx, del, valA)
	require.True(t, found)
	require.Equal(t, int64(2), unbonding.Balance.Amount.Int64())

	// redelegation should have been slashed by half
	redelegation, found := keeper.GetRedelegation(ctx, del, valA, valB)
	require.True(t, found)
	require.Equal(t, int64(3), redelegation.Balance.Amount.Int64())

	// destination delegation should have been slashed by half
	delegation, found = keeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewRat(3), delegation.Shares)

	// validator power should have been reduced by half
	validator, found := keeper.GetValidator(ctx, valA)
	require.True(t, found)
	require.Equal(t, sdk.NewRat(5), validator.GetPower())

	// slash the validator for an infraction committed after the unbonding and redelegation begin
	ctx = ctx.WithBlockHeight(3)
	keeper.Slash(ctx, keep.PKs[0], 2, 10, sdk.NewRat(1, 2))

	// unbonding delegation should be unchanged
	unbonding, found = keeper.GetUnbondingDelegation(ctx, del, valA)
	require.True(t, found)
	require.Equal(t, int64(2), unbonding.Balance.Amount.Int64())

	// redelegation should be unchanged
	redelegation, found = keeper.GetRedelegation(ctx, del, valA, valB)
	require.True(t, found)
	require.Equal(t, int64(3), redelegation.Balance.Amount.Int64())

	// destination delegation should be unchanged
	delegation, found = keeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewRat(3), delegation.Shares)

	// validator power should have been reduced to zero
	// ergo validator should have been removed from the store
	_, found = keeper.GetValidator(ctx, valA)
	require.False(t, found)
}
