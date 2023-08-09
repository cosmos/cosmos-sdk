package keeper_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	sdkmath "cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/codec/address"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestTally(t *testing.T) {
	type suite struct {
		t        *testing.T
		proposal v1.Proposal
		valAddrs []sdk.ValAddress
		delAddrs []sdk.AccAddress
		keeper   *keeper.Keeper
		ctx      sdk.Context
		mocks    mocks
	}

	var (
		// handy functions
		setTotalBonded = func(s suite, n int64) {
			s.mocks.stakingKeeper.EXPECT().ValidatorAddressCodec().Return(address.NewBech32Codec("cosmosvaloper")).AnyTimes()
			s.mocks.stakingKeeper.EXPECT().TotalBondedTokens(gomock.Any()).
				Return(sdkmath.NewInt(n), nil)
		}
		delegatorVote = func(s suite, voter sdk.AccAddress, delegations []stakingtypes.Delegation, vote v1.VoteOption) {
			err := s.keeper.AddVote(s.ctx, s.proposal.Id, voter, v1.NewNonSplitVoteOption(vote), "")
			require.NoError(s.t, err)
			s.mocks.stakingKeeper.EXPECT().
				IterateDelegations(s.ctx, voter, gomock.Any()).
				DoAndReturn(
					func(ctx context.Context, voter sdk.AccAddress, fn func(index int64, d stakingtypes.DelegationI) bool) error {
						for i, d := range delegations {
							fn(int64(i), d)
						}
						return nil
					})
		}
		validatorVote = func(s suite, voter sdk.ValAddress, vote v1.VoteOption) {
			// validatorVote is like delegatorVote but without delegations
			delegatorVote(s, sdk.AccAddress(voter), nil, vote)
		}
	)
	tests := []struct {
		name          string
		expedited     bool
		setup         func(suite)
		expectedPass  bool
		expectedBurn  bool
		expectedTally v1.TallyResult
		expectedError string
	}{
		{
			name: "no votes, no bonded tokens: prop fails",
			setup: func(s suite) {
				setTotalBonded(s, 0)
			},
			expectedPass: false,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:        "0",
				AbstainCount:    "0",
				NoCount:         "0",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "no votes: prop fails/burn deposit",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
			},
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:        "0",
				AbstainCount:    "0",
				NoCount:         "0",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "one validator votes: prop fails/burn deposit",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
			},
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:        "0",
				AbstainCount:    "0",
				NoCount:         "1000000",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "one account votes without delegation: prop fails/burn deposit",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				delegatorVote(s, s.delAddrs[0], nil, v1.VoteOption_VOTE_OPTION_YES)
			},
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:        "0",
				AbstainCount:    "0",
				NoCount:         "0",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "one delegator votes: prop fails/burn deposit",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				delegations := []stakingtypes.Delegation{{
					DelegatorAddress: s.delAddrs[0].String(),
					ValidatorAddress: s.valAddrs[0].String(),
					Shares:           sdkmath.LegacyNewDec(42),
				}}
				delegatorVote(s, s.delAddrs[0], delegations, v1.VoteOption_VOTE_OPTION_YES)
			},
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:        "42",
				AbstainCount:    "0",
				NoCount:         "0",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "one delegator votes yes, validator votes also yes: prop fails/burn deposit",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				delegations := []stakingtypes.Delegation{{
					DelegatorAddress: s.delAddrs[0].String(),
					ValidatorAddress: s.valAddrs[0].String(),
					Shares:           sdkmath.LegacyNewDec(42),
				}}
				delegatorVote(s, s.delAddrs[0], delegations, v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
			},
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:        "1000000",
				AbstainCount:    "0",
				NoCount:         "0",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "one delegator votes yes, validator votes no: prop fails/burn deposit",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				delegations := []stakingtypes.Delegation{{
					DelegatorAddress: s.delAddrs[0].String(),
					ValidatorAddress: s.valAddrs[0].String(),
					Shares:           sdkmath.LegacyNewDec(42),
				}}
				delegatorVote(s, s.delAddrs[0], delegations, v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
			},
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:        "42",
				AbstainCount:    "0",
				NoCount:         "999958",
				NoWithVetoCount: "0",
			},
		},
		{
			// one delegator delegates 42 shares to 2 different validators (21 each)
			// delegator votes yes
			// first validator votes yes
			// second validator votes no
			// third validator (no delegation) votes abstain
			name: "delegator with mixed delegations: prop fails/burn deposit",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				delegations := []stakingtypes.Delegation{
					{
						DelegatorAddress: s.delAddrs[0].String(),
						ValidatorAddress: s.valAddrs[0].String(),
						Shares:           sdkmath.LegacyNewDec(21),
					},
					{
						DelegatorAddress: s.delAddrs[0].String(),
						ValidatorAddress: s.valAddrs[1].String(),
						Shares:           sdkmath.LegacyNewDec(21),
					},
				}
				delegatorVote(s, s.delAddrs[0], delegations, v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
				validatorVote(s, s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[2], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:        "1000021",
				AbstainCount:    "1000000",
				NoCount:         "999979",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "quorum reached with only abstain: prop fails",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				validatorVote(s, s.valAddrs[1], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				validatorVote(s, s.valAddrs[2], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				validatorVote(s, s.valAddrs[3], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			expectedPass: false,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:        "0",
				AbstainCount:    "4000000",
				NoCount:         "0",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "quorum reached with veto>1/3: prop fails/burn deposit",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[2], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[4], v1.VoteOption_VOTE_OPTION_NO_WITH_VETO)
				validatorVote(s, s.valAddrs[5], v1.VoteOption_VOTE_OPTION_NO_WITH_VETO)
				validatorVote(s, s.valAddrs[6], v1.VoteOption_VOTE_OPTION_NO_WITH_VETO)
			},
			expectedPass: false,
			expectedBurn: true,
			expectedTally: v1.TallyResult{
				YesCount:        "4000000",
				AbstainCount:    "0",
				NoCount:         "0",
				NoWithVetoCount: "3000000",
			},
		},
		{
			name: "quorum reached with yes<=.5: prop fails",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
				validatorVote(s, s.valAddrs[3], v1.VoteOption_VOTE_OPTION_NO)
			},
			expectedPass: false,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:        "2000000",
				AbstainCount:    "0",
				NoCount:         "2000000",
				NoWithVetoCount: "0",
			},
		},
		{
			name: "quorum reached with yes>.5: prop succeeds",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[2], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[4], v1.VoteOption_VOTE_OPTION_NO)
				validatorVote(s, s.valAddrs[5], v1.VoteOption_VOTE_OPTION_NO)
				validatorVote(s, s.valAddrs[6], v1.VoteOption_VOTE_OPTION_NO_WITH_VETO)
			},
			expectedPass: true,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:        "4000000",
				AbstainCount:    "0",
				NoCount:         "2000000",
				NoWithVetoCount: "1000000",
			},
		},
		{
			name: "quorum reached thanks to abstain, yes>.5: prop succeeds",
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
				validatorVote(s, s.valAddrs[3], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				validatorVote(s, s.valAddrs[4], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				validatorVote(s, s.valAddrs[5], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			expectedPass: true,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:        "2000000",
				AbstainCount:    "3000000",
				NoCount:         "1000000",
				NoWithVetoCount: "0",
			},
		},
		{
			name:      "quorum reached with yes<=.667: expedited prop fails",
			expedited: true,
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[2], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[4], v1.VoteOption_VOTE_OPTION_NO)
				validatorVote(s, s.valAddrs[5], v1.VoteOption_VOTE_OPTION_NO)
				validatorVote(s, s.valAddrs[6], v1.VoteOption_VOTE_OPTION_NO_WITH_VETO)
			},
			expectedPass: false,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:        "4000000",
				AbstainCount:    "0",
				NoCount:         "2000000",
				NoWithVetoCount: "1000000",
			},
		},
		{
			name:      "quorum reached with yes>.667: expedited prop succeeds",
			expedited: true,
			setup: func(s suite) {
				setTotalBonded(s, 10000000)
				validatorVote(s, s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[2], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[4], v1.VoteOption_VOTE_OPTION_YES)
				validatorVote(s, s.valAddrs[5], v1.VoteOption_VOTE_OPTION_NO)
				validatorVote(s, s.valAddrs[6], v1.VoteOption_VOTE_OPTION_NO_WITH_VETO)
			},
			expectedPass: true,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:        "5000000",
				AbstainCount:    "0",
				NoCount:         "1000000",
				NoWithVetoCount: "1000000",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			govKeeper, mocks, _, ctx := setupGovKeeper(t, mockAccountKeeperExpectations)
			params := v1.DefaultParams()
			// Ensure params value are different than false
			params.BurnVoteQuorum = true
			params.BurnVoteVeto = true
			err := govKeeper.Params.Set(ctx, params)
			require.NoError(t, err)
			var (
				numVals       = 10
				numDelegators = 5
				addrs         = simtestutil.CreateRandomAccounts(numVals + numDelegators)
				valAddrs      = simtestutil.ConvertAddrsToValAddrs(addrs[:numVals])
				delAddrs      = addrs[numVals:]
			)
			// Mocks a bunch of validators
			mocks.stakingKeeper.EXPECT().
				IterateBondedValidatorsByPower(ctx, gomock.Any()).
				DoAndReturn(
					func(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) error {
						for i := int64(0); i < int64(numVals); i++ {
							fn(i, stakingtypes.Validator{
								OperatorAddress: valAddrs[i].String(),
								Status:          stakingtypes.Bonded,
								Tokens:          sdkmath.NewInt(1000000),
								DelegatorShares: sdkmath.LegacyNewDec(1000000),
							})
						}
						return nil
					})
			// Submit and activate a proposal
			proposal, err := govKeeper.SubmitProposal(ctx, TestProposal, "", "title", "summary", delAddrs[0], tt.expedited)
			require.NoError(t, err)
			err = govKeeper.ActivateVotingPeriod(ctx, proposal)
			require.NoError(t, err)
			suite := suite{
				t:        t,
				proposal: proposal,
				valAddrs: valAddrs,
				delAddrs: delAddrs,
				ctx:      ctx,
				keeper:   govKeeper,
				mocks:    mocks,
			}
			tt.setup(suite)

			pass, burn, tally, err := govKeeper.Tally(ctx, proposal)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPass, pass, "wrong pass")
			assert.Equal(t, tt.expectedBurn, burn, "wrong burn")
			assert.Equal(t, tt.expectedTally, tally)
			// Assert votes removal after tally
			rng := collections.NewPrefixedPairRange[uint64, sdk.AccAddress](proposal.Id)
			_, err = suite.keeper.Votes.Iterate(suite.ctx, rng)
			assert.Nil(t, err, collections.ErrInvalidIterator, "votes must be removed after tally")
		})
	}
}
