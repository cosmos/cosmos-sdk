package v1_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func TestEmptyGenesis(t *testing.T) {
	state1 := v1.GenesisState{}
	require.True(t, state1.Empty())

	state2 := v1.DefaultGenesisState()
	require.False(t, state2.Empty())
}

func TestValidateGenesis(t *testing.T) {
	testCases := []struct {
		name         string
		genesisState func() *v1.GenesisState
		expErrMsg    string
	}{
		{
			name: "valid",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
		},
		{
			name: "invalid StartingProposalId",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				return v1.NewGenesisState(0, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "starting proposal id must be greater than 0",
		},
		{
			name: "min deposit throttler is nil",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.MinDepositThrottler = nil
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "min deposit throttler must not be nil",
		},
		{
			name: "invalid min deposit throttler floor",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				minDepositThrottler1 := *params.MinDepositThrottler
				minDepositThrottler1.FloorValue = sdk.Coins{{
					Denom:  sdk.DefaultBondDenom,
					Amount: math.NewInt(-100),
				}}
				params.MinDepositThrottler = &minDepositThrottler1

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "invalid minimum deposit floor",
		},
		{
			name: "min deposit throttler update period is nil",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				mdt.UpdatePeriod = nil
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "minimum deposit update period must not be nil",
		},
		{
			name: "min deposit throttler update period is 0",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				d := time.Duration(0)
				mdt.UpdatePeriod = &d
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "minimum deposit update period must be positive: 0s",
		},
		{
			name: "min deposit throttler update period is greater than voting period",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				d := *params.VotingPeriod + 1
				mdt.UpdatePeriod = &d
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "minimum deposit update period must be less than or equal to the voting period: 504h0m0.000000001s",
		},
		{
			name: "min deposit throttler sensitivity target distance is 0",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				mdt.DecreaseSensitivityTargetDistance = 0
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "minimum deposit sensitivity target distance must be positive: 0",
		},
		{
			name: "invalid min deposit throttler increase ratio",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				mdt.IncreaseRatio = ""
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "invalid minimum deposit increase ratio: decimal string cannot be empty",
		},
		{
			name: "min deposit throttler increase ratio is 0",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				mdt.IncreaseRatio = "0"
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "minimum deposit increase ratio must be positive: 0.000000000000000000",
		},
		{
			name: "min deposit throttler increase ratio is 1",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				mdt.IncreaseRatio = "1"
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "minimum deposit increase ratio too large: 1.000000000000000000",
		},
		{
			name: "invalid min deposit throttler decrease ratio",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				mdt.DecreaseRatio = ""
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "invalid minimum deposit decrease ratio: decimal string cannot be empty",
		},
		{
			name: "min deposit throttler decrease ratio is 0",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				mdt.DecreaseRatio = "0"
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "minimum deposit decrease ratio must be positive: 0.000000000000000000",
		},
		{
			name: "min deposit throttler decrease ratio is 1",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				mdt := *params.MinDepositThrottler
				mdt.DecreaseRatio = "1"
				params.MinDepositThrottler = &mdt
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "minimum deposit decrease ratio too large: 1.000000000000000000",
		},
		{
			name: "min deposit is deprecated",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.MinDeposit = sdk.NewCoins(sdk.NewInt64Coin("xxx", 1))

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "manually setting min deposit is deprecated in favor of a dynamic min deposit",
		},
		{
			name: "quorum timeout is nil",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.QuorumCheckCount = 1
				params.QuorumTimeout = nil

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "quorum timeout must not be nil",
		},
		{
			name: "quorum timeout is negative",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.QuorumCheckCount = 1
				d := time.Duration(-1)
				params.QuorumTimeout = &d

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "quorum timeout must be 0 or greater: -1ns",
		},
		{
			name: "quorum timeout is equal to voting period",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.QuorumCheckCount = 1
				params.QuorumTimeout = params.VotingPeriod

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "quorum timeout 504h0m0s must be strictly less than the voting period 504h0m0s",
		},
		{
			name: "max voting period extension is nil",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.QuorumCheckCount = 1
				params.MaxVotingPeriodExtension = nil
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "max voting period extension must not be nil",
		},
		{
			name: "max voting period extension less than voting period - quorum time out",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.QuorumCheckCount = 1
				d := time.Duration(-1)
				params.MaxVotingPeriodExtension = &d
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "max voting period extension -1ns must be greater than or equal to the difference between the voting period 504h0m0s and the quorum timeout 480h0m0s",
		},
		{
			name: "invalid max deposit period",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.MaxDepositPeriod = nil

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "maximum deposit period must not be nil",
		},
		{
			name: "invalid min quorum",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.QuorumRange.Min = "2"

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "quorumRange.min too large",
		},
		{
			name: "invalid max quorum",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.QuorumRange.Max = "2"

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "quorumRange.max too large",
		},
		{
			name: "invalid threshold",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.Threshold = "2"

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "vote threshold too large",
		},
		{
			name: "invalid min constitution amendment quorum",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.ConstitutionAmendmentQuorumRange.Min = "2"

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "constitutionAmendmentQuorumRange.min too large",
		},
		{
			name: "invalid max constitution amendment quorum",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.ConstitutionAmendmentQuorumRange.Max = "2"

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "constitutionAmendmentQuorumRange.max too large",
		},
		{
			name: "invalid constitution amendment threshold",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.ConstitutionAmendmentThreshold = "-1"

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "constitution amendment threshold must be positive",
		},
		{
			name: "invalid min law quorum",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.LawQuorumRange.Min = "2"

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "lawQuorumRange.min too large",
		},
		{
			name: "invalid max law quorum",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.LawQuorumRange.Max = "2"

				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "lawQuorumRange.max too large",
		},
		{
			name: "invalid law threshold",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				params.LawThreshold = "-2"
				return v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
			},
			expErrMsg: "law threshold must be positive",
		},
		{
			name: "duplicate proposals",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				state := v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
				state.Proposals = append(state.Proposals, &v1.Proposal{Id: 1})
				state.Proposals = append(state.Proposals, &v1.Proposal{Id: 1})

				return state
			},
			expErrMsg: "duplicate proposal id: 1",
		},
		{
			name: "duplicate votes",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				state := v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
				state.Proposals = append(state.Proposals, &v1.Proposal{Id: 1})
				state.Votes = append(state.Votes,
					&v1.Vote{
						ProposalId: 1,
						Voter:      "voter",
					},
					&v1.Vote{
						ProposalId: 1,
						Voter:      "voter",
					})

				return state
			},
			expErrMsg: "duplicate vote",
		},
		{
			name: "duplicate deposits",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				state := v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
				state.Proposals = append(state.Proposals, &v1.Proposal{Id: 1})
				state.Deposits = append(state.Deposits,
					&v1.Deposit{
						ProposalId: 1,
						Depositor:  "depositor",
					},
					&v1.Deposit{
						ProposalId: 1,
						Depositor:  "depositor",
					})

				return state
			},
			expErrMsg: "duplicate deposit: proposal_id:1 depositor:\"depositor\"",
		},
		{
			name: "non-existent proposal id in votes",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				state := v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
				state.Votes = append(state.Votes,
					&v1.Vote{
						ProposalId: 1,
						Voter:      "voter",
					})

				return state
			},
			expErrMsg: "vote proposal_id:1 voter:\"voter\"  has non-existent proposal id: 1",
		},
		{
			name: "non-existent proposal id in deposits",
			genesisState: func() *v1.GenesisState {
				params := v1.DefaultParams()
				state := v1.NewGenesisState(v1.DefaultStartingProposalID, v1.DefaultParticipationEma, v1.DefaultParticipationEma, v1.DefaultParticipationEma, params)
				state.Deposits = append(state.Deposits,
					&v1.Deposit{
						ProposalId: 1,
						Depositor:  "depositor",
					})

				return state
			},
			expErrMsg: "deposit proposal_id:1 depositor:\"depositor\"",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := v1.ValidateGenesis(tc.genesisState())
			if tc.expErrMsg != "" {
				require.Error(t, err)
				require.ErrorContains(t, err, tc.expErrMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
