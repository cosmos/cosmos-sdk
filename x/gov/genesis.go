package gov

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

// InitGenesis - store genesis parameters
func InitGenesis(ctx sdk.Context, ak types.AccountKeeper, bk types.BankKeeper, k *keeper.Keeper, data *v1.GenesisState) {
	k.SetProposalID(ctx, data.StartingProposalId)
	k.SetParams(ctx, *data.Params)

	// check if the deposits pool account exists
	moduleAcc := k.GetGovernanceAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.ModuleName))
	}

	var totalDeposits sdk.Coins
	for _, deposit := range data.Deposits {
		k.SetDeposit(ctx, *deposit)
		totalDeposits = totalDeposits.Add(deposit.Amount...)
	}

	for _, vote := range data.Votes {
		k.SetVote(ctx, *vote)
	}

	for _, proposal := range data.Proposals {
		switch proposal.Status {
		case v1.StatusDepositPeriod:
			k.InsertInactiveProposalQueue(ctx, proposal.Id, *proposal.DepositEndTime)
		case v1.StatusVotingPeriod:
			k.InsertActiveProposalQueue(ctx, proposal.Id, *proposal.VotingEndTime)
		}
		k.SetProposal(ctx, *proposal)
	}

	// if account has zero balance it probably means it's not set, so we set it
	balance := bk.GetAllBalances(ctx, moduleAcc.GetAddress())
	if balance.IsZero() {
		ak.SetModuleAccount(ctx, moduleAcc)
	}

	// check if total deposits equals balance, if it doesn't panic because there were export/import errors
	if !balance.IsEqual(totalDeposits) {
		panic(fmt.Sprintf("expected module account was %s but we got %s", balance.String(), totalDeposits.String()))
	}
}

// ExportGenesis - output genesis parameters
//
//nolint:staticcheck
func ExportGenesis(ctx sdk.Context, k *keeper.Keeper) *v1.GenesisState {
	startingProposalID, _ := k.GetProposalID(ctx)
	proposals := k.GetProposals(ctx)
	params := k.GetParams(ctx)

	depositParams := v1.NewDepositParams(params.MinDeposit, params.MaxDepositPeriod)
	votingParams := v1.NewVotingParams(params.VotingPeriod)
	tallyParams := v1.NewTallyParams(params.Quorum, params.Threshold, params.VetoThreshold)

	var proposalsDeposits v1.Deposits
	var proposalsVotes v1.Votes
	for _, proposal := range proposals {
		deposits := k.GetDeposits(ctx, proposal.Id)
		proposalsDeposits = append(proposalsDeposits, deposits...)

		votes := k.GetVotes(ctx, proposal.Id)
		proposalsVotes = append(proposalsVotes, votes...)
	}

	return &v1.GenesisState{
		StartingProposalId: startingProposalID,
		Deposits:           proposalsDeposits,
		Votes:              proposalsVotes,
		Proposals:          proposals,
		DepositParams:      &depositParams,
		VotingParams:       &votingParams,
		TallyParams:        &tallyParams,
		Params:             &params,
	}
}

func InitGenesisFrom(ctx sdk.Context, cdc codec.JSONCodec, ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper, importPath string) error {
	f, err := module.OpenGenesisModuleFile(importPath, types.ModuleName)
	if err != nil {
		return err
	}
	defer f.Close()

	bz, err := module.FileRead(f)
	if err != nil {
		return err
	}

	var gs v1.GenesisState
	cdc.MustUnmarshalJSON(bz, &gs)
	InitGenesis(ctx, ak, bk, &k, &gs)
	return nil
}

func ExportGenesisTo(ctx sdk.Context, cdc codec.JSONCodec, k keeper.Keeper, exportPath string) error {
	f, err := module.CreateGenesisExportFile(exportPath, types.ModuleName)
	if err != nil {
		return err
	}
	defer f.Close()

	gs := ExportGenesis(ctx, &k)
	bz := cdc.MustMarshalJSON(gs)
	return module.FileWrite(f, bz)
}
