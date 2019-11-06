package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type KeeperTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	querier sdk.Querier
	keeper  keeper.Keeper
}

func (suite *KeeperTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	// get the app's codec and register custom testing types
	cdc := app.Codec()
	cdc.RegisterConcrete(types.TestEquivocationEvidence{}, "test/TestEquivocationEvidence", nil)

	// recreate keeper in order to use custom testing types
	evidenceKeeper := evidence.NewKeeper(
		cdc, app.GetKey(evidence.StoreKey), app.GetSubspace(evidence.ModuleName),
		evidence.DefaultCodespace,
	)
	router := evidence.NewRouter()
	router = router.AddRoute(types.TestEvidenceRouteEquivocation, types.TestEquivocationHandler(*evidenceKeeper))
	evidenceKeeper.SetRouter(router)

	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1})
	suite.querier = keeper.NewQuerier(*evidenceKeeper)
	suite.keeper = *evidenceKeeper
}

func (suite *KeeperTestSuite) populateEvidence(ctx sdk.Context, numEvidence int) []exported.Evidence {
	evidence := make([]exported.Evidence, numEvidence)

	for i := 0; i < numEvidence; i++ {
		pk := ed25519.GenPrivKey()
		sv := types.TestVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           int64(i),
			Round:            0,
		}

		sig, err := pk.Sign(sv.SignBytes(ctx.ChainID()))
		suite.NoError(err)
		sv.Signature = sig

		evidence[i] = types.TestEquivocationEvidence{
			Power:      100,
			TotalPower: 100000,
			PubKey:     pk.PubKey(),
			VoteA:      sv,
			VoteB:      sv,
		}

		suite.Nil(suite.keeper.SubmitEvidence(ctx, evidence[i]))
	}

	return evidence
}

func (suite *KeeperTestSuite) TestSubmitValidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()
	sv := types.TestVote{
		ValidatorAddress: pk.PubKey().Address(),
		Height:           11,
		Round:            0,
	}

	sig, err := pk.Sign(sv.SignBytes(ctx.ChainID()))
	suite.NoError(err)
	sv.Signature = sig

	e := types.TestEquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA:      sv,
		VoteB:      sv,
	}

	suite.Nil(suite.keeper.SubmitEvidence(ctx, e))

	res, ok := suite.keeper.GetEvidence(ctx, e.Hash())
	suite.True(ok)
	suite.Equal(e, res)
}

func (suite *KeeperTestSuite) TestSubmitValidEvidence_Duplicate() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()
	sv := types.TestVote{
		ValidatorAddress: pk.PubKey().Address(),
		Height:           11,
		Round:            0,
	}

	sig, err := pk.Sign(sv.SignBytes(ctx.ChainID()))
	suite.NoError(err)
	sv.Signature = sig

	e := types.TestEquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA:      sv,
		VoteB:      sv,
	}

	suite.Nil(suite.keeper.SubmitEvidence(ctx, e))
	suite.Error(suite.keeper.SubmitEvidence(ctx, e))

	res, ok := suite.keeper.GetEvidence(ctx, e.Hash())
	suite.True(ok)
	suite.Equal(e, res)
}

func (suite *KeeperTestSuite) TestSubmitInvalidEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	pk := ed25519.GenPrivKey()
	e := types.TestEquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA: types.TestVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           10,
			Round:            0,
		},
		VoteB: types.TestVote{
			ValidatorAddress: pk.PubKey().Address(),
			Height:           11,
			Round:            0,
		},
	}

	suite.Error(suite.keeper.SubmitEvidence(ctx, e))

	res, ok := suite.keeper.GetEvidence(ctx, e.Hash())
	suite.False(ok)
	suite.Nil(res)
}

func (suite *KeeperTestSuite) TestIterateEvidence() {
	ctx := suite.ctx.WithIsCheckTx(false)
	numEvidence := 100
	suite.populateEvidence(ctx, numEvidence)

	evidence := suite.keeper.GetAllEvidence(ctx)
	suite.Len(evidence, numEvidence)
}

func (suite *KeeperTestSuite) TestGetEvidenceHandler() {
	handler, err := suite.keeper.GetEvidenceHandler(types.TestEquivocationEvidence{}.Route())
	suite.NoError(err)
	suite.NotNil(handler)

	handler, err = suite.keeper.GetEvidenceHandler("invalidHandler")
	suite.Error(err)
	suite.Nil(handler)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
