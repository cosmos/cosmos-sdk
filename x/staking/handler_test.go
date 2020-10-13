package staking_test

import (
	"strings"
	"testing"
	"time"

	gogotypes "github.com/gogo/protobuf/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func bootstrapHandlerGenesisTest(t *testing.T, power int64, numAddrs int, accAmount int64) (*simapp.SimApp, sdk.Context, []sdk.AccAddress, []sdk.ValAddress) {
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
	initBond := sdk.TokensFromConsensusPower(initPower)

	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 10, 10000000000000)

	validatorAddr, validatorAddr3 := valAddrs[0], valAddrs[1]

	handler := staking.NewHandler(app.StakingKeeper)

	// create validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], initBond)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// must end-block
	updates := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
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
	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr3, PKs[2], initBond)
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// must end-block
	updates = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
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
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, totalBond)
	msgUndelegate := types.NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)

	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	ts := &gogotypes.Timestamp{}
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

	finishTime, err := gogotypes.TimestampFromProto(ts)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(finishTime)
	staking.EndBlocker(ctx, app.StakingKeeper)
	staking.EndBlocker(ctx, app.StakingKeeper)

	// verify that by power key nolonger exists
	_, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.False(t, found)
	require.False(t, keeper.ValidatorByPowerIndexExists(ctx, app.StakingKeeper, power3))
}

func TestDuplicatesMsgCreateValidator(t *testing.T) {
	initPower := int64(1000000)
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 10, 10000000000000)

	handler := staking.NewHandler(app.StakingKeeper)

	addr1, addr2 := valAddrs[0], valAddrs[1]
	pk1, pk2 := PKs[0], PKs[1]

	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator1 := NewTestMsgCreateValidator(addr1, pk1, valTokens)
	res, err := handler(ctx, msgCreateValidator1)
	require.NoError(t, err)
	require.NotNil(t, res)

	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator, found := app.StakingKeeper.GetValidator(ctx, addr1)
	require.True(t, found)
	assert.Equal(t, types.Bonded, validator.Status)
	assert.Equal(t, addr1.String(), validator.OperatorAddress)
	assert.Equal(t, pk1.(cryptotypes.IntoTmPubKey).AsTmPubKey(), validator.GetConsPubKey())
	assert.Equal(t, valTokens, validator.BondedTokens())
	assert.Equal(t, valTokens.ToDec(), validator.DelegatorShares)
	assert.Equal(t, types.Description{}, validator.Description)

	// two validators can't have the same operator address
	msgCreateValidator2 := NewTestMsgCreateValidator(addr1, pk2, valTokens)
	res, err = handler(ctx, msgCreateValidator2)
	require.Error(t, err)
	require.Nil(t, res)

	// two validators can't have the same pubkey
	msgCreateValidator3 := NewTestMsgCreateValidator(addr2, pk1, valTokens)
	res, err = handler(ctx, msgCreateValidator3)
	require.Error(t, err)
	require.Nil(t, res)

	// must have different pubkey and operator
	msgCreateValidator4 := NewTestMsgCreateValidator(addr2, pk2, valTokens)
	res, err = handler(ctx, msgCreateValidator4)
	require.NoError(t, err)
	require.NotNil(t, res)

	// must end-block
	updates := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	validator, found = app.StakingKeeper.GetValidator(ctx, addr2)

	require.True(t, found)
	assert.Equal(t, types.Bonded, validator.Status)
	assert.Equal(t, addr2.String(), validator.OperatorAddress)
	assert.Equal(t, pk2.(cryptotypes.IntoTmPubKey).AsTmPubKey(), validator.GetConsPubKey())
	assert.True(sdk.IntEq(t, valTokens, validator.Tokens))
	assert.True(sdk.DecEq(t, valTokens.ToDec(), validator.DelegatorShares))
	assert.Equal(t, types.Description{}, validator.Description)
}

func TestInvalidPubKeyTypeMsgCreateValidator(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 1, 1000)
	handler := staking.NewHandler(app.StakingKeeper)
	ctx = ctx.WithConsensusParams(&abci.ConsensusParams{
		Validator: &tmproto.ValidatorParams{PubKeyTypes: []string{tmtypes.ABCIPubKeyTypeEd25519}},
	})

	addr := valAddrs[0]
	invalidPk := secp256k1.GenPrivKey().PubKey()

	// invalid pukKey type should not be allowed
	msgCreateValidator := NewTestMsgCreateValidator(addr, invalidPk, sdk.NewInt(10))
	res, err := handler(ctx, msgCreateValidator)
	require.Error(t, err)
	require.Nil(t, res)
}

func TestLegacyValidatorDelegations(t *testing.T) {
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 2, 100000000)
	handler := staking.NewHandler(app.StakingKeeper)

	bondAmount := sdk.TokensFromConsensusPower(10)
	valAddr := valAddrs[0]
	valConsPubKey, valConsAddr := PKs[0], sdk.ConsAddress(PKs[0].Address())
	delAddr := delAddrs[1]

	// create validator
	msgCreateVal := NewTestMsgCreateValidator(valAddr, valConsPubKey, bondAmount)
	res, err := handler(ctx, msgCreateVal)
	require.NoError(t, err)
	require.NotNil(t, res)

	// must end-block
	updates := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(updates))

	// verify the validator exists and has the correct attributes
	validator, found := app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, types.Bonded, validator.Status)
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount, validator.BondedTokens())

	// delegate tokens to the validator
	msgDelegate := NewTestMsgDelegate(delAddr, valAddr, bondAmount)
	res, err = handler(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// verify validator bonded shares
	validator, found = app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(2), validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(2), validator.BondedTokens())

	// unbond validator total self-delegations (which should jail the validator)
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, bondAmount)
	msgUndelegate := types.NewMsgUndelegate(sdk.AccAddress(valAddr), valAddr, unbondAmt)

	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	ts := &gogotypes.Timestamp{}
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

	finishTime, err := gogotypes.TimestampFromProto(ts)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(finishTime)
	staking.EndBlocker(ctx, app.StakingKeeper)

	// verify the validator record still exists, is jailed, and has correct tokens
	validator, found = app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.True(t, validator.Jailed)
	require.Equal(t, bondAmount, validator.Tokens)

	// verify delegation still exists
	bond, found := app.StakingKeeper.GetDelegation(ctx, delAddr, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount, bond.Shares.RoundInt())
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())

	// verify the validator can still self-delegate
	msgSelfDelegate := NewTestMsgDelegate(sdk.AccAddress(valAddr), valAddr, bondAmount)
	res, err = handler(ctx, msgSelfDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// verify validator bonded shares
	validator, found = app.StakingKeeper.GetValidator(ctx, valAddr)
	require.True(t, found)
	require.Equal(t, bondAmount.MulRaw(2), validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount.MulRaw(2), validator.Tokens)

	// unjail the validator now that is has non-zero self-delegated shares
	app.StakingKeeper.Unjail(ctx, valConsAddr)

	// verify the validator can now accept delegations
	msgDelegate = NewTestMsgDelegate(delAddr, valAddr, bondAmount)
	res, err = handler(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

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
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)

	params := app.StakingKeeper.GetParams(ctx)

	bondAmount := sdk.TokensFromConsensusPower(10)
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]

	// first create validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], bondAmount)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.Equal(t, types.Bonded, validator.Status)
	require.Equal(t, bondAmount, validator.DelegatorShares.RoundInt())
	require.Equal(t, bondAmount, validator.BondedTokens(), "validator: %v", validator)

	_, found = app.StakingKeeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
	require.False(t, found)

	bond, found := app.StakingKeeper.GetDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found)
	require.Equal(t, bondAmount, bond.Shares.RoundInt())

	bondedTokens := app.StakingKeeper.TotalBondedTokens(ctx)
	require.Equal(t, bondAmount.Int64(), bondedTokens.Int64())

	// just send the same msgbond multiple times
	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, bondAmount)

	for i := int64(0); i < 5; i++ {
		ctx = ctx.WithBlockHeight(i)

		res, err := handler(ctx, msgDelegate)
		require.NoError(t, err)
		require.NotNil(t, res)

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
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 1, 1000000000)

	validatorAddr := valAddrs[0]
	handler := staking.NewHandler(app.StakingKeeper)

	// create validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], initBond)
	msgCreateValidator.MinSelfDelegation = sdk.NewInt(2)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// must end-block
	updates := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
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
	res, err = handler(ctx, msgEditValidator)
	require.Error(t, err)
	require.Nil(t, res)
}

func TestEditValidatorIncreaseMinSelfDelegationBeyondCurrentBond(t *testing.T) {
	initPower := int64(100)
	initBond := sdk.TokensFromConsensusPower(100)

	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, 1000000000)
	validatorAddr := valAddrs[0]

	handler := staking.NewHandler(app.StakingKeeper)

	// create validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], initBond)
	msgCreateValidator.MinSelfDelegation = sdk.NewInt(2)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// must end-block
	updates := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
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
	res, err = handler(ctx, msgEditValidator)
	require.Error(t, err)
	require.Nil(t, res)
}

func TestIncrementsMsgUnbond(t *testing.T) {
	initPower := int64(1000)
	initBond := sdk.TokensFromConsensusPower(initPower)

	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 2, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)

	params := app.StakingKeeper.GetParams(ctx)
	denom := params.BondDenom

	// create validator, delegate
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]

	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], initBond)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// initial balance
	amt1 := app.BankKeeper.GetBalance(ctx, delegatorAddr, denom).Amount

	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, initBond)
	res, err = handler(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

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
		res, err := handler(ctx, msgUndelegate)
		require.NoError(t, err)
		require.NotNil(t, res)

		ts := &gogotypes.Timestamp{}
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

		finishTime, err := gogotypes.TimestampFromProto(ts)
		require.NoError(t, err)

		ctx = ctx.WithBlockTime(finishTime)
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

	for _, c := range errorCases {
		unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, c)
		msgUndelegate := types.NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)
		res, err = handler(ctx, msgUndelegate)
		require.Error(t, err)
		require.Nil(t, res)
	}

	leftBonded := initBond.Sub(unbondAmt.Amount.Mul(sdk.NewInt(numUnbonds)))

	// should be able to unbond remaining
	unbondAmt = sdk.NewCoin(sdk.DefaultBondDenom, leftBonded)
	msgUndelegate = types.NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err, "msgUnbond: %v\nshares: %s\nleftBonded: %s\n", msgUndelegate, unbondAmt, leftBonded)
	require.NotNil(t, res, "msgUnbond: %v\nshares: %s\nleftBonded: %s\n", msgUndelegate, unbondAmt, leftBonded)
}

func TestMultipleMsgCreateValidator(t *testing.T) {
	initPower := int64(1000)
	initTokens := sdk.TokensFromConsensusPower(initPower)
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, initPower, 3, 1000000000)

	handler := staking.NewHandler(app.StakingKeeper)

	params := app.StakingKeeper.GetParams(ctx)
	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

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
	for i, validatorAddr := range validatorAddrs {
		valTokens := sdk.TokensFromConsensusPower(10)
		msgCreateValidatorOnBehalfOf := NewTestMsgCreateValidator(validatorAddr, PKs[i], valTokens)

		res, err := handler(ctx, msgCreateValidatorOnBehalfOf)
		require.NoError(t, err)
		require.NotNil(t, res)

		// verify that the account is bonded
		validators := app.StakingKeeper.GetValidators(ctx, 100)
		require.Equal(t, (i + 1), len(validators))

		val := validators[i]
		balanceExpd := initTokens.Sub(valTokens)
		balanceGot := app.BankKeeper.GetBalance(ctx, delegatorAddrs[i], params.BondDenom).Amount

		require.Equal(t, i+1, len(validators), "expected %d validators got %d, validators: %v", i+1, len(validators), validators)
		require.Equal(t, valTokens, val.DelegatorShares.RoundInt(), "expected %d shares, got %d", 10, val.DelegatorShares)
		require.Equal(t, balanceExpd, balanceGot, "expected account to have %d, got %d", balanceExpd, balanceGot)
	}

	staking.EndBlocker(ctx, app.StakingKeeper)

	// unbond them all by removing delegation
	for i, validatorAddr := range validatorAddrs {
		_, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
		require.True(t, found)

		unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))
		msgUndelegate := types.NewMsgUndelegate(delegatorAddrs[i], validatorAddr, unbondAmt) // remove delegation
		res, err := handler(ctx, msgUndelegate)
		require.NoError(t, err)
		require.NotNil(t, res)

		ts := &gogotypes.Timestamp{}
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

		_, err = gogotypes.TimestampFromProto(ts)
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
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 50, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)
	validatorAddr, delegatorAddrs := valAddrs[0], delAddrs[1:]

	// first make a validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], sdk.NewInt(10))
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// delegate multiple parties
	for _, delegatorAddr := range delegatorAddrs {
		msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, sdk.NewInt(10))
		res, err := handler(ctx, msgDelegate)
		require.NoError(t, err)
		require.NotNil(t, res)

		// check that the account is bonded
		bond, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.True(t, found)
		require.NotNil(t, bond, "expected delegatee bond %d to exist", bond)
	}

	// unbond them all
	for _, delegatorAddr := range delegatorAddrs {
		unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
		msgUndelegate := types.NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)

		res, err := handler(ctx, msgUndelegate)
		require.NoError(t, err)
		require.NotNil(t, res)

		ts := &gogotypes.Timestamp{}
		types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

		finishTime, err := gogotypes.TimestampFromProto(ts)
		require.NoError(t, err)

		ctx = ctx.WithBlockTime(finishTime)
		staking.EndBlocker(ctx, app.StakingKeeper)

		// check that the account is unbonded
		_, found := app.StakingKeeper.GetDelegation(ctx, delegatorAddr, validatorAddr)
		require.False(t, found)
	}
}

func TestJailValidator(t *testing.T) {
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 2, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]

	// create the validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], sdk.NewInt(10))
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// bond a delegator
	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, sdk.NewInt(10))
	res, err = handler(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// unbond the validators bond portion
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgUndelegateValidator := types.NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)
	res, err = handler(ctx, msgUndelegateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	ts := &gogotypes.Timestamp{}
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

	finishTime, err := gogotypes.TimestampFromProto(ts)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(finishTime)
	staking.EndBlocker(ctx, app.StakingKeeper)

	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.Jailed, "%v", validator)

	// test that the delegator can still withdraw their bonds
	msgUndelegateDelegator := types.NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)

	res, err = handler(ctx, msgUndelegateDelegator)
	require.NoError(t, err)
	require.NotNil(t, res)

	ts = &gogotypes.Timestamp{}
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

	finishTime, err = gogotypes.TimestampFromProto(ts)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(finishTime)
	staking.EndBlocker(ctx, app.StakingKeeper)

	// verify that the pubkey can now be reused
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestValidatorQueue(t *testing.T) {
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 2, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validator
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], valTokens)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// bond a delegator
	delTokens := sdk.TokensFromConsensusPower(10)
	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, delTokens)
	res, err = handler(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	// unbond the all self-delegation to put validator in unbonding state
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, delTokens)
	msgUndelegateValidator := types.NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)
	res, err = handler(ctx, msgUndelegateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	ts := &gogotypes.Timestamp{}
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

	finishTime, err := gogotypes.TimestampFromProto(ts)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(finishTime)
	staking.EndBlocker(ctx, app.StakingKeeper)

	origHeader := ctx.BlockHeader()

	validator, found := app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonding(), "%v", validator)

	// should still be unbonding at time 6 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 6))
	staking.EndBlocker(ctx, app.StakingKeeper)

	validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonding(), "%v", validator)

	// should be in unbonded state at time 7 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 7))
	staking.EndBlocker(ctx, app.StakingKeeper)

	validator, found = app.StakingKeeper.GetValidator(ctx, validatorAddr)
	require.True(t, found)
	require.True(t, validator.IsUnbonded(), "%v", validator)
}

func TestUnbondingPeriod(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 1, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)
	validatorAddr := valAddrs[0]

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validator
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], valTokens)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin unbonding
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(10))
	msgUndelegate := types.NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	origHeader := ctx.BlockHeader()

	_, found := app.StakingKeeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// cannot complete unbonding at same time
	staking.EndBlocker(ctx, app.StakingKeeper)
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// cannot complete unbonding at time 6 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 6))
	staking.EndBlocker(ctx, app.StakingKeeper)
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.True(t, found, "should not have unbonded")

	// can complete unbonding at time 7 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 7))
	staking.EndBlocker(ctx, app.StakingKeeper)
	_, found = app.StakingKeeper.GetUnbondingDelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr)
	require.False(t, found, "should have unbonded")
}

func TestUnbondingFromUnbondingValidator(t *testing.T) {
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 2, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)
	validatorAddr, delegatorAddr := valAddrs[0], delAddrs[1]

	// create the validator
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], sdk.NewInt(10))
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// bond a delegator
	msgDelegate := NewTestMsgDelegate(delegatorAddr, validatorAddr, sdk.NewInt(10))
	res, err = handler(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// unbond the validators bond portion
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgUndelegateValidator := types.NewMsgUndelegate(sdk.AccAddress(validatorAddr), validatorAddr, unbondAmt)
	res, err = handler(ctx, msgUndelegateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// change the ctx to Block Time one second before the validator would have unbonded
	ts := &gogotypes.Timestamp{}
	types.ModuleCdc.MustUnmarshalBinaryLengthPrefixed(res.Data, ts)

	finishTime, err := gogotypes.TimestampFromProto(ts)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(finishTime.Add(time.Second * -1))

	// unbond the delegator from the validator
	msgUndelegateDelegator := types.NewMsgUndelegate(delegatorAddr, validatorAddr, unbondAmt)
	res, err = handler(ctx, msgUndelegateDelegator)
	require.NoError(t, err)
	require.NotNil(t, res)

	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(app.StakingKeeper.UnbondingTime(ctx)))

	// Run the EndBlocker
	staking.EndBlocker(ctx, app.StakingKeeper)

	// Check to make sure that the unbonding delegation is no longer in state
	// (meaning it was deleted in the above EndBlocker)
	_, found := app.StakingKeeper.GetUnbondingDelegation(ctx, delegatorAddr, validatorAddr)
	require.False(t, found, "should be removed from state")
}

func TestRedelegationPeriod(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 2, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)
	validatorAddr, validatorAddr2 := valAddrs[0], valAddrs[1]
	denom := app.StakingKeeper.GetParams(ctx).BondDenom

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 7 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validators
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], sdk.NewInt(10))

	// initial balance
	amt1 := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(validatorAddr), denom).Amount

	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// balance should have been subtracted after creation
	amt2 := app.BankKeeper.GetBalance(ctx, sdk.AccAddress(validatorAddr), denom).Amount
	require.Equal(t, amt1.Sub(sdk.NewInt(10)).Int64(), amt2.Int64(), "expected coins to be subtracted")

	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr2, PKs[1], sdk.NewInt(10))
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	bal1 := app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(validatorAddr))

	// begin redelegate
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2, redAmt)
	res, err = handler(ctx, msgBeginRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// origin account should not lose tokens as with a regular delegation
	bal2 := app.BankKeeper.GetAllBalances(ctx, sdk.AccAddress(validatorAddr))
	require.Equal(t, bal1, bal2)

	origHeader := ctx.BlockHeader()

	// cannot complete redelegation at same time
	staking.EndBlocker(ctx, app.StakingKeeper)
	_, found := app.StakingKeeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.True(t, found, "should not have unbonded")

	// cannot complete redelegation at time 6 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 6))
	staking.EndBlocker(ctx, app.StakingKeeper)
	_, found = app.StakingKeeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.True(t, found, "should not have unbonded")

	// can complete redelegation at time 7 seconds later
	ctx = ctx.WithBlockTime(origHeader.Time.Add(time.Second * 7))
	staking.EndBlocker(ctx, app.StakingKeeper)
	_, found = app.StakingKeeper.GetRedelegation(ctx, sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2)
	require.False(t, found, "should have unbonded")
}

func TestTransitiveRedelegation(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 3, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)

	validatorAddr := valAddrs[0]
	validatorAddr2 := valAddrs[1]
	validatorAddr3 := valAddrs[2]

	blockTime := time.Now().UTC()
	ctx = ctx.WithBlockTime(blockTime)

	// create the validators
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr, PKs[0], sdk.NewInt(10))
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr2, PKs[1], sdk.NewInt(10))
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr3, PKs[2], sdk.NewInt(10))
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// begin redelegate
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(10))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(sdk.AccAddress(validatorAddr), validatorAddr, validatorAddr2, redAmt)
	res, err = handler(ctx, msgBeginRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// cannot redelegation to next validator while first delegation exists
	msgBeginRedelegate = types.NewMsgBeginRedelegate(sdk.AccAddress(validatorAddr), validatorAddr2, validatorAddr3, redAmt)
	res, err = handler(ctx, msgBeginRedelegate)
	require.Error(t, err)
	require.Nil(t, res)

	params := app.StakingKeeper.GetParams(ctx)
	ctx = ctx.WithBlockTime(blockTime.Add(params.UnbondingTime))

	// complete first redelegation
	staking.EndBlocker(ctx, app.StakingKeeper)

	// now should be able to redelegate from the second validator to the third
	res, err = handler(ctx, msgBeginRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)
}

func TestMultipleRedelegationAtSameTime(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 2, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)

	valAddr := valAddrs[0]
	valAddr2 := valAddrs[1]

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 1 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validators
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valAddr, PKs[0], valTokens)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgCreateValidator = NewTestMsgCreateValidator(valAddr2, PKs[1], valTokens)
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond them
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin a redelegate
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(selfDelAddr, valAddr, valAddr2, redAmt)
	res, err = handler(ctx, msgBeginRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// there should only be one entry in the redelegation object
	rd, found := app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 1)

	// start a second redelegation at this same time as the first
	res, err = handler(ctx, msgBeginRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// now there should be two entries
	rd, found = app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 2)

	// move forward in time, should complete both redelegations
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(1 * time.Second))
	staking.EndBlocker(ctx, app.StakingKeeper)

	rd, found = app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.False(t, found)
}

func TestMultipleRedelegationAtUniqueTimes(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 2, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)

	valAddr := valAddrs[0]
	valAddr2 := valAddrs[1]

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 10 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validators
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valAddr, PKs[0], valTokens)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgCreateValidator = NewTestMsgCreateValidator(valAddr2, PKs[1], valTokens)
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond them
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin a redelegate
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(selfDelAddr, valAddr, valAddr2, redAmt)
	res, err = handler(ctx, msgBeginRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// move forward in time and start a second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	res, err = handler(ctx, msgBeginRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// now there should be two entries
	rd, found := app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 2)

	// move forward in time, should complete the first redelegation, but not the second
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	staking.EndBlocker(ctx, app.StakingKeeper)
	rd, found = app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.True(t, found)
	require.Len(t, rd.Entries, 1)

	// move forward in time, should complete the second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	staking.EndBlocker(ctx, app.StakingKeeper)
	rd, found = app.StakingKeeper.GetRedelegation(ctx, selfDelAddr, valAddr, valAddr2)
	require.False(t, found)
}

func TestMultipleUnbondingDelegationAtSameTime(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 1, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)

	valAddr := valAddrs[0]

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 1 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validator
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valAddr, PKs[0], valTokens)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin an unbonding delegation
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgUndelegate := types.NewMsgUndelegate(selfDelAddr, valAddr, unbondAmt)
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// there should only be one entry in the ubd object
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// start a second ubd at this same time as the first
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// now there should be two entries
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 2)

	// move forwaubd in time, should complete both ubds
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(1 * time.Second))
	staking.EndBlocker(ctx, app.StakingKeeper)

	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.False(t, found)
}

func TestMultipleUnbondingDelegationAtUniqueTimes(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 1, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)
	valAddr := valAddrs[0]

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.UnbondingTime = 10 * time.Second
	app.StakingKeeper.SetParams(ctx, params)

	// create the validator
	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valAddr, PKs[0], valTokens)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// end block to bond
	staking.EndBlocker(ctx, app.StakingKeeper)

	// begin an unbonding delegation
	selfDelAddr := sdk.AccAddress(valAddr) // (the validator is it's own delegator)
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens.QuoRaw(2))
	msgUndelegate := types.NewMsgUndelegate(selfDelAddr, valAddr, unbondAmt)
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// there should only be one entry in the ubd object
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// move forwaubd in time and start a second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// now there should be two entries
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 2)

	// move forwaubd in time, should complete the first redelegation, but not the second
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	staking.EndBlocker(ctx, app.StakingKeeper)
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)

	// move forwaubd in time, should complete the second redelegation
	ctx = ctx.WithBlockTime(ctx.BlockHeader().Time.Add(5 * time.Second))
	staking.EndBlocker(ctx, app.StakingKeeper)
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, selfDelAddr, valAddr)
	require.False(t, found)
}

func TestUnbondingWhenExcessValidators(t *testing.T) {
	app, ctx, _, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 3, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)

	validatorAddr1 := valAddrs[0]
	validatorAddr2 := valAddrs[1]
	validatorAddr3 := valAddrs[2]

	// set the unbonding time
	params := app.StakingKeeper.GetParams(ctx)
	params.MaxValidators = 2
	app.StakingKeeper.SetParams(ctx, params)

	// add three validators
	valTokens1 := sdk.TokensFromConsensusPower(50)
	msgCreateValidator := NewTestMsgCreateValidator(validatorAddr1, PKs[0], valTokens1)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 1, len(app.StakingKeeper.GetLastValidators(ctx)))

	valTokens2 := sdk.TokensFromConsensusPower(30)
	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr2, PKs[1], valTokens2)
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(app.StakingKeeper.GetLastValidators(ctx)))

	valTokens3 := sdk.TokensFromConsensusPower(10)
	msgCreateValidator = NewTestMsgCreateValidator(validatorAddr3, PKs[2], valTokens3)
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(app.StakingKeeper.GetLastValidators(ctx)))

	// unbond the validator-2
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, valTokens2)
	msgUndelegate := types.NewMsgUndelegate(sdk.AccAddress(validatorAddr2), validatorAddr2, unbondAmt)
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// apply TM updates
	app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)

	// because there are extra validators waiting to get in, the queued
	// validator (aka. validator-1) should make it into the bonded group, thus
	// the total number of validators should stay the same
	vals := app.StakingKeeper.GetLastValidators(ctx)
	require.Equal(t, 2, len(vals), "vals %v", vals)
	val1, found := app.StakingKeeper.GetValidator(ctx, validatorAddr1)
	require.True(t, found)
	require.Equal(t, types.Bonded, val1.Status, "%v", val1)
}

func TestBondUnbondRedelegateSlashTwice(t *testing.T) {
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 3, 1000000000)

	handler := staking.NewHandler(app.StakingKeeper)

	valA, valB, del := valAddrs[0], valAddrs[1], delAddrs[2]
	consAddr0 := sdk.ConsAddress(PKs[0].Address())

	valTokens := sdk.TokensFromConsensusPower(10)
	msgCreateValidator := NewTestMsgCreateValidator(valA, PKs[0], valTokens)
	res, err := handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgCreateValidator = NewTestMsgCreateValidator(valB, PKs[1], valTokens)
	res, err = handler(ctx, msgCreateValidator)
	require.NoError(t, err)
	require.NotNil(t, res)

	// delegate 10 stake
	msgDelegate := NewTestMsgDelegate(del, valA, valTokens)
	res, err = handler(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// apply Tendermint updates
	updates := app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))

	// a block passes
	ctx = ctx.WithBlockHeight(1)

	// begin unbonding 4 stake
	unbondAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(4))
	msgUndelegate := types.NewMsgUndelegate(del, valA, unbondAmt)
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// begin redelegate 6 stake
	redAmt := sdk.NewCoin(sdk.DefaultBondDenom, sdk.TokensFromConsensusPower(6))
	msgBeginRedelegate := types.NewMsgBeginRedelegate(del, valA, valB, redAmt)
	res, err = handler(ctx, msgBeginRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	// destination delegation should have 6 shares
	delegation, found := app.StakingKeeper.GetDelegation(ctx, del, valB)
	require.True(t, found)
	require.Equal(t, sdk.NewDecFromInt(redAmt.Amount), delegation.Shares)

	// must apply validator updates
	updates = app.StakingKeeper.ApplyAndReturnValidatorSetUpdates(ctx)
	require.Equal(t, 2, len(updates))

	// slash the validator by half
	app.StakingKeeper.Slash(ctx, consAddr0, 0, 20, sdk.NewDecWithPrec(5, 1))

	// unbonding delegation should have been slashed by half
	ubd, found := app.StakingKeeper.GetUnbondingDelegation(ctx, del, valA)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	require.Equal(t, unbondAmt.Amount.QuoRaw(2), ubd.Entries[0].Balance)

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

	// unbonding delegation should be unchanged
	ubd, found = app.StakingKeeper.GetUnbondingDelegation(ctx, del, valA)
	require.True(t, found)
	require.Len(t, ubd.Entries, 1)
	require.Equal(t, unbondAmt.Amount.QuoRaw(2), ubd.Entries[0].Balance)

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
	app, ctx, delAddrs, valAddrs := bootstrapHandlerGenesisTest(t, 1000, 3, 1000000000)
	handler := staking.NewHandler(app.StakingKeeper)

	valA, valB, delAddr := valAddrs[0], valAddrs[1], delAddrs[2]

	valTokens := sdk.TokensFromConsensusPower(100)
	invalidCoin := sdk.NewCoin("churros", valTokens)
	validCoin := sdk.NewCoin(sdk.DefaultBondDenom, valTokens)
	oneCoin := sdk.NewCoin(sdk.DefaultBondDenom, sdk.OneInt())

	commission := types.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.ZeroDec())

	msgCreate := types.NewMsgCreateValidator(valA, PKs[0], invalidCoin, types.Description{}, commission, sdk.OneInt())
	res, err := handler(ctx, msgCreate)
	require.Error(t, err)
	require.Nil(t, res)

	msgCreate = types.NewMsgCreateValidator(valA, PKs[0], validCoin, types.Description{}, commission, sdk.OneInt())
	res, err = handler(ctx, msgCreate)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgCreate = types.NewMsgCreateValidator(valB, PKs[1], validCoin, types.Description{}, commission, sdk.OneInt())
	res, err = handler(ctx, msgCreate)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgDelegate := types.NewMsgDelegate(delAddr, valA, invalidCoin)
	res, err = handler(ctx, msgDelegate)
	require.Error(t, err)
	require.Nil(t, res)

	msgDelegate = types.NewMsgDelegate(delAddr, valA, validCoin)
	res, err = handler(ctx, msgDelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgUndelegate := types.NewMsgUndelegate(delAddr, valA, invalidCoin)
	res, err = handler(ctx, msgUndelegate)
	require.Error(t, err)
	require.Nil(t, res)

	msgUndelegate = types.NewMsgUndelegate(delAddr, valA, oneCoin)
	res, err = handler(ctx, msgUndelegate)
	require.NoError(t, err)
	require.NotNil(t, res)

	msgRedelegate := types.NewMsgBeginRedelegate(delAddr, valA, valB, invalidCoin)
	res, err = handler(ctx, msgRedelegate)
	require.Error(t, err)
	require.Nil(t, res)

	msgRedelegate = types.NewMsgBeginRedelegate(delAddr, valA, valB, oneCoin)
	res, err = handler(ctx, msgRedelegate)
	require.NoError(t, err)
	require.NotNil(t, res)
}
