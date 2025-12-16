package keeper_test

import (
	"time"

	"cosmossdk.io/math"

	"github.com/cosmos/cosmos-sdk/x/protocolpool/types"
)

func (suite *KeeperTestSuite) TestInitExportGenesis() {
	suite.bankKeeper.EXPECT().BlockedAddr(recipientAddr).Return(false).Times(1)

	gs := types.NewGenesisState(
		[]types.ContinuousFund{
			{
				Recipient:  recipientAddr.String(),
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

func (suite *KeeperTestSuite) TestInitExportGenesis_BlockedAddress() {
	suite.bankKeeper.EXPECT().BlockedAddr(recipientAddr).Return(true).Times(1)

	gs := types.NewGenesisState(
		[]types.ContinuousFund{
			{
				Recipient:  recipientAddr.String(),
				Percentage: math.LegacyMustNewDecFromStr("0.1"),
				Expiry:     nil,
			},
		},
	)

	err := suite.poolKeeper.InitGenesis(suite.ctx, gs)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestInitGenesis_InvalidRecipient() {
	gs := types.NewGenesisState([]types.ContinuousFund{
		{
			Recipient:  "invalid_address", // This should fail the address decoding.
			Percentage: math.LegacyMustNewDecFromStr("0.1"),
			Expiry:     nil,
		},
	})

	err := suite.poolKeeper.InitGenesis(suite.ctx, gs)
	suite.Require().Error(err)
}

func (suite *KeeperTestSuite) TestInitGenesis_SkipsExpiredFunds() {
	suite.bankKeeper.EXPECT().BlockedAddr(recipientAddr).Return(false).Times(1)
	suite.bankKeeper.EXPECT().BlockedAddr(recipientAddr2).Return(false).Times(1)

	// Set up block time for the test
	currentTime := suite.ctx.BlockTime()
	expiredTime := currentTime.Add(-time.Hour)
	futureTime := currentTime.Add(time.Hour)

	gs := types.NewGenesisState([]types.ContinuousFund{
		{
			Recipient:  recipientAddr.String(),
			Percentage: math.LegacyMustNewDecFromStr("0.1"),
			Expiry:     &expiredTime, // This fund should be ignored.
		},
		{
			Recipient:  recipientAddr2.String(),
			Percentage: math.LegacyMustNewDecFromStr("0.2"),
			Expiry:     &futureTime, // This fund should be accepted.
		},
	})

	err := suite.poolKeeper.InitGenesis(suite.ctx, gs)
	suite.Require().NoError(err)

	// Export and verify only the valid continuous fund is present.
	exportedGenState, err := suite.poolKeeper.ExportGenesis(suite.ctx)
	suite.Require().NoError(err)
	suite.Require().Len(exportedGenState.ContinuousFunds, 1)
	suite.Require().Equal(recipientAddr2.String(), exportedGenState.ContinuousFunds[0].Recipient)
}
