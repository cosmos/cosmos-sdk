package keeper_test

import (
	"context"

	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *KeeperTestSuite) TestQueryConnection() {
	counterparty := types.NewCounterparty(testClientIDA, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
	expConn := types.NewConnectionEnd(types.INIT, testConnectionIDB, testClientIDB, counterparty, types.GetCompatibleVersions())
	suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), testConnectionIDA, expConn)

	req := &types.QueryConnectionRequest{
		ConnectionID: testConnectionIDB,
	}

	connectionRes, err := suite.grpcQueryClient.Connection(context.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(connectionRes)
	suite.Require().Equal(expConn, connectionRes.Connection)
}

// func TestAllProposal() {

// 	// check for the proposals with no proposal added should return null.
// 	pageReq := &query.PageRequest{
// 		Key:        nil,
// 		Limit:      1,
// 		CountTotal: false,
// 	}

// 	req := types.NewQueryProposalsRequest(0, nil, nil, pageReq)

// 	proposals, err := queryClient.Proposals(gocontext.Background(), req)
// 	require.NoError(t, err)
// 	require.Empty(t, proposals.Proposals)

// 	// create 2 test proposals
// 	for i := 0; i < 2; i++ {
// 		num := strconv.Itoa(i + 1)
// 		testProposal := types.NewTextProposal("Proposal"+num, "testing proposal "+num)
// 		_, err := app.GovKeeper.SubmitProposal(ctx, testProposal)
// 		require.NoError(t, err)
// 	}

// 	// Query for proposals after adding 2 proposals to the store.
// 	// give page limit as 1 and expect NextKey should not to be empty
// 	proposals, err = queryClient.Proposals(gocontext.Background(), req)
// 	require.NoError(t, err)
// 	require.NotEmpty(t, proposals.Proposals)
// 	require.NotEmpty(t, proposals.Res.NextKey)

// 	pageReq = &query.PageRequest{
// 		Key:        proposals.Res.NextKey,
// 		Limit:      1,
// 		CountTotal: false,
// 	}

// 	req = types.NewQueryProposalsRequest(0, nil, nil, pageReq)

// 	// query for the next page which is 2nd proposal at present context.
// 	// and expect NextKey should be empty
// 	proposals, err = queryClient.Proposals(gocontext.Background(), req)

// 	require.NoError(t, err)
// 	require.NotEmpty(t, proposals.Proposals)
// 	require.Empty(t, proposals.Res)

// 	pageReq = &query.PageRequest{
// 		Key:        nil,
// 		Limit:      2,
// 		CountTotal: false,
// 	}

// 	req = types.NewQueryProposalsRequest(0, nil, nil, pageReq)

// 	// Query the page with limit 2 and expect NextKey should ne nil
// 	proposals, err = queryClient.Proposals(gocontext.Background(), req)

// 	require.NoError(t, err)
// 	require.NotEmpty(t, proposals.Proposals)
// 	require.Empty(t, proposals.Res)
// }
