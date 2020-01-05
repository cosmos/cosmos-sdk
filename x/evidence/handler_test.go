package evidence_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/internal/types"

	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/ed25519"
)

type HandlerTestSuite struct {
	suite.Suite

	ctx     sdk.Context
	handler sdk.Handler
	keeper  evidence.Keeper
}

func (suite *HandlerTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	// get the app's codec and register custom testing types
	cdc := app.Codec()
	cdc.RegisterConcrete(types.TestEquivocationEvidence{}, "test/TestEquivocationEvidence", nil)

	// recreate keeper in order to use custom testing types
	evidenceKeeper := evidence.NewKeeper(
		cdc, app.GetKey(evidence.StoreKey), app.GetSubspace(evidence.ModuleName), app.StakingKeeper, app.SlashingKeeper,
	)
	router := evidence.NewRouter()
	router = router.AddRoute(types.TestEvidenceRouteEquivocation, types.TestEquivocationHandler(*evidenceKeeper))
	evidenceKeeper.SetRouter(router)

	suite.ctx = app.BaseApp.NewContext(checkTx, abci.Header{Height: 1})
	suite.handler = evidence.NewHandler(*evidenceKeeper)
	suite.keeper = *evidenceKeeper
}

func (suite *HandlerTestSuite) TestMsgSubmitEvidence_Valid() {
	pk := ed25519.GenPrivKey()
	sv := types.TestVote{
		ValidatorAddress: pk.PubKey().Address(),
		Height:           11,
		Round:            0,
	}

	sig, err := pk.Sign(sv.SignBytes(suite.ctx.ChainID()))
	suite.NoError(err)
	sv.Signature = sig

	s := sdk.AccAddress("test")
	e := types.TestEquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA:      sv,
		VoteB:      sv,
	}

	ctx := suite.ctx.WithIsCheckTx(false)
	msg := evidence.NewMsgSubmitEvidence(e, s)
	res, err := suite.handler(ctx, msg)
	suite.NoError(err)
	suite.NotNil(res)
	suite.Equal(e.Hash().Bytes(), res.Data)
}

func (suite *HandlerTestSuite) TestMsgSubmitEvidence_Invalid() {
	pk := ed25519.GenPrivKey()
	sv := types.TestVote{
		ValidatorAddress: pk.PubKey().Address(),
		Height:           11,
		Round:            0,
	}

	sig, err := pk.Sign(sv.SignBytes(suite.ctx.ChainID()))
	suite.NoError(err)
	sv.Signature = sig

	s := sdk.AccAddress("test")
	e := types.TestEquivocationEvidence{
		Power:      100,
		TotalPower: 100000,
		PubKey:     pk.PubKey(),
		VoteA:      sv,
		VoteB:      types.TestVote{Height: 10, Round: 1},
	}

	ctx := suite.ctx.WithIsCheckTx(false)
	msg := evidence.NewMsgSubmitEvidence(e, s)
	res, err := suite.handler(ctx, msg)
	suite.Error(err)
	suite.Nil(res)
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
