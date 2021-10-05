package v045

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdktx "github.com/cosmos/cosmos-sdk/types/tx"
	v043gov "github.com/cosmos/cosmos-sdk/x/gov/migrations/v043"
	v045gov "github.com/cosmos/cosmos-sdk/x/gov/types"
)

func Migrate(oldGovState v043gov.GenesisState) *v045gov.GenesisState {
	return &v045gov.GenesisState{
		Deposits:      oldGovState.Deposits,
		Votes:         oldGovState.Votes,
		Proposals:     migrateProposals(oldGovState.Proposals),
		DepositParams: oldGovState.DepositParams,
		VotingParams:  oldGovState.VotingParams,
		TallyParams:   oldGovState.TallyParams,
	}
}

func migrateProposals(oldProposals v043gov.Proposals) v045gov.Proposals {
	newProposals := make(v045gov.Proposals, len(oldProposals))
	for idx, proposal := range oldProposals {
		msgs, err := sdktx.SetMsgs(migrateContent(proposal.Content))
		if err != nil {
			panic(fmt.Sprintf("failed to marshal proposal msgs: %v", err))
		}

		newProposals[idx] = v045gov.Proposal{
			ProposalId:       proposal.ProposalId,
			Messages:         msgs,
			Status:           proposal.Status,
			FinalTallyResult: proposal.FinalTallyResult,
			SubmitTime:       proposal.SubmitTime,
			DepositEndTime:   proposal.DepositEndTime,
			VotingStartTime:  proposal.VotingStartTime,
			VotingEndTime:    proposal.VotingEndTime,
		}
	}
	return newProposals
}

func migrateContent(content v040gov.Content) []sdk.Msg {
	switch content.ProposalType() {
	case v040gov.ProposalTypeText:
		return []sdk.Msg{v045gov.NewMsgSignal(content.GetTitle(), content.GetDescription())}
	// TODO: enter the other proposal content types
	default:
		// NOTE: If a network is using a unique content type that isn't recognisable then it will not be possible to migrate it to the new proposal type. The best thing to do in this situation, rather than silently ignore it, is to convert it to a signal proposal
		return []sdk.Msg{v045gov.NewMsgSignal(content.GetTitle(), content.GetDescription())}
	}
}
