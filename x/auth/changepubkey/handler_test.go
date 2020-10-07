package changepubkey_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	changepubkey "github.com/cosmos/cosmos-sdk/x/auth/changepubkey"
	"github.com/cosmos/cosmos-sdk/x/auth/changepubkey/types"
)

type HandlerTestSuite struct {
	suite.Suite

	handler sdk.Handler
	app     *simapp.SimApp
}

func (suite *HandlerTestSuite) SetupTest() {
	checkTx := false
	app := simapp.Setup(checkTx)

	suite.handler = changepubkey.NewHandler(app.AccountKeeper)
	suite.app = app
}

func (suite *HandlerTestSuite) TestMsgChangePubKey() {
	ctx := suite.app.BaseApp.NewContext(false, tmproto.Header{Height: suite.app.LastBlockHeight() + 1})

	addr1 := sdk.AccAddress([]byte("any----------address"))
	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(suite.app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))))

	addr2 := sdk.AccAddress([]byte("some---------address"))
	acc2 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	suite.Require().NotNil(acc2)
	suite.Require().Equal(addr2, acc2.GetAddress())

	suite.app.AccountKeeper.SetAccount(ctx, acc2)
	suite.Require().NotNil(suite.app.AccountKeeper.GetAccount(ctx, addr2))
	suite.Require().NoError(suite.app.BankKeeper.SetBalances(ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("stake", 10000))))

	privKey := secp256k1.GenPrivKeyFromSecret([]byte("mySecret"))
	pubKey := privKey.PubKey()

	testCases := []struct {
		name      string
		msg       *types.MsgChangePubKey
		expectErr bool
	}{
		{
			name:      "try changing pubkey",
			msg:       types.NewMsgChangePubKey(addr2, pubKey),
			expectErr: false,
		},
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
