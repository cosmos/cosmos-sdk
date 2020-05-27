package evidence_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/types/time"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

type GenesisTestSuite struct {
	suite.Suite

	ctx    sdk.Context
	keeper evidence.Keeper
}

func (suite *GenesisTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1})
	suite.keeper = app.EvidenceKeeper
}

func (suite *GenesisTestSuite) TestInitGenesis_Valid() {
	pk := ed25519.GenPrivKey()

	testEvidence := make([]exported.Evidence, 100)
	for i := 0; i < 100; i++ {
		testEvidence[i] = &types.Equivocation{
			Height:           int64(i + 1),
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: pk.PubKey().Address().Bytes(),
		}
	}

	suite.NotPanics(func() {
		evidence.InitGenesis(suite.ctx, suite.keeper, evidence.NewGenesisState(testEvidence))
	})

	for _, e := range testEvidence {
		_, ok := suite.keeper.GetEvidence(suite.ctx, e.Hash())
		suite.True(ok)
	}
}

func (suite *GenesisTestSuite) TestInitGenesis_Invalid() {
	pk := ed25519.GenPrivKey()

	testEvidence := make([]exported.Evidence, 100)
	for i := 0; i < 100; i++ {
		testEvidence[i] = &types.Equivocation{
			Power:            100,
			Time:             time.Now().UTC(),
			ConsensusAddress: pk.PubKey().Address().Bytes(),
		}
	}

	suite.Panics(func() {
		evidence.InitGenesis(suite.ctx, suite.keeper, evidence.NewGenesisState(testEvidence))
	})

	suite.Empty(suite.keeper.GetAllEvidence(suite.ctx))
}

func TestGenesisTestSuite(t *testing.T) {
	suite.Run(t, new(GenesisTestSuite))
}
