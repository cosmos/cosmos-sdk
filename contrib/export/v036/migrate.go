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
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/staking"
)

func migrateGovernance(initialState v034gov.GenesisState) v036gov.GenesisState {
	var targetGov v036gov.GenesisState
	targetGov.StartingProposalID = initialState.StartingProposalID
	targetGov.DepositParams = initialState.DepositParams
	targetGov.TallyParams = initialState.TallyParams
	targetGov.VotingParams = initialState.VotingParams

	var deposits gov.Deposits
	for _, p := range initialState.Deposits {
		deposits = append(deposits, p.Deposit)
	}
	targetGov.Deposits = deposits

	var votes gov.Votes
	for _, p := range initialState.Votes {
		votes = append(votes, p.Vote)
	}
	targetGov.Votes = votes

	var proposals v036gov.Proposals
	for _, p := range initialState.Proposals {
		proposal := v036gov.Proposal{
			Content:          p.Content,
			ProposalID:       p.ProposalID,
			Status:           p.Status,
			FinalTallyResult: p.FinalTallyResult,
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

func Migrate(appState extypes.AppMap, cdc *codec.Codec) extypes.AppMap {
	var govState v034gov.GenesisState
	v034gov.RegisterCodec(cdc)
	cdc.MustUnmarshalJSON(appState[gov.ModuleName], &govState)
	appState[gov.ModuleName] = cdc.MustMarshalJSON(migrateGovernance(govState))

	// migration below are only sanity check:
	// used as unmarshal guarantee,
	// so if anyone change the types without a migrations we panic
	// We could move this next steps to a test OR use the outcome on CCI for re-import

	var authState auth.GenesisState
	cdc.MustUnmarshalJSON(appState[auth.ModuleName], &authState)
	appState[auth.ModuleName] = cdc.MustMarshalJSON(authState)

	var bankState bank.GenesisState
	cdc.MustUnmarshalJSON(appState[bank.ModuleName], &bankState)
	appState[bank.ModuleName] = cdc.MustMarshalJSON(bankState)

	var crisisState crisis.GenesisState
	cdc.MustUnmarshalJSON(appState[crisis.ModuleName], &crisisState)
	appState[crisis.ModuleName] = cdc.MustMarshalJSON(crisisState)

	var distributionState distribution.GenesisState
	cdc.MustUnmarshalJSON(appState[distribution.ModuleName], &distributionState)
	appState[distribution.ModuleName] = cdc.MustMarshalJSON(distributionState)

	// Cannot decode empty bytes
	//var genutilState genutil.GenesisState
	//cdc.MustUnmarshalJSON(appState[genutil.ModuleName], &genutilState)
	//appState[genutil.ModuleName] = cdc.MustMarshalJSON(genutilState)

	var mintState mint.GenesisState
	cdc.MustUnmarshalJSON(appState[mint.ModuleName], &mintState)
	appState[mint.ModuleName] = cdc.MustMarshalJSON(mintState)

	var slashingState slashing.GenesisState
	cdc.MustUnmarshalJSON(appState[slashing.ModuleName], &slashingState)
	appState[slashing.ModuleName] = cdc.MustMarshalJSON(slashingState)

	var stakingState staking.GenesisState
	cdc.MustUnmarshalJSON(appState[staking.ModuleName], &stakingState)
	appState[staking.ModuleName] = cdc.MustMarshalJSON(stakingState)

	return appState
}
