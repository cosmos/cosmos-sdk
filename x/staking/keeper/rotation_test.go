package keeper_test

import (
	"errors"
	"testing"
	"time"

	abci "github.com/cometbft/cometbft/abci/types"
	cmtproto "github.com/cometbft/cometbft/proto/tendermint/types"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// bondedValidator stores and bonds a validator with the given consensus
// pubkey, returns the validator record together with its operator address.
func (s *KeeperTestSuite) bondedValidator(pk cryptotypes.PubKey) (stakingtypes.Validator, sdk.ValAddress) {
	require := s.Require()
	valAddr := sdk.ValAddress(pk.Address())
	v, err := stakingtypes.NewValidator(valAddr.String(), pk, stakingtypes.Description{Moniker: "v"})
	require.NoError(err)
	v.Status = stakingtypes.Bonded
	v.Tokens = sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
	v.DelegatorShares = math.LegacyNewDecFromInt(v.Tokens)
	require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
	require.NoError(s.stakingKeeper.SetValidatorByConsAddr(s.ctx, v))
	return v, valAddr
}

func (s *KeeperTestSuite) TestConsKeyRotationUpdate() {
	require := s.Require()

	s.T().Run("emits old at zero power and new at full power without mutating state", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		validator, valAddr := s.bondedValidator(oldPk)

		updates, err := s.stakingKeeper.ConsKeyRotationUpdate(validator, newPk, sdk.DefaultPowerReduction)
		require.NoError(err)
		require.Len(updates, 2)
		require.Equal(int64(0), updates[0].Power)
		require.Equal(int64(10), updates[1].Power)

		// state must not have been touched
		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		gotConsAddr, err := stored.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(oldPk.Address()).Bytes(), gotConsAddr)

		_, err = s.stakingKeeper.GetValidatorByConsAddr(s.ctx, sdk.ConsAddress(newPk.Address()))
		require.Error(err)
	})
}

func (s *KeeperTestSuite) TestApplyConsKeyRotationState() {
	require := s.Require()

	s.T().Run("swaps stored ConsensusPubkey and byConsAddr index", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)

		require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, newPk))

		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		gotConsAddr, err := stored.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(newPk.Address()).Bytes(), gotConsAddr)

		byNew, err := s.stakingKeeper.GetValidatorByConsAddr(s.ctx, sdk.ConsAddress(newPk.Address()))
		require.NoError(err)
		require.Equal(valAddr.String(), byNew.OperatorAddress)

		_, err = s.stakingKeeper.GetValidatorByConsAddr(s.ctx, sdk.ConsAddress(oldPk.Address()))
		require.Error(err)
	})

	s.T().Run("returns nil when validator no longer exists", func(t *testing.T) {
		s.SetupTest()

		missing := sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
		newPk := ed25519.GenPrivKey().PubKey()
		require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, missing, newPk))
	})

	s.T().Run("preserves jailed flag and status across the swap", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		v, valAddr := s.bondedValidator(oldPk)
		v.Jailed = true
		v.Status = stakingtypes.Unbonding
		require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))

		require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, newPk))

		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		require.True(stored.Jailed)
		require.Equal(stakingtypes.Unbonding, stored.Status)
	})
}

func (s *KeeperTestSuite) TestPendingConsKeyRotations() {
	require := s.Require()

	s.T().Run("empty queue returns empty map", func(t *testing.T) {
		s.SetupTest()

		got, err := s.stakingKeeper.PendingConsKeyRotations(s.ctx)
		require.NoError(err)
		require.Empty(got)
	})

	s.T().Run("entry in flight is returned keyed by valAddr", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		got, err := s.stakingKeeper.PendingConsKeyRotations(s.ctx)
		require.NoError(err)
		require.Len(got, 1)
		rotation, ok := got[string(valAddr)]
		require.True(ok)
		require.True(rotation.NewPubKey.Equals(newPk))
		require.Equal(100+stakingtypes.ConsensusUpdateDelay, rotation.ApplyHeight)
	})

	s.T().Run("drained entry is no longer returned", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		s.ctx = s.ctx.WithBlockHeight(100 + stakingtypes.ConsensusUpdateDelay)
		require.NoError(s.stakingKeeper.ApplyConsKeyRotations(s.ctx))

		got, err := s.stakingKeeper.PendingConsKeyRotations(s.ctx)
		require.NoError(err)
		require.Empty(got)
	})
}

func (s *KeeperTestSuite) TestPrepareConsKeyRotationsForZeroHeightExport() {
	require := s.Require()
	s.SetupTest()

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	_, valAddr := s.bondedValidator(oldPk)

	s.ctx = s.ctx.WithBlockHeight(100)
	maturity := s.ctx.BlockTime().Add(stakingtypes.DefaultUnbondingTime)
	originalApplyHeight := s.ctx.BlockHeight() + stakingtypes.ConsensusUpdateDelay
	zeroHeightApplyHeight := int64(1) + stakingtypes.ConsensusUpdateDelay
	require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

	hasOriginalQueue, err := s.stakingKeeper.HasConsKeyRotationApplyQueueEntry(s.ctx, originalApplyHeight, valAddr)
	require.NoError(err)
	require.True(hasOriginalQueue)

	require.NoError(s.stakingKeeper.PrepareConsKeyRotationsForZeroHeightExport(s.ctx))

	hasOriginalQueue, err = s.stakingKeeper.HasConsKeyRotationApplyQueueEntry(s.ctx, originalApplyHeight, valAddr)
	require.NoError(err)
	require.False(hasOriginalQueue)
	hasRebasedQueue, err := s.stakingKeeper.HasConsKeyRotationApplyQueueEntry(s.ctx, zeroHeightApplyHeight, valAddr)
	require.NoError(err)
	require.True(hasRebasedQueue)

	pending, err := s.stakingKeeper.PendingConsKeyRotations(s.ctx)
	require.NoError(err)
	require.Len(pending, 1)
	require.Equal(zeroHeightApplyHeight, pending[string(valAddr)].ApplyHeight)
	require.True(pending[string(valAddr)].NewPubKey.Equals(newPk))

	stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
	require.NoError(err)
	storedConsAddr, err := stored.GetConsAddr()
	require.NoError(err)
	require.Equal(sdk.ConsAddress(oldPk.Address()).Bytes(), storedConsAddr)

	hasHistory, err := s.stakingKeeper.HasConsKeyRotationInUnbondingWindow(s.ctx, valAddr)
	require.NoError(err)
	require.True(hasHistory)
	hasMaturityQueue, err := s.stakingKeeper.HasConsKeyRotationQueueEntry(s.ctx, maturity, valAddr)
	require.NoError(err)
	require.True(hasMaturityQueue)
	hasOldLock, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, sdk.ConsAddress(oldPk.Address()))
	require.NoError(err)
	require.True(hasOldLock)
	hasNewLock, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, sdk.ConsAddress(newPk.Address()))
	require.NoError(err)
	require.True(hasNewLock)
}

func (s *KeeperTestSuite) TestApplyConsKeyRotations() {
	require := s.Require()

	s.T().Run("at apply height swaps state", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		s.ctx = s.ctx.WithBlockHeight(100 + stakingtypes.ConsensusUpdateDelay)
		require.NoError(s.stakingKeeper.ApplyConsKeyRotations(s.ctx))

		// swap is now visible in state
		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		gotConsAddr, err := stored.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(newPk.Address()).Bytes(), gotConsAddr)

		byNew, err := s.stakingKeeper.GetValidatorByConsAddr(s.ctx, sdk.ConsAddress(newPk.Address()))
		require.NoError(err)
		require.Equal(valAddr.String(), byNew.OperatorAddress)

		_, err = s.stakingKeeper.GetValidatorByConsAddr(s.ctx, sdk.ConsAddress(oldPk.Address()))
		require.Error(err)

		// the new addr rotation lock is released; old addr lock remains
		hasNew, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, sdk.ConsAddress(newPk.Address()))
		require.NoError(err)
		require.False(hasNew)
		hasOld, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, sdk.ConsAddress(oldPk.Address()))
		require.NoError(err)
		require.True(hasOld)

		kind, gotValAddr, found, err := s.stakingKeeper.GetRotationLockedConsAddr(s.ctx, sdk.ConsAddress(oldPk.Address()))
		require.NoError(err)
		require.True(found)
		require.Equal(stakingtypes.ConsAddrLockRotatedFrom, kind)
		require.Equal(valAddr, gotValAddr)
	})

	s.T().Run("drain skips removed validator and still clears queue and lock", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		v, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		// drop the validator before the apply height. RemoveValidator only
		// works on unbonded records, so flip the status before deleting.
		v.Status = stakingtypes.Unbonded
		v.Tokens = math.ZeroInt()
		v.DelegatorShares = math.LegacyZeroDec()
		require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
		require.NoError(s.stakingKeeper.RemoveValidator(s.ctx, valAddr))

		s.ctx = s.ctx.WithBlockHeight(100 + stakingtypes.ConsensusUpdateDelay)
		require.NoError(s.stakingKeeper.ApplyConsKeyRotations(s.ctx))

		hasNew, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, sdk.ConsAddress(newPk.Address()))
		require.NoError(err)
		require.False(hasNew)

		kind, gotValAddr, found, err := s.stakingKeeper.GetRotationLockedConsAddr(s.ctx, sdk.ConsAddress(oldPk.Address()))
		require.NoError(err)
		require.True(found)
		require.Equal(stakingtypes.ConsAddrLockPendingFrom, kind)
		require.Equal(valAddr, gotValAddr)

		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(stakingtypes.DefaultUnbondingTime + time.Second))
		require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(s.ctx))

		hasOld, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, sdk.ConsAddress(oldPk.Address()))
		require.NoError(err)
		require.False(hasOld)
	})
}

func (s *KeeperTestSuite) TestRotationLockedConsAddrIndex() {
	require := s.Require()
	s.SetupTest()

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	_, valAddr := s.bondedValidator(oldPk)
	oldConsAddr := sdk.ConsAddress(oldPk.Address())
	newConsAddr := sdk.ConsAddress(newPk.Address())

	require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

	kind, gotValAddr, found, err := s.stakingKeeper.GetRotationLockedConsAddr(s.ctx, oldConsAddr)
	require.NoError(err)
	require.True(found)
	require.Equal(stakingtypes.ConsAddrLockPendingFrom, kind)
	require.Equal(valAddr, gotValAddr)

	kind, gotValAddr, found, err = s.stakingKeeper.GetRotationLockedConsAddr(s.ctx, newConsAddr)
	require.NoError(err)
	require.True(found)
	require.Equal(stakingtypes.ConsAddrLockPendingTo, kind)
	require.Equal(valAddr, gotValAddr)

	require.NoError(s.stakingKeeper.ApplyConsKeyRotations(s.ctx.WithBlockHeight(s.ctx.BlockHeight() + stakingtypes.ConsensusUpdateDelay)))

	kind, gotValAddr, found, err = s.stakingKeeper.GetRotationLockedConsAddr(s.ctx, oldConsAddr)
	require.NoError(err)
	require.True(found)
	require.Equal(stakingtypes.ConsAddrLockRotatedFrom, kind)
	require.Equal(valAddr, gotValAddr)

	_, _, found, err = s.stakingKeeper.GetRotationLockedConsAddr(s.ctx, newConsAddr)
	require.NoError(err)
	require.False(found)
}

func (s *KeeperTestSuite) TestValidatorByHistoricalConsAddr() {
	require := s.Require()

	s.T().Run("resolves rotated-from key after rotation applies", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)
		oldConsAddr := sdk.ConsAddress(oldPk.Address())
		newConsAddr := sdk.ConsAddress(newPk.Address())

		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
		require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, newPk))

		validator, err := s.stakingKeeper.ValidatorByHistoricalConsAddr(s.ctx, oldConsAddr)
		require.NoError(err)
		require.Equal(valAddr.String(), validator.OperatorAddress)
		currentConsAddr, err := validator.GetConsAddr()
		require.NoError(err)
		require.Equal(newConsAddr.Bytes(), currentConsAddr)
	})

	s.T().Run("rejects pending-to key", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)
		newConsAddr := sdk.ConsAddress(newPk.Address())

		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		_, err := s.stakingKeeper.ValidatorByHistoricalConsAddr(s.ctx, newConsAddr)
		require.True(errors.Is(err, stakingtypes.ErrNoValidatorFound))
	})

	s.T().Run("rejects pending-from key", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)
		oldConsAddr := sdk.ConsAddress(oldPk.Address())

		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		_, err := s.stakingKeeper.ValidatorByHistoricalConsAddr(s.ctx, oldConsAddr)
		require.True(errors.Is(err, stakingtypes.ErrNoValidatorFound))
	})

	s.T().Run("stops resolving after maturity pruning without evidence params", func(t *testing.T) {
		s.SetupTest()

		// no consensus evidence params are set, so retirement collapses to the
		// unbonding maturity (the pre-fix behavior).
		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)
		oldConsAddr := sdk.ConsAddress(oldPk.Address())

		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
		require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, newPk))

		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(stakingtypes.DefaultUnbondingTime + time.Second))
		require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(s.ctx))

		_, err := s.stakingKeeper.ValidatorByHistoricalConsAddr(s.ctx, oldConsAddr)
		require.True(errors.Is(err, stakingtypes.ErrNoValidatorFound))
	})

	s.T().Run("keeps resolving until the evidence block window closes", func(t *testing.T) {
		s.SetupTest()

		// evidence stays admissible for a long block window that outlasts
		// unbonding in wall-clock time; the lock must survive until the block
		// window also closes.
		const maxAgeNumBlocks = int64(1_000_000)
		s.ctx = s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Evidence: &cmtproto.EvidenceParams{
				MaxAgeDuration:  stakingtypes.DefaultUnbondingTime,
				MaxAgeNumBlocks: maxAgeNumBlocks,
			},
		}).WithBlockHeight(100)

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)
		oldConsAddr := sdk.ConsAddress(oldPk.Address())
		applyHeight := s.ctx.BlockHeight() + stakingtypes.ConsensusUpdateDelay

		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
		require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, newPk))

		// advance time past unbonding and the evidence time window, but keep the
		// height inside the evidence block window.
		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(2 * stakingtypes.DefaultUnbondingTime))
		require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(s.ctx))

		// the re-rotation gate is retired, but the evidence lock still resolves.
		hasGate, err := s.stakingKeeper.HasConsKeyRotationInUnbondingWindow(s.ctx, valAddr)
		require.NoError(err)
		require.False(hasGate)

		validator, err := s.stakingKeeper.ValidatorByHistoricalConsAddr(s.ctx, oldConsAddr)
		require.NoError(err)
		require.Equal(valAddr.String(), validator.OperatorAddress)

		// once the height passes the evidence block window too, the lock retires.
		s.ctx = s.ctx.WithBlockHeight(applyHeight + maxAgeNumBlocks + 1)
		require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(s.ctx))

		_, err = s.stakingKeeper.ValidatorByHistoricalConsAddr(s.ctx, oldConsAddr)
		require.True(errors.Is(err, stakingtypes.ErrNoValidatorFound))
	})

	s.T().Run("re-rotation after gate retirement keeps independent evidence locks", func(t *testing.T) {
		s.SetupTest()

		const maxAgeNumBlocks = int64(1_000_000)
		s.ctx = s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
			Evidence: &cmtproto.EvidenceParams{
				MaxAgeDuration:  stakingtypes.DefaultUnbondingTime,
				MaxAgeNumBlocks: maxAgeNumBlocks,
			},
		}).WithBlockHeight(100)

		pk0 := ed25519.GenPrivKey().PubKey()
		pk1 := ed25519.GenPrivKey().PubKey()
		pk2 := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(pk0)
		cons0 := sdk.ConsAddress(pk0.Address())
		cons1 := sdk.ConsAddress(pk1.Address())

		// first rotation pk0 -> pk1, applied.
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, pk0, pk1))
		require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, pk1))

		// advance past unbonding so the first gate retires, then prune.
		s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(stakingtypes.DefaultUnbondingTime + time.Second))
		require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(s.ctx))
		hasGate, err := s.stakingKeeper.HasConsKeyRotationInUnbondingWindow(s.ctx, valAddr)
		require.NoError(err)
		require.False(hasGate)

		// second rotation pk1 -> pk2 is now allowed and must set a fresh gate.
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, pk1, pk2))
		require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, pk2))
		hasGate, err = s.stakingKeeper.HasConsKeyRotationInUnbondingWindow(s.ctx, valAddr)
		require.NoError(err)
		require.True(hasGate, "second rotation must set a fresh re-rotation gate")

		// pruning again (the first rotation is still lingering for its block
		// window) must not clobber the fresh gate, and both old keys must resolve.
		require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(s.ctx))
		hasGate, err = s.stakingKeeper.HasConsKeyRotationInUnbondingWindow(s.ctx, valAddr)
		require.NoError(err)
		require.True(hasGate, "fresh gate must survive pruning of the earlier rotation")

		for _, cons := range []sdk.ConsAddress{cons0, cons1} {
			validator, err := s.stakingKeeper.ValidatorByHistoricalConsAddr(s.ctx, cons)
			require.NoError(err)
			require.Equal(valAddr.String(), validator.OperatorAddress)
		}
	})
}

func (s *KeeperTestSuite) TestPendingConsKeyRotationUpdates() {
	require := s.Require()

	const H = int64(100)

	oldPk := ed25519.GenPrivKey().PubKey()
	newPk := ed25519.GenPrivKey().PubKey()
	valAddr := sdk.ValAddress(oldPk.Address())

	testCases := []struct {
		name           string                                       // subtest name
		setup          func(t *testing.T)                           // state setup before scanning
		lastValidators map[string]int64                             // last validator powers keyed by operator address
		wantRotations  []stakingkeeper.PendingConsKeyRotationUpdate // expected pending rotations
	}{
		{
			name: "future rotation is skipped before validator lookup",
			setup: func(t *testing.T) {
				t.Helper()

				// Queue a rotation for a missing validator, then scan before the
				// Comet-visible emit height so Pending... exits before lookup.
				s.ctx = s.ctx.WithBlockHeight(H)
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
				s.ctx = s.ctx.WithBlockHeight(H - 1)
			},
			lastValidators: map[string]int64{string(valAddr): 10},
			wantRotations:  nil,
		},
		{
			name: "emit height rotation carries last validator power",
			setup: func(t *testing.T) {
				t.Helper()

				// Queue a rotation for an existing bonded validator and scan at
				// the emit height, before SDK state applies the new key.
				_, valAddr = s.bondedValidator(oldPk)
				s.ctx = s.ctx.WithBlockHeight(H)
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
			},
			lastValidators: map[string]int64{string(valAddr): 10},
			wantRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
		},
		{
			name: "in-flight rotation remains visible after emit height",
			setup: func(t *testing.T) {
				t.Helper()

				// Queue a rotation for an existing validator, then scan after
				// the emit height but before the apply height drains the queue.
				_, valAddr = s.bondedValidator(oldPk)
				s.ctx = s.ctx.WithBlockHeight(H)
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
				s.ctx = s.ctx.WithBlockHeight(H + 1)
			},
			lastValidators: map[string]int64{string(valAddr): 10},
			wantRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
		},
		{
			name: "production drain at apply height removes pending rotation",
			setup: func(t *testing.T) {
				t.Helper()

				// Queue a rotation, advance to the SDK apply height, and drain
				// mature rotations before scanning, matching EndBlock order.
				_, valAddr = s.bondedValidator(oldPk)
				s.ctx = s.ctx.WithBlockHeight(H)
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
				s.ctx = s.ctx.WithBlockHeight(H + stakingtypes.ConsensusUpdateDelay)
				require.NoError(s.stakingKeeper.ApplyConsKeyRotations(s.ctx))
			},
			lastValidators: map[string]int64{string(valAddr): 10},
		},
		{
			name: "missing validator is skipped",
			setup: func(t *testing.T) {
				t.Helper()

				// Queue a rotation for an operator address with no validator
				// record and scan at the emit height.
				s.ctx = s.ctx.WithBlockHeight(H)
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
			},
			lastValidators: map[string]int64{string(valAddr): 10},
		},
		{
			name: "validator outside last set is returned with zero last power",
			setup: func(t *testing.T) {
				t.Helper()

				// Queue a rotation for an existing validator but pass no last
				// validator power entry, modeling an inactive validator.
				_, valAddr = s.bondedValidator(oldPk)
				s.ctx = s.ctx.WithBlockHeight(H)
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
			},
			wantRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  0,
			}},
		},
		{
			name: "last power comes from last validator set not current tokens",
			setup: func(t *testing.T) {
				t.Helper()

				// Queue a rotation for a validator whose current tokens imply
				// power 20; Pending... should still report lastValidators power.
				validator, addr := s.bondedValidator(oldPk)
				valAddr = addr
				validator.Tokens = sdk.TokensFromConsensusPower(20, sdk.DefaultPowerReduction)
				validator.DelegatorShares = math.LegacyNewDecFromInt(validator.Tokens)
				require.NoError(s.stakingKeeper.SetValidator(s.ctx, validator))
				s.ctx = s.ctx.WithBlockHeight(H)
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
			},
			lastValidators: map[string]int64{string(valAddr): 10},
			wantRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()
			tc.setup(t)

			rotations, err := s.stakingKeeper.PendingConsKeyRotationUpdates(s.ctx, tc.lastValidators)
			require.NoError(err)
			require.Equal(tc.wantRotations, rotations)
		})
	}
}

func (s *KeeperTestSuite) TestProcessValidatorUpdatesForConsKeyRotations() {
	require := s.Require()

	const H = int64(100)
	power0 := int64(0)
	power10 := int64(10)
	power25 := int64(25)

	oldPk := ed25519.GenPrivKey().PubKey()
	oldCmtPk := cmtPk(s.T(), oldPk)

	newPk := ed25519.GenPrivKey().PubKey()
	newCmtPk := cmtPk(s.T(), newPk)

	oldPk2 := ed25519.GenPrivKey().PubKey()
	oldCmtPk2 := cmtPk(s.T(), oldPk2)

	newPk2 := ed25519.GenPrivKey().PubKey()
	newCmtPk2 := cmtPk(s.T(), newPk2)

	otherPk := ed25519.GenPrivKey().PubKey()
	otherCmtPk := cmtPk(s.T(), otherPk)

	trailingPk := ed25519.GenPrivKey().PubKey()
	trailingCmtPk := cmtPk(s.T(), trailingPk)

	testCases := []struct {
		name             string
		height           int64
		pendingRotations []stakingkeeper.PendingConsKeyRotationUpdate
		previousUpdates  []abci.ValidatorUpdate
		wantErr          bool
		wantUpdates      []abci.ValidatorUpdate
	}{
		{
			name:            "no rotations returns normal updates unchanged",
			height:          H,
			previousUpdates: []abci.ValidatorUpdate{{PubKey: otherCmtPk, Power: 7}},
			wantUpdates:     []abci.ValidatorUpdate{{PubKey: otherCmtPk, Power: 7}},
		},
		{
			name:   "baseline rotation emits old zero and new last power",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			wantUpdates: []abci.ValidatorUpdate{
				{PubKey: oldCmtPk, Power: 0},
				{PubKey: newCmtPk, Power: 10},
			},
		},
		{
			name:   "same-height positive update becomes rotation pair at updated power",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{{PubKey: oldCmtPk, Power: power25}},
			wantUpdates: []abci.ValidatorUpdate{
				{PubKey: oldCmtPk, Power: 0},
				{PubKey: newCmtPk, Power: 25},
			},
		},
		{
			name:   "same-height zero update keeps old zero and does not add new zero",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{{PubKey: oldCmtPk, Power: power0}},
			wantUpdates:     []abci.ValidatorUpdate{{PubKey: oldCmtPk, Power: 0}},
		},
		{
			name:   "post-emit positive update is translated to new key",
			height: H + 1,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{{PubKey: oldCmtPk, Power: power25}},
			wantUpdates:     []abci.ValidatorUpdate{{PubKey: newCmtPk, Power: 25}},
		},
		{
			name:   "post-emit zero update is translated to new zero",
			height: H + 1,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{{PubKey: oldCmtPk, Power: power0}},
			wantUpdates:     []abci.ValidatorUpdate{{PubKey: newCmtPk, Power: 0}},
		},
		{
			name:   "post-emit rotation without matching update leaves unrelated updates unchanged",
			height: H + 1,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{{PubKey: otherCmtPk, Power: 7}},
			wantUpdates:     []abci.ValidatorUpdate{{PubKey: otherCmtPk, Power: 7}},
		},
		{
			name:   "validator outside last set entering active set is translated to new key",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  0,
			}},
			previousUpdates: []abci.ValidatorUpdate{{PubKey: oldCmtPk, Power: power25}},
			wantUpdates:     []abci.ValidatorUpdate{{PubKey: newCmtPk, Power: 25}},
		},
		{
			name:   "baseline rotation is appended after non-rotating updates",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{
				{PubKey: otherCmtPk, Power: 7},
			},
			wantUpdates: []abci.ValidatorUpdate{
				{PubKey: otherCmtPk, Power: 7},
				{PubKey: oldCmtPk, Power: 0},
				{PubKey: newCmtPk, Power: 10},
			},
		},
		{
			name:   "rewritten updates preserve the replaced update position",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{
				{PubKey: otherCmtPk, Power: 7},
				{PubKey: oldCmtPk, Power: power25},
				{PubKey: trailingCmtPk, Power: 9},
			},
			wantUpdates: []abci.ValidatorUpdate{
				{PubKey: otherCmtPk, Power: 7},
				{PubKey: oldCmtPk, Power: 0},
				{PubKey: newCmtPk, Power: 25},
				{PubKey: trailingCmtPk, Power: 9},
			},
		},
		{
			name:   "multiple baseline rotations are appended in rotation order",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{
				{
					OldPubKey:  oldPk,
					NewPubKey:  newPk,
					EmitHeight: H,
					LastPower:  10,
				},
				{
					OldPubKey:  oldPk2,
					NewPubKey:  newPk2,
					EmitHeight: H,
					LastPower:  6,
				},
			},
			wantUpdates: []abci.ValidatorUpdate{
				{PubKey: oldCmtPk, Power: 0},
				{PubKey: newCmtPk, Power: 10},
				{PubKey: oldCmtPk2, Power: 0},
				{PubKey: newCmtPk2, Power: 6},
			},
		},
		{
			name:   "duplicate normal updates return an error",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{
				{PubKey: oldCmtPk, Power: power10},
				{PubKey: oldCmtPk, Power: power10},
			},
			wantErr: true,
		},
		{
			name:   "invalid normal update pubkey returns an error",
			height: H,
			pendingRotations: []stakingkeeper.PendingConsKeyRotationUpdate{{
				OldPubKey:  oldPk,
				NewPubKey:  newPk,
				EmitHeight: H,
				LastPower:  10,
			}},
			previousUpdates: []abci.ValidatorUpdate{{Power: 1}},
			wantErr:         true,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()

			s.ctx = s.ctx.WithBlockHeight(tc.height)
			processed, err := s.stakingKeeper.ProcessValidatorUpdatesForConsKeyRotations(s.ctx, tc.pendingRotations, tc.previousUpdates)
			if tc.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)

			require.Len(processed, len(tc.wantUpdates))
			for i, want := range tc.wantUpdates {
				require.Equal(want.PubKey, processed[i].PubKey)
				require.Equal(want.Power, processed[i].Power)
			}
		})
	}
}

func (s *KeeperTestSuite) TestPruneMaturedConsKeyRotations() {
	require := s.Require()

	type rec struct {
		valAddr  sdk.ValAddress
		consAddr sdk.ConsAddress
		maturity time.Time
	}

	queueRotation := func() rec {
		oldPk := ed25519.GenPrivKey().PubKey()
		valAddr := sdk.ValAddress(oldPk.Address())
		newPk := ed25519.GenPrivKey().PubKey()
		maturity := s.ctx.BlockTime().Add(stakingtypes.DefaultUnbondingTime)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))
		return rec{valAddr, sdk.ConsAddress(oldPk.Address()), maturity}
	}

	testCases := []struct {
		name       string
		matured    int
		notMatured int
	}{
		{name: "empty queue", matured: 0, notMatured: 0},
		{name: "single matured entry", matured: 1, notMatured: 0},
		{name: "single not yet matured entry", matured: 0, notMatured: 1},
		{name: "mixed matured and future", matured: 2, notMatured: 2},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()

			baseTime := s.ctx.BlockTime()
			var maturedEntries, futureEntries []rec

			// queue matured entries by rewinding the context so their maturity
			// (queueTime + unbondingTime) falls strictly before baseTime
			s.ctx = s.ctx.WithBlockTime(baseTime.Add(-stakingtypes.DefaultUnbondingTime - time.Hour))
			for i := 0; i < tc.matured; i++ {
				maturedEntries = append(maturedEntries, queueRotation())
			}

			// queue future entries at baseTime so their maturity is in the future
			s.ctx = s.ctx.WithBlockTime(baseTime)
			for i := 0; i < tc.notMatured; i++ {
				futureEntries = append(futureEntries, queueRotation())
			}

			require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(s.ctx))

			for _, e := range maturedEntries {
				hasQueue, err := s.stakingKeeper.HasConsKeyRotationQueueEntry(s.ctx, e.maturity, e.valAddr)
				require.NoError(err)
				require.False(hasQueue, "matured queue entry should be pruned")

				hasPending, err := s.stakingKeeper.HasConsKeyRotationInUnbondingWindow(s.ctx, e.valAddr)
				require.NoError(err)
				require.False(hasPending, "matured per-validator entry should be pruned")

				hasCons, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, e.consAddr)
				require.NoError(err)
				require.False(hasCons, "matured rotated cons addr entry should be pruned")
			}

			for _, e := range futureEntries {
				hasQueue, err := s.stakingKeeper.HasConsKeyRotationQueueEntry(s.ctx, e.maturity, e.valAddr)
				require.NoError(err)
				require.True(hasQueue, "future queue entry should remain")

				hasPending, err := s.stakingKeeper.HasConsKeyRotationInUnbondingWindow(s.ctx, e.valAddr)
				require.NoError(err)
				require.True(hasPending, "future per-validator entry should remain")

				hasCons, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, e.consAddr)
				require.NoError(err)
				require.True(hasCons, "future rotated cons addr entry should remain")
			}
		})
	}
}

func (s *KeeperTestSuite) TestConsKeyRotationGenesisRoundTrip() {
	require := s.Require()
	s.SetupTest()

	const maxAgeNumBlocks = int64(1_000_000)
	s.ctx = s.ctx.WithConsensusParams(cmtproto.ConsensusParams{
		Evidence: &cmtproto.EvidenceParams{
			MaxAgeDuration:  stakingtypes.DefaultUnbondingTime,
			MaxAgeNumBlocks: maxAgeNumBlocks,
		},
	}).WithBlockHeight(100)

	pk0 := ed25519.GenPrivKey().PubKey()
	pk1 := ed25519.GenPrivKey().PubKey()
	_, valAddr := s.bondedValidator(pk0)
	consAddr := sdk.ConsAddress(pk0.Address())
	applyHeight := s.ctx.BlockHeight() + stakingtypes.ConsensusUpdateDelay

	require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, pk0, pk1))
	require.NoError(s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, pk1))

	// export carries the gate maturity plus the evidence-lock horizon.
	histories, err := s.stakingKeeper.ExportConsKeyRotationHistory(s.ctx)
	require.NoError(err)
	require.Len(histories, 1)
	require.Equal(valAddr.String(), histories[0].ValidatorAddress)
	require.Equal(consAddr.String(), histories[0].OldConsensusAddress)
	require.False(histories[0].MaturityTime.IsZero(), "active gate maturity must be exported")
	require.Equal(applyHeight+maxAgeNumBlocks, histories[0].EvidenceExpiryHeight)

	// import into a fresh store and confirm the gate, the evidence lock, and its
	// retirement queue are all restored.
	s.SetupTest()
	s.ctx = s.ctx.WithBlockHeight(100)
	require.NoError(s.stakingKeeper.ImportConsKeyRotations(s.ctx, histories, nil))

	hasGate, err := s.stakingKeeper.HasConsKeyRotationInUnbondingWindow(s.ctx, valAddr)
	require.NoError(err)
	require.True(hasGate)

	kind, gotVal, found, err := s.stakingKeeper.GetRotationLockedConsAddr(s.ctx, consAddr)
	require.NoError(err)
	require.True(found)
	require.Equal(stakingtypes.ConsAddrLockRotatedFrom, kind)
	require.Equal(valAddr.Bytes(), gotVal.Bytes())

	// the retirement queue survived: pruning before the block window closes keeps
	// the lock; pruning after retires it.
	beforeCtx := s.ctx.WithBlockTime(s.ctx.BlockTime().Add(2 * stakingtypes.DefaultUnbondingTime)).WithBlockHeight(100)
	require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(beforeCtx))
	locked, err := s.stakingKeeper.IsConsAddrLockedByRotation(beforeCtx, consAddr)
	require.NoError(err)
	require.True(locked, "imported lock must be kept while its block window is open")

	afterCtx := beforeCtx.WithBlockHeight(applyHeight + maxAgeNumBlocks + 1)
	require.NoError(s.stakingKeeper.PruneMaturedConsKeyRotations(afterCtx))
	locked, err = s.stakingKeeper.IsConsAddrLockedByRotation(afterCtx, consAddr)
	require.NoError(err)
	require.False(locked, "imported lock must retire once its block window closes")
}
