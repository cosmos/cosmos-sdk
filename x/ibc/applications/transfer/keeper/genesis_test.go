package keeper_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/ibc/applications/transfer/types"
)

func (suite *KeeperTestSuite) TestGenesis() {
	var (
		path   string
		traces types.Traces
	)

	for i := 0; i < 5; i++ {
		prefix := fmt.Sprintf("transfer/channelToChain%d", i)
		if i == 0 {
			path = prefix
		} else {
			path = prefix + "/" + path
		}

		denomTrace := types.DenomTrace{
			BaseDenom: "uatom",
			Path:      path,
		}
		traces = append(types.Traces{denomTrace}, traces...)
		suite.chainA.App.TransferKeeper.SetDenomTrace(suite.chainA.GetContext(), denomTrace)
	}

	genesis := suite.chainA.App.TransferKeeper.ExportGenesis(suite.chainA.GetContext())

	suite.Require().Equal(types.PortID, genesis.PortId)
	suite.Require().Equal(traces.Sort(), genesis.DenomTraces)

	suite.Require().NotPanics(func() {
		suite.chainA.App.TransferKeeper.InitGenesis(suite.chainA.GetContext(), *genesis)
	})
}
