package v036

import (
	"github.com/cosmos/cosmos-sdk/codec"
	extypes "github.com/cosmos/cosmos-sdk/contrib/migrate/types"
	v034gov "github.com/cosmos/cosmos-sdk/contrib/migrate/v034/gov"
	v036gov "github.com/cosmos/cosmos-sdk/contrib/migrate/v036/gov"
)

func migrateGovernance(initialState v034gov.GenesisState) v036gov.GenesisState {
	targetGov := v036gov.GenesisState{
		StartingProposalID: initialState.StartingProposalID,
		DepositParams: v036gov.DepositParams{
			MinDeposit:       initialState.DepositParams.MinDeposit,
			MaxDepositPeriod: initialState.DepositParams.MaxDepositPeriod,
		},
		TallyParams: v036gov.TallyParams{
			Quorum:    initialState.TallyParams.Quorum,
			Threshold: initialState.TallyParams.Threshold,
			Veto:      initialState.TallyParams.Veto,
		},
		VotingParams: v036gov.VotingParams{
			VotingPeriod: initialState.VotingParams.VotingPeriod,
		},
	}

	var deposits v036gov.Deposits
	for _, p := range initialState.Deposits {
		deposits = append(deposits, v036gov.Deposit{
			ProposalID: p.Deposit.ProposalID,
			Amount:     p.Deposit.Amount,
			Depositor:  p.Deposit.Depositor,
		})
	}

	targetGov.Deposits = deposits

	var votes v036gov.Votes
	for _, p := range initialState.Votes {
		votes = append(votes, v036gov.Vote{
			ProposalID: p.Vote.ProposalID,
			Option:     v036gov.VoteOption(p.Vote.Option),
			Voter:      p.Vote.Voter,
		})
	}

	targetGov.Votes = votes

	var proposals v036gov.Proposals
	for _, p := range initialState.Proposals {
		proposal := v036gov.Proposal{
			Content:          migrateContent(p.ProposalContent),
			ProposalID:       p.ProposalID,
			Status:           v036gov.ProposalStatus(p.Status),
			FinalTallyResult: v036gov.TallyResult(p.FinalTallyResult),
			SubmitTime:       p.SubmitTime,
			DepositEndTime:   p.DepositEndTime,
			TotalDeposit:     p.TotalDeposit,
			VotingStartTime:  p.VotingStartTime,
			VotingEndTime:    p.VotingEndTime,
		}

		proposals = append(proposals, proposal)
	}

	targetGov.Proposals = proposals
	return targetGov
}

func migrateContent(proposalContent v034gov.ProposalContent) (content v036gov.Content) {
	switch proposalContent.ProposalType() {
	case v034gov.ProposalTypeText:
		return v036gov.NewTextProposal(proposalContent.GetTitle(), proposalContent.GetDescription())
	case v034gov.ProposalTypeSoftwareUpgrade:
		return v036gov.NewSoftwareUpgradeProposal(proposalContent.GetTitle(), proposalContent.GetDescription())
	default:
		return nil
	}
}

// Migrate - unmarshal with the previous version and marshal with the new types
func Migrate(appState extypes.AppMap, cdc *codec.Codec) extypes.AppMap {
	v034Codec := codec.New()
	codec.RegisterCrypto(v034Codec)
	v036Codec := codec.New()
	codec.RegisterCrypto(v036Codec)

	if appState[v034gov.ModuleName] != nil {
		var govState v034gov.GenesisState
		v034gov.RegisterCodec(v034Codec)
		v034Codec.MustUnmarshalJSON(appState[v034gov.ModuleName], &govState)
		v036gov.RegisterCodec(v036Codec)
		delete(appState, v034gov.ModuleName) // Drop old key, in case it changed name
		appState[v036gov.ModuleName] = v036Codec.MustMarshalJSON(migrateGovernance(govState))
	}
	return appState
}
