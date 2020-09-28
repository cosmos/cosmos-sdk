package types_test

import (
	clienttypes "github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite SoloMachineTestSuite) TestCanUnmarshalDataByType() {
	var (
		data []byte
	)

	// test singlesig and multisig public keys
	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {

		cdc := suite.chainA.App.AppCodec()
		cases := []struct {
			name     string
			dataType types.DataType
			malleate func()
			expPass  bool
		}{
			{
				"unspecified", types.UNSPECIFIED, func() {
					data = solomachine.GetClientStateDataBytes(counterpartyClientIdentifier)
				}, false,
			},
			{
				"client", types.CLIENT, func() {
					data = solomachine.GetClientStateDataBytes(counterpartyClientIdentifier)
				}, true,
			},
			{
				"consensus", types.CONSENSUS, func() {
					data = solomachine.GetConsensusStateDataBytes(counterpartyClientIdentifier, clienttypes.NewHeight(0, 5))
				}, true,
			},
		}

		for _, tc := range cases {
			tc := tc

			suite.Run(tc.name, func() {
				res := types.CanUnmarshalDataByType(cdc, tc.dataType, data)

				suite.Require().Equal(tc.expPass, res)
			})
		}
	}

}
