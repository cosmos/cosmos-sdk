package keeper_test

import (
	gocontext "context"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
	tmbytes "github.com/tendermint/tendermint/libs/bytes"
)

func (suite *KeeperTestSuite) TestQueryEvidence() {
	// ctx := suite.ctx.WithIsCheckTx(false)
	app, ctx := suite.app, suite.ctx

	queryHelper := baseapp.NewQueryServerTestHelper(ctx)
	types.RegisterQueryServer(queryHelper, app.EvidenceKeeper)
	queryClient := types.NewQueryClient(queryHelper)

	_, err := queryClient.Evidence(gocontext.Background(), &types.QueryEvidenceRequest{})
	suite.Require().Error(err)

	_, err = queryClient.Evidence(gocontext.Background(), &types.QueryEvidenceRequest{EvidenceHash: tmbytes.HexBytes{}})
	suite.Require().Error(err)

	numEvidence := 100
	evidences := suite.populateEvidence(ctx, numEvidence)

	req := types.NewQueryEvidenceRequest(evidences[0].Hash())
	res, err := queryClient.Evidence(gocontext.Background(), req)
	suite.Require().NoError(err)
	suite.Require().NotNil(res)

	// TODO check returned evidence
}

// TODO
func (suite *KeeperTestSuite) TestQueryAllEvidences() {
}
