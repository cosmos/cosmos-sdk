package v040

import (
	"fmt"

	proto "github.com/gogo/protobuf/proto"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	v036distr "github.com/cosmos/cosmos-sdk/x/distribution/legacy/v036"
	v040distr "github.com/cosmos/cosmos-sdk/x/distribution/types"
	v034gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v034"
	v036gov "github.com/cosmos/cosmos-sdk/x/gov/legacy/v036"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v036params "github.com/cosmos/cosmos-sdk/x/params/legacy/v036"
	v040params "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	v038upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v038"
	v040upgrade "github.com/cosmos/cosmos-sdk/x/upgrade/legacy/v040"
)

func migrateVoteOption(oldVoteOption v034gov.VoteOption) VoteOption {
	switch oldVoteOption {
	case v034gov.OptionEmpty:
		return OptionEmpty

	case v034gov.OptionYes:
		return OptionYes

	case v034gov.OptionAbstain:
		return OptionAbstain

	case v034gov.OptionNo:
		return OptionNo

	case v034gov.OptionNoWithVeto:
		return OptionNoWithVeto

	default:
		panic(fmt.Errorf("'%s' is not a valid vote option", oldVoteOption))
	}
}

func migrateProposalStatus(oldProposalStatus v034gov.ProposalStatus) ProposalStatus {
	switch oldProposalStatus {

	case v034gov.StatusNil:
		return StatusNil

	case v034gov.StatusDepositPeriod:
		return StatusDepositPeriod

	case v034gov.StatusVotingPeriod:
		return StatusVotingPeriod

	case v034gov.StatusPassed:
		return StatusPassed

	case v034gov.StatusRejected:
		return StatusRejected

	case v034gov.StatusFailed:
		return StatusFailed

	default:
		panic(fmt.Errorf("'%s' is not a valid proposal status", oldProposalStatus))
	}
}

func migrateContent(oldContent v036gov.Content) *codectypes.Any {
	var protoProposal proto.Message

	switch oldContent := oldContent.(type) {
	case v036gov.TextProposal:
		{
			protoProposal = &TextProposal{
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
func Migrate(oldGovState v036gov.GenesisState) *GenesisState {
	newDeposits := make([]Deposit, len(oldGovState.Deposits))
	for i, oldDeposit := range oldGovState.Deposits {
		newDeposits[i] = Deposit{
			ProposalId: oldDeposit.ProposalID,
			Depositor:  oldDeposit.Depositor.String(),
			Amount:     oldDeposit.Amount,
		}
	}

	newVotes := make([]Vote, len(oldGovState.Votes))
	for i, oldVote := range oldGovState.Votes {
		newVotes[i] = Vote{
			ProposalId: oldVote.ProposalID,
			Voter:      oldVote.Voter.String(),
			Option:     migrateVoteOption(oldVote.Option),
		}
	}

	newProposals := make([]Proposal, len(oldGovState.Proposals))
	for i, oldProposal := range oldGovState.Proposals {
		newProposals[i] = Proposal{
			ProposalId: oldProposal.ProposalID,
			Content:    migrateContent(oldProposal.Content),
			Status:     migrateProposalStatus(oldProposal.Status),
			FinalTallyResult: TallyResult{
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

	return &GenesisState{
		StartingProposalId: oldGovState.StartingProposalID,
		Deposits:           newDeposits,
		Votes:              newVotes,
		Proposals:          newProposals,
		DepositParams: DepositParams{
			MinDeposit:       oldGovState.DepositParams.MinDeposit,
			MaxDepositPeriod: oldGovState.DepositParams.MaxDepositPeriod,
		},
		VotingParams: VotingParams{
			VotingPeriod: oldGovState.VotingParams.VotingPeriod,
		},
		TallyParams: TallyParams{
			Quorum:        oldGovState.TallyParams.Quorum,
			Threshold:     oldGovState.TallyParams.Threshold,
			VetoThreshold: oldGovState.TallyParams.Veto,
		},
	}
}

func RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	registry.RegisterInterface(
		"cosmos.gov.v1beta1.Content",
		(*types.Content)(nil),
		&TextProposal{},
		&v040upgrade.SoftwareUpgradeProposal{},
		&v040upgrade.CancelSoftwareUpgradeProposal{},
		&v040distr.CommunityPoolSpendProposal{},
	)
}
