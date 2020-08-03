package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc-transfer/types"
)

func (suite *KeeperTestSuite) TestGenesis() {
	var trace string
	for i := 0; i < 5; i++ {
		prefix := fmt.Sprintf("transfer/channelToChain%d", i)
		if i == 0 {
			trace = prefix
		} else {
			trace = prefix + trace
		}

		denomTrace := types.DenomTrace{
			BaseDenom: "uatom",
			Trace:     trace,
		}
		suite.chainA.App.TransferKeeper.SetDenomTrace(suite.chainA.GetContext(), denomTrace)
	}

	genesis := suite.chainA.App.TransferKeeper.ExportGenesis(suite.chainA.GetContext())

	suite.Require().NotPanics(func() {
		suite.chainA.App.TransferKeeper.InitGenesis(suite.chainA.GetContext(), genesis)
	})
}
