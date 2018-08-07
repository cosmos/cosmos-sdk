package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
	"strings"
)

// nolint
const (
	ParamStoreKeyDepositProcedureDeposit          = "gov/depositprocedure/deposit"
	ParamStoreKeyDepositProcedureMaxDepositPeriod = "gov/depositprocedure/maxDepositPeriod"

	ParamStoreKeyVotingProcedureVotingPeriod = "gov/votingprocedure/votingPeriod"

	ParamStoreKeyTallyingProcedureThreshold = "gov/tallyingprocedure/threshold"
	ParamStoreKeyTallyingProcedureVeto      = "gov/tallyingprocedure/veto"
	ParamStoreKeyTallyingProcedurePenalty   = "gov/tallyingprocedure/penalty"
)

// =====================================================
// Procedures

// Returns the current Deposit Procedure from the global param store
func (keeper Keeper) GetDepositProcedure(ctx sdk.Context) DepositProcedure {
	return DepositProcedure{
		MinDeposit:       keeper.getDepositProcedureDeposit(ctx),
		MaxDepositPeriod: keeper.getDepositProcedureMaxDepositPeriod(ctx),
	}
}

// Returns the current Voting Procedure from the global param store
func (keeper Keeper) GetVotingProcedure(ctx sdk.Context) VotingProcedure {
	return VotingProcedure{
		VotingPeriod: keeper.getVotingProcedureVotingPeriod(ctx),
	}
}

// Returns the current Tallying Procedure from the global param store
func (keeper Keeper) GetTallyingProcedure(ctx sdk.Context) TallyingProcedure {
	return TallyingProcedure{
		Threshold:         keeper.getTallyingProcedure(ctx, ParamStoreKeyTallyingProcedureThreshold),
		Veto:              keeper.getTallyingProcedure(ctx, ParamStoreKeyTallyingProcedureVeto),
		GovernancePenalty: keeper.getTallyingProcedure(ctx, ParamStoreKeyTallyingProcedurePenalty),
	}
}

func (keeper Keeper) setDepositProcedureDeposit(ctx sdk.Context, Deposit sdk.Coins) {
	data := Deposit.String()
	keeper.ps.Set(ctx, ParamStoreKeyDepositProcedureDeposit, &data)
}

func (keeper Keeper) setDepositProcedureMaxDepositPeriod(ctx sdk.Context, MaxDepositPeriod int64) {
	keeper.ps.Set(ctx, ParamStoreKeyDepositProcedureMaxDepositPeriod, &MaxDepositPeriod)
}

func (keeper Keeper) setVotingProcedureVotingPeriod(ctx sdk.Context, VotingPeriod int64) {
	keeper.ps.Set(ctx, ParamStoreKeyVotingProcedureVotingPeriod, &VotingPeriod)
}

func (keeper Keeper) setTallyingProcedure(ctx sdk.Context, key string, rat sdk.Rat) {
	str := rat.String()
	keeper.ps.Set(ctx, key, &str)
}

func (keeper Keeper) getDepositProcedureDeposit(ctx sdk.Context) (Deposit sdk.Coins) {
	var data string
	keeper.ps.Get(ctx, ParamStoreKeyDepositProcedureDeposit, &data)
	Deposit, _ = sdk.ParseCoins(data)
	return
}

func (keeper Keeper) getDepositProcedureMaxDepositPeriod(ctx sdk.Context) (MaxDepositPeriod int64) {
	keeper.ps.Get(ctx, ParamStoreKeyDepositProcedureMaxDepositPeriod, &MaxDepositPeriod)
	return
}

func (keeper Keeper) getVotingProcedureVotingPeriod(ctx sdk.Context) (VotingPeriod int64) {
	keeper.ps.Get(ctx, ParamStoreKeyVotingProcedureVotingPeriod, &VotingPeriod)
	return
}

func (keeper Keeper) getTallyingProcedure(ctx sdk.Context, key string) sdk.Rat {
	var data string
	keeper.ps.Get(ctx, key, &data)
	str := strings.Split(data, "/")
	x, _ := strconv.ParseInt(str[0], 10, 64)
	y, _ := strconv.ParseInt(str[1], 10, 64)
	return sdk.NewRat(x, y)

}
