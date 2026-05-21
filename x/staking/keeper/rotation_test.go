package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestApplyConsKeyRotation() {
	require := s.Require()

	createValidator := func() (sdk.ValAddress, cryptotypes.PubKey) {
		pk := ed25519.GenPrivKey().PubKey()
		valAddr := sdk.ValAddress(pk.Address())
		v, err := stakingtypes.NewValidator(valAddr.String(), pk, stakingtypes.Description{Moniker: "v"})
		require.NoError(err)
		v.Status = stakingtypes.Bonded
		require.NoError(s.stakingKeeper.SetValidator(s.ctx, v))
		require.NoError(s.stakingKeeper.SetValidatorByConsAddr(s.ctx, v))
		return valAddr, pk
	}

	testCases := []struct {
		name      string
		setup     func() (valAddr sdk.ValAddress, oldPk, newPk cryptotypes.PubKey)
		expErr    bool
		expErrMsg string
	}{
		{
			name: "successful rotation",
			setup: func() (sdk.ValAddress, cryptotypes.PubKey, cryptotypes.PubKey) {
				valAddr, oldPk := createValidator()
				return valAddr, oldPk, ed25519.GenPrivKey().PubKey()
			},
		},
		{
			name: "validator not found",
			setup: func() (sdk.ValAddress, cryptotypes.PubKey, cryptotypes.PubKey) {
				missing := sdk.ValAddress(ed25519.GenPrivKey().PubKey().Address())
				return missing, nil, ed25519.GenPrivKey().PubKey()
			},
			expErr:    true,
			expErrMsg: stakingtypes.ErrNoValidatorFound.Error(),
		},
		{
			name: "rotate to same key is a no-op",
			setup: func() (sdk.ValAddress, cryptotypes.PubKey, cryptotypes.PubKey) {
				valAddr, oldPk := createValidator()
				return valAddr, oldPk, oldPk
			},
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			s.SetupTest()

			valAddr, oldPk, newPk := tc.setup()

			err := s.stakingKeeper.ApplyConsKeyRotation(s.ctx, valAddr, newPk)
			if tc.expErr {
				require.Error(err)
				require.Contains(err.Error(), tc.expErrMsg)
				return
			}
			require.NoError(err)

			// validator's stored ConsensusPubkey must now resolve to newPk
			v, err := s.stakingKeeper.GetValidator(s.ctx, valAddr)
			require.NoError(err)
			gotConsAddr, err := v.GetConsAddr()
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

	testCases := []struct {
		name      string
		seedCount int
		stopAfter int
		expectLen int
	}{
		{"empty store", 0, 0, 0},
		{"single entry", 1, 0, 1},
		{"three entries", 3, 0, 3},
		{"stop after first of three", 3, 1, 1},
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
				require.NoError(s.stakingKeeper.SetConsKeyRotation(s.ctx, seeded[i].valAddr, oldPk, seeded[i].newPk, stakingtypes.DefaultKeyRotationFee))
			}

			var observed []entry
			err := s.stakingKeeper.IterateUnappliedConsKeyRotations(s.ctx, func(valAddr sdk.ValAddress, newPk cryptotypes.PubKey) bool {
				observed = append(observed, entry{valAddr, newPk})
				return tc.stopAfter > 0 && len(observed) >= tc.stopAfter
			})
			require.NoError(err)
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

			// non-stop cases must observe every seeded entry
			if tc.stopAfter == 0 {
				require.Len(observed, len(seeded))
			}
		})
	}
}
