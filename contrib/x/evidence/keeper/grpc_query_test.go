package keeper_test

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/contrib/x/evidence/exported"
	types2 "github.com/cosmos/cosmos-sdk/contrib/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

func (suite *KeeperTestSuite) TestQueryEvidence() {
	var (
		req      *types2.QueryEvidenceRequest
		evidence []exported.Evidence
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		expErrMsg string
		posttests func(res *types2.QueryEvidenceResponse)
	}{
		{
			"invalid request with empty evidence hash",
			func() {
				req = &types2.QueryEvidenceRequest{Hash: ""}
			},
			false,
			"invalid request; hash is empty",
			func(res *types2.QueryEvidenceResponse) {},
		},
		{
			"evidence not found",
			func() {
				numEvidence := 1
				evidence = suite.populateEvidence(suite.ctx, numEvidence)
				evidenceHash := strings.ToUpper(hex.EncodeToString(evidence[0].Hash()))
				reqHash := strings.Repeat("a", len(evidenceHash))
				req = types2.NewQueryEvidenceRequest(reqHash)
			},
			false,
			"not found",
			func(res *types2.QueryEvidenceResponse) {
			},
		},
		{
			"non-existent evidence",
			func() {
				reqHash := "DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660"
				req = types2.NewQueryEvidenceRequest(reqHash)
			},
			false,
			"evidence DF0C23E8634E480F84B9D5674A7CDC9816466DEC28A3358F73260F68D28D7660 not found",
			func(res *types2.QueryEvidenceResponse) {
			},
		},
		{
			"success",
			func() {
				numEvidence := 100
				evidence = suite.populateEvidence(suite.ctx, numEvidence)
				req = types2.NewQueryEvidenceRequest(strings.ToUpper(hex.EncodeToString(evidence[0].Hash())))
			},
			true,
			"",
			func(res *types2.QueryEvidenceResponse) {
				var evi exported.Evidence
				err := suite.encCfg.InterfaceRegistry.UnpackAny(res.Evidence, &evi)
				suite.Require().NoError(err)
				suite.Require().NotNil(evi)
				suite.Require().Equal(evi, evidence[0])
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.malleate()
			res, err := suite.queryClient.Evidence(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Contains(err.Error(), tc.expErrMsg)
				suite.Require().Nil(res)
			}

			tc.posttests(res)
		})
	}
}

func (suite *KeeperTestSuite) TestQueryAllEvidence() {
	var req *types2.QueryAllEvidenceRequest

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func(res *types2.QueryAllEvidenceResponse)
	}{
		{
			"success without evidence",
			func() {
				req = &types2.QueryAllEvidenceRequest{}
			},
			true,
			func(res *types2.QueryAllEvidenceResponse) {
				suite.Require().Empty(res.Evidence)
			},
		},
		{
			"success",
			func() {
				numEvidence := 100
				_ = suite.populateEvidence(suite.ctx, numEvidence)
				pageReq := &query.PageRequest{
					Key:        nil,
					Limit:      50,
					CountTotal: false,
				}
				req = types2.NewQueryAllEvidenceRequest(pageReq)
			},
			true,
			func(res *types2.QueryAllEvidenceResponse) {
				suite.Equal(len(res.Evidence), 50)
				suite.NotNil(res.Pagination.NextKey)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.malleate()
			res, err := suite.queryClient.AllEvidence(suite.ctx, req)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)
			} else {
				suite.Require().Error(err)
				suite.Require().Nil(res)
			}

			tc.posttests(res)
		})
	}
}
