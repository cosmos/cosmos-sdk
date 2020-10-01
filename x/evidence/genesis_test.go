package evidence_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper keeper.Keeper
}

func (suite *GenesisTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{Height: 1})
	suite.keeper = app.EvidenceKeeper
}

func (suite *GenesisTestSuite) TestInitGenesis() {
	var (
		genesisState *types.GenesisState
		testEvidence []exported.Evidence
		pk           = ed25519.GenPrivKey()
	)

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func()
	}{
		{
			"valid",
			func() {
				testEvidence = make([]exported.Evidence, 100)
				for i := 0; i < 100; i++ {
					testEvidence[i] = &types.Equivocation{
						Height:           int64(i + 1),
						Power:            100,
						Time:             time.Now().UTC(),
						ConsensusAddress: pk.PubKey().Address().String(),
					}
				}
				genesisState = types.NewGenesisState(testEvidence)
			},
			true,
			func() {
				for _, e := range testEvidence {
					_, ok := suite.keeper.GetEvidence(suite.ctx, e.Hash())
					suite.True(ok)
				}
			},
		},
		{
			"invalid",
			func() {
				testEvidence = make([]exported.Evidence, 100)
				for i := 0; i < 100; i++ {
					testEvidence[i] = &types.Equivocation{
						Power:            100,
						Time:             time.Now().UTC(),
						ConsensusAddress: pk.PubKey().Address().String(),
					}
				}
				genesisState = types.NewGenesisState(testEvidence)
			},
			false,
			func() {
				suite.Empty(suite.keeper.GetAllEvidence(suite.ctx))
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.malleate()

			if tc.expPass {
				suite.NotPanics(func() {
					evidence.InitGenesis(suite.ctx, suite.keeper, genesisState)
				})
			} else {
				suite.Panics(func() {
					evidence.InitGenesis(suite.ctx, suite.keeper, genesisState)
				})
			}

			tc.posttests()
		})
	}
}

func (suite *GenesisTestSuite) TestExportGenesis() {
	pk := ed25519.GenPrivKey()

	testCases := []struct {
		msg       string
		malleate  func()
		expPass   bool
		posttests func()
	}{
		{
			"success",
			func() {
				suite.keeper.SetEvidence(suite.ctx, &types.Equivocation{
					Height:           1,
					Power:            100,
					Time:             time.Now().UTC(),
					ConsensusAddress: pk.PubKey().Address().String(),
				})
			},
			true,
			func() {},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.malleate()

			if tc.expPass {
				suite.NotPanics(func() {
					evidence.ExportGenesis(suite.ctx, suite.keeper)
				})
			} else {
				suite.Panics(func() {
					evidence.ExportGenesis(suite.ctx, suite.keeper)
				})
			}

			tc.posttests()
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
