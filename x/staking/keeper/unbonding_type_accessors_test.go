package keeper_test

import (
	"time"

	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func (s *KeeperTestSuite) TestIncrementUnbondingID() {
	for i := 1; i < 10; i++ {
		s.Require().Equal(uint64(i), s.app.StakingKeeper.IncrementUnbondingID(s.ctx))
	}
}

func (s *KeeperTestSuite) TestUnbondingTypeAccessors() {
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
				s.app.StakingKeeper.SetUnbondingType(s.ctx, uint64(i), tc.expected)
			}

			unbondingType, found := s.app.StakingKeeper.GetUnbondingType(s.ctx, uint64(i))
			if tc.exists {
				s.Require().True(found)
				s.Require().Equal(tc.expected, unbondingType)
			} else {
				s.Require().False(found)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnbondingDelegationByUnbondingIDAccessors() {
	delAddrs := simapp.AddTestAddrsIncremental(s.app, s.ctx, 2, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(delAddrs)

	type exists struct {
		setUnbondingDelegation              bool
		setUnbondingDelegationByUnbondingID bool
	}

	cases := []struct {
		exists   exists
		name     string
		expected types.UnbondingDelegation
	}{
		{
			name:   "existing 1",
			exists: exists{true, true},
			expected: types.NewUnbondingDelegation(
				delAddrs[0],
				valAddrs[0],
				0,
				time.Unix(0, 0).UTC(),
				sdk.NewInt(5),
				0,
			),
		},
		{
			name:   "not existing 1",
			exists: exists{false, true},
			expected: types.NewUnbondingDelegation(
				delAddrs[1],
				valAddrs[1],
				0,
				time.Unix(0, 0).UTC(),
				sdk.NewInt(5),
				0,
			),
		},
		{
			name:   "not existing 2",
			exists: exists{false, false},
			expected: types.NewUnbondingDelegation(
				delAddrs[0],
				valAddrs[0],
				0,
				time.Unix(0, 0).UTC(),
				sdk.NewInt(5),
				0,
			),
		},
	}

	for i, tc := range cases {
		s.Run(tc.name, func() {
			if tc.exists.setUnbondingDelegation {
				s.app.StakingKeeper.SetUnbondingDelegation(s.ctx, tc.expected)
			}

			if tc.exists.setUnbondingDelegationByUnbondingID {
				s.app.StakingKeeper.SetUnbondingDelegationByUnbondingID(s.ctx, tc.expected, uint64(i))
			}

			ubd, found := s.app.StakingKeeper.GetUnbondingDelegationByUnbondingID(s.ctx, uint64(i))
			if tc.exists.setUnbondingDelegation && tc.exists.setUnbondingDelegationByUnbondingID {
				s.Require().True(found)
				s.Require().Equal(tc.expected, ubd)
			} else {
				s.Require().False(found)
			}
		})
	}
}

func (s *KeeperTestSuite) TestRedelegationByUnbondingIDAccessors() {
	delAddrs := simapp.AddTestAddrsIncremental(s.app, s.ctx, 2, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(delAddrs)

	type exists struct {
		setRedelegation              bool
		setRedelegationByUnbondingID bool
	}

	cases := []struct {
		exists   exists
		name     string
		expected types.Redelegation
	}{
		{
			name:   "existing 1",
			exists: exists{true, true},
			expected: types.NewRedelegation(
				delAddrs[0],
				valAddrs[0],
				valAddrs[1],
				0,
				time.Unix(5, 0).UTC(),
				sdk.NewInt(10),
				sdk.NewDec(10),
				0,
			),
		},
		{
			name:   "not existing 1",
			exists: exists{false, true},
			expected: types.NewRedelegation(
				delAddrs[1],
				valAddrs[0],
				valAddrs[1],
				0,
				time.Unix(5, 0).UTC(),
				sdk.NewInt(10),
				sdk.NewDec(10),
				0,
			),
		},
		{
			name:   "not existing 2",
			exists: exists{false, false},
			expected: types.NewRedelegation(
				delAddrs[1],
				valAddrs[1],
				valAddrs[0],
				0,
				time.Unix(5, 0).UTC(),
				sdk.NewInt(10),
				sdk.NewDec(10),
				0,
			),
		},
	}

	for i, tc := range cases {
		s.Run(tc.name, func() {
			if tc.exists.setRedelegation {
				s.app.StakingKeeper.SetRedelegation(s.ctx, tc.expected)
			}

			if tc.exists.setRedelegationByUnbondingID {
				s.app.StakingKeeper.SetRedelegationByUnbondingID(s.ctx, tc.expected, uint64(i))
			}

			red, found := s.app.StakingKeeper.GetRedelegationByUnbondingID(s.ctx, uint64(i))
			if tc.exists.setRedelegation && tc.exists.setRedelegationByUnbondingID {
				s.Require().True(found)
				s.Require().Equal(tc.expected, red)
			} else {
				s.Require().False(found)
			}
		})
	}
}

func (s *KeeperTestSuite) TestValidatorByUnbondingIDAccessors() {
	delAddrs := simapp.AddTestAddrsIncremental(s.app, s.ctx, 3, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(delAddrs)

	type exists struct {
		setValidator              bool
		setValidatorByUnbondingID bool
	}

	newVal := func(valAddr sdk.ValAddress, pk cryptotypes.PubKey) types.Validator {
		val, err := types.NewValidator(valAddr, pk, types.Description{})
		val.MinSelfDelegation = sdk.ZeroInt()
		s.Require().NoError(err)
		return val
	}

	cases := []struct {
		exists    exists
		name      string
		validator types.Validator
	}{
		{
			name:      "existing 1",
			exists:    exists{true, true},
			validator: newVal(valAddrs[0], PKs[0]),
		},
		{
			name:      "not existing 1",
			exists:    exists{false, true},
			validator: newVal(valAddrs[2], PKs[0]),
		},
		{
			name:      "not existing 2",
			exists:    exists{false, false},
			validator: newVal(valAddrs[2], PKs[0]),
		},
	}

	for i, tc := range cases {
		s.Run(tc.name, func() {
			if tc.exists.setValidator {
				s.app.StakingKeeper.SetValidator(s.ctx, tc.validator)
			}

			if tc.exists.setValidatorByUnbondingID {
				s.app.StakingKeeper.SetValidatorByUnbondingID(s.ctx, tc.validator, uint64(i))
			}

			val, found := s.app.StakingKeeper.GetValidatorByUnbondingID(s.ctx, uint64(i))
			if tc.exists.setValidator && tc.exists.setValidatorByUnbondingID {
				s.Require().True(found)
				s.Require().Equal(tc.validator, val)
			} else {
				s.Require().False(found)
			}
		})
	}
}

func (s *KeeperTestSuite) TestUnbondingCanComplete() {
	delAddrs := simapp.AddTestAddrsIncremental(s.app, s.ctx, 3, sdk.NewInt(10000))
	valAddrs := simapp.ConvertAddrsToValAddrs(delAddrs)
	unbondingID := uint64(1)

	// no unbondingID set
	err := s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().ErrorIs(err, types.ErrUnbondingNotFound)

	// unbonding delegation
	s.app.StakingKeeper.SetUnbondingType(s.ctx, unbondingID, types.UnbondingType_UnbondingDelegation)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().ErrorIs(err, types.ErrUnbondingNotFound)

	ubd := types.NewUnbondingDelegation(
		delAddrs[0],
		valAddrs[0],
		0,
		time.Unix(0, 0).UTC(),
		sdk.NewInt(5),
		unbondingID,
	)
	s.app.StakingKeeper.SetUnbondingDelegation(s.ctx, ubd)
	s.app.StakingKeeper.SetUnbondingDelegationByUnbondingID(s.ctx, ubd, unbondingID)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().ErrorIs(err, types.ErrUnbondingOnHoldRefCountNegative)

	err = s.app.StakingKeeper.PutUnbondingOnHold(s.ctx, unbondingID)
	s.Require().NoError(err)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().NoError(err)

	// redelegation
	unbondingID++
	s.app.StakingKeeper.SetUnbondingType(s.ctx, unbondingID, types.UnbondingType_Redelegation)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().ErrorIs(err, types.ErrUnbondingNotFound)

	red := types.NewRedelegation(
		delAddrs[0],
		valAddrs[0],
		valAddrs[1],
		0,
		time.Unix(5, 0).UTC(),
		sdk.NewInt(10),
		sdk.NewDec(10),
		unbondingID,
	)
	s.app.StakingKeeper.SetRedelegation(s.ctx, red)
	s.app.StakingKeeper.SetRedelegationByUnbondingID(s.ctx, red, unbondingID)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().ErrorIs(err, types.ErrUnbondingOnHoldRefCountNegative)

	err = s.app.StakingKeeper.PutUnbondingOnHold(s.ctx, unbondingID)
	s.Require().NoError(err)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().NoError(err)

	// validator unbonding
	unbondingID++
	s.app.StakingKeeper.SetUnbondingType(s.ctx, unbondingID, types.UnbondingType_ValidatorUnbonding)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().ErrorIs(err, types.ErrUnbondingNotFound)

	val, err := types.NewValidator(valAddrs[0], PKs[0], types.Description{})
	s.Require().NoError(err)
	s.app.StakingKeeper.SetValidator(s.ctx, val)
	s.app.StakingKeeper.SetValidatorByUnbondingID(s.ctx, val, unbondingID)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().ErrorIs(err, types.ErrUnbondingOnHoldRefCountNegative)

	err = s.app.StakingKeeper.PutUnbondingOnHold(s.ctx, unbondingID)
	s.Require().NoError(err)
	err = s.app.StakingKeeper.UnbondingCanComplete(s.ctx, unbondingID)
	s.Require().NoError(err)
}
