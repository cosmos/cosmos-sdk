package keeper_test

import (
	"context"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type fixture struct {
	t *testing.T

	valAddrs    []sdk.ValAddress
	delAddrs    []sdk.AccAddress
	govAddrs    []types.GovernorAddress
	totalBonded int64
	validators  []stakingtypes.Validator
	delegations []stakingtypes.Delegation

	keeper *keeper.Keeper
	ctx    sdk.Context
	mocks  mocks

	activeGovernors  []v1.Governor
	inactiveGovernor v1.Governor
	proposal         v1.Proposal
}

// newFixture returns a configured fixture for testing the govKeeper methods.
// - track staking delegations and ensure the mock staking keeper replies
// accordingly.
// - setup 1 active and 1 inactive governors
// - initiates the validators with a self delegation of 1:
//   - setup IterateBondedValidatorsByPower call
//   - setup IterateDelegations call for validators
func newFixture(t *testing.T, ctx sdk.Context, numVals, numDelegators,
	numGovernors int, govKeeper *keeper.Keeper, mocks mocks,
) *fixture {
	var (
		addrs    = simtestutil.CreateRandomAccounts(numVals + numDelegators + numGovernors)
		valAddrs = simtestutil.ConvertAddrsToValAddrs(addrs[:numVals])
		delAddrs = addrs[numVals : numVals+numDelegators]
		govAddrs = convertAddrsToGovAddrs(addrs[numVals+numDelegators:])
		s        = &fixture{
			t:        t,
			ctx:      ctx,
			valAddrs: valAddrs,
			delAddrs: delAddrs,
			govAddrs: govAddrs,
			keeper:   govKeeper,
			mocks:    mocks,
		}
	)
	mocks.stakingKeeper.EXPECT().TotalBondedTokens(gomock.Any()).
		DoAndReturn(func(_ context.Context) (math.Int, error) {
			return math.NewInt(s.totalBonded), nil
		}).MaxTimes(1)

	// Mocks a bunch of validators
	for i := 0; i < len(valAddrs); i++ {
		s.validators = append(s.validators, stakingtypes.Validator{
			OperatorAddress: valAddrs[i].String(),
			Status:          stakingtypes.Bonded,
			Tokens:          math.ZeroInt(),
			DelegatorShares: math.LegacyZeroDec(),
		})
		// validator self delegation
		s.delegate(sdk.AccAddress(valAddrs[i]), valAddrs[i], 1)
	}
	mocks.stakingKeeper.EXPECT().
		IterateBondedValidatorsByPower(ctx, gomock.Any()).
		DoAndReturn(
			func(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) error {
				for i := 0; i < len(valAddrs); i++ {
					fn(int64(i), s.validators[i])
				}
				return nil
			}).AnyTimes()
	mocks.stakingKeeper.EXPECT().
		IterateDelegations(ctx, gomock.Any(), gomock.Any()).
		DoAndReturn(
			func(ctx context.Context, voter sdk.AccAddress, fn func(index int64, d stakingtypes.DelegationI) bool) error {
				for i, d := range s.delegations {
					if d.DelegatorAddress == voter.String() {
						fn(int64(i), d)
					}
				}
				return nil
			}).AnyTimes()
	mocks.stakingKeeper.EXPECT().GetValidator(ctx, gomock.Any()).DoAndReturn(
		func(ctx context.Context, addr sdk.ValAddress) (stakingtypes.ValidatorI, bool) {
			for i := 0; i < len(valAddrs); i++ {
				if valAddrs[i].String() == addr.String() {
					return s.validators[i], true
				}
			}
			return nil, false
		}).AnyTimes()
	mocks.stakingKeeper.EXPECT().GetDelegation(ctx, gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, del sdk.AccAddress, val sdk.ValAddress) (stakingtypes.Delegation, bool) {
			for _, d := range s.delegations {
				if d.DelegatorAddress == del.String() && d.ValidatorAddress == val.String() {
					return d, true
				}
			}
			return stakingtypes.Delegation{}, false
		}).AnyTimes()

	// Create active governors
	for i := 0; i < len(govAddrs)-1; i++ {
		governor, err := v1.NewGovernor(govAddrs[i].String(), v1.GovernorDescription{}, time.Now())
		require.NoError(t, err)
		err = govKeeper.Governors.Set(ctx, governor.GetAddress(), governor)
		require.NoError(t, err)
		s.activeGovernors = append(s.activeGovernors, governor)
	}
	// Create one inactive governor
	inactiveGovAddr := govAddrs[len(govAddrs)-1]
	governor, err := v1.NewGovernor(inactiveGovAddr.String(), v1.GovernorDescription{}, time.Now())
	require.NoError(t, err)
	governor.Status = v1.Inactive
	err = govKeeper.Governors.Set(ctx, governor.GetAddress(), governor)
	require.NoError(t, err)
	s.inactiveGovernor = governor
	return s
}

// delegate updates the tallyFixture delegations and validators fields.
// WARNING: delegate must be called *after* any calls to govKeeper.DelegateToGovernor
// because the hooks are not invoked in this test setup.
func (s *fixture) delegate(delegator sdk.AccAddress, validator sdk.ValAddress, m int64) {
	// Increment total bonded according to each delegations
	s.totalBonded += m
	delegation := stakingtypes.Delegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
	}
	// Increase validator shares and tokens, compute delegation.Shares
	for i := 0; i < len(s.validators); i++ {
		if s.validators[i].OperatorAddress == validator.String() {
			s.validators[i], delegation.Shares = s.validators[i].AddTokensFromDel(math.NewInt(m))
			break
		}
	}
	s.delegations = append(s.delegations, delegation)
}

// vote calls govKeeper.Vote()
func (s *fixture) vote(voter sdk.AccAddress, vote v1.VoteOption) {
	err := s.keeper.AddVote(s.ctx, s.proposal.Id, voter, v1.NewNonSplitVoteOption(vote), "")
	require.NoError(s.t, err)
}

func (s *fixture) validatorVote(voter sdk.ValAddress, vote v1.VoteOption) {
	s.vote(sdk.AccAddress(voter), vote)
}

func (s *fixture) governorVote(voter types.GovernorAddress, vote v1.VoteOption) {
	s.vote(sdk.AccAddress(voter), vote)
}
