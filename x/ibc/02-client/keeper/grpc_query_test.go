package keeper_test

// func (suite *KeeperTestSuite) TestQueryClientStates() {
// 	var (
// 		req             *types.QueryClientStatesRequest
// 		expClientStates = []*types.IdentifiedClientState{}
// 	)

// 	testCases := []struct {
// 		msg      string
// 		malleate func()
// 		expPass  bool
// 	}{
// 		{
// 			"empty request",
// 			func() {
// 				req = nil
// 			},
// 			false,
// 		},
// 		{
// 			"empty pagination",
// 			func() {
// 				req = &types.QueryClientStatesRequest{}
// 			},
// 			true,
// 		},
// 		{
// 			"success",
// 			func() {
// 				clientA, clientB, connA0, connB0 := suite.coordinator.SetupClientClientStates(suite.chainA, suite.chainB, exported.Tendermint)
// 				connA1, connB1, err := suite.coordinator.ConnOpenInit(suite.chainA, suite.chainB, clientA, clientB)
// 				suite.Require().NoError(err)

// 				clientA1, clientB1, connA2, connB2 := suite.coordinator.SetupClientClientStates(suite.chainA, suite.chainB, exported.Tendermint)

// 				conn1 := types.NewClientStateEnd(types.OPEN, clientA, counterparty1, types.GetCompatibleEncodedVersions())
// 				conn2 := types.NewClientStateEnd(types.INIT, clientA, counterparty2, types.GetCompatibleEncodedVersions())
// 				conn3 := types.NewClientStateEnd(types.OPEN, clientA1, counterparty3, types.GetCompatibleEncodedVersions())

// 				iconn1 := types.NewIdentifiedClientState(connA0.ID, conn1)
// 				iconn2 := types.NewIdentifiedClientState(connA1.ID, conn2)
// 				iconn3 := types.NewIdentifiedClientState(connA2.ID, conn3)

// 				expClientStates = []*types.IdentifiedClientState{&iconn1, &iconn2, &iconn3}

// 				req = &types.QueryClientStatesRequest{
// 					Pagination: &query.PageRequest{
// 						Limit:      3,
// 						CountTotal: true,
// 					},
// 				}
// 			},
// 			true,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
// 			suite.SetupTest() // reset

// 			tc.malleate()
// 			ctx := sdk.WrapSDKContext(suite.chainA.GetContext())

// 			res, err := suite.chainA.QueryServer.ClientStates(ctx, req)

// 			if tc.expPass {
// 				suite.Require().NoError(err)
// 				suite.Require().NotNil(res)
// 				suite.Require().Equal(expClientStates, res.ClientStates)
// 			} else {
// 				suite.Require().Error(err)
// 			}
// 		})
// 	}
// }
