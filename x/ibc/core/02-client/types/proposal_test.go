package types_test

import (
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/02-client/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/core/exported"
	ibctesting "github.com/cosmos/cosmos-sdk/x/ibc/testing"
)

func (suite *TypesTestSuite) TestValidateBasic() {
	subject, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	subjectClientState := suite.chainA.GetClientState(subject)
	substitute, _ := suite.coordinator.SetupClients(suite.chainA, suite.chainB, exported.Tendermint)
	initialHeight := types.NewHeight(subjectClientState.GetLatestHeight().GetRevisionNumber(), subjectClientState.GetLatestHeight().GetRevisionHeight()+1)

	testCases := []struct {
		name     string
		proposal govtypes.Content
		expPass  bool
	}{
		{
			"success",
			types.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, substitute, initialHeight),
			true,
		},
		{
			"fails validate abstract - empty title",
			types.NewClientUpdateProposal("", ibctesting.Description, subject, substitute, initialHeight),
			false,
		},
		{
			"subject and substitute use the same identifier",
			types.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, subject, initialHeight),
			false,
		},
		{
			"invalid subject clientID",
			types.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, ibctesting.InvalidID, substitute, initialHeight),
			false,
		},
		{
			"invalid substitute clientID",
			types.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, ibctesting.InvalidID, initialHeight),
			false,
		},
		{
			"initial height is zero",
			types.NewClientUpdateProposal(ibctesting.Title, ibctesting.Description, subject, substitute, types.ZeroHeight()),
			false,
		},
	}

	for _, tc := range testCases {

		err := tc.proposal.ValidateBasic()

		if tc.expPass {
			suite.Require().NoError(err, tc.name)
		} else {
			suite.Require().Error(err, tc.name)
		}
	}
}

// tests a client update proposal can be marshaled and unmarshaled
func (suite *TypesTestSuite) TestMarshalClientUpdateProposalProposal() {
	// create proposal
	proposal := types.NewClientUpdateProposal("update IBC client", "description", "subject", "substitute", types.NewHeight(1, 0))

	// create codec
	ir := codectypes.NewInterfaceRegistry()
	types.RegisterInterfaces(ir)
	govtypes.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)

	// marshal message
	bz, err := cdc.MarshalJSON(proposal)
	suite.Require().NoError(err)

	// unmarshal proposal
	newProposal := &types.ClientUpdateProposal{}
	err = cdc.UnmarshalJSON(bz, newProposal)
	suite.Require().NoError(err)
}
