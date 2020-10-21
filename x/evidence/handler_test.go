package evidence_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/evidence/exported"
	"github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	"github.com/cosmos/cosmos-sdk/x/evidence/types"
)

type HandlerTestSuite struct {
	suite.Suite

	handler sdk.Handler
	app     *simapp.SimApp
}

func testMsgSubmitEvidence(r *require.Assertions, e exported.Evidence, s sdk.AccAddress) exported.MsgSubmitEvidenceI {
	msg, err := types.NewMsgSubmitEvidence(s, e)
	r.NoError(err)
	return msg
}

func testEquivocationHandler(k interface{}) types.Handler {
	return func(ctx sdk.Context, e exported.Evidence) error {
		if err := e.ValidateBasic(); err != nil {
			return err
		}

		ee, ok := e.(*types.Equivocation)
		if !ok {
			return fmt.Errorf("unexpected evidence type: %T", e)
		}
		if ee.Height%2 == 0 {
			return fmt.Errorf("unexpected even evidence height: %d", ee.Height)
		}

		return nil
	}
}

func (suite *HandlerTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	// recreate keeper in order to use custom testing types
	evidenceKeeper := keeper.NewKeeper(
		app.AppCodec(), app.GetKey(types.StoreKey), app.StakingKeeper, app.SlashingKeeper,
	)
	router := types.NewRouter()
	router = router.AddRoute(types.RouteEquivocation, testEquivocationHandler(*evidenceKeeper))
	evidenceKeeper.SetRouter(router)

	app.EvidenceKeeper = *evidenceKeeper

	suite.handler = evidence.NewHandler(*evidenceKeeper)
	suite.app = app
}

func (suite *HandlerTestSuite) TestMsgSubmitEvidence() {
	pk := ed25519.GenPrivKey()
	s := sdk.AccAddress("test________________")

	testCases := []struct {
		msg       sdk.Msg
		expectErr bool
	}{
		{
			testMsgSubmitEvidence(
				suite.Require(),
				&types.Equivocation{
					Height:           11,
					Time:             time.Now().UTC(),
					Power:            100,
					ConsensusAddress: pk.PubKey().Address().String(),
				},
				s,
			),
			false,
		},
		{
			testMsgSubmitEvidence(
				suite.Require(),
				&types.Equivocation{
					Height:           10,
					Time:             time.Now().UTC(),
					Power:            100,
					ConsensusAddress: pk.PubKey().Address().String(),
				},
				s,
			),
			true,
		},
	}

	for i, tc := range testCases {
		ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

		res, err := suite.handler(ctx, tc.msg)
		if tc.expectErr {
			suite.Require().Error(err, "expected error; tc #%d", i)
		} else {
			suite.Require().NoError(err, "unexpected error; tc #%d", i)
			suite.Require().NotNil(res, "expected non-nil result; tc #%d", i)

			msg := tc.msg.(exported.MsgSubmitEvidenceI)

			var resultData types.MsgSubmitEvidenceResponse
			suite.app.AppCodec().UnmarshalBinaryBare(res.Data, &resultData)
			suite.Require().Equal(msg.GetEvidence().Hash().Bytes(), resultData.Hash, "invalid hash; tc #%d", i)
		}
	}
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
