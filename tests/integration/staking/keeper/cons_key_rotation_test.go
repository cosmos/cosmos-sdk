package keeper_test

import (
	"testing"
	"time"

	cmtabcitypes "github.com/cometbft/cometbft/abci/types"
	cmttypes "github.com/cometbft/cometbft/types"
	"gotest.tools/v3/assert"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/keeper"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

// Covers msg-server queuing plus the EndBlocker's two-phase drain: the fee
// transfer, all four store indexes, the queue entry at the correct apply
// height, the pre-apply read-back of the unchanged ConsensusPubkey, and the
// post-apply swap of the validator's ConsensusPubkey and byConsAddr index.
func TestRotateConsPubKey_MsgServerQueuesAndEndBlockerApplies(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)
	bondDenom, err := f.stakingKeeper.BondDenom(f.sdkCtx)
	assert.NilError(t, err)

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	valAddr, accAddr := bondConsKeyRotationValidator(t, f, oldPk)
	oldConsAddr := sdk.ConsAddress(oldPk.Address())
	newConsAddr := sdk.ConsAddress(newPk.Address())

	accBalBefore := f.bankKeeper.GetBalance(f.sdkCtx, accAddr, bondDenom)
	supplyBefore := f.bankKeeper.GetSupply(f.sdkCtx, bondDenom)

	valBefore, err := f.stakingKeeper.GetValidator(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	powerReduction := f.stakingKeeper.PowerReduction(f.sdkCtx)
	powerBefore := valBefore.ConsensusPower(powerReduction)
	assert.Assert(t, powerBefore > 0)

	writeCtx := f.sdkCtx.WithBlockHeight(f.app.LastBlockHeight() + 1)
	writeHeight := writeCtx.BlockHeight()
	_, err = msgServer.RotateConsPubKey(writeCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, newPk),
	})
	assert.NilError(t, err)

	// fee debited from the operator account and burned (total supply
	// decreases by exactly the fee)
	fee := sdk.NewCoin(bondDenom, types.DefaultKeyRotationFeeAmount)
	assert.DeepEqual(t, accBalBefore.Sub(fee), f.bankKeeper.GetBalance(f.sdkCtx, accAddr, bondDenom))
	assert.DeepEqual(t, supplyBefore.Sub(fee), f.bankKeeper.GetSupply(f.sdkCtx, bondDenom))

	// per-validator pending index recorded (gates further rotations inside the
	// unbonding window)
	hasPending, err := f.stakingKeeper.HasConsKeyRotationInUnbondingWindow(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasPending)

	// maturity queue entry recorded at BlockTime + UnbondingTime
	unbondingTime, err := f.stakingKeeper.UnbondingTime(f.sdkCtx)
	assert.NilError(t, err)
	maturity := f.sdkCtx.BlockHeader().Time.Add(unbondingTime)
	hasQueue, err := f.stakingKeeper.HasConsKeyRotationQueueEntry(f.sdkCtx, maturity, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasQueue)

	// rotated cons addr index recorded so the old key still resolves to this
	// validator for slashing/evidence routing
	hasRotated, err := f.stakingKeeper.IsConsAddrLockedByRotation(f.sdkCtx, oldConsAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasRotated)

	// validators stored ConsensusPubkey is unchanged before the apply height.
	// The deferred design keeps the SDK view aligned with CometBFT's active
	// key, which only switches at writeHeight + ConsensusUpdateDelay.
	preEndBlocker, err := f.stakingKeeper.GetValidator(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	preConsAddr, err := preEndBlocker.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, oldConsAddr.Bytes(), preConsAddr)

	// advance one block at the current block time. With the deferred apply
	// design the EndBlocker emit pass runs but the SDK-side state swap does
	// not yet land (applyHeight = writeHeight + 2 has not been reached).
	advanceBlock(t, f, f.sdkCtx.BlockHeader().Time)

	stillOld, err := f.stakingKeeper.GetValidator(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	stillOldConsAddr, err := stillOld.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, oldConsAddr.Bytes(), stillOldConsAddr)
	byOld, err := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, oldConsAddr)
	assert.NilError(t, err)
	assert.Equal(t, valAddr.String(), byOld.OperatorAddress)

	// at the apply height (writeHeight + 2), the drain pass swaps state.
	// invoke the EndBlocker pass directly at the apply height because the
	// integration fixture's EndBlocker runs with the captured initial block
	// height and does not advance with FinalizeBlock.
	applyCtx := f.sdkCtx.WithBlockHeight(writeHeight + types.ConsensusUpdateDelay)
	assert.NilError(t, f.stakingKeeper.ApplyConsKeyRotations(applyCtx))

	// old by-cons-addr index is gone
	_, err = f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, oldConsAddr)
	assert.ErrorContains(t, err, types.ErrNoValidatorFound.Error())

	// new by-cons-addr index resolves to this validator
	byNew, err := f.stakingKeeper.GetValidatorByConsAddr(f.sdkCtx, newConsAddr)
	assert.NilError(t, err)
	assert.Equal(t, valAddr.String(), byNew.OperatorAddress)

	// validators stored ConsensusPubkey now reflects newPk and power is
	// unchanged
	stored, err := f.stakingKeeper.GetValidator(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	storedConsAddr, err := stored.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, newConsAddr.Bytes(), storedConsAddr)
	assert.Equal(t, powerBefore, stored.ConsensusPower(powerReduction))

	// the per-validator pending index intentionally persists past the apply
	// so that further rotations are gated until the end blocker prunes it
	// after maturity
	hasPendingAfter, err := f.stakingKeeper.HasConsKeyRotationInUnbondingWindow(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, hasPendingAfter)
}

// Covers PruneMaturedConsKeyRotations (called from the end blocker) clearing
// the maturity queue, the per-validator pending index, and the rotated cons
// addr index once the unbonding window has elapsed.
func TestRotateConsPubKey_PruneClearsRotationStateAfterUnbonding(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	valAddr, _ := bondConsKeyRotationValidator(t, f, oldPk)
	oldConsAddr := sdk.ConsAddress(oldPk.Address())

	_, err := msgServer.RotateConsPubKey(f.sdkCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, newPk),
	})
	assert.NilError(t, err)

	unbondingTime, err := f.stakingKeeper.UnbondingTime(f.sdkCtx)
	assert.NilError(t, err)
	maturity := f.sdkCtx.BlockHeader().Time.Add(unbondingTime)

	// first block at current time: maturity is in the future so no pruning
	// happens
	advanceBlock(t, f, f.sdkCtx.BlockHeader().Time)

	has, err := f.stakingKeeper.HasConsKeyRotationQueueEntry(f.sdkCtx, maturity, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, has)

	// second block past maturity: the end blocker prunes
	advanceBlock(t, f, maturity.Add(time.Second))

	has, err = f.stakingKeeper.HasConsKeyRotationQueueEntry(f.sdkCtx, maturity, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, !has, "maturity queue entry should be pruned")

	hasPending, err := f.stakingKeeper.HasConsKeyRotationInUnbondingWindow(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	assert.Assert(t, !hasPending, "per-validator pending index should be pruned")

	hasRotated, err := f.stakingKeeper.IsConsAddrLockedByRotation(f.sdkCtx, oldConsAddr)
	assert.NilError(t, err)
	assert.Assert(t, !hasRotated, "rotated cons addr index should be pruned")
}

// Covers the per-window rotation cap lifting after pruning, and that the
// original consensus pubkey can be reused once it leaves the rotation history.
func TestRotateConsPubKey_SecondRotationAfterPruningSucceeds(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	pkA := ed25519.GenPrivKey().PubKey()
	pkB := ed25519.GenPrivKey().PubKey()
	pkC := ed25519.GenPrivKey().PubKey()
	valAddr, _ := bondConsKeyRotationValidator(t, f, pkA)

	// first rotation A -> B
	writeCtx := f.sdkCtx.WithBlockHeight(f.app.LastBlockHeight() + 1)
	writeHeight := writeCtx.BlockHeight()
	_, err := msgServer.RotateConsPubKey(writeCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, pkB),
	})
	assert.NilError(t, err)
	advanceBlock(t, f, f.sdkCtx.BlockHeader().Time)

	// a second rotation inside the unbonding window is rejected
	_, err = msgServer.RotateConsPubKey(f.sdkCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, pkC),
	})
	assert.ErrorContains(t, err, types.ErrExceedingMaxConsPubKeyRotations.Error())

	// drive the SDK-side state swap at writeHeight + ConsensusUpdateDelay so
	// the validators stored ConsensusPubkey becomes pkB and the byConsAddr
	// index moves accordingly. The integration fixture's EndBlocker keeps the
	// captured block height, so the drain pass is invoked here directly.
	applyCtx := f.sdkCtx.WithBlockHeight(writeHeight + types.ConsensusUpdateDelay)
	assert.NilError(t, f.stakingKeeper.ApplyConsKeyRotations(applyCtx))

	// advance past maturity and let the end blocker prune
	unbondingTime, err := f.stakingKeeper.UnbondingTime(f.sdkCtx)
	assert.NilError(t, err)
	advanceBlock(t, f, f.sdkCtx.BlockHeader().Time.Add(unbondingTime).Add(time.Second))

	// second rotation back to pkA (the original key) succeeds: the rotation
	// history was cleared by pruning
	writeCtx = f.sdkCtx.WithBlockHeight(f.app.LastBlockHeight() + 1)
	writeHeight = writeCtx.BlockHeight()
	_, err = msgServer.RotateConsPubKey(writeCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, pkA),
	})
	assert.NilError(t, err)
	advanceBlock(t, f, f.sdkCtx.BlockHeader().Time)

	// drive the drain again for this second rotation
	applyCtx = f.sdkCtx.WithBlockHeight(writeHeight + types.ConsensusUpdateDelay)
	assert.NilError(t, f.stakingKeeper.ApplyConsKeyRotations(applyCtx))

	stored, err := f.stakingKeeper.GetValidator(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	storedConsAddr, err := stored.GetConsAddr()
	assert.NilError(t, err)
	assert.DeepEqual(t, sdk.ConsAddress(pkA.Address()).Bytes(), storedConsAddr)
}

func TestRotateConsPubKey_PowerChangeUpdatesAcceptedByComet(t *testing.T) {
	t.Parallel()
	f := initFixture(t)
	msgServer := keeper.NewMsgServerImpl(f.stakingKeeper)

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	valAddr, _ := bondConsKeyRotationValidator(t, f, oldPk)
	validator, err := f.stakingKeeper.GetValidator(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	initialCmtVals, err := cmttypes.PB2TM.ValidatorUpdates([]cmtabcitypes.ValidatorUpdate{
		validator.ABCIValidatorUpdate(f.stakingKeeper.PowerReduction(f.sdkCtx)),
	})
	assert.NilError(t, err)
	valSet := cmttypes.NewValidatorSet(initialCmtVals)

	_, err = msgServer.RotateConsPubKey(f.sdkCtx, &types.MsgRotateConsPubKey{
		ValidatorAddress: valAddr.String(),
		NewPubkey:        newPubKeyAny(t, newPk),
	})
	assert.NilError(t, err)

	// Force a same-height power update for the rotating validator. Staking must
	// coalesce this with the rotation's new-key update before returning the
	// batch to CometBFT.
	validator, err = f.stakingKeeper.GetValidator(f.sdkCtx, valAddr)
	assert.NilError(t, err)
	validator, _ = validator.AddTokensFromDel(f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 1))
	keeper.TestingUpdateValidator(f.stakingKeeper, f.sdkCtx, validator, false)

	updates, err := f.stakingKeeper.ApplyAndReturnValidatorSetUpdates(f.sdkCtx)
	assert.NilError(t, err)
	assert.Equal(t, 2, len(updates), "%v", updates)

	cmtUpdates, err := cmttypes.PB2TM.ValidatorUpdates(updates)
	assert.NilError(t, err)

	assert.NilError(t, valSet.UpdateWithChangeSet(cmtUpdates))
	assert.Assert(t, !valSet.HasAddress(oldPk.Address()))
	_, newCmtVal := valSet.GetByAddress(newPk.Address())
	assert.Assert(t, newCmtVal != nil)
	assert.Equal(t, int64(101), newCmtVal.VotingPower)
}

// bondConsKeyRotationValidator creates and bonds a single validator under
// consPk, funding the operator account with enough tokens to cover several
// rotation fees plus the self delegation.
func bondConsKeyRotationValidator(t *testing.T, f *fixture, consPk cryptotypes.PubKey) (sdk.ValAddress, sdk.AccAddress) {
	t.Helper()
	addrs := simtestutil.AddTestAddrsIncremental(f.bankKeeper, f.stakingKeeper, f.sdkCtx, 1, f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 300))
	valAddr := sdk.ValAddress(addrs[0])

	v, err := types.NewValidator(valAddr.String(), consPk, types.NewDescription("v", "", "", "", ""))
	assert.NilError(t, err)
	assert.NilError(t, f.stakingKeeper.SetValidator(f.sdkCtx, v))
	assert.NilError(t, f.stakingKeeper.SetValidatorByConsAddr(f.sdkCtx, v))
	assert.NilError(t, f.stakingKeeper.SetNewValidatorByPowerIndex(f.sdkCtx, v))

	_, err = f.stakingKeeper.Delegate(f.sdkCtx, addrs[0], f.stakingKeeper.TokensFromConsensusPower(f.sdkCtx, 100), types.Unbonded, v, true)
	assert.NilError(t, err)

	applyValidatorSetUpdates(t, f.sdkCtx, f.stakingKeeper, 1)

	return valAddr, addrs[0]
}

func newPubKeyAny(t *testing.T, pk cryptotypes.PubKey) *codectypes.Any {
	t.Helper()
	a, err := codectypes.NewAnyWithValue(pk)
	assert.NilError(t, err)
	return a
}

// advanceBlock advances the chain by one block at blockTime, driving the
// staking end blocker through the real ABCI flow so that any matured rotation
// entries are pruned.
func advanceBlock(t *testing.T, f *fixture, blockTime time.Time) *cmtabcitypes.ResponseFinalizeBlock {
	t.Helper()
	res, err := f.app.FinalizeBlock(&cmtabcitypes.RequestFinalizeBlock{
		Height: f.app.LastBlockHeight() + 1,
		Time:   blockTime,
	})
	assert.NilError(t, err)
	_, err = f.app.Commit()
	assert.NilError(t, err)
	return res
}
