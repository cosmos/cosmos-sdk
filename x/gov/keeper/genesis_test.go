package keeper_test

import (
	"go.uber.org/mock/gomock"

	sdkmath "cosmossdk.io/math"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/keeper"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (suite *KeeperTestSuite) TestImportExportQueues_ErrorUnconsistentState() {
	suite.reset()
	suite.acctKeeper.EXPECT().SetModuleAccount(suite.ctx, gomock.Any()).AnyTimes()
	suite.Require().Panics(func() {
		keeper.InitGenesis(suite.ctx, suite.acctKeeper, suite.bankKeeper, suite.govKeeper, &v1.GenesisState{
			Deposits: v1.Deposits{
				{
					ProposalId: 1234,
					Depositor:  "me",
					Amount: sdk.Coins{
						sdk.NewCoin(
							"stake",
							sdkmath.NewInt(1234),
						),
					},
				},
			},
		})
	})
	keeper.InitGenesis(suite.ctx, suite.acctKeeper, suite.bankKeeper, suite.govKeeper, v1.DefaultGenesisState())
	genState, err := keeper.ExportGenesis(suite.ctx, suite.govKeeper)
	suite.Require().NoError(err)
	suite.Require().Equal(genState, v1.DefaultGenesisState())
}
