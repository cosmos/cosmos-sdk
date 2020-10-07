package v040

import (
	"fmt"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func migrateVoteOption(oldVoteOption v034gov.VoteOption) v040gov.VoteOption {
	switch oldVoteOption {
	case v034gov.OptionEmpty:
		return v040gov.OptionEmpty

	case v034gov.OptionYes:
		return v040gov.OptionYes

	case v034gov.OptionAbstain:
		return v040gov.OptionAbstain

	case v034gov.OptionNo:
		return v040gov.OptionNo

	case v034gov.OptionNoWithVeto:
		return v040gov.OptionNoWithVeto

	default:
		panic(fmt.Errorf("'%s' is not a valid vote option", oldVoteOption))
	}
}

func migrateProposalStatus(oldProposalStatus v034gov.ProposalStatus) v040gov.ProposalStatus {
	switch oldProposalStatus {

	case v034gov.StatusNil:
		return v040gov.StatusNil

	case v034gov.StatusDepositPeriod:
		return v040gov.StatusDepositPeriod

	case v034gov.StatusVotingPeriod:
		return v040gov.StatusVotingPeriod

	case v034gov.StatusPassed:
		return v040gov.StatusPassed

	case v034gov.StatusRejected:
		return v040gov.StatusRejected

	case v034gov.StatusFailed:
		return v040gov.StatusFailed

	default:
		panic(fmt.Errorf("'%s' is not a valid proposal status", oldProposalStatus))
	}
}

func migrateContent(oldContent v036gov.Content) *codectypes.Any {
	switch oldContent := oldContent.(type) {
	case *v040gov.TextProposal:
		{
			// Convert the content into Any.
			contentAny, err := codectypes.NewAnyWithValue(oldContent)
			if err != nil {
				panic(err)
			}

			return contentAny
		}
	default:
		panic(fmt.Errorf("'%T' is not a valid proposal content type", oldContent))
	}
}

// Migrate accepts exported v0.36 x/gov genesis state and migrates it to
// v0.40 x/gov genesis state. The migration includes:
//
// - Convert vote option & proposal status from byte to enum.
// - Migrate proposal content to Any.
// - Convert addresses from bytes to bech32 strings.
// - Re-encode in v0.40 GenesisState.
func Migrate(oldGovState v036gov.GenesisState) *v040gov.GenesisState {
	newDeposits := make([]v040gov.Deposit, len(oldGovState.Deposits))
	for i, oldDeposit := range oldGovState.Deposits {
		newDeposits[i] = v040gov.Deposit{
			ProposalId: oldDeposit.ProposalID,
			Depositor:  oldDeposit.Depositor.String(),
			Amount:     oldDeposit.Amount,
		}
	}

	newVotes := make([]v040gov.Vote, len(oldGovState.Votes))
	for i, oldVote := range oldGovState.Votes {
		newVotes[i] = v040gov.Vote{
			ProposalId: oldVote.ProposalID,
			Voter:      oldVote.Voter.String(),
			Option:     migrateVoteOption(oldVote.Option),
		}
	}

	newProposals := make([]v040gov.Proposal, len(oldGovState.Proposals))
	for i, oldProposal := range oldGovState.Proposals {
		newProposals[i] = v040gov.Proposal{
			ProposalId: oldProposal.ProposalID,
			Content:    migrateContent(oldProposal.Content),
			Status:     migrateProposalStatus(oldProposal.Status),
			FinalTallyResult: v040gov.TallyResult{
				Yes:        oldProposal.FinalTallyResult.Yes,
				Abstain:    oldProposal.FinalTallyResult.Abstain,
				No:         oldProposal.FinalTallyResult.No,
				NoWithVeto: oldProposal.FinalTallyResult.NoWithVeto,
			},
			SubmitTime:      oldProposal.SubmitTime,
			DepositEndTime:  oldProposal.DepositEndTime,
			TotalDeposit:    oldProposal.TotalDeposit,
			VotingStartTime: oldProposal.VotingStartTime,
			VotingEndTime:   oldProposal.VotingEndTime,
		}
	}

	return &v040gov.GenesisState{
		StartingProposalId: oldGovState.StartingProposalID,
		Deposits:           newDeposits,
		Votes:              newVotes,
		Proposals:          newProposals,
		DepositParams: v040gov.DepositParams{
			MinDeposit:       oldGovState.DepositParams.MinDeposit,
			MaxDepositPeriod: oldGovState.DepositParams.MaxDepositPeriod,
		},
		VotingParams: v040gov.VotingParams{
			VotingPeriod: oldGovState.VotingParams.VotingPeriod,
		},
		TallyParams: v040gov.TallyParams{
			Quorum:        oldGovState.TallyParams.Quorum,
			Threshold:     oldGovState.TallyParams.Threshold,
			VetoThreshold: oldGovState.TallyParams.Veto,
		},
	}
}
