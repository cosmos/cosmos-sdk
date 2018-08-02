package gov

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/magiconair/properties/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
)

const (
	ParamStoreKeyDepositProcedureDepositAmount = "gov/depositprocedure/depositAmount"
	ParamStoreKeyDepositProcedureDepositDenom  = "gov/depositprocedure/depositDenom"
)

func TestParameterProposal(t *testing.T) {
	mapp, keeper, _, _, _, _ := getMockApp(t, 0)

	mapp.BeginBlock(abci.RequestBeginBlock{})
	ctx := mapp.BaseApp.NewContext(false, abci.Header{})

	fmt.Println(keeper.GetDepositProcedure(ctx))
	fmt.Println(keeper.GetTallyingProcedure(ctx))
	fmt.Println(keeper.GetVotingProcedure(ctx))

	pp := ParameterProposal{
		Datas: []Data{
			{Key: ParamStoreKeyDepositProcedureDeposit, Value: "200iris", Op: Update},
			{Key: ParamStoreKeyDepositProcedureDepositAmount, Value: "200", Op: Update},
			{Key: ParamStoreKeyDepositProcedureDepositDenom, Value: "iris", Op: Update},
			{Key: ParamStoreKeyDepositProcedureMaxDepositPeriod, Value: 100, Op: Update},
			{Key: ParamStoreKeyVotingProcedureVotingPeriod, Value: 200, Op: Update},
			{Key: ParamStoreKeyTallyingProcedurePenalty, Value: "1/50", Op: Update},
			{Key: ParamStoreKeyTallyingProcedureVeto, Value: "1/4", Op: Update},
			{Key: ParamStoreKeyTallyingProcedureThreshold, Value: "2/8", Op: Update},
		},
	}

	pp.Execute(ctx, keeper)
	assert.Equal(t, keeper.GetDepositProcedure(ctx),
		DepositProcedure{
			MinDeposit:       sdk.Coins{sdk.NewCoin("iris", 200)},
			MaxDepositPeriod: 100})

	assert.Equal(t, keeper.GetVotingProcedure(ctx),
		VotingProcedure{
			VotingPeriod: 200,
		})
	fmt.Println(keeper.GetTallyingProcedure(ctx))
	assert.Equal(t, keeper.GetTallyingProcedure(ctx),
		TallyingProcedure{
			Threshold:         sdk.NewRat(2, 8),
			Veto:              sdk.NewRat(1, 4),
			GovernancePenalty: sdk.NewRat(1, 50),
		})
}
