package v036

import (
	"github.com/cosmos/cosmos-sdk/codec"
	extypes "github.com/cosmos/cosmos-sdk/contrib/export/types"
	v034gov "github.com/cosmos/cosmos-sdk/contrib/export/v034/gov"
	v036gov "github.com/cosmos/cosmos-sdk/contrib/export/v036/gov"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	"github.com/cosmos/cosmos-sdk/x/distribution"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
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
	// migration below are only sanity check:
	// used as unmarshal guarantee,
	// so if anyone change the types without a migrations we panic
	// We could move this next steps to a test OR use the outcome on CCI for re-import.

	if appState[auth.ModuleName] != nil {
		var authState auth.GenesisState
		cdc.MustUnmarshalJSON(appState[auth.ModuleName], &authState)
		appState[auth.ModuleName] = cdc.MustMarshalJSON(authState)
	}

	if appState[bank.ModuleName] != nil {
		var bankState bank.GenesisState
		cdc.MustUnmarshalJSON(appState[bank.ModuleName], &bankState)
		appState[bank.ModuleName] = cdc.MustMarshalJSON(bankState)
	}

	if appState[crisis.ModuleName] != nil {
		var crisisState crisis.GenesisState
		cdc.MustUnmarshalJSON(appState[crisis.ModuleName], &crisisState)
		appState[crisis.ModuleName] = cdc.MustMarshalJSON(crisisState)
	}

	if appState[distribution.ModuleName] != nil {
		var distributionState distribution.GenesisState
		cdc.MustUnmarshalJSON(appState[distribution.ModuleName], &distributionState)
		appState[distribution.ModuleName] = cdc.MustMarshalJSON(distributionState)
	}

	if appState[mint.ModuleName] != nil {
		var mintState mint.GenesisState
		cdc.MustUnmarshalJSON(appState[mint.ModuleName], &mintState)
		appState[mint.ModuleName] = cdc.MustMarshalJSON(mintState)
	}

	if appState[slashing.ModuleName] != nil {
		var slashingState slashing.GenesisState
		cdc.MustUnmarshalJSON(appState[slashing.ModuleName], &slashingState)
		appState[slashing.ModuleName] = cdc.MustMarshalJSON(slashingState)
	}

	if appState[staking.ModuleName] != nil {
		var stakingState staking.GenesisState
		cdc.MustUnmarshalJSON(appState[staking.ModuleName], &stakingState)
		appState[staking.ModuleName] = cdc.MustMarshalJSON(stakingState)
	}
	return appState
}
