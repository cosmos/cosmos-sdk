package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
)

type GenesisState struct {
	StartingProposalID uint64                    `json:"starting_proposal_id"`
	Deposits           []gov.DepositWithMetadata `json:"deposits"`
	Votes              []gov.VoteWithMetadata    `json:"votes"`
	Proposals          []Proposal                `json:"proposals"`
	DepositParams      gov.DepositParams         `json:"deposit_params"`
	VotingParams       gov.VotingParams          `json:"voting_params"`
	TallyParams        gov.TallyParams           `json:"tally_params"`
}

func InitGenesis(ctx sdk.Context, k gov.Keeper, data GenesisState) {
	gov.InitGenesis(ctx, k, MigrateGenesis(data))
}

func MigrateGenesis(data GenesisState) (res gov.GenesisState) {
	res = gov.GenesisState{
		StartingProposalID: data.StartingProposalID,
		Deposits:           data.Deposits,
		Votes:              data.Votes,
		Proposals:          make([]gov.Proposal, len(data.Proposals)),
		DepositParams:      data.DepositParams,
		VotingParams:       data.VotingParams,
		TallyParams:        data.TallyParams,
	}

	var err error
	for i, p := range data.Proposals {
		res.Proposals[i], err = p.Migrate()
		if err != nil {
			panic(err)
		}
	}

	return
}
