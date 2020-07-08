package keeper_test

import (
	gocontext "context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"

	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

func (suite *KeeperTestSuite) TestQueryEvidence() {
	app, ctx := suite.app, suite.ctx

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.EvidenceKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	_, err := queryClient.Evidence(gocontext.Background(), &types.QueryEvidenceRequest{})
	suite.Require().Error(err)

	_, err = queryClient.Evidence(gocontext.Background(), &types.QueryEvidenceRequest{EvidenceHash: tmbytes.HexBytes{}})
	suite.Require().Error(err)

	numEvidence := 100
	evidence := suite.populateEvidence(ctx, numEvidence)

	req := types.NewQueryEvidenceRequest(evidence[0].Hash())
	res, err := queryClient.Evidence(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().NotNil(res.Evidence)
}

func (suite *KeeperTestSuite) TestQueryAllEvidence() {
	app, ctx := suite.app, suite.ctx

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, app.EvidenceKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	res, err := queryClient.AllEvidence(gocontext.Background(), &types.QueryAllEvidenceRequest{})
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Require().Empty(res.Evidence)

	numEvidence := 100
	_ = suite.populateEvidence(ctx, numEvidence)
	pageReq := &query.PageRequest{
		Key:        nil,
		Limit:      50,
		CountTotal: false,
	}
	req := types.NewQueryAllEvidenceRequest(pageReq)
	res, err = queryClient.AllEvidence(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)
	suite.Equal(len(res.Evidence), 50)
	suite.NotNil(res.Res.NextKey)

	pageReq = &query.PageRequest{
		Key:        res.Res.NextKey,
		Limit:      50,
		CountTotal: true,
	}
	req = types.NewQueryAllEvidenceRequest(pageReq)
	res, err = queryClient.AllEvidence(gocontext.Background(), req)
	suite.Equal(len(res.Evidence), 50)
	suite.Nil(res.Res.NextKey)
}
