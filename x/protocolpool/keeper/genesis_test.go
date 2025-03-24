package keeper_test

import (
	"go.uber.org/mock/gomock"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func (suite *KeeperTestSuite) TestInitExportGenesis() {
	suite.bankKeeper.EXPECT().SendCoinsFromModuleToModule(gomock.Any(), types.ProtocolPoolDistrAccount, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	gs := types.NewGenesisState(
		[]types.ContinuousFund{
			{
				Recipient:  "cosmos1qy3529yj3v4xw2z3vz3vz3vz3vz3vz3v3k0vyf",
				Percentage: math.LegacyMustNewDecFromStr("0.1"),
				Expiry:     nil,
			},
		},
	)

	err := suite.poolKeeper.InitGenesis(suite.ctx, gs)
	suite.Require().NoError(err)

	// Export
	exportedGenState, err := suite.poolKeeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Equal(gs.ContinuousFunds, exportedGenState.ContinuousFunds)
}
