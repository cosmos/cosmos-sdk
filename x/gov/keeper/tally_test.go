package keeper_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"cosmossdk.io/collections"
	"cosmossdk.io/math"

	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtestutil "github.com/cosmos/cosmos-sdk/x/gov/testutil"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type tallyFixture struct {
	t *testing.T

	proposal    v1.Proposal
	valAddrs    []sdk.ValAddress
	delAddrs    []sdk.AccAddress
	totalBonded int64
	validators  []stakingtypes.Validator
	delegations []stakingtypes.Delegation

	keeper        *keeper.Keeper
	ctx           sdk.Context
	acctKeeper    *govtestutil.MockAccountKeeper
	bankKeeper    *govtestutil.MockBankKeeper
	stakingKeeper *govtestutil.MockStakingKeeper
}

// newTallyFixture returns a configured fixture for testing the govKeeper.Tally
// method.
// - setup TotalBondedTokens call
// - initiates the validators with a self delegation of 1:
//   - setup IterateBondedValidatorsByPower call
//   - setup IterateDelegations call for validators
func newTallyFixture(t *testing.T, ctx sdk.Context, proposal v1.Proposal,
	valAddrs []sdk.ValAddress, delAddrs []sdk.AccAddress,
	govKeeper *keeper.Keeper,
	acctKeeper *govtestutil.MockAccountKeeper,
	bankKeeper *govtestutil.MockBankKeeper,
	stakingKeeper *govtestutil.MockStakingKeeper,
	modifier func(*tallyFixture),
) *tallyFixture {
	s := &tallyFixture{
		t:             t,
		ctx:           ctx,
		proposal:      proposal,
		valAddrs:      valAddrs,
		delAddrs:      delAddrs,
		keeper:        govKeeper,
		acctKeeper:    acctKeeper,
		bankKeeper:    bankKeeper,
		stakingKeeper: stakingKeeper,
	}

	stakingKeeper.EXPECT().TotalBondedTokens(gomock.Any()).
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
	stakingKeeper.EXPECT().
		IterateBondedValidatorsByPower(ctx, gomock.Any()).
		DoAndReturn(
			func(ctx context.Context, fn func(index int64, validator stakingtypes.ValidatorI) bool) error {
				for i := 0; i < len(valAddrs); i++ {
					fn(int64(i), s.validators[i])
				}
				return nil
			}).AnyTimes()
	stakingKeeper.EXPECT().
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

	if modifier != nil {
		modifier(s)
	}

	return s
}

// delegate updates the tallyFixture delegations and validators fields.
func (s *tallyFixture) delegate(delegator sdk.AccAddress, validator sdk.ValAddress, m int64) {
	// Increment total bonded according to each delegations
	s.totalBonded += m
	delegation := stakingtypes.Delegation{
		DelegatorAddress: delegator.String(),
		ValidatorAddress: validator.String(),
	}
	for i := 0; i < len(s.validators); i++ {
		if s.validators[i].OperatorAddress == validator.String() {
			s.validators[i], delegation.Shares = s.validators[i].AddTokensFromDel(math.NewInt(m))
			break
		}
	}
	s.delegations = append(s.delegations, delegation)
}

// vote calls govKeeper.Vote()
func (s *tallyFixture) vote(voter sdk.AccAddress, vote v1.VoteOption) {
	err := s.keeper.AddVote(s.ctx, s.proposal.Id, voter, v1.NewNonSplitVoteOption(vote), "")
	require.NoError(s.t, err)
}

func (s *tallyFixture) validatorVote(voter sdk.ValAddress, vote v1.VoteOption) {
	s.vote(sdk.AccAddress(voter), vote)
}

func TestTally(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*tallyFixture)
		proposalMsgs  []sdk.Msg
		expectedPass  bool
		expectedBurn  bool
		expectedTally v1.TallyResult
		expectedError string
		endorse       bool
	}{
		{
			name:         "no votes: prop fails/burn deposit",
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: true,
			expectedTally: v1.TallyResult{
				YesCount:     "0",
				AbstainCount: "0",
				NoCount:      "0",
			},
		},
		{
			name: "one validator votes: prop fails/burn deposit",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:     "0",
				AbstainCount: "0",
				NoCount:      "1",
			},
		},
		{
			name: "one account votes without delegation: prop fails/burn deposit",
			setup: func(s *tallyFixture) {
				s.vote(s.delAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:     "0",
				AbstainCount: "0",
				NoCount:      "0",
			},
		},
		{
			name: "one delegator votes: prop fails/burn deposit",
			setup: func(s *tallyFixture) {
				s.delegate(s.delAddrs[0], s.valAddrs[0], 2)
				s.vote(s.delAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:     "2",
				AbstainCount: "0",
				NoCount:      "0",
			},
		},
		{
			name: "one delegator votes yes, validator votes also yes: prop fails/burn deposit",
			setup: func(s *tallyFixture) {
				s.delegate(s.delAddrs[0], s.valAddrs[0], 1)
				s.vote(s.delAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:     "2",
				AbstainCount: "0",
				NoCount:      "0",
			},
		},
		{
			name: "one delegator votes yes, validator votes no: prop fails/burn deposit",
			setup: func(s *tallyFixture) {
				s.delegate(s.delAddrs[0], s.valAddrs[0], 1)
				s.vote(s.delAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:     "1",
				AbstainCount: "0",
				NoCount:      "1",
			},
		},
		{
			name: "validator votes yes, doesn't inherit delegations",
			setup: func(s *tallyFixture) {
				s.delegate(s.delAddrs[0], s.valAddrs[0], 2)
				s.delegate(s.delAddrs[1], s.valAddrs[0], 2)
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: true, // burn because quorum not reached
			expectedTally: v1.TallyResult{
				YesCount:     "0",
				AbstainCount: "0",
				NoCount:      "1",
			},
		},
		{
			// one delegator delegates 42 shares to 2 different validators (21 each)
			// delegator votes yes
			// first validator votes yes
			// second validator votes no
			// third validator (no delegation) votes abstain
			name: "delegator with mixed delegations: prop pass",
			setup: func(s *tallyFixture) {
				s.delegate(s.delAddrs[0], s.valAddrs[0], 2)
				s.delegate(s.delAddrs[0], s.valAddrs[1], 2)
				s.vote(s.delAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestProposal,
			expectedPass: true,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "5",
				AbstainCount: "1",
				NoCount:      "1",
			},
		},
		{
			name: "quorum reached with only abstain: prop fails",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "0",
				AbstainCount: "4",
				NoCount:      "0",
			},
		},
		{
			name: "quorum reached with yes<=.667: prop rejected",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[4], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[5], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[6], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[7], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[8], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "4",
				AbstainCount: "3",
				NoCount:      "2",
			},
		},
		{
			name: "quorum reached with yes>.667: prop passes",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[4], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[5], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[6], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[7], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[8], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestProposal,
			expectedPass: true,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "5",
				AbstainCount: "3",
				NoCount:      "1",
			},
		},
		{
			name: "quorum reached with no>.7: prop rejected and deposit burned",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[4], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[5], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[6], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[7], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[8], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestProposal,
			expectedPass: false,
			expectedBurn: true,
			expectedTally: v1.TallyResult{
				YesCount:     "1",
				AbstainCount: "3",
				NoCount:      "5",
			},
		},
		{
			name: "quorum reached thanks to abstain, yes>.667: prop passes",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[4], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[5], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestProposal,
			expectedPass: true,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "3",
				AbstainCount: "2",
				NoCount:      "1",
			},
		},
		{
			name: "amendment quorum not reached: prop rejected",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
			},
			proposalMsgs: TestAmendmentProposal,
			expectedPass: false,
			expectedBurn: true,
			expectedTally: v1.TallyResult{
				YesCount:     "2",
				AbstainCount: "0",
				NoCount:      "0",
			},
		},
		{
			name: "amendment quorum reached and threshold not reached: prop rejected",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[4], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[5], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestAmendmentProposal,
			expectedPass: false,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "2",
				AbstainCount: "3",
				NoCount:      "1",
			},
		},
		{
			name: "amendment quorum reached and threshold reached: prop passes",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[4], v1.VoteOption_VOTE_OPTION_YES)
				s.delegate(s.delAddrs[0], s.valAddrs[5], 2)
				s.delegate(s.delAddrs[0], s.valAddrs[6], 2)
				s.vote(s.delAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.delegate(s.delAddrs[1], s.valAddrs[5], 1)
				s.delegate(s.delAddrs[1], s.valAddrs[6], 1)
				s.vote(s.delAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
			},
			proposalMsgs: TestAmendmentProposal,
			expectedPass: true,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "10",
				AbstainCount: "0",
				NoCount:      "1",
			},
		},
		{
			name: "law quorum not reached: prop rejected",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
			},
			proposalMsgs: TestLawProposal,
			expectedPass: false,
			expectedBurn: true,
			expectedTally: v1.TallyResult{
				YesCount:     "1",
				AbstainCount: "0",
				NoCount:      "1",
			},
		},
		{
			name: "law quorum reached and threshold not reached: prop rejected",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[4], v1.VoteOption_VOTE_OPTION_ABSTAIN)
				s.validatorVote(s.valAddrs[5], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestLawProposal,
			expectedPass: false,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "2",
				AbstainCount: "3",
				NoCount:      "1",
			},
		},
		{
			name: "law quorum reached and threshold reached: prop passes",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[5], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs: TestLawProposal,
			expectedPass: true,
			expectedBurn: false,
			expectedTally: v1.TallyResult{
				YesCount:     "3",
				AbstainCount: "1",
				NoCount:      "0",
			},
		},
		{
			name: "law quorum reached and threshold not reached no endorse",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_NO)
			},
			proposalMsgs: TestLawProposal,
			expectedPass: false,
			expectedBurn: false,
			endorse:      false,
			expectedTally: v1.TallyResult{
				YesCount:     "3",
				AbstainCount: "0",
				NoCount:      "1",
			},
		},
		{
			name: "law quorum reached and threshold lowered with endorse",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_NO)
			},
			proposalMsgs: TestLawProposal,
			expectedPass: true,
			expectedBurn: false,
			endorse:      true,
			expectedTally: v1.TallyResult{
				YesCount:     "3",
				AbstainCount: "0",
				NoCount:      "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			govKeeper, mockAccountKeeper, mockBankKeeper, mockStakingKeeper, _, _, ctx := setupGovKeeper(t, mockAccountKeeperExpectations)

			params := v1.DefaultParams()
			// Ensure params value are different than false
			params.BurnVoteQuorum = true
			err := govKeeper.Params.Set(ctx, params)
			require.NoError(t, err)
			var (
				numVals       = 10
				numDelegators = 5
				addrs         = simtestutil.CreateRandomAccounts(numVals + numDelegators)
				valAddrs      = simtestutil.ConvertAddrsToValAddrs(addrs[:numVals])
				delAddrs      = addrs[numVals:]
			)
			// Submit and activate a proposal
			proposal, err := govKeeper.SubmitProposal(ctx, tt.proposalMsgs, "", "title", "summary", delAddrs[0])
			require.NoError(t, err)
			govKeeper.ActivateVotingPeriod(ctx, proposal)
			if tt.endorse {
				proposal.Endorsed = true
			}

			// Create the test fixture
			_ = newTallyFixture(t, ctx, proposal, valAddrs, delAddrs, govKeeper, mockAccountKeeper, mockBankKeeper, mockStakingKeeper, tt.setup)

			pass, burn, _, tally, err := govKeeper.Tally(ctx, proposal)

			require.NoError(t, err)
			assert.Equal(t, tt.expectedPass, pass, "wrong pass")
			assert.Equal(t, tt.expectedBurn, burn, "wrong burn")
			assert.Equal(t, tt.expectedTally, tally)
			votes := make([]v1.Vote, 0)
			err = govKeeper.Votes.Walk(ctx, nil, func(key collections.Pair[uint64, sdk.AccAddress], value v1.Vote) (bool, error) {
				votes = append(votes, value)
				return false, nil
			})
			assert.NoError(t, err)
			assert.Empty(t, votes, "votes not be removed after tally")
		})
	}
}

func TestHasReachedQuorum(t *testing.T) {
	tests := []struct {
		name           string
		proposalMsgs   []sdk.Msg
		setup          func(*tallyFixture)
		expectedQuorum bool
	}{
		{
			name:         "no votes: no quorum",
			proposalMsgs: TestProposal,
			setup: func(s *tallyFixture) {
			},
			expectedQuorum: false,
		},
		{
			name:         "not enough votes: no quorum",
			proposalMsgs: TestProposal,
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
			},
			expectedQuorum: false,
		},
		{
			name:         "enough votes: quorum",
			proposalMsgs: TestProposal,
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.delegate(s.delAddrs[0], s.valAddrs[2], 500000)
				s.delegate(s.delAddrs[0], s.valAddrs[3], 500000)
				s.delegate(s.delAddrs[0], s.valAddrs[4], 500000)
				s.delegate(s.delAddrs[0], s.valAddrs[5], 500000)
				s.vote(s.delAddrs[0], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			expectedQuorum: true,
		},
		{
			name: "amendment quorum not reached",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
			},
			proposalMsgs:   TestAmendmentProposal,
			expectedQuorum: false,
		},
		{
			name: "amendment quorum reached",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[4], v1.VoteOption_VOTE_OPTION_YES)
				s.delegate(s.delAddrs[0], s.valAddrs[5], 2)
				s.delegate(s.delAddrs[0], s.valAddrs[6], 2)
				s.vote(s.delAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.delegate(s.delAddrs[1], s.valAddrs[5], 1)
				s.delegate(s.delAddrs[1], s.valAddrs[6], 1)
				s.vote(s.delAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
			},
			proposalMsgs:   TestAmendmentProposal,
			expectedQuorum: true,
		},
		{
			name: "law quorum not reached",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[2], v1.VoteOption_VOTE_OPTION_NO)
			},
			proposalMsgs:   TestLawProposal,
			expectedQuorum: false,
		},
		{
			name: "law quorum reached",
			setup: func(s *tallyFixture) {
				s.validatorVote(s.valAddrs[0], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[1], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[3], v1.VoteOption_VOTE_OPTION_YES)
				s.validatorVote(s.valAddrs[5], v1.VoteOption_VOTE_OPTION_ABSTAIN)
			},
			proposalMsgs:   TestLawProposal,
			expectedQuorum: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			govKeeper, mockAccountKeeper, mockBankKeeper, mockStakingKeeper, _, _, ctx := setupGovKeeper(t, mockAccountKeeperExpectations)
			var (
				numVals       = 10
				numDelegators = 5
				addrs         = simtestutil.CreateRandomAccounts(numVals + numDelegators)
				valAddrs      = simtestutil.ConvertAddrsToValAddrs(addrs[:numVals])
				delAddrs      = addrs[numVals:]
			)
			// Submit and activate a proposal
			proposal, err := govKeeper.SubmitProposal(ctx, tt.proposalMsgs, "", "title", "summary", delAddrs[0])
			require.NoError(t, err)
			govKeeper.ActivateVotingPeriod(ctx, proposal)

			suite := newTallyFixture(t, ctx, proposal, valAddrs, delAddrs, govKeeper, mockAccountKeeper, mockBankKeeper, mockStakingKeeper, tt.setup)

			quorum, err := govKeeper.HasReachedQuorum(ctx, proposal)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedQuorum, quorum)
			if tt.expectedQuorum {
				// Assert votes are still here after HasReachedQuorum
				votes := make([]v1.Vote, 0)
				err := suite.keeper.Votes.Walk(suite.ctx, nil, func(key collections.Pair[uint64, sdk.AccAddress], value v1.Vote) (bool, error) {
					votes = append(votes, value)
					return false, nil
				})
				require.NoError(t, err)
				assert.NotEmpty(t, votes, "votes must be kept after HasReachedQuorum")
			}
		})
	}
}
