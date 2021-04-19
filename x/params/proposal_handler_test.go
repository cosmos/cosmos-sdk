package params_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
)

type HandlerTestSuite struct {
	suite.Suite

	app        *simapp.SimApp
	ctx        sdk.Context
	govHandler govtypes.Handler
}

func (suite *HandlerTestSuite) SetupTest() {
	suite.app = simapp.Setup(false)
	suite.ctx = suite.app.BaseApp.NewContext(false, tmproto.Header{})
	suite.govHandler = params.NewParamChangeProposalHandler(suite.app.ParamsKeeper)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}

func testProposal(changes ...proposal.ParamChange) *proposal.ParameterChangeProposal {
	return proposal.NewParameterChangeProposal("title", "description", changes)
}

func (suite *HandlerTestSuite) TestProposalHandlerPassed() {
	// tp := testProposal(proposal.NewParamChange(testSubspace, keyMaxValidators, "1"))
	// suite.Require().NoError(suite.govHandler(suite.ctx, tp))

	// var param uint16
	// ss.Get(suite.ctx, []byte(keyMaxValidators), &param)
	// suite.Require().Equal(param, uint16(1))
}

func (suite *HandlerTestSuite) TestProposalHandlerFailed() {

	// tp := testProposal(proposal.NewParamChange(testSubspace, keyMaxValidators, "invalidType"))
	// suite.Require().Error(suite.govHandler(suite.ctx, tp))

	// suite.Require().False(ss.Has(suite.ctx, []byte(keyMaxValidators)))
}

func (suite *HandlerTestSuite) TestProposalHandlerUpdateOmitempty() {

	// var param testParamsSlashingRate

	// tp := testProposal(proposal.NewParamChange(testSubspace, keySlashingRate, `{"downtime": 7}`))
	// suite.Require().NoError(suite.govHandler(suite.ctx, tp))

	// ss.Get(suite.ctx, []byte(keySlashingRate), &param)
	// suite.Require().Equal(testParamsSlashingRate{0, 7}, param)

	// tp = testProposal(proposal.NewParamChange(testSubspace, keySlashingRate, `{"double_sign": 10}`))
	// suite.Require().NoError(suite.govHandler(suite.ctx, tp))

	// ss.Get(suite.ctx, []byte(keySlashingRate), &param)
	// suite.Require().Equal(testParamsSlashingRate{10, 7}, param)

	tp := testProposal(proposal.ParamChange{
		Subspace: "gov",
		Key:      "depositparams",
		Value:    `{"min_deposit": [{"denom": "uatom","amount": "64000000"}]}`,
	})

	suite.Require().NoError(suite.govHandler(suite.ctx, tp))

	depositParams := suite.app.GovKeeper.GetDepositParams(suite.ctx)
	suite.Require().Equal(govtypes.DepositParams{
		MinDeposit:       sdk.NewCoins(sdk.NewCoin("uatom", sdk.NewInt(64000000))),
		MaxDepositPeriod: govtypes.DefaultPeriod,
	}, depositParams)
}
