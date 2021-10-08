package v040

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/migrations/v036"
	v040distr "github.com/cosmos/cosmos-sdk/x/distribution/types"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v034"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v036"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	v036params "github.com/cosmos/cosmos-sdk/x/params/migrations/v036"
	v040params "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/migrations/v038"
	v040upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/types"
)

func migrateVoteOption(oldVoteOption v034gov.VoteOption) v1beta1.VoteOption {
	switch oldVoteOption {
	case v034gov.OptionEmpty:
		return v1beta1.OptionEmpty

	case v034gov.OptionYes:
		return v1beta1.OptionYes

	case v034gov.OptionAbstain:
		return v1beta1.OptionAbstain

	case v034gov.OptionNo:
		return v1beta1.OptionNo

	case v034gov.OptionNoWithVeto:
		return v1beta1.OptionNoWithVeto

	default:
		panic(fmt.Errorf("'%s' is not a valid vote option", oldVoteOption))
	}
}

func migrateProposalStatus(oldProposalStatus v034gov.ProposalStatus) v1beta1.ProposalStatus {
	switch oldProposalStatus {

	case v034gov.StatusNil:
		return v1beta1.StatusNil

	case v034gov.StatusDepositPeriod:
		return v1beta1.StatusDepositPeriod

	case v034gov.StatusVotingPeriod:
		return v1beta1.StatusVotingPeriod

	case v034gov.StatusPassed:
		return v1beta1.StatusPassed

	case v034gov.StatusRejected:
		return v1beta1.StatusRejected

	case v034gov.StatusFailed:
		return v1beta1.StatusFailed

	default:
		panic(fmt.Errorf("'%s' is not a valid proposal status", oldProposalStatus))
	}
}

func migrateContent(oldContent v036gov.Content) *codectypes.Any {
	var protoProposal proto.Message

	switch oldContent := oldContent.(type) {
	case v036gov.TextProposal:
		{
			protoProposal = &v1beta1.TextProposal{
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
func Migrate(oldGovState v036gov.GenesisState) *v1beta1.GenesisState {
	newDeposits := make([]v1beta1.Deposit, len(oldGovState.Deposits))
	for i, oldDeposit := range oldGovState.Deposits {
		newDeposits[i] = v1beta1.Deposit{
			ProposalId: oldDeposit.ProposalID,
			Depositor:  oldDeposit.Depositor.String(),
			Amount:     oldDeposit.Amount,
		}
	}

	newVotes := make([]v1beta1.Vote, len(oldGovState.Votes))
	for i, oldVote := range oldGovState.Votes {
		newVotes[i] = v1beta1.Vote{
			ProposalId: oldVote.ProposalID,
			Voter:      oldVote.Voter.String(),
			Option:     migrateVoteOption(oldVote.Option),
		}
	}

	newProposals := make([]v1beta1.Proposal, len(oldGovState.Proposals))
	for i, oldProposal := range oldGovState.Proposals {
		newProposals[i] = v1beta1.Proposal{
			ProposalId: oldProposal.ProposalID,
			Content:    migrateContent(oldProposal.Content),
			Status:     migrateProposalStatus(oldProposal.Status),
			FinalTallyResult: v1beta1.TallyResult{
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

	return &v1beta1.GenesisState{
		StartingProposalId: oldGovState.StartingProposalID,
		Deposits:           newDeposits,
		Votes:              newVotes,
		Proposals:          newProposals,
		DepositParams: v1beta1.DepositParams{
			MinDeposit:       oldGovState.DepositParams.MinDeposit,
			MaxDepositPeriod: oldGovState.DepositParams.MaxDepositPeriod,
		},
		VotingParams: v1beta1.VotingParams{
			VotingPeriod: oldGovState.VotingParams.VotingPeriod,
		},
		TallyParams: v1beta1.TallyParams{
			Quorum:        oldGovState.TallyParams.Quorum,
			Threshold:     oldGovState.TallyParams.Threshold,
			VetoThreshold: oldGovState.TallyParams.Veto,
		},
	}
}
