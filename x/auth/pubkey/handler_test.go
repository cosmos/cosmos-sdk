package pubkey_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/pubkey"
	"github.com/cosmos/cosmos-sdk/x/auth/pubkey/types"
)

type HandlerTestSuite struct {
	suite.Suite

	handler sdk.Handler
	app     *simapp.SimApp
}

func (suite *HandlerTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.handler = pubkey.NewHandler(app.AccountKeeper, app.BankKeeper, app.StakingKeeper)
	suite.app = app
}

func (suite *HandlerTestSuite) TestMsgChangePubKey() {
	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

	balances := sdk.NewCoins(sdk.NewInt64Coin("test", 1000))
	addr1 := sdk.AccAddress([]byte("addr1"))

	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(suite.app.BankKeeper.SetBalances(ctx, addr1, balances))

	var pubKey crypto.PubKey // TODO should define pubKey to use for testing

	testCases := []struct {
		name      string
		msg       *types.MsgChangePubKey
		expectErr bool
	}{
		{
			name:      "try changing pubkey",
			msg:       types.NewMsgChangePubKey(addr1, pubKey),
			expectErr: false,
		},
		// TODO should add more tests
	}

	for _, tc := range testCases {
		tc := tc

		suite.Run(tc.name, func() {
			res, err := suite.handler(ctx, tc.msg)
			if tc.expectErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().NotNil(res)

				accI := suite.app.AccountKeeper.GetAccount(ctx, tc.msg.Address)
				suite.Require().NotNil(accI)
			}
		})
	}
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
