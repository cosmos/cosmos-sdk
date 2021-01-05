package staking_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/golang/protobuf/proto"

	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/teststaking"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func bootstrapHandlerGenesisTest(t *testing.T, power int64, numAddrs int, accAmount sdk.Int) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
	_, app, ctx := getBaseSimappWithCustomKeeper()

	addrDels, addrVals := generateAddresses(app, ctx, numAddrs, accAmount)

	amt := sdk.TokensFromConsensusPower(power)
	totalSupply := sdk.NewCoins(sdk.NewCoin(app.StakingKeeper.BondDenom(ctx), amt.MulRaw(int64(len(addrDels)))))

	notBondedPool := app.StakingKeeper.GetNotBondedPool(ctx)
	err := app.BankKeeper.SetBalances(ctx, notBondedPool.GetAddress(), totalSupply)
	require.NoError(t, err)

	app.AccountKeeper.SetModuleAccount(ctx, notBondedPool)
	app.BankKeeper.SetSupply(ctx, banktypes.NewSupply(totalSupply))

	return app, ctx, addrDels, addrVals
}

func TestValidatorByPowerIndex(t *testing.T) {
	initPower := int64(1000000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 10, sdk.TokensFromConsensusPower(initPower))
	validatorAddr, validatorAddr3 := valAddrs[0], valAddrs[1]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create validator
	initBond := tstaking.CreateValidatorWithValPower(validatorAddr, PKs[0], initPower, true)

	// must end-block
	updates, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(updates))

	// verify the self-delegation exists
	bond, found := app.StakingKeeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	gotBond := bond.Shares.RoundInt()
	require.Equal(t, initBond, gotBond)

	// verify that the by power index exists
	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	power := types.GetValidatorsByPowerIndexKey(validator)
	require.True(t, keeper.ValidatorByPowerIndexExists(ctx, app.StakingKeeper, power))

	// create a second validator keep it bonded
	tstaking.CreateValidatorWithValPower(validatorAddr3, PKs[2], initPower, true)

	// must end-block
	updates, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(updates))

	// slash and jail the first validator
	consAddr0 := sdk.ConsAddress(PKs[0].Address())
	app.StakingKeeper.Slash(ctx, consAddr0, 0, initPower, sdk.NewDecWithPrec(5, 1))
	app.StakingKeeper.Jail(ctx, consAddr0)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, types.Unbonding, validator.Status)    // ensure is unbonding
	require.Equal(t, initBond.QuoRaw(2), validator.Tokens) // ensure tokens slashed
	app.StakingKeeper.Unjail(ctx, consAddr0)

	// the old power record should have been deleted as the power changed
	require.False(t, keeper.ValidatorByPowerIndexExists(ctx, app.StakingKeeper, power))

	// but the new power record should have been created
	validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	power2 := types.GetValidatorsByPowerIndexKey(validator)
	require.True(t, keeper.ValidatorByPowerIndexExists(ctx, app.StakingKeeper, power2))

	// now the new record power index should be the same as the original record
	power3 := types.GetValidatorsByPowerIndexKey(validator)
	require.Equal(t, power2, power3)

	// unbond self-delegation
	totalBond := validator.TokensFromShares(bond.GetShares()).TruncateInt()
	res := tstaking.Undelegate(sdk.AccAddress(validatorAddr), validatorAddr, totalBond, true)

	var resData types.MsgUndelegateResponse
	err = proto.Unmarshal(res.Data, &resData)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(resData.CompletionTime)
	staking.EndBlocker(ctx, app.StakingKeeper)
	staking.EndBlocker(ctx, app.StakingKeeper)

	// verify that by power key nolonger exists
	_, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.False(t, found)
	require.False(t, keeper.ValidatorByPowerIndexExists(ctx, app.StakingKeeper, power3))
}

func TestDuplicatesMsgCreateValidator(t *testing.T) {
	initPower := int64(1000000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 10, sdk.TokensFromConsensusPower(initPower))

	addr1, addr2 := valAddrs[0], valAddrs[1]
	pk1, pk2 := PKs[0], PKs[1]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	valTokens := tstaking.CreateValidatorWithValPower(addr1, pk1, 10, true)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator := tstaking.CheckValidator(addr1, types.Bonded, false)
	assert.Equal(t, addr1.String(), validator.OperatorAddress)
	consKey, err := validator.TmConsPublicKey()
	require.NoError(t, err)
	tmPk1, err := cryptocodec.ToTmProtoPublicKey(pk1)
	require.NoError(t, err)
	assert.Equal(t, tmPk1, consKey)
	assert.Equal(t, valTokens, validator.BondedTokens())
	assert.Equal(t, valTokens.ToDec(), validator.DelegatorShares)
	assert.Equal(t, types.Description{}, validator.Description)

	// two validators can't have the same operator address
	tstaking.CreateValidator(addr1, pk2, valTokens, false)

	// two validators can't have the same pubkey
	tstaking.CreateValidator(addr2, pk1, valTokens, false)

	// must have different pubkey and operator
	tstaking.CreateValidator(addr2, pk2, valTokens, true)

	// must end-block
	updates, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(updates))

	validator = tstaking.CheckValidator(addr2, types.Bonded, false)
	assert.Equal(t, addr2.String(), validator.OperatorAddress)
	consPk, err := validator.TmConsPublicKey()
	require.NoError(t, err)
	tmPk2, err := cryptocodec.ToTmProtoPublicKey(pk2)
	require.NoError(t, err)
	assert.Equal(t, tmPk2, consPk)
	assert.True(sdk.IntEq(t, valTokens, validator.Tokens))
	assert.True(sdk.DecEq(t, valTokens.ToDec(), validator.DelegatorShares))
	assert.Equal(t, types.Description{}, validator.Description)
}

func TestInvalidPubKeyTypeMsgCreateValidator(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 1, sdk.TokensFromConsensusPower(initPower))
	ctx = ctx.WithConsensusParams(&abci.ConsensusParams{
		Validator: &tmproto.ValidatorParams{PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519}},
	})

	addr := valAddrs[0]
	invalidPk := secp256k1.GenPrivKey().PubKey()
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// invalid pukKey type should not be allowed
	tstaking.CreateValidator(addr, invalidPk, sdk.NewInt(10), false)
}

func TestBothPubKeyTypesMsgCreateValidator(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 2, sdk.NewInt(1000))
	ctx = ctx.WithConsensusParams(&abci.ConsensusParams{
		Validator: &tmproto.ValidatorParams{PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519, tmtypes.ABCIPubKeyTypeSecp256k1}},
	})

	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	testCases := []struct {
		name string
		addr sdk.ValAddress
		pk   cryptotypes.PubKey
	}{
		{
			"can create a validator with ed25519 pubkey",
			valAddrs[0],
			ed25519.GenPrivKey().PubKey(),
		},
		{
			"can create a validator with secp256k1 pubkey",
			valAddrs[1],
			secp256k1.GenPrivKey().PubKey(),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(*testing.T) {
			tstaking.CreateValidator(tc.addr, tc.pk, sdk.NewInt(10), true)
		})
	}
}

func TestLegacyValidatorDelegations(t *testing.T) {
	initPower := int64(1000)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))

	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	valAddr := valAddrs[0]
	valConsPubKey, valConsAddr := PKs[0], sdk.ConsAddress(PKs[0].Address())
	delAddr := delAddrs[1]

	// create validator
	bondAmount := tstaking.CreateValidatorWithValPower(valAddr, valConsPubKey, 10, true)

	// must end-block
	updates, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(updates))

	// verify the validator exists and has the correct attributes
	validator := tstaking.CheckValidator(valAddr, types.Bonded, false)
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount, validator.BondedTokens())

	// delegate tokens to the validator
	tstaking.Delegate(delAddr, valAddr, bondAmount)

	// verify validator bonded shares
	validator = tstaking.CheckValidator(valAddr, types.Bonded, false)
	require.Equal(t, bondAmount.MulRaw(2), validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(2), validator.BondedTokens())

	// unbond validator total self-delegations (which should jail the validator)
	res := tstaking.Undelegate(sdk.AccAddress(valAddr), valAddr, bondAmount, true)

	var resData types.MsgUndelegateResponse
	err = proto.Unmarshal(res.Data, &resData)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(resData.CompletionTime)
	tstaking.Ctx = ctx
	staking.EndBlocker(ctx, app.StakingKeeper)

	// verify the validator record still exists, is jailed, and has correct tokens
	validator = tstaking.CheckValidator(valAddr, -1, true)
	require.Equal(t, bondAmount, validator.Tokens)

	// verify delegation still exists
	bond, found := app.StakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount, bond.Shares.RoundInt())
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())

	// verify the validator can still self-delegate
	tstaking.Delegate(sdk.AccAddress(valAddr), valAddr, bondAmount)

	// verify validator bonded shares
	validator, found = app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(2), validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(2), validator.Tokens)

	// unjail the validator now that is has non-zero self-delegated shares
	app.StakingKeeper.Unjail(ctx, valConsAddr)

	// verify the validator can now accept delegations
	tstaking.Delegate(delAddr, valAddr, bondAmount)

	// verify validator bonded shares
	validator, found = app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(3), validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(3), validator.Tokens)

	// verify new delegation
	bond, found = app.StakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(2), bond.Shares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(3), validator.DelegatorShares.RoundInt())
}

func TestIncrementsMsgDelegate(t *testing.T) {
	initPower := int64(1000)
	initBond := sdk.TokensFromConsensusPower(initPower)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))

	params := app.StakingKeeper.GetParams(ctx)
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// first create validator
	bondAmount := tstaking.CreateValidatorWithValPower(validatorAddr, PKs[0], 10, true)

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator := tstaking.CheckValidator(validatorAddr, types.Bonded, false)
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount, validator.BondedTokens(), "validator: %v", validator)

	tstaking.CheckDelegator(delegatorAddr, validatorAddr, false)

	bond, found := app.StakingKeeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	require.Equal(t, bondAmount, bond.Shares.RoundInt())

	bondedTokens := app.StakingKeeper.TotalBondedTokens(ctx)
	require.Equal(t, bondAmount, bondedTokens)

	for i := int64(0); i < 5; i++ {
		ctx = ctx.WithBlockHeight(i)
		tstaking.Ctx = ctx
		tstaking.Delegate(delegatorAddr, validatorAddr, bondAmount)

		//Check that the accounts and the bond account have the appropriate values
		validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		bond, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)

		expBond := bondAmount.MulRaw(i + 1)
		expDelegatorShares := bondAmount.MulRaw(i + 2) // (1 self delegation)
		expDelegatorAcc := initBond.Sub(expBond)

		gotBond := bond.Shares.RoundInt()
		gotDelegatorShares := validator.DelegatorShares.RoundInt()
		gotDelegatorAcc := app.BankKeeper.GetBalance(ctx, delegatorAddr, params.BondDenom).Amount

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
	initPower := int64(100)
	initBond := sdk.TokensFromConsensusPower(100)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 1, sdk.TokensFromConsensusPower(initPower))

	validatorAddr := valAddrs[0]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create validator
	msgCreateValidator := tstaking.CreateValidatorMsg(validatorAddr, PKs[0], initBond)
	msgCreateValidator.MinSelfDelegation = sdk.NewInt(2)
	tstaking.Handle(msgCreateValidator, true)

	// must end-block
	updates, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(updates))

	// verify the self-delegation exists
	bond, found := app.StakingKeeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	gotBond := bond.Shares.RoundInt()
	require.Equal(t, initBond, gotBond,
		"initBond: %v\ngotBond: %v\nbond: %v\n",
		initBond, gotBond, bond)

	newMinSelfDelegation := sdk.OneInt()
	msgEditValidator := types.NewMsgEditValidator(validatorAddr, types.Description{}, nil, &newMinSelfDelegation)
	tstaking.Handle(msgEditValidator, false)
}

func TestEditValidatorIncreaseMinSelfDelegationBeyondCurrentBond(t *testing.T) {
	initPower := int64(100)
	initBond := sdk.TokensFromConsensusPower(100)

	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))
	validatorAddr := valAddrs[0]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create validator
	msgCreateValidator := tstaking.CreateValidatorMsg(validatorAddr, PKs[0], initBond)
	msgCreateValidator.MinSelfDelegation = sdk.NewInt(2)
	tstaking.Handle(msgCreateValidator, true)

	// must end-block
	updates, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 1, len(updates))

	// verify the self-delegation exists
	bond, found := app.StakingKeeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	gotBond := bond.Shares.RoundInt()
	require.Equal(t, initBond, gotBond,
		"initBond: %v\ngotBond: %v\nbond: %v\n",
		initBond, gotBond, bond)

	newMinSelfDelegation := initBond.Add(sdk.OneInt())
	msgEditValidator := types.NewMsgEditValidator(validatorAddr, types.Description{}, nil, &newMinSelfDelegation)
	tstaking.Handle(msgEditValidator, false)
}

func TestIncrementsMsgUnbond(t *testing.T) {
	initPower := int64(1000)

	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	params := app.StakingKeeper.GetParams(ctx)
	denom := params.BondDenom

	// create validator, delegate
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]
	initBond := tstaking.CreateValidatorWithValPower(validatorAddr, PKs[0], initPower, true)

	// initial balance
	amt1 := app.BankKeeper.GetBalance(ctx, delegatorAddr, denom).Amount

	tstaking.Delegate(delegatorAddr, validatorAddr, initBond)

	// balance should have been subtracted after delegation
	amt2 := app.BankKeeper.GetBalance(ctx, delegatorAddr, denom).Amount
	require.True(sdk.IntEq(t, amt1.Sub(initBond), amt2))

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, initBond.MulRaw(2), validator.DelegatorShares.RoundInt())
	require.Equal(t, initBond.MulRaw(2), validator.BondedTokens())

	// just send the same msgUnbond multiple times
	// TODO use decimals here
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgUndelegate := types.NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)
	numUnbonds := int64(5)

	for i := int64(0); i < numUnbonds; i++ {
		res := tstaking.Handle(msgUndelegate, true)

		var resData types.MsgUndelegateResponse
		err := proto.Unmarshal(res.Data, &resData)
		require.NoError(t, err)

		ctx = ctx.WithBlockTime(resData.CompletionTime)
		tstaking.Ctx = ctx
		staking.EndBlocker(ctx, app.StakingKeeper)

		// check that the accounts and the bond account have the appropriate values
		validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)
		bond, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)

		expBond := initBond.Sub(unbondAmt.Amount.Mul(sdk.NewInt(i + 1)))
		expDelegatorShares := initBond.MulRaw(2).Sub(unbondAmt.Amount.Mul(sdk.NewInt(i + 1)))
		expDelegatorAcc := initBond.Sub(expBond)

		gotBond := bond.Shares.RoundInt()
		gotDelegatorShares := validator.DelegatorShares.RoundInt()
		gotDelegatorAcc := app.BankKeeper.GetBalance(ctx, delegatorAddr, params.BondDenom).Amount

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
	errorCases := []sdk.Int{
		//1<<64 - 1, // more than int64 power
		//1<<63 + 1, // more than int64 power
		sdk.TokensFromConsensusPower(1<<63 - 1),
		sdk.TokensFromConsensusPower(1 << 31),
		initBond,
	}

	for _, c := range errorCases {
		tstaking.Undelegate(delegatorAddr, validatorAddr, c, false)
	}

	// should be able to unbond remaining
	leftBonded := initBond.Sub(unbondAmt.Amount.Mul(sdk.NewInt(numUnbonds)))
	tstaking.Undelegate(delegatorAddr, validatorAddr, leftBonded, true)
}

func TestMultipleMsgCreateValidator(t *testing.T) {
	initPower := int64(1000)
	initTokens := sdk.TokensFromConsensusPower(initPower)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, sdk.TokensFromConsensusPower(initPower))

	params := app.StakingKeeper.GetParams(ctx)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	validatorAddrs := []sdk.ValAddress{
		valAddrs[0],
		valAddrs[1],
		valAddrs[2],
	}
	delegatorAddrs := []sdk.AccAddress{
		delAddrs[0],
		delAddrs[1],
		delAddrs[2],
	}

	// bond them all
	amt := sdk.TokensFromConsensusPower(10)
	for i, validatorAddr := range validatorAddrs {
		tstaking.CreateValidator(validatorAddr, PKs[i], amt, true)
		// verify that the account is bonded
		validators := app.StakingKeeper.GetValidators(ctx, 100)
		require.Equal(t, (i + 1), len(validators))

		val := validators[i]
		balanceExpd := initTokens.Sub(amt)
		balanceGot := app.BankKeeper.GetBalance(ctx, delegatorAddrs[i], params.BondDenom).Amount

		require.Equal(t, i+1, len(validators), "expected %d validators got %d, validators: %v", i+1, len(validators), validators)
		require.Equal(t, amt, val.DelegatorShares.RoundInt(), "expected %d shares, got %d", amt, val.DelegatorShares)
		require.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	staking.EndBlocker(ctx, app.StakingKeeper)

	// unbond them all by removing delegation
	for i, validatorAddr := range validatorAddrs {
		_, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)

		res := tstaking.Undelegate(delegatorAddrs[i], validatorAddr, amt, true)

		var resData types.MsgUndelegateResponse
		err := proto.Unmarshal(res.Data, &resData)
		require.NoError(t, err)

		// adds validator into unbonding queue
		staking.EndBlocker(ctx, app.StakingKeeper)

		// removes validator from queue and set
		staking.EndBlocker(ctx.WithBlockTime(blockTime.Add(params.UnbondingTime)), app.StakingKeeper)

		// Check that the validator is deleted from state
		validators := app.StakingKeeper.GetValidators(ctx, 100)
		require.Equal(t, len(validatorAddrs)-(i+1), len(validators),
			"expected %d validators got %d", len(validatorAddrs)-(i+1), len(validators))

		_, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
		require.False(t, found)

		gotBalance := app.BankKeeper.GetBalance(ctx, delegatorAddrs[i], params.BondDenom).Amount
		require.Equal(t, initTokens, gotBalance, "expected account to have %d, got %d", initTokens, gotBalance)
	}
}

func TestMultipleMsgDelegate(t *testing.T) {
	initPower := int64(1000)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 50, sdk.TokensFromConsensusPower(initPower))
	validatorAddr, delegatorAddrs := valAddrs[0], delAddrs[1:]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	var amount int64 = 10

	// first make a validator
	tstaking.CreateValidator(validatorAddr, PKs[0], sdk.NewInt(amount), true)

	// delegate multiple parties
	for _, delegatorAddr := range delegatorAddrs {
		tstaking.Delegate(delegatorAddr, validatorAddr, sdk.NewInt(10))
		tstaking.CheckDelegator(delegatorAddr, validatorAddr, true)
	}

	// unbond them all
	for _, delegatorAddr := range delegatorAddrs {
		res := tstaking.Undelegate(delegatorAddr, validatorAddr, sdk.NewInt(amount), true)

		var resData types.MsgUndelegateResponse
		err := proto.Unmarshal(res.Data, &resData)
		require.NoError(t, err)

		ctx = ctx.WithBlockTime(resData.CompletionTime)
		staking.EndBlocker(ctx, app.StakingKeeper)
		tstaking.Ctx = ctx

		// check that the account is unbonded
		_, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.False(t, found)
	}
}

func TestJailValidator(t *testing.T) {
	initPower := int64(1000)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)
	var amt int64 = 10

	// create the validator and delegate
	tstaking.CreateValidator(validatorAddr, PKs[0], sdk.NewInt(amt), true)
	tstaking.Delegate(delegatorAddr, validatorAddr, sdk.NewInt(amt))

	// unbond the validators bond portion
	unamt := sdk.NewInt(amt)
	res := tstaking.Undelegate(sdk.AccAddress(validatorAddr), validatorAddr, unamt, true)

	var resData types.MsgUndelegateResponse
	err := proto.Unmarshal(res.Data, &resData)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(resData.CompletionTime)
	staking.EndBlocker(ctx, app.StakingKeeper)
	tstaking.Ctx = ctx

	tstaking.CheckValidator(validatorAddr, -1, true)

	// test that the delegator can still withdraw their bonds
	tstaking.Undelegate(delegatorAddr, validatorAddr, unamt, true)

	err = proto.Unmarshal(res.Data, &resData)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(resData.CompletionTime)
	staking.EndBlocker(ctx, app.StakingKeeper)
	tstaking.Ctx = ctx

	// verify that the pubkey can now be reused
	tstaking.CreateValidator(validatorAddr, PKs[0], sdk.NewInt(amt), true)
}

func TestValidatorQueue(t *testing.T) {
	initPower := int64(1000)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validator and make a bond
	amt := tstaking.CreateValidatorWithValPower(validatorAddr, PKs[0], 10, true)
	tstaking.Delegate(delegatorAddr, validatorAddr, amt)
	staking.EndBlocker(ctx, app.StakingKeeper)

	// unbond the all self-delegation to put validator in unbonding state
	res := tstaking.Undelegate(sdk.AccAddress(validatorAddr), validatorAddr, amt, true)

	var resData types.MsgUndelegateResponse
	err := proto.Unmarshal(res.Data, &resData)
	require.NoError(t, err)

	finishTime := resData.CompletionTime

	ctx = tstaking.TurnBlock(finishTime)
	origHeader := ctx.BlockHeader()

	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonding(), "%v", validator)

	// should still be unbonding at time 6 seconds later
	ctx = tstaking.TurnBlock(origHeader.Time.Add(time.Second * 6))

	validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonding(), "%v", validator)

	// should be in unbonded state at time 7 seconds later
	ctx = tstaking.TurnBlock(origHeader.Time.Add(time.Second * 7))

	validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonded(), "%v", validator)
}

func TestUnbondingPeriod(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 1, sdk.TokensFromConsensusPower(initPower))
	validatorAddr := valAddrs[0]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validator
	amt := tstaking.CreateValidatorWithValPower(validatorAddr, PKs[0], 10, true)
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin unbonding
	tstaking.Undelegate(sdk.AccAddress(validatorAddr), validatorAddr, amt, true)

	origHeader := ctx.BlockHeader()

	_, found := app.StakingKeeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// cannot complete unbonding at same time
	staking.EndBlocker(ctx, app.StakingKeeper)
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// cannot complete unbonding at time 6 seconds later
	ctx = tstaking.TurnBlock(origHeader.Time.Add(time.Second * 6))
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// can complete unbonding at time 7 seconds later
	ctx = tstaking.TurnBlock(origHeader.Time.Add(time.Second * 7))
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.False(t, found, "should have unbonded")
}

func TestUnbondingFromUnbondingValidator(t *testing.T) {
	initPower := int64(1000)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create the validator and delegate
	tstaking.CreateValidator(validatorAddr, PKs[0], sdk.NewInt(10), true)
	tstaking.Delegate(delegatorAddr, validatorAddr, sdk.NewInt(10))

	// unbond the validators bond portion
	unbondAmt := sdk.NewInt(10)
	res := tstaking.Undelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt, true)

	// change the ctx to Block Time one second before the validator would have unbonded
	var resData types.MsgUndelegateResponse
	err := proto.Unmarshal(res.Data, &resData)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(resData.CompletionTime.Add(time.Second * -1))

	// unbond the delegator from the validator
	res = tstaking.Undelegate(delegatorAddr, validatorAddr, unbondAmt, true)

	ctx = tstaking.TurnBlockTimeDiff(app.StakingKeeper.UnbondingTime(ctx))
	tstaking.Ctx = ctx

	// Check to make sure that the unbonding delegation is no longer in state
	// (meaning it was deleted in the above EndBlocker)
	_, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	require.False(t, found, "should be removed from state")
}

func TestRedelegationPeriod(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))
	validatorAddr, validatorAddr2 := valAddrs[0], valAddrs[1]
	denom := app.StakingKeeper.GetParams(ctx).BondDenom
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	app.StakingKeeper.SetParams(ctx, params)
	// initial balance
	amt1 := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(validatorAddr), denom).Amount

	// create the validators
	tstaking.CreateValidator(validatorAddr, PKs[0], sdk.NewInt(10), true)

	// balance should have been subtracted after creation
	amt2 := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(validatorAddr), denom).Amount
	require.Equal(t, amt1.Sub(sdk.NewInt(10)), amt2, "expected coins to be subtracted")

	tstaking.CreateValidator(validatorAddr2, PKs[1], sdk.NewInt(10), true)
	bal1 := app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(validatorAddr))

	// begin redelegate
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2, redAmt)
	tstaking.Handle(msgBeginRedelegate, true)

	// origin account should not lose tokens as with a regular delegation
	bal2 := app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(validatorAddr))
	require.Equal(t, bal1, bal2)

	origHeader := ctx.BlockHeader()

	// cannot complete redelegation at same time
	staking.EndBlocker(ctx, app.StakingKeeper)
	_, found := app.StakingKeeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.True(t, found, "should not have unbonded")

	// cannot complete redelegation at time 6 seconds later
	ctx = tstaking.TurnBlock(origHeader.Time.Add(time.Second * 6))
	_, found = app.StakingKeeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.True(t, found, "should not have unbonded")

	// can complete redelegation at time 7 seconds later
	ctx = tstaking.TurnBlock(origHeader.Time.Add(time.Second * 7))
	_, found = app.StakingKeeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.False(t, found, "should have unbonded")
}

func TestTransitiveRedelegation(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, sdk.TokensFromConsensusPower(initPower))

	val1, val2, val3 := valAddrs[0], valAddrs[1], valAddrs[2]
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// create the validators
	tstaking.CreateValidator(val1, PKs[0], sdk.NewInt(10), true)
	tstaking.CreateValidator(val2, PKs[1], sdk.NewInt(10), true)
	tstaking.CreateValidator(val3, PKs[2], sdk.NewInt(10), true)

	// begin redelegate
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(sdk.AccAddress(val1), val1, val2, redAmt)
	tstaking.Handle(msgBeginRedelegate, true)

	// cannot redelegation to next validator while first delegation exists
	msgBeginRedelegate = types.NewMsgBeginRedelegate(sdk.AccAddress(val1), val2, val3, redAmt)
	tstaking.Handle(msgBeginRedelegate, false)

	params := app.StakingKeeper.GetParams(ctx)
	ctx = ctx.WithBlockTime(blockTime.Add(params.UnbondingTime))
	tstaking.Ctx = ctx

	// complete first redelegation
	staking.EndBlocker(ctx, app.StakingKeeper)

	// now should be able to redelegate from the second validator to the third
	tstaking.Handle(msgBeginRedelegate, true)
}

func TestMultipleRedelegationAtSameTime(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))
	valAddr := valAddrs[0]
	valAddr2 := valAddrs[1]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 1 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validators
	valTokens := tstaking.CreateValidatorWithValPower(valAddr, PKs[0], 10, true)
	tstaking.CreateValidator(valAddr2, PKs[1], valTokens, true)

	// end block to bond them
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin a redelegate
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(selfDelAddr, valAddr, valAddr2, redAmt)
	tstaking.Handle(msgBeginRedelegate, true)

	// there should only be one entry in the redelegation object
	rd, found := app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 1)

	// start a second redelegation at this same time as the first
	tstaking.Handle(msgBeginRedelegate, true)

	// now there should be two entries
	rd, found = app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 2)

	// move forward in time, should complete both redelegations
	ctx = tstaking.TurnBlockTimeDiff(1 * time.Second)
	rd, found = app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.False(t, found)
}

func TestMultipleRedelegationAtUniqueTimes(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, sdk.TokensFromConsensusPower(initPower))
	valAddr := valAddrs[0]
	valAddr2 := valAddrs[1]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 10 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validators
	valTokens := tstaking.CreateValidatorWithValPower(valAddr, PKs[0], 10, true)
	tstaking.CreateValidator(valAddr2, PKs[1], valTokens, true)

	// end block to bond them
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin a redelegate
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(selfDelAddr, valAddr, valAddr2, redAmt)
	tstaking.Handle(msgBeginRedelegate, true)

	// move forward in time and start a second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	tstaking.Ctx = ctx
	tstaking.Handle(msgBeginRedelegate, true)

	// now there should be two entries
	rd, found := app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 2)

	// move forward in time, should complete the first redelegation, but not the second
	ctx = tstaking.TurnBlockTimeDiff(5 * time.Second)
	rd, found = app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 1)

	// move forward in time, should complete the second redelegation
	ctx = tstaking.TurnBlockTimeDiff(5 * time.Second)
	rd, found = app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.False(t, found)
}

func TestMultipleUnbondingDelegationAtSameTime(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 1, sdk.TokensFromConsensusPower(initPower))
	valAddr := valAddrs[0]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 1 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validators
	valTokens := tstaking.CreateValidatorWithValPower(valAddr, PKs[0], 10, true)

	// end block to bond
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin an unbonding delegation
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	tstaking.Undelegate(selfDelAddr, valAddr, valTokens.QuoRaw(2), true)

	// there should only be one entry in the ubd object
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// start a second ubd at this same time as the first
	tstaking.Undelegate(selfDelAddr, valAddr, valTokens.QuoRaw(2), true)

	// now there should be two entries
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 2)

	// move forwaubd in time, should complete both ubds
	ctx = tstaking.TurnBlockTimeDiff(1 * time.Second)
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.False(t, found)
}

func TestMultipleUnbondingDelegationAtUniqueTimes(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 1, sdk.TokensFromConsensusPower(initPower))
	valAddr := valAddrs[0]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 10 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validator
	valTokens := tstaking.CreateValidatorWithValPower(valAddr, PKs[0], 10, true)

	// end block to bond
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin an unbonding delegation
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	tstaking.Undelegate(selfDelAddr, valAddr, valTokens.QuoRaw(2), true)

	// there should only be one entry in the ubd object
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// move forwaubd in time and start a second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	tstaking.Ctx = ctx
	tstaking.Undelegate(selfDelAddr, valAddr, valTokens.QuoRaw(2), true)

	// now there should be two entries
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 2)

	// move forwaubd in time, should complete the first redelegation, but not the second
	ctx = tstaking.TurnBlockTimeDiff(5 * time.Second)
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// move forwaubd in time, should complete the second redelegation
	ctx = tstaking.TurnBlockTimeDiff(5 * time.Second)
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.False(t, found)
}

func TestUnbondingWhenExcessValidators(t *testing.T) {
	initPower := int64(1000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, sdk.TokensFromConsensusPower(initPower))
	val1 := valAddrs[0]
	val2 := valAddrs[1]
	val3 := valAddrs[2]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = 2
	app.StakingKeeper.SetParams(ctx, params)

	// add three validators
	tstaking.CreateValidatorWithValPower(val1, PKs[0], 50, true)
	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(app.StakingKeeper.GetLastValidators(ctx)))

	valTokens2 := tstaking.CreateValidatorWithValPower(val2, PKs[1], 30, true)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(app.StakingKeeper.GetLastValidators(ctx)))

	tstaking.CreateValidatorWithValPower(val3, PKs[2], 10, true)
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(app.StakingKeeper.GetLastValidators(ctx)))

	// unbond the validator-2
	tstaking.Undelegate(sdk.AccAddress(val2), val2, valTokens2, true)
	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	// because there are extra validators waiting to get in, the queued
	// validator (aka. validator-1) should make it into the bonded group, thus
	// the total number of validators should stay the same
	vals := app.StakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 2, len(vals), "vals %v", vals)
	tstaking.CheckValidator(val1, types.Bonded, false)
}

func TestBondUnbondRedelegateSlashTwice(t *testing.T) {
	initPower := int64(1000)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, sdk.TokensFromConsensusPower(initPower))
	valA, valB, del := valAddrs[0], valAddrs[1], delAddrs[2]
	consAddr0 := sdk.ConsAddress(PKs[0].Address())
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	valTokens := tstaking.CreateValidatorWithValPower(valA, PKs[0], 10, true)
	tstaking.CreateValidator(valB, PKs[1], valTokens, true)

	// delegate 10 stake
	tstaking.Delegate(del, valA, valTokens)

	// apply Tendermint updates
	updates, err := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(updates))

	// a block passes
	ctx = ctx.WithBlockHeight(1)
	tstaking.Ctx = ctx

	// begin unbonding 4 stake
	unbondAmt := sdk.TokensFromConsensusPower(4)
	tstaking.Undelegate(del, valA, unbondAmt, true)

	// begin redelegate 6 stake
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(6))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(del, valA, valB, redAmt)
	tstaking.Handle(msgBeginRedelegate, true)

	// destination delegation should have 6 shares
	delegation, found := app.StakingKeeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewDecFromInt(redAmt.Amount), delegation.Shares)

	// must apply validator updates
	updates, err = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.NoError(t, err)
	require.Equal(t, 2, len(updates))

	// slash the validator by half
	app.StakingKeeper.Slash(ctx, consAddr0, 0, 20, sdk.NewDecWithPrec(5, 1))

	// unbonding delegation should have been slashed by half
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, del, valA)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	require.Equal(t, unbondAmt.QuoRaw(2), ubd.Entries[0].Balance)

	// redelegation should have been slashed by half
	redelegation, found := app.StakingKeeper.GetRedelegation(ctx, del, valA, valB)
	require.True(t, found)
	require.Len(t, redelegation.Entries, 1)

	// destination delegation should have been slashed by half
	delegation, found = app.StakingKeeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewDecFromInt(redAmt.Amount.QuoRaw(2)), delegation.Shares)

	// validator power should have been reduced by half
	validator, found := app.StakingKeeper.GetValidator(ctx, valA)
	require.True(t, found)
	require.Equal(t, valTokens.QuoRaw(2), validator.GetBondedTokens())

	// slash the validator for an infraction committed after the unbonding and redelegation begin
	ctx = ctx.WithBlockHeight(3)
	app.StakingKeeper.Slash(ctx, consAddr0, 2, 10, sdk.NewDecWithPrec(5, 1))
	tstaking.Ctx = ctx

	// unbonding delegation should be unchanged
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, del, valA)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	require.Equal(t, unbondAmt.QuoRaw(2), ubd.Entries[0].Balance)

	// redelegation should be unchanged
	redelegation, found = app.StakingKeeper.GetRedelegation(ctx, del, valA, valB)
	require.True(t, found)
	require.Len(t, redelegation.Entries, 1)

	// destination delegation should be unchanged
	delegation, found = app.StakingKeeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewDecFromInt(redAmt.Amount.QuoRaw(2)), delegation.Shares)

	// end blocker
	staking.EndBlocker(ctx, app.StakingKeeper)

	// validator power should have been reduced to zero
	// validator should be in unbonding state
	validator, _ = app.StakingKeeper.GetValidator(ctx, valA)
	require.Equal(t, validator.GetStatus(), types.Unbonding)
}

func TestInvalidMsg(t *testing.T) {
	k := keeper.Keeper{}
	h := staking.NewHandler(k)

	res, err := h(sdk.NewContext(nil, tmproto.Header{}, false, nil), testdata.NewTestMsg())
	require.Error(t, err)
	require.Nil(t, res)
	require.True(t, strings.Contains(err.Error(), "unrecognized staking message type"))
}

func TestInvalidCoinDenom(t *testing.T) {
	initPower := int64(1000)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, sdk.TokensFromConsensusPower(initPower))
	valA, valB, delAddr := valAddrs[0], valAddrs[1], delAddrs[2]
	tstaking := teststaking.NewHelper(t, ctx, app.StakingKeeper)

	valTokens := sdk.TokensFromConsensusPower(100)
	invalidCoin := sdk.NewCoin("churros", valTokens)
	validCoin := sdk.NewCoin(sdk.DefaultBondDenom, valTokens)
	oneCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())

	commission := types.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.ZeroDec())
	msgCreate, err := types.NewMsgCreateValidator(valA, PKs[0], invalidCoin, types.Description{}, commission, sdk.OneInt())
	require.NoError(t, err)
	tstaking.Handle(msgCreate, false)

	msgCreate, err = types.NewMsgCreateValidator(valA, PKs[0], validCoin, types.Description{}, commission, sdk.OneInt())
	require.NoError(t, err)
	tstaking.Handle(msgCreate, true)

	msgCreate, err = types.NewMsgCreateValidator(valB, PKs[1], validCoin, types.Description{}, commission, sdk.OneInt())
	require.NoError(t, err)
	tstaking.Handle(msgCreate, true)

	msgDelegate := types.NewMsgDelegate(delAddr, valA, invalidCoin)
	tstaking.Handle(msgDelegate, false)

	msgDelegate = types.NewMsgDelegate(delAddr, valA, validCoin)
	tstaking.Handle(msgDelegate, true)

	msgUndelegate := types.NewMsgUndelegate(delAddr, valA, invalidCoin)
	tstaking.Handle(msgUndelegate, false)

	msgUndelegate = types.NewMsgUndelegate(delAddr, valA, oneCoin)
	tstaking.Handle(msgUndelegate, true)

	msgRedelegate := types.NewMsgBeginRedelegate(delAddr, valA, valB, invalidCoin)
	tstaking.Handle(msgRedelegate, false)

	msgRedelegate = types.NewMsgBeginRedelegate(delAddr, valA, valB, oneCoin)
	tstaking.Handle(msgRedelegate, true)
}
