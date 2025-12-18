package keeper_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestGovernor(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)
	govKeeper, _, _, _, _, _, ctx := setupGovKeeper(t)
	addrs := simtestutil.CreateRandomAccounts(3)
	govAddrs := convertAddrsToGovAddrs(addrs)

	// Add 2 governors
	gov1Desc := v1.NewGovernorDescription("moniker1", "id1", "website1", "sec1", "detail1")
	gov1, err := v1.NewGovernor(govAddrs[0].String(), gov1Desc, time.Now().UTC())
	require.NoError(err)
	gov2Desc := v1.NewGovernorDescription("moniker2", "id2", "website2", "sec2", "detail2")
	gov2, err := v1.NewGovernor(govAddrs[1].String(), gov2Desc, time.Now().UTC())
	require.NoError(err)
	gov2.Status = v1.Inactive
	govKeeper.Governors.Set(ctx, gov1.GetAddress(), gov1)
	govKeeper.Governors.Set(ctx, gov2.GetAddress(), gov2)

	// Get gov1
	gov, err := govKeeper.Governors.Get(ctx, govAddrs[0])
	assert.NoError(err, "cant find gov1")
	assert.Equal(gov1, gov)

	// Get gov2
	gov, err = govKeeper.Governors.Get(ctx, govAddrs[1])
	assert.NoError(err, "cant find gov2")
	assert.Equal(gov2, gov)

	// Get all govs
	var govs []*v1.Governor
	err = govKeeper.Governors.Walk(ctx, nil, func(_ types.GovernorAddress, gov v1.Governor) (stop bool, err error) {
		govs = append(govs, &gov)
		return false, nil
	})
	require.NoError(err)
	if assert.Len(govs, 2, "expected 2 governors") {
		// Insert order is not preserved, order is related to the address which is
		// generated randomly, so the order of govs is random.
		for i := 0; i < 2; i++ {
			switch govs[i].GetAddress().String() {
			case gov1.GetAddress().String():
				assert.Equal(gov1, *govs[i])
			case gov2.GetAddress().String():
				assert.Equal(gov2, *govs[i])
			}
		}
	}

	// Get all active govs
	govs = nil
	err = govKeeper.Governors.Walk(ctx, nil, func(_ types.GovernorAddress, gov v1.Governor) (stop bool, err error) {
		if gov.IsActive() {
			govs = append(govs, &gov)
		}
		return false, nil
	})
	require.NoError(err)
	if assert.Len(govs, 1, "expected 1 active governor") {
		assert.Equal(gov1, *govs[0])
	}

	// Remove gov2
	err = govKeeper.Governors.Remove(ctx, govAddrs[1])
	require.NoError(err)
	_, err = govKeeper.Governors.Get(ctx, govAddrs[1])
	assert.ErrorIs(err, collections.ErrNotFound, "expected gov2 to be removed")

	// Get all govs after removal
	govs = nil
	err = govKeeper.Governors.Walk(ctx, nil, func(_ types.GovernorAddress, gov v1.Governor) (stop bool, err error) {
		govs = append(govs, &gov)
		return false, nil
	})
	require.NoError(err)
	if assert.Len(govs, 1, "expected 1 governor after removal") {
		assert.Equal(gov1, *govs[0])
	}
}

func TestValidateGovernorMinSelfDelegation(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*fixture) v1.Governor
		selfDelegation bool
		valDelegations []stakingtypes.Delegation
		expectedPanic  bool
		expectedValid  bool
	}{
		{
			name: "inactive governor",
			setup: func(s *fixture) v1.Governor {
				return s.inactiveGovernor
			},
			expectedPanic: false,
			expectedValid: false,
		},
		{
			name: "active governor w/o self delegation w/o validator delegation",
			setup: func(s *fixture) v1.Governor {
				return s.activeGovernors[0]
			},
			expectedPanic: true,
			expectedValid: false,
		},
		{
			name: "active governor w self delegation w/o validator delegation",
			setup: func(s *fixture) v1.Governor {
				govAddr := s.activeGovernors[0].GetAddress()
				delAddr := sdk.AccAddress(govAddr)
				err := s.keeper.DelegateToGovernor(s.ctx, delAddr, govAddr)
				require.NoError(s.t, err)
				return s.activeGovernors[0]
			},
			expectedPanic: false,
			expectedValid: false,
		},
		{
			name: "active governor w self delegation w not enough validator delegation",
			setup: func(s *fixture) v1.Governor {
				govAddr := s.activeGovernors[0].GetAddress()
				delAddr := sdk.AccAddress(govAddr)
				err := s.keeper.DelegateToGovernor(s.ctx, delAddr, govAddr)
				require.NoError(s.t, err)
				s.delegate(delAddr, s.valAddrs[0], 1)
				return s.activeGovernors[0]
			},
			expectedPanic: false,
			expectedValid: false,
		},
		{
			name: "active governor w self delegation w enough validator delegation",
			setup: func(s *fixture) v1.Governor {
				govAddr := s.activeGovernors[0].GetAddress()
				delAddr := sdk.AccAddress(govAddr)
				err := s.keeper.DelegateToGovernor(s.ctx, delAddr, govAddr)
				require.NoError(s.t, err)
				s.delegate(delAddr, s.valAddrs[0], v1.DefaultMinGovernorSelfDelegation.Int64())
				return s.activeGovernors[0]
			},
			expectedPanic: false,
			expectedValid: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			govKeeper, accKeeper, bankKeeper, stakingKeeper, distrKeeper, _, ctx := setupGovKeeper(t, mockAccountKeeperExpectations)
			mocks := mocks{
				accKeeper:          accKeeper,
				bankKeeper:         bankKeeper,
				stakingKeeper:      stakingKeeper,
				distributionKeeper: distrKeeper,
			}
			s := newFixture(t, ctx, 2, 2, 2, govKeeper, mocks)
			governor := tt.setup(s)

			if tt.expectedPanic {
				assert.Panics(t, func() { govKeeper.ValidateGovernorMinSelfDelegation(ctx, governor) })
			} else {
				valid := govKeeper.ValidateGovernorMinSelfDelegation(ctx, governor)

				assert.Equal(t, tt.expectedValid, valid, "return of ValidateGovernorMinSelfDelegation")
			}
		})
	}
}
