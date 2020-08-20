package keeper_test

import (
	"fmt"
	"time"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/ibc/02-client/types"
	ibctmtypes "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *KeeperTestSuite) TestQueryClientStates() {
	var (
		req             *types.QueryClientStatesRequest
		expClientStates = []*types.IdentifiedClientState(nil)
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"empty pagination",
			func() {
				req = &types.QueryClientStatesRequest{}
			},
			true,
		},
		{
			"success, no results",
			func() {
				req = &types.QueryClientStatesRequest{
					Pagination: &query.PageRequest{
						Limit:      3,
						CountTotal: true,
					},
				}
			},
			true,
		},
		{
			"success",
			func() {
				clientState := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs())
				clientState2 := ibctmtypes.NewClientState(testChainID, ibctmtypes.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, 0, commitmenttypes.GetSDKSpecs())
				suite.keeper.SetClientState(suite.ctx, testClientID, clientState)
				suite.keeper.SetClientState(suite.ctx, testClientID2, clientState2)

				idcs := types.NewIdentifiedClientState(testClientID, clientState)
				idcs2 := types.NewIdentifiedClientState(testClientID2, clientState2)

				// order is swapped because the res is sorted by client id
				expClientStates = []*types.IdentifiedClientState{&idcs2, &idcs}
				req = &types.QueryClientStatesRequest{
					Pagination: &query.PageRequest{
						Limit:      3,
						CountTotal: true,
					},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.ClientStates(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(len(expClientStates), len(res.ClientStates))
				for i := range expClientStates {
					suite.Require().Equal(expClientStates[i].ClientId, res.ClientStates[i].ClientId)
					suite.Require().NotNil(res.ClientStates[i].ClientState)
					// can't use suite.Require().Equal because of cached value mismatch
					suite.Require().True(res.ClientStates[i].ClientState.Equal(expClientStates[i].ClientState))
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *KeeperTestSuite) TestQueryConsensusStates() {
	var (
		req                *types.QueryConsensusStatesRequest
		expConsensusStates = []*codectypes.Any(nil)
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"invalid client identifier",
			func() {
				req = &types.QueryConsensusStatesRequest{}
			},
			false,
		},
		{
			"empty pagination",
			func() {
				req = &types.QueryConsensusStatesRequest{
					ClientId: testClientID,
				}
			},
			true,
		},
		{
			"success, no results",
			func() {
				req = &types.QueryConsensusStatesRequest{
					ClientId: testClientID,
					Pagination: &query.PageRequest{
						Limit:      3,
						CountTotal: true,
					},
				}
			},
			true,
		},
		{
			"success",
			func() {
				cs := ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp, commitmenttypes.NewMerkleRoot([]byte("hash1")), suite.consensusState.GetHeight(), nil,
				)
				cs2 := ibctmtypes.NewConsensusState(
					suite.consensusState.Timestamp.Add(time.Second), commitmenttypes.NewMerkleRoot([]byte("hash2")), suite.consensusState.GetHeight(), nil,
				)

				suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight, cs)
				suite.keeper.SetClientConsensusState(suite.ctx, testClientID, testClientHeight+1, cs2)

				any, err := types.PackConsensusState(cs)
				suite.Require().NoError(err)
				any2, err := types.PackConsensusState(cs2)
				suite.Require().NoError(err)

				// order is swapped because the res is sorted by client id
				expConsensusStates = []*codectypes.Any{any, any2}
				req = &types.QueryConsensusStatesRequest{
					ClientId: testClientID,
					Pagination: &query.PageRequest{
						Limit:      3,
						CountTotal: true,
					},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest() // reset

			tc.malleate()
			ctx := sdk.WrapSDKContext(suite.ctx)

			res, err := suite.queryClient.ConsensusStates(ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
				suite.Require().Equal(len(expConsensusStates), len(res.ConsensusStates))
				for i := range expConsensusStates {
					suite.Require().NotNil(res.ConsensusStates[i])
					// can't use suite.Require().Equal because of cached value mismatch
					suite.Require().True(res.ConsensusStates[i].Equal(expConsensusStates[i]))
				}
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
