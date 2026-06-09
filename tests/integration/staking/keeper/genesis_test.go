package keeper_test

import (
	"bytes"
	"fmt"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/stretchr/testify/require"
	"gotest.tools/v3/assert"

	"cosmossdk.io/math"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktestutil "github.com/cosmos/cosmos-sdk/x/bank/testutil"
	"github.com/cosmos/cosmos-sdk/x/staking"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func bootstrapGenesisTest(t *testing.T, numAddrs int) (*fixture, []sdk.AccAddress) {
	t.Helper()
	t.Parallel()

	f := initFixture(t)

	addrDels, _ := generateAddresses(f, numAddrs)
	return f, addrDels
}

func TestInitGenesis(t *testing.T) {
	f, addrs := bootstrapGenesisTest(t, 10)

	valTokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1)

	pk0, err := codectypes.NewAnyWithValue(PKs[0])
	assert.NilError(t, err)

	bondedVal := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[0]).String(),
		ConsensusPubkey: pk0,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: math.LegacyNewDecFromInt(valTokens),
		Description:     types.NewDescription("hoop", "", "", "", ""),
	}
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, bondedVal))

	params, err := f.stakingKeeper.GetParams(f.sdkCtx)
	assert.NilError(t, err)

	validators, err := f.stakingKeeper.GetAllValidators(f.sdkCtx)
	assert.NilError(t, err)

	assert.Assert(t, len(validators) == 1)
	var delegations []types.Delegation

	pk1, err := codectypes.NewAnyWithValue(PKs[1])
	assert.NilError(t, err)

	pk2, err := codectypes.NewAnyWithValue(PKs[2])
	assert.NilError(t, err)

	// initialize the validators
	bondedVal1 := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[1]).String(),
		ConsensusPubkey: pk1,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: math.LegacyNewDecFromInt(valTokens),
		Description:     types.NewDescription("hoop", "", "", "", ""),
	}
	bondedVal2 := types.Validator{
		OperatorAddress: sdk.ValAddress(addrs[2]).String(),
		ConsensusPubkey: pk2,
		Status:          types.Bonded,
		Tokens:          valTokens,
		DelegatorShares: math.LegacyNewDecFromInt(valTokens),
		Description:     types.NewDescription("bloop", "", "", "", ""),
	}

	// append new bonded validators to the list
	validators = append(validators, bondedVal1, bondedVal2)

	// mint coins in the bonded pool representing the validators coins
	i2 := len(validators)
	assert.NilError(t,
		banktestutil.FundModuleAccount(
			f.sdkCtx,
			f.bankKeeper,
			types.BondedPoolName,
			sdk.NewCoins(
				sdk.NewCoin(params.BondDenom, valTokens.MulRaw((int64)(i2))),
			),
		),
	)

	genesisDelegations, err := f.stakingKeeper.GetAllDelegations(f.sdkCtx)
	assert.NilError(t, err)
	delegations = append(delegations, genesisDelegations...)

	genesisState := types.NewGenesisState(params, validators, delegations)
	vals := (f.stakingKeeper.InitGenesis(f.sdkCtx, genesisState))

	actualGenesis := (f.stakingKeeper.ExportGenesis(f.sdkCtx))
	assert.DeepEqual(t, genesisState.Params, actualGenesis.Params)
	assert.DeepEqual(t, genesisState.Delegations, actualGenesis.Delegations)

	allvals, err := f.stakingKeeper.GetAllValidators(f.sdkCtx)
	assert.NilError(t, err)
	assert.DeepEqual(t, allvals, actualGenesis.Validators)

	// Ensure validators have addresses.
	vals2, err := staking.WriteValidators(f.sdkCtx, (f.stakingKeeper))
	assert.NilError(t, err)

	for _, val := range vals2 {
		assert.Assert(t, val.Address.String() != "")
	}

	// now make sure the validators are bonded and intra-tx counters are correct
	resVal, found := (f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[1])))
	assert.Assert(t, found)
	assert.Equal(t, types.Bonded, resVal.Status)

	resVal, found = (f.stakingKeeper.GetValidator(f.sdkCtx, sdk.ValAddress(addrs[2])))
	assert.Assert(t, found)
	assert.Equal(t, types.Bonded, resVal.Status)

	abcivals := make([]abci.ValidatorUpdate, len(vals))

	for i, val := range validators {
		abcivals[i] = val.ABCIValidatorUpdate((f.stakingKeeper.PowerReduction(f.sdkCtx)))
	}

	assert.DeepEqual(t, abcivals, vals)
}

func TestInitGenesis_PoolsBalanceMismatch(t *testing.T) {
	t.Parallel()
	f := initFixture(t)

	consPub, err := codectypes.NewAnyWithValue(PKs[0])
	assert.NilError(t, err)

	validator := types.Validator{
		OperatorAddress: sdk.ValAddress("12345678901234567890").String(),
		ConsensusPubkey: consPub,
		Jailed:          false,
		Tokens:          math.NewInt(10),
		DelegatorShares: math.LegacyNewDecFromInt(math.NewInt(10)),
		Description:     types.NewDescription("bloop", "", "", "", ""),
	}

	params := types.Params{
		UnbondingTime: 10000,
		MaxValidators: 1,
		MaxEntries:    10,
		BondDenom:     "stake",
	}

	require.Panics(t, func() {
		// setting validator status to bonded so the balance counts towards bonded pool
		validator.Status = types.Bonded
		f.stakingKeeper.InitGenesis(f.sdkCtx, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	},
	// "should panic because bonded pool balance is different from bonded pool coins",
	)

	require.Panics(t, func() {
		// setting validator status to unbonded so the balance counts towards not bonded pool
		validator.Status = types.Unbonded
		f.stakingKeeper.InitGenesis(f.sdkCtx, &types.GenesisState{
			Params:     params,
			Validators: []types.Validator{validator},
		})
	},
	// "should panic because not bonded pool balance is different from not bonded pool coins",
	)
}

func TestInitGenesisLargeValidatorSet(t *testing.T) {
	size := 200
	assert.Assert(t, size > 100)

	f, addrs := bootstrapGenesisTest(t, 200)
	genesisValidators, err := f.stakingKeeper.GetAllValidators(f.sdkCtx)
	assert.NilError(t, err)

	params, err := f.stakingKeeper.GetParams(f.sdkCtx)
	assert.NilError(t, err)
	delegations := []types.Delegation{}
	validators := make([]types.Validator, size)

	bondedPoolAmt := math.ZeroInt()
	for i := range validators {
		validators[i], err = types.NewValidator(
			sdk.ValAddress(addrs[i]).String(),
			PKs[i],
			types.NewDescription(fmt.Sprintf("#%d", i), "", "", "", ""),
		)
		assert.NilError(t, err)
		validators[i].Status = types.Bonded

		tokens := f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1)
		if i < 100 {
			tokens = f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 2)
		}

		validators[i].Tokens = tokens
		validators[i].DelegatorShares = math.LegacyNewDecFromInt(tokens)

		// add bonded coins
		bondedPoolAmt = bondedPoolAmt.Add(tokens)
	}

	validators = append(validators, genesisValidators...)
	genesisState := types.NewGenesisState(params, validators, delegations)

	// mint coins in the bonded pool representing the validators coins
	assert.NilError(t,
		banktestutil.FundModuleAccount(
			f.sdkCtx,
			f.bankKeeper,
			types.BondedPoolName,
			sdk.NewCoins(sdk.NewCoin(params.BondDenom, bondedPoolAmt)),
		),
	)

	vals := f.stakingKeeper.InitGenesis(f.sdkCtx, genesisState)

	abcivals := make([]abci.ValidatorUpdate, 100)
	for i, val := range validators[:100] {
		abcivals[i] = val.ABCIValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx))
	}

	// remove genesis validator
	vals = vals[:100]
	assert.DeepEqual(t, abcivals, vals)
}

func TestExportImportGenesisWithPendingConsensusKeyRotation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	valAddr, _ := bondConsKeyRotationValidator(t, f, oldPk)

	// enqueue a rotation at the current height, exporting here means the
	// imported chain starts before comet activates the new key
	writeHeight := f.app.LastBlockHeight()
	writeCtx := f.sdkCtx.WithBlockHeight(writeHeight)
	_, err := msgServer.RotateConsPubKey(writeCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, newPk),
	})
	assert.NilError(t, err)

	// export should keep the unbonding window history and pending apply queue
	// since the rotation is not active yet, the apply height gets pushed
	// forward
	exported := f.stakingKeeper.ExportGenesis(writeCtx)
	require.Len(t, exported.ConsensusKeyRotationHistory, 1)
	require.Len(t, exported.PendingConsensusKeyRotations, 1)
	importedInitialHeight := writeHeight + 1
	assert.Equal(t, importedInitialHeight+types.ConsensusUpdateDelay, exported.PendingConsensusKeyRotations[0].ApplyHeight)

	// the top level Comet genesis validator set should still use the old key
	// because the rotation is not active at the exported initial height
	topLevelVals, err := staking.WriteValidators(writeCtx, f.stakingKeeper)
	assert.NilError(t, err)
	require.Len(t, topLevelVals, 1)
	assert.Assert(t, bytes.Equal(oldPk.Address().Bytes(), topLevelVals[0].PubKey.Address()))

	// create a new fixture to import the genesis onto
	imported := initFixture(t)

	// typically bank genesis import would fund these pools, however we are not
	// exporting bank here so we manually fund the bonded pools and not bonded
	// pools to match the validators in the staking genesis
	fundStakingPoolsForGenesis(t, imported, exported)

	importCtx := imported.sdkCtx.WithBlockHeight(importedInitialHeight)
	initUpdates := imported.stakingKeeper.InitGenesis(importCtx, exported)
	require.Len(t, initUpdates, 1)

	// ensure the validator updates returned from the init are still using the
	// old pubkey
	initUpdatePk, err := cryptocodec.FromCmtProtoPublicKey(initUpdates[0].PubKey)
	assert.NilError(t, err)
	assert.DeepEqual(t, oldPk.Address().Bytes(), initUpdatePk.Address().Bytes())

	// before the pushed apply height, staking state should still have the old key
	// the rotation should also still be tracked in the unbonding window indexes
	stored, err := imported.stakingKeeper.GetValidator(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	storedConsAddr, err := stored.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, oldPk.Address().Bytes(), storedConsAddr)

	// ensure we have a pending key rotation in the staking keeper
	hasPending, err := imported.stakingKeeper.HasConsKeyRotationInUnbondingWindow(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasPending)

	// on the imported chain's first block, staking should re-emit the update
	// this recreates the Comet update that was not persisted in app genesis
	finalizeRes := advanceBlock(t, imported, imported.sdkCtx.BlockHeader().Time)
	updates := finalizeRes.ValidatorUpdates
	require.Len(t, updates, 2)
	reemittedPk, err := cryptocodec.FromCmtProtoPublicKey(updates[1].PubKey)
	assert.NilError(t, err)
	assert.DeepEqual(t, newPk.Address().Bytes(), reemittedPk.Address().Bytes())

	// at the pushed apply height, staking should apply the key swap
	// the pending apply queue should be empty after that
	for imported.app.LastBlockHeight() < exported.PendingConsensusKeyRotations[0].ApplyHeight {
		advanceBlock(t, imported, imported.sdkCtx.BlockHeader().Time)
	}

	// ensure the key rotated for this validator
	stored, err = imported.stakingKeeper.GetValidator(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	storedConsAddr, err = stored.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, newPk.Address().Bytes(), storedConsAddr)

	// ensure there are no more pending rotations
	pending, err := imported.stakingKeeper.PendingConsKeyRotations(imported.sdkCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(pending) == 0)

	// the validator marker and old key lock should remain until maturity
	hasRotationHistory, err := imported.stakingKeeper.HasConsKeyRotationInUnbondingWindow(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasRotationHistory)
	hasOldLock, err := imported.stakingKeeper.IsConsAddrLockedByRotation(imported.sdkCtx, sdk.ConsAddress(oldPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, hasOldLock)

	// the new key is live now, so the temporary new key lock should be released
	hasNewLock, err := imported.stakingKeeper.IsConsAddrLockedByRotation(imported.sdkCtx, sdk.ConsAddress(newPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, !hasNewLock)
}

func TestExportImportZeroHeightGenesisWithPendingConsensusKeyRotation(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	valAddr, _ := bondConsKeyRotationValidator(t, f, oldPk)

	// queue a rotation before the zero height prep runs
	writeHeight := int64(100)
	writeCtx := f.sdkCtx.WithBlockHeight(writeHeight)
	_, err := msgServer.RotateConsPubKey(writeCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, newPk),
	})
	assert.NilError(t, err)

	// mimic simapp zero height prep by rebasing the pending queue in state
	originalApplyHeight := writeHeight + types.ConsensusUpdateDelay
	zeroHeightApplyHeight := int64(1) + types.ConsensusUpdateDelay
	assert.NilError(t, f.stakingKeeper.PrepareConsKeyRotationsForZeroHeightExport(writeCtx))

	hasOriginalQueue, err := f.stakingKeeper.HasConsKeyRotationApplyQueueEntry(writeCtx, originalApplyHeight, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, !hasOriginalQueue)
	hasZeroHeightQueue, err := f.stakingKeeper.HasConsKeyRotationApplyQueueEntry(writeCtx, zeroHeightApplyHeight, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasZeroHeightQueue)

	// match simapp zero height export by using a zero height context after prep
	zeroHeightCtx := writeCtx.WithBlockHeight(0)
	exported := f.stakingKeeper.ExportGenesis(zeroHeightCtx)
	require.Len(t, exported.ConsensusKeyRotationHistory, 1)
	require.Len(t, exported.PendingConsensusKeyRotations, 1)
	assert.Equal(t, zeroHeightApplyHeight, exported.PendingConsensusKeyRotations[0].ApplyHeight)

	// the exported comet validator set should still use the old key
	topLevelVals, err := staking.WriteValidators(zeroHeightCtx, f.stakingKeeper)
	assert.NilError(t, err)
	require.Len(t, topLevelVals, 1)
	assert.Assert(t, bytes.Equal(oldPk.Address().Bytes(), topLevelVals[0].PubKey.Address()))

	// import onto a fresh chain at height zero
	imported := initFixture(t)
	fundStakingPoolsForGenesis(t, imported, exported)
	initUpdates := imported.stakingKeeper.InitGenesis(imported.sdkCtx.WithBlockHeight(0), exported)
	require.Len(t, initUpdates, 1)

	// init should also report the old key to match the top level validator set
	initUpdatePk, err := cryptocodec.FromCmtProtoPublicKey(initUpdates[0].PubKey)
	assert.NilError(t, err)
	assert.DeepEqual(t, oldPk.Address().Bytes(), initUpdatePk.Address().Bytes())

	// before the apply height, the imported staking state should still use the old key
	stored, err := imported.stakingKeeper.GetValidator(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	storedConsAddr, err := stored.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, oldPk.Address().Bytes(), storedConsAddr)

	// advance through the normal abci flow until the rebased apply height
	for imported.app.LastBlockHeight() < zeroHeightApplyHeight {
		advanceBlock(t, imported, imported.sdkCtx.BlockHeader().Time)
	}

	// after the apply height, staking should have moved to the new key
	stored, err = imported.stakingKeeper.GetValidator(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	storedConsAddr, err = stored.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, newPk.Address().Bytes(), storedConsAddr)

	// the apply queue drains while history and the old key lock remain
	pending, err := imported.stakingKeeper.PendingConsKeyRotations(imported.sdkCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(pending) == 0)
	hasRotationHistory, err := imported.stakingKeeper.HasConsKeyRotationInUnbondingWindow(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasRotationHistory)
	hasOldLock, err := imported.stakingKeeper.IsConsAddrLockedByRotation(imported.sdkCtx, sdk.ConsAddress(oldPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, hasOldLock)
	hasNewLock, err := imported.stakingKeeper.IsConsAddrLockedByRotation(imported.sdkCtx, sdk.ConsAddress(newPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, !hasNewLock)
}

func TestExportGenesisWithPendingConsensusKeyRotationAtActivationHeight(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	valAddr, _ := bondConsKeyRotationValidator(t, f, oldPk)

	// enqueue a rotation, then export from the height right before sdk apply
	// the imported initial height is the activation height
	writeHeight := f.app.LastBlockHeight()
	writeCtx := f.sdkCtx.WithBlockHeight(writeHeight)
	_, err := msgServer.RotateConsPubKey(writeCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, newPk),
	})
	assert.NilError(t, err)

	// because the rotation is active at the exported initial height
	// export should keep the original apply height
	activationExportCtx := f.sdkCtx.WithBlockHeight(writeHeight + types.ConsensusUpdateDelay - 1)
	exported := f.stakingKeeper.ExportGenesis(activationExportCtx)
	require.Len(t, exported.PendingConsensusKeyRotations, 1)
	assert.Equal(t, writeHeight+types.ConsensusUpdateDelay, exported.PendingConsensusKeyRotations[0].ApplyHeight)

	// the top level Comet genesis validator set should use the new key
	// this is the key Comet would have active at this height
	topLevelVals, err := staking.WriteValidators(activationExportCtx, f.stakingKeeper)
	assert.NilError(t, err)
	require.Len(t, topLevelVals, 1)
	assert.Assert(t, bytes.Equal(newPk.Address().Bytes(), topLevelVals[0].PubKey.Address()))

	// create a new fixture to import the genesis onto
	imported := initFixture(t)

	// typically bank genesis import would fund these pools, however we are not
	// exporting bank here so we manually fund the bonded pools and not bonded
	// pools to match the validators in the staking genesis
	fundStakingPoolsForGenesis(t, imported, exported)

	importCtx := imported.sdkCtx.WithBlockHeight(writeHeight + types.ConsensusUpdateDelay)
	initUpdates := imported.stakingKeeper.InitGenesis(importCtx, exported)
	require.Len(t, initUpdates, 1)

	// ensure update is emitted with the new key (one we rotated to)
	initUpdatePk, err := cryptocodec.FromCmtProtoPublicKey(initUpdates[0].PubKey)
	assert.NilError(t, err)
	assert.DeepEqual(t, newPk.Address().Bytes(), initUpdatePk.Address().Bytes())

	// advancing to the activation height should drain the pending rotation
	// staking state should end up on the new key
	for imported.app.LastBlockHeight() < importCtx.BlockHeight() {
		advanceBlock(t, imported, imported.sdkCtx.BlockHeader().Time)
	}

	// ensure the key did rotate for this validator
	stored, err := imported.stakingKeeper.GetValidator(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	storedConsAddr, err := stored.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, newPk.Address().Bytes(), storedConsAddr)

	// the pending apply queue should be drained after the key swap
	pending, err := imported.stakingKeeper.PendingConsKeyRotations(imported.sdkCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(pending) == 0)

	// the validator marker and old key lock should remain until maturity
	hasRotationHistory, err := imported.stakingKeeper.HasConsKeyRotationInUnbondingWindow(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasRotationHistory)
	hasOldLock, err := imported.stakingKeeper.IsConsAddrLockedByRotation(imported.sdkCtx, sdk.ConsAddress(oldPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, hasOldLock)

	// the new key is live now, so the temporary new key lock should be released
	hasNewLock, err := imported.stakingKeeper.IsConsAddrLockedByRotation(imported.sdkCtx, sdk.ConsAddress(newPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, !hasNewLock)
}

func TestExportImportGenesisWithAppliedConsensusKeyRotationHistory(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	nextPk := ed25519.GenPrivKey().PubKey()
	valAddr, _ := bondConsKeyRotationValidator(t, f, oldPk)

	// apply a rotation before export so only the unbonding window history remains
	// there should be no pending apply queue entry
	writeHeight := f.app.LastBlockHeight()
	writeCtx := f.sdkCtx.WithBlockHeight(writeHeight)
	_, err := msgServer.RotateConsPubKey(writeCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, newPk),
	})
	assert.NilError(t, err)

	// advance to the apply height so the end blocker swaps staking state
	for f.app.LastBlockHeight() < writeHeight+types.ConsensusUpdateDelay {
		advanceBlock(t, f, f.sdkCtx.BlockHeader().Time)
	}

	// after apply, staking state should be on the new key
	stored, err := f.stakingKeeper.GetValidator(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	storedConsAddr, err := stored.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, newPk.Address().Bytes(), storedConsAddr)

	// the apply queue should be empty while history remains until maturity
	pending, err := f.stakingKeeper.PendingConsKeyRotations(f.sdkCtx)
	assert.NilError(t, err)
	assert.Assert(t, len(pending) == 0)
	hasRotationHistory, err := f.stakingKeeper.HasConsKeyRotationInUnbondingWindow(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasRotationHistory)

	// old key stays locked until maturity, new key lock should be released
	hasOldLock, err := f.stakingKeeper.IsConsAddrLockedByRotation(f.sdkCtx, sdk.ConsAddress(oldPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, hasOldLock)
	hasNewLock, err := f.stakingKeeper.IsConsAddrLockedByRotation(f.sdkCtx, sdk.ConsAddress(newPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, !hasNewLock)

	// export and import should preserve the old consensus address history
	// it should not recreate a pending rotation
	exportCtx := f.sdkCtx.WithBlockHeight(f.app.LastBlockHeight())
	exported := f.stakingKeeper.ExportGenesis(exportCtx)
	require.Len(t, exported.ConsensusKeyRotationHistory, 1)
	require.Empty(t, exported.PendingConsensusKeyRotations)

	// create a new fixture to import the genesis onto
	imported := initFixture(t)

	// typically bank genesis import would fund these pools, however we are not
	// exporting bank here so we manually fund the bonded pools and not bonded
	// pools to match the validators in the staking genesis
	fundStakingPoolsForGenesis(t, imported, exported)
	imported.stakingKeeper.InitGenesis(imported.sdkCtx, exported)

	hasOldLock, err = imported.stakingKeeper.IsConsAddrLockedByRotation(imported.sdkCtx, sdk.ConsAddress(oldPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, hasOldLock)

	// the restored history should still block another rotation
	// this should hold until the unbonding window has passed
	_, err = keeper.NewMsgServerImpl(imported.stakingKeeper).RotateConsPubKey(imported.sdkCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, nextPk),
	})
	assert.ErrorContains(t, err, types.ErrExceedingMaxConsPubKeyRotations.Error())

	// once the restored history matures, the end blocker clears the marker
	// it should also clear the old consensus address lock
	advanceBlock(t, imported, exported.ConsensusKeyRotationHistory[0].MaturityTime.AddDate(0, 0, 1))

	// the rotation has matured, so the validator marker and old key lock
	// should be gone
	hasPending, err := imported.stakingKeeper.HasConsKeyRotationInUnbondingWindow(imported.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, !hasPending)
	hasOldLock, err = imported.stakingKeeper.IsConsAddrLockedByRotation(imported.sdkCtx, sdk.ConsAddress(oldPk.Address()))
	assert.NilError(t, err)
	assert.Assert(t, !hasOldLock)
}

func fundStakingPoolsForGenesis(t *testing.T, f *fixture, genesis *types.GenesisState) {
	t.Helper()

	bondedTokens := math.ZeroInt()
	notBondedTokens := math.ZeroInt()
	for _, validator := range genesis.Validators {
		switch validator.GetStatus() {
		case types.Bonded:
			bondedTokens = bondedTokens.Add(validator.GetTokens())
		case types.Unbonded, types.Unbonding:
			notBondedTokens = notBondedTokens.Add(validator.GetTokens())
		}
	}
	for _, ubd := range genesis.UnbondingDelegations {
		for _, entry := range ubd.Entries {
			notBondedTokens = notBondedTokens.Add(entry.Balance)
		}
	}

	if !bondedTokens.IsZero() {
		assert.NilError(t, banktestutil.FundModuleAccount(
			f.sdkCtx,
			f.bankKeeper,
			types.BondedPoolName,
			sdk.NewCoins(sdk.NewCoin(genesis.Params.BondDenom, bondedTokens)),
		))
	}
	if !notBondedTokens.IsZero() {
		assert.NilError(t, banktestutil.FundModuleAccount(
			f.sdkCtx,
			f.bankKeeper,
			types.NotBondedPoolName,
			sdk.NewCoins(sdk.NewCoin(genesis.Params.BondDenom, notBondedTokens)),
		))
	}
}
