package keeper_test

import (
	"time"

	"cosmossdk.io/math"
	"cosmossdk.io/x/protocolpool/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (suite *KeeperTestSuite) TestInitGenesis() {
	hour := time.Hour
	gs := types.NewGenesisState(
		[]*types.ContinuousFund{
			{
				Recipient:  "cosmos1qy3529yj3v4xw2z3vz3vz3vz3vz3vz3v3k0vyf",
				Percentage: math.LegacyMustNewDecFromStr("0.1"),
				Expiry:     nil,
			},
		},
		[]*types.Budget{
			{
				RecipientAddress: "cosmos1qy3529yj3v4xw2z3vz3vz3vz3vz3vz3v3k0vyf",
				ClaimedAmount:    &sdk.Coin{},
				LastClaimedAt:    &time.Time{},
				TranchesLeft:     10,
				BudgetPerTranche: &sdk.Coin{Denom: "stake", Amount: math.NewInt(100)},
				Period:           &hour,
			},
		},
	)

	err := suite.poolKeeper.InitGenesis(suite.ctx, gs)
	suite.Require().NoError(err)

	// Export
	exportedGenState, err := suite.poolKeeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(gs.ContinuousFund, exportedGenState.ContinuousFund)
	suite.Require().Equal(gs.Budget, exportedGenState.Budget)
}
