package keeper_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
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

func (s *KeeperTestSuite) TestConsKeyRotationUpdates() {
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
		pk, ok := got[string(valAddr)]
		require.True(ok)
		require.True(pk.Equals(newPk))
	})

	s.T().Run("drained entry is no longer returned", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		s.ctx = s.ctx.WithBlockHeight(100 + stakingtypes.ConsensusUpdateDelay)
		_, err := s.stakingKeeper.ProcessConsKeyRotations(s.ctx, sdk.DefaultPowerReduction)
		require.NoError(err)

		got, err := s.stakingKeeper.PendingConsKeyRotations(s.ctx)
		require.NoError(err)
		require.Empty(got)
	})
}

func (s *KeeperTestSuite) TestProcessConsKeyRotations() {
	require := s.Require()

	s.T().Run("at write height emits rotation pair, drain is no-op", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		updates, err := s.stakingKeeper.ProcessConsKeyRotations(s.ctx, sdk.DefaultPowerReduction)
		require.NoError(err)
		require.Len(updates, 2)
		require.Equal(int64(0), updates[0].Power)
		require.Equal(int64(10), updates[1].Power)

		// state is still unchanged: drain has not yet matured
		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		gotConsAddr, err := stored.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(oldPk.Address()).Bytes(), gotConsAddr)
	})

	s.T().Run("between write and apply heights both passes are no-ops", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		s.ctx = s.ctx.WithBlockHeight(100 + stakingtypes.ConsensusUpdateDelay - 1)
		updates, err := s.stakingKeeper.ProcessConsKeyRotations(s.ctx, sdk.DefaultPowerReduction)
		require.NoError(err)
		require.Empty(updates)

		stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
		require.NoError(err)
		gotConsAddr, err := stored.GetConsAddr()
		require.NoError(err)
		require.Equal(sdk.ConsAddress(oldPk.Address()).Bytes(), gotConsAddr)
	})

	s.T().Run("at apply height drain swaps state and emit is no-op", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		_, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		s.ctx = s.ctx.WithBlockHeight(100 + stakingtypes.ConsensusUpdateDelay)
		updates, err := s.stakingKeeper.ProcessConsKeyRotations(s.ctx, sdk.DefaultPowerReduction)
		require.NoError(err)
		require.Empty(updates)

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
	})

	s.T().Run("emit skips removed validator", func(t *testing.T) {
		s.SetupTest()

		oldPk := ed25519.GenPrivKey().PubKey()
		newPk := ed25519.GenPrivKey().PubKey()
		v, valAddr := s.bondedValidator(oldPk)

		s.ctx = s.ctx.WithBlockHeight(100)
		require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, valAddr, oldPk, newPk))

		// drop the validator in the same block, after the rotation is queued
		// but before emit runs. RemoveValidator only works on unbonded records.
		v.Status = stakingtypes.Unbonded
		v.Tokens = math.ZeroInt()
		v.DelegatorShares = math.LegacyZeroDec()
		require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
		require.NoError(s.stakingKeeper.RemoveValidator(s.ctx, valAddr))

		// still at the write height, so emit runs against the queued entry
		updates, err := s.stakingKeeper.ProcessConsKeyRotations(s.ctx, sdk.DefaultPowerReduction)
		require.NoError(err)
		require.Empty(updates)
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
		updates, err := s.stakingKeeper.ProcessConsKeyRotations(s.ctx, sdk.DefaultPowerReduction)
		require.NoError(err)
		require.Empty(updates)

		hasNew, err := s.stakingKeeper.IsConsAddrLockedByRotation(s.ctx, sdk.ConsAddress(newPk.Address()))
		require.NoError(err)
		require.False(hasNew)
	})
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
