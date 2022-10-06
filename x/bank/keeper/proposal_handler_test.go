package keeper_test

import (
	"errors"
	"github.com/cosmos/cosmos-sdk/simapp"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"testing"

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
	k   bankkeeper.Keeper

	accountAddr sdk.AccAddress
}

func (s *ProposalIntegrationTestSuite) SetupSuite() {
	s.accountAddr = sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address())
}

func (s *ProposalIntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")
}

func (s *ProposalIntegrationTestSuite) TestMsgFeeProposals() {
	testCases := []struct {
		name string
		prop govtypesv1beta1.Content
		err  error
	}{
		{
			"set denom metadata - bad denom",
			banktypes.NewUpdateDenomMetadataProposal("title", "description",
				banktypes.Metadata{
					Description: "some denom description",
					Base:        "bad$char",
					Display:     "badchar",
					Name:        "Bad Char",
					Symbol:      "BC",
					DenomUnits: []*banktypes.DenomUnit{
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
			banktypes.NewUpdateDenomMetadataProposal("title", "description",
				banktypes.Metadata{
					Description: "the best denom description",
					Base:        "test1",
					Display:     "test1",
					Name:        "Test One",
					Symbol:      "TONE",
					DenomUnits: []*banktypes.DenomUnit{
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

		s.T().Run(tc.name, func(t *testing.T) {
			var err error
			switch c := tc.prop.(type) {
			case *banktypes.UpdateDenomMetadataProposal:
				err = bankkeeper.HandleUpdateDenomMetadataProposal(s.ctx, s.k, c, s.app.InterfaceRegistry())
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
