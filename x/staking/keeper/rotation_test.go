package keeper_test

import (
	"errors"
	"testing"
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestApplyConsKeyRotation() {
	require := s.Require()

	createValidator := func() (stakingtypes.Validator, sdk.ValAddress, cryptotypes.PubKey) {
		pk := ed25519.GenPrivKey().PubKey()
		valAddr := sdk.ValAddress(pk.Address())
		v, err := stakingtypes.NewValidator(valAddr.String(), pk, stakingtypes.Description{Moniker: "v"})
		require.NoError(err)
		v.Status = stakingtypes.Bonded
		v.Tokens = sdk.TokensFromConsensusPower(10, sdk.DefaultPowerReduction)
		v.DelegatorShares = math.LegacyNewDecFromInt(v.Tokens)
		require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
		require.NoError(s.stakingKeeper.SetValidatorByConsAddr(s.ctx, v))
		return v, valAddr, pk
	}

	testCases := []struct {
		name  string
		setup func() (validator stakingtypes.Validator, valAddr sdk.ValAddress, oldPk, newPk cryptotypes.PubKey)
	}{
		{
			name: "successful rotation",
			setup: func() (stakingtypes.Validator, sdk.ValAddress, cryptotypes.PubKey, cryptotypes.PubKey) {
				v, valAddr, oldPk := createValidator()
				return v, valAddr, oldPk, ed25519.GenPrivKey().PubKey()
			},
		},
		{
			name: "rotate to same key is a no-op",
			setup: func() (stakingtypes.Validator, sdk.ValAddress, cryptotypes.PubKey, cryptotypes.PubKey) {
				v, valAddr, oldPk := createValidator()
				return v, valAddr, oldPk, oldPk
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()

			validator, valAddr, oldPk, newPk := tc.setup()

			updates, err := s.stakingKeeper.ApplyConsKeyRotation(s.ctx, validator, newPk, sdk.DefaultPowerReduction)
			require.NoError(err)
			require.Len(updates, 2)
			require.Equal(int64(0), updates[0].Power)
			require.Equal(int64(10), updates[1].Power)

			// validator's stored ConsensusPubkey must now resolve to newPk
			stored, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
			require.NoError(err)
			gotConsAddr, err := stored.GetConsAddr()
			require.NoError(err)
			require.Equal(sdk.ConsAddress(newPk.Address()).Bytes(), gotConsAddr)

			// new by cons address lookup resolves to this validator
			byNew, err := s.stakingKeeper.GetValidatorByConsAddr(s.ctx, sdk.ConsAddress(newPk.Address()))
			require.NoError(err)
			require.Equal(valAddr.String(), byNew.OperatorAddress)

			// old by cons address lookup is gone unless the rotation was a no-op
			if !oldPk.Equals(newPk) {
				_, err = s.stakingKeeper.GetValidatorByConsAddr(s.ctx, sdk.ConsAddress(oldPk.Address()))
				require.Error(err)
			}
		})
	}
}

func (s *KeeperTestSuite) TestIterateUnappliedConsKeyRotations() {
	require := s.Require()

	type entry struct {
		valAddr sdk.ValAddress
		newPk   cryptotypes.PubKey
	}

	errStop := errors.New("stop")

	testCases := []struct {
		name      string
		seedCount int
		stopAfter int // 0 = no stop
		expectLen int
		expectErr error
	}{
		{name: "empty store", seedCount: 0, expectLen: 0},
		{name: "single entry", seedCount: 1, expectLen: 1},
		{name: "three entries", seedCount: 3, expectLen: 3},
		{name: "stop with error after first of three", seedCount: 3, stopAfter: 1, expectLen: 1, expectErr: errStop},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()

			seeded := make([]entry, tc.seedCount)
			for i := range seeded {
				seeded[i] = entry{
					valAddr: sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address()),
					newPk:   ed25519.GenPrivKey().PubKey(),
				}
				oldPk := ed25519.GenPrivKey().PubKey()
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, seeded[i].valAddr, oldPk, seeded[i].newPk))
			}

			var observed []entry
			err := s.stakingKeeper.IterateUnappliedConsKeyRotations(s.ctx, func(valAddr sdk.ValAddress, newPk cryptotypes.PubKey) error {
				observed = append(observed, entry{valAddr, newPk})
				if tc.stopAfter > 0 && len(observed) >= tc.stopAfter {
					return errStop
				}
				return nil
			})
			if tc.expectErr != nil {
				require.ErrorIs(err, tc.expectErr)
			} else {
				require.NoError(err)
			}
			require.Len(observed, tc.expectLen)

			// each observed entry must round trip to one of the seeded entries
			for _, got := range observed {
				found := false
				for _, want := range seeded {
					if got.valAddr.Equals(want.valAddr) && got.newPk.Equals(want.newPk) {
						found = true
						break
					}
				}
				require.True(found, "observed entry not in seeded set: %s", got.valAddr)
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
