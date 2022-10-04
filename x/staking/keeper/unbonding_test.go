package keeper_test

import "github.com/cosmos/cosmos-sdk/x/staking/types"

func (s *KeeperTestSuite) TestIncrementUnbondingId() {
	for i := 1; i < 10; i++ {
		s.Require().Equal(uint64(i), s.stakingKeeper.IncrementUnbondingId(s.ctx))
	}
}

func (s *KeeperTestSuite) TestUnbondingTypeSetters() {
	cases := []struct {
		exists   bool
		name     string
		expected types.UnbondingType
	}{
		{
			name:     "existing 1",
			exists:   true,
			expected: types.UnbondingType_UnbondingDelegation,
		},
		{
			name:     "existing 2",
			exists:   true,
			expected: types.UnbondingType_Redelegation,
		},
		{
			name:   "not existing",
			exists: false,
		},
	}

	for i, tc := range cases {
		s.Run(tc.name, func() {
			if tc.exists {
				s.stakingKeeper.SetUnbondingType(s.ctx, uint64(i), tc.expected)
			}

			unbondingType, found := s.stakingKeeper.GetUnbondingType(s.ctx, uint64(i))
			if tc.exists {
				s.Require().True(found)
				s.Require().Equal(tc.expected, unbondingType)
			} else {
				s.Require().False(found)
			}
		})
	}
}

func (s *KeeperTestSuite) TestGetUnbondingDelegationByUnbondingId() {}

func (s *KeeperTestSuite) TestGetRedelegationByUnbondingId() {}

func (s *KeeperTestSuite) TestGetValidatorByUnbondingId() {}

func (s *KeeperTestSuite) TestSetUnbondingDelegationByUnbondingId() {}

func (s *KeeperTestSuite) TestSetRedelegationByUnbondingId() {}

func (s *KeeperTestSuite) TestUnbondingCanComplete() {}

func (s *KeeperTestSuite) TestPutUnbondingOnHold() {}
