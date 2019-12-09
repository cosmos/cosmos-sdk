package keeper_test

import (
	"strings"

	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"

	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	custom = "custom"
)

func (suite *KeeperTestSuite) TestQueryEvidence_Existing() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100

	evidence := suite.populateEvidence(ctx, numEvidence)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryEvidence}, "/"),
		Data: types.TestingCdc.MustMarshalJSON(types.NewQueryEvidenceParams(evidence[0].Hash().String())),
	}

	bz, err := suite.querier(ctx, []string{types.QueryEvidence}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var e exported.Evidence
	suite.Nil(types.TestingCdc.UnmarshalJSON(bz, &e))
	suite.Equal(evidence[0], e)
}

func (suite *KeeperTestSuite) TestQueryEvidence_NonExisting() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100

	suite.populateEvidence(ctx, numEvidence)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryEvidence}, "/"),
		Data: types.TestingCdc.MustMarshalJSON(types.NewQueryEvidenceParams("0000000000000000000000000000000000000000000000000000000000000000")),
	}

	bz, err := suite.querier(ctx, []string{types.QueryEvidence}, query)
	suite.NotNil(err)
	suite.Nil(bz)
}

func (suite *KeeperTestSuite) TestQueryAllEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100

	suite.populateEvidence(ctx, numEvidence)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryAllEvidence}, "/"),
		Data: types.TestingCdc.MustMarshalJSON(types.NewQueryAllEvidenceParams(1, numEvidence)),
	}

	bz, err := suite.querier(ctx, []string{types.QueryAllEvidence}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var e []exported.Evidence
	suite.Nil(types.TestingCdc.UnmarshalJSON(bz, &e))
	suite.Len(e, numEvidence)
}

func (suite *KeeperTestSuite) TestQueryAllEvidence_InvalidPagination() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100

	suite.populateEvidence(ctx, numEvidence)
	query := abci.RequestQuery{
		Path: strings.Join([]string{custom, types.QuerierRoute, types.QueryAllEvidence}, "/"),
		Data: types.TestingCdc.MustMarshalJSON(types.NewQueryAllEvidenceParams(0, numEvidence)),
	}

	bz, err := suite.querier(ctx, []string{types.QueryAllEvidence}, query)
	suite.Nil(err)
	suite.NotNil(bz)

	var e []exported.Evidence
	suite.Nil(types.TestingCdc.UnmarshalJSON(bz, &e))
	suite.Len(e, 0)
}

func (suite *KeeperTestSuite) TestQueryParams() {
	ctx := suite.ctx.WithIsCheckTx(false)

	bz, err := suite.querier(ctx, []string{types.QueryParameters}, abci.RequestQuery{})
	suite.Nil(err)
	suite.NotNil(bz)
	suite.Equal("{\n  \"max_evidence_age\": \"120000000000\"\n}", string(bz))
}
