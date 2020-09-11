package types_test

import (
	"reflect"

	"github.com/golang/protobuf/proto"

	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *TypesTestSuite) TestMarshalConsensusStateWithHeight() {
	var (
		cswh *types.ConsensusStateWithHeight
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"solo machine client", func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "")
				cs := types.NewConsensusStateWithHeight(types.NewHeight(0, soloMachine.Sequence), soloMachine.ConsensusState())
				cswh = &cs
			},
		},
		{
			"tendermint client", func() {
				clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				clientState := suite.chainA.GetClientState(clientA)
				consensusState, ok := suite.chainA.GetConsensusState(clientA, clientState.GetLatestHeight())
				suite.Require().True(ok)

				cs := types.NewConsensusStateWithHeight(clientState.GetLatestHeight().(types.Height), consensusState)
				cswh = &cs
			},
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			suite.SetupTest()

			tc.malleate()

			cdc := suite.chainA.App.AppCodec()

			// marshal message
			bz, err := cdc.MarshalJSON(cswh)
			suite.Require().NoError(err)

			// unmarshal message
			newCswh := &types.ConsensusStateWithHeight{}
			err = cdc.UnmarshalJSON(bz, newCswh)
			suite.Require().NoError(err)

			suite.Require().True(reflect.DeepEqual(cswh, newCswh)) // fails
			suite.Require().True(proto.Equal(cswh, newCswh))       // also fails
		})
	}
}
