package gov

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/magiconair/properties/assert"
	abci "github.com/tendermint/tendermint/abci/types"
	"testing"
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
			{Key: ParamStoreKeyDepositProcedureMaxDepositPeriod, Value: "20", Op: Update},
			{Key: ParamStoreKeyTallyingProcedurePenalty, Value: "1/50", Op: Update},
			{Key: ParamStoreKeyTallyingProcedureVeto, Value: "1/4", Op: Update},
			{Key: ParamStoreKeyTallyingProcedureThreshold, Value: "2/8", Op: Update},
			{Key: "upgrade", Value: "false", Op: Update},
			{Key: "version", Value: "2", Op: Update},
		},
	}

	keeper.ps.Set(ctx,"upgrade",true)
	keeper.ps.Set(ctx,"version",int16(1))

	pp.Execute(ctx, keeper)
	assert.Equal(t, keeper.GetDepositProcedure(ctx).MinDeposit,
		sdk.Coins{sdk.NewCoin("iris", 200)})

	assert.Equal(t, keeper.GetDepositProcedure(ctx).MaxDepositPeriod,int64(20))

	upgrade,_ := keeper.ps.GetBool(ctx,"upgrade")
	assert.Equal(t, upgrade,false)

	version,_ := keeper.ps.GetInt16(ctx,"version")
	assert.Equal(t, version,int16(2))



	assert.Equal(t, keeper.GetTallyingProcedure(ctx),
		TallyingProcedure{
			Threshold:         sdk.NewRat(2, 8),
			Veto:              sdk.NewRat(1, 4),
			GovernancePenalty: sdk.NewRat(1, 50),
		})
}