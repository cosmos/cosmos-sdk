package v040

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
	v040distr "github.com/cosmos/cosmos-sdk/x/distribution/types"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	v040gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"
	v040params "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v038"
	v040upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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
	var protoProposal proto.Message

	switch oldContent := oldContent.(type) {
	case v036gov.TextProposal:
		{
			protoProposal = &v040gov.TextProposal{
				Title:       oldContent.Title,
				Description: oldContent.Description,
			}
			// Convert the content into Any.
			contentAny, err := codectypes.NewAnyWithValue(protoProposal)
			if err != nil {
				panic(err)
			}

			return contentAny
		}
	case v036distr.CommunityPoolSpendProposal:
		{
			protoProposal = &v040distr.CommunityPoolSpendProposal{
				Title:       oldContent.Title,
				Description: oldContent.Description,
				Recipient:   oldContent.Recipient.String(),
				Amount:      oldContent.Amount,
			}
		}
	case v038upgrade.CancelSoftwareUpgradeProposal:
		{
			protoProposal = &v040upgrade.CancelSoftwareUpgradeProposal{
				Description: oldContent.Description,
				Title:       oldContent.Title,
			}
		}
	case v038upgrade.SoftwareUpgradeProposal:
		{
			protoProposal = &v040upgrade.SoftwareUpgradeProposal{
				Description: oldContent.Description,
				Title:       oldContent.Title,
				Plan: v040upgrade.Plan{
					Name:   oldContent.Plan.Name,
					Time:   oldContent.Plan.Time,
					Height: oldContent.Plan.Height,
					Info:   oldContent.Plan.Info,
				},
			}
		}
	case v036params.ParameterChangeProposal:
		{
			newChanges := make([]v040params.ParamChange, len(oldContent.Changes))
			for i, oldChange := range oldContent.Changes {
				newChanges[i] = v040params.ParamChange{
					Subspace: oldChange.Subspace,
					Key:      oldChange.Key,
					Value:    oldChange.Value,
				}
			}

			protoProposal = &v040params.ParameterChangeProposal{
				Description: oldContent.Description,
				Title:       oldContent.Title,
				Changes:     newChanges,
			}
		}
	default:
		panic(fmt.Errorf("%T is not a valid proposal content type", oldContent))
	}

	// Convert the content into Any.
	contentAny, err := codectypes.NewAnyWithValue(protoProposal)
	if err != nil {
		panic(err)
	}

	return contentAny
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
