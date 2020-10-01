package types_test

import (
	"github.com/cosmos/cosmos-sdk/x/ibc/light-clients/06-solomachine/types"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *SoloMachineTestSuite) TestClientStateSignBytes() {
	cdc := suite.chainA.App.AppCodec()

	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {
		// success
		path := solomachine.GetClientStatePath(counterpartyClientIdentifier)
		bz, err := types.ClientStateSignBytes(cdc, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, solomachine.ClientState())
		suite.Require().NoError(err)
		suite.Require().NotNil(bz)

		// nil client state
		bz, err = types.ClientStateSignBytes(cdc, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, nil)
		suite.Require().Error(err)
		suite.Require().Nil(bz)
	}
}

func (suite *SoloMachineTestSuite) TestConsensusStateSignBytes() {
	cdc := suite.chainA.App.AppCodec()

	for _, solomachine := range []*ibctesting.Solomachine{suite.solomachine, suite.solomachineMulti} {
		// success
		path := solomachine.GetConsensusStatePath(counterpartyClientIdentifier, consensusHeight)
		bz, err := types.ConsensusStateSignBytes(cdc, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, solomachine.ConsensusState())
		suite.Require().NoError(err)
		suite.Require().NotNil(bz)

		// nil consensus state
		bz, err = types.ConsensusStateSignBytes(cdc, solomachine.Sequence, solomachine.Time, solomachine.Diversifier, path, nil)
		suite.Require().Error(err)
		suite.Require().Nil(bz)
	}
}
