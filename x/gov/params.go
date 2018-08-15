package gov

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"strconv"
	"strings"
)

const Prefix = "gov/"

const (
	ParamStoreKeyDepositProcedureDeposit          = "depositprocedure/deposit"
	ParamStoreKeyDepositProcedureMaxDepositPeriod = "depositprocedure/maxDepositPeriod"
	ParamStoreKeyVotingProcedureVotingPeriod      = "votingprocedure/votingPeriod"
	ParamStoreKeyTallyingProcedureThreshold       = "tallyingprocedure/threshold"
	ParamStoreKeyTallyingProcedureVeto            = "tallyingprocedure/veto"
	ParamStoreKeyTallyingProcedurePenalty         = "tallyingprocedure/penalty"
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
	keeper.ps.GovSetter().Set(ctx, ParamStoreKeyDepositProcedureDeposit, &data)
}

func (keeper Keeper) setDepositProcedureMaxDepositPeriod(ctx sdk.Context, MaxDepositPeriod int64) {
	maxDepositPeriod := strconv.FormatInt(MaxDepositPeriod, 10)
	keeper.ps.GovSetter().Set(ctx, ParamStoreKeyDepositProcedureMaxDepositPeriod, &maxDepositPeriod)
}

func (keeper Keeper) setVotingProcedureVotingPeriod(ctx sdk.Context, VotingPeriod int64) {
	votingPeriod := strconv.FormatInt(VotingPeriod, 10)
	keeper.ps.GovSetter().Set(ctx, ParamStoreKeyVotingProcedureVotingPeriod, &votingPeriod)
}

func (keeper Keeper) setTallyingProcedure(ctx sdk.Context, key string, rat sdk.Rat) {
	str := rat.String()
	keeper.ps.GovSetter().Set(ctx, key, &str)
}

func (keeper Keeper) getDepositProcedureDeposit(ctx sdk.Context) (Deposit sdk.Coins) {
	var data string
	keeper.ps.GovSetter().Get(ctx, ParamStoreKeyDepositProcedureDeposit, &data)
	Deposit, _ = sdk.ParseCoins(data)
	return
}

func (keeper Keeper) getDepositProcedureMaxDepositPeriod(ctx sdk.Context) (MaxDepositPeriod int64) {
	var maxDepositPeriod string
	if keeper.ps.GovSetter().Get(ctx, ParamStoreKeyDepositProcedureMaxDepositPeriod, &maxDepositPeriod) == nil {
		MaxDepositPeriod, _ = strconv.ParseInt(maxDepositPeriod, 10, 64)
	}
	return
}

func (keeper Keeper) getVotingProcedureVotingPeriod(ctx sdk.Context) (VotingPeriod int64) {
	var votingPeriod string
	if keeper.ps.GovSetter().Get(ctx, ParamStoreKeyVotingProcedureVotingPeriod, &votingPeriod) == nil {
		VotingPeriod, _ = strconv.ParseInt(votingPeriod, 10, 64)
	}
	return
}

func (keeper Keeper) getTallyingProcedure(ctx sdk.Context, key string) sdk.Rat {
	var data string
	keeper.ps.GovSetter().Get(ctx, key, &data)
	str := strings.Split(data, "/")
	x, _ := strconv.ParseInt(str[0], 10, 64)
	y, _ := strconv.ParseInt(str[1], 10, 64)
	return sdk.NewRat(x, y)

}
