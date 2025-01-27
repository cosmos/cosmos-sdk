package evidence

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"cosmossdk.io/x/evidence"
	"cosmossdk.io/x/evidence/exported"
	"cosmossdk.io/x/evidence/keeper"
	"cosmossdk.io/x/evidence/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    context.Context
	keeper keeper.Keeper
}

func (suite *GenesisTestSuite) SetupTest() {
	f := initFixture(suite.T())

	suite.ctx = f.ctx
	suite.keeper = f.evidenceKeeper
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
					_, err := suite.keeper.Evidences.Get(suite.ctx, e.Hash())
					suite.NoError(err)
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
				_, err := suite.keeper.Evidences.Iterate(suite.ctx, nil)
				suite.Require().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.malleate()

			if tc.expPass {
				err := evidence.InitGenesis(suite.ctx, suite.keeper, genesisState)
				suite.NoError(err)
			} else {
				err := evidence.InitGenesis(suite.ctx, suite.keeper, genesisState)
				suite.Error(err)

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
				ev := &types.Equivocation{
					Height:           1,
					Power:            100,
					Time:             time.Now().UTC(),
					ConsensusAddress: pk.PubKey().Address().String(),
				}
				suite.Require().NoError(suite.keeper.Evidences.Set(suite.ctx, ev.Hash(), ev))
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

				_, err := evidence.ExportGenesis(suite.ctx, suite.keeper)
				suite.Require().NoError(err)

			} else {
				_, err := evidence.ExportGenesis(suite.ctx, suite.keeper)
				suite.Require().Error(err)
			}

			tc.posttests()
		})
	}
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
