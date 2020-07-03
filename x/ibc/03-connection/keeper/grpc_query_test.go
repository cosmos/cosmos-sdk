package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/03-connection/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *KeeperTestSuite) TestQueryConnection() {
	var (
		req           *types.QueryConnectionRequest
		expConnection types.ConnectionEnd
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{"invalid connectionID",
			func() {
				req = &types.QueryConnectionRequest{}
			},
			false,
		},
		{"connection not found",
			func() {
				req = &types.QueryConnectionRequest{
					ConnectionID: testConnectionIDB,
				}
			},
			false,
		},
		{
			"sucess",
			func() {
				counterparty := types.NewCounterparty(testClientIDA, testConnectionIDA, commitmenttypes.NewMerklePrefix(suite.chainA.App.IBCKeeper.ConnectionKeeper.GetCommitmentPrefix().Bytes()))
				expConnection = types.NewConnectionEnd(types.INIT, testConnectionIDB, testClientIDB, counterparty, types.GetCompatibleVersions())
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetConnection(suite.chainA.GetContext(), testConnectionIDB, expConnection)

				req = &types.QueryConnectionRequest{
					ConnectionID: testConnectionIDB,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.Connection(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(&expConnection, res.Connection)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryClientConnections() {
	var (
		req      *types.QueryClientConnectionsRequest
		expPaths []string
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty request",
			func() {
				req = nil
			},
			false,
		},
		{"invalid connectionID",
			func() {
				req = &types.QueryClientConnectionsRequest{}
			},
			false,
		},
		{"connection not found",
			func() {
				req = &types.QueryClientConnectionsRequest{
					ClientID: testClientIDA,
				}
			},
			false,
		},
		{
			"sucess",
			func() {
				expPaths = []string{testConnectionIDA, testConnectionIDB}
				suite.chainA.App.IBCKeeper.ConnectionKeeper.SetClientConnectionPaths(suite.chainA.GetContext(), testClientIDA, expPaths)

				req = &types.QueryClientConnectionsRequest{
					ClientID: testClientIDA,
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

			res, err := suite.chainA.QueryServer.ClientConnections(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(expPaths, res.ConnectionPaths)
			} else {
				suite.Require().Error(err)
			}
		})
	}
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
