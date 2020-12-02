package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *TypesTestSuite) TestMarshalConsensusStateWithHeight() {
	var (
		cswh types.ConsensusStateWithHeight
	)

	testCases := []struct {
		name     string
		malleate func()
	}{
		{
			"solo machine client", func() {
				soloMachine := ibctesting.NewSolomachine(suite.T(), suite.chainA.Codec, "solomachine", "", 1)
				cswh = types.NewConsensusStateWithHeight(types.NewHeight(0, soloMachine.Sequence), soloMachine.ConsensusState())
			},
		},
		{
			"tendermint client", func() {
				clientA, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
				clientState := suite.chainA.GetClientState(clientA)
				consensusState, ok := suite.chainA.GetConsensusState(clientA, clientState.GetLatestHeight())
				suite.Require().True(ok)

				cswh = types.NewConsensusStateWithHeight(clientState.GetLatestHeight().(types.Height), consensusState)
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
			bz, err := cdc.MarshalJSON(&cswh)
			suite.Require().NoError(err)

			// unmarshal message
			newCswh := &types.ConsensusStateWithHeight{}
			err = cdc.UnmarshalJSON(bz, newCswh)
			suite.Require().NoError(err)
		})
	}
}

func TestValidateClientType(t *testing.T) {
	testCases := []struct {
		name       string
		clientType string
		expPass    bool
	}{
		{"valid", "tendermint", true},
		{"valid solomachine", "solomachine-v1", true},
		{"too large", "tenderminttenderminttenderminttenderminttendermintt", false},
		{"too short", "t", false},
		{"blank id", "               ", false},
		{"empty id", "", false},
		{"ends with dash", "tendermint-", false},
	}

	for _, tc := range testCases {

		err := types.ValidateClientType(tc.clientType)

		if tc.expPass {
			require.NoError(t, err, tc.name)
		} else {
			require.Error(t, err, tc.name)
		}
	}
}
