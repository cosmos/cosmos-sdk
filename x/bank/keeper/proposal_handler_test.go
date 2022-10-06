package keeper_test

import (
	"errors"
	"github.com/cosmos/cosmos-sdk/simapp"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/tendermint/tendermint/crypto/secp256k1"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypesv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

type ProposalIntegrationTestSuite struct {
	suite.Suite

	app *simapp.SimApp
	ctx sdk.Context
	accountAddr sdk.AccAddress
}

func (suite *ProposalIntegrationTestSuite) SetupSuite() {
	app := simapp.Setup(suite.T(), false)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{Time: time.Now()})

	suite.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
	app.BankKeeper.SetParams(ctx, types.DefaultParams())

	suite.app = app
	suite.ctx = ctx
}

func (suite *ProposalIntegrationTestSuite) TearDownSuite() {
	suite.T().Log("tearing down integration test suite")
}

func (suite *ProposalIntegrationTestSuite) TestMsgFeeProposals() {
	testCases := []struct {
		name string
		prop govtypesv1beta1.Content
		err  error
	}{
		{
			"set denom metadata - bad denom",
			types.NewUpdateDenomMetadataProposal("title", "description",
					types.Metadata{
					Description: "some denom description",
					Base:        "bad$char",
					Display:     "badchar",
					Name:        "Bad Char",
					Symbol:      "BC",
					DenomUnits: []*types.DenomUnit{
						{
							Denom:    "bad$char",
							Exponent: 0,
							Aliases:  nil,
						},
					},
				},
			),
			errors.New("invalid denom: bad$char"),
		},
		{
			"set denom metadata - valid",
			types.NewUpdateDenomMetadataProposal("title", "description",
					types.Metadata{
					Description: "the best denom description",
					Base:        "test1",
					Display:     "test1",
					Name:        "Test One",
					Symbol:      "TONE",
					DenomUnits: []*types.DenomUnit{
						{
							Denom:    "test1",
							Exponent: 0,
							Aliases:  []string{"tone"},
						},
					},
				},
			),
			nil,
		},
	}

	for _, tc := range testCases {
		tc := tc

		suite.T().Run(tc.name, func(t *testing.T) {
			var err error
			switch c := tc.prop.(type) {
			case *types.UpdateDenomMetadataProposal:
				err = bankkeeper.HandleUpdateDenomMetadataProposal(suite.ctx, suite.app.BankKeeper, c, suite.app.InterfaceRegistry())
			default:
				panic("invalid proposal type")
			}

			if tc.err != nil {
				require.Error(t, err)
				require.Equal(t, tc.err.Error(), err.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}

}

func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ProposalIntegrationTestSuite))
}
