package pubkey_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/stretchr/testify/suite"
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

	addr1 := sdk.AccAddress([]byte("addr1"))
	acc1 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr1)
	suite.app.AccountKeeper.SetAccount(ctx, acc1)
	suite.Require().NoError(suite.app.BankKeeper.SetBalances(ctx, addr1, sdk.NewCoins(sdk.NewInt64Coin("stake", 1000))))

	addr2 := sdk.AccAddress([]byte("addr2"))
	acc2 := suite.app.AccountKeeper.NewAccountWithAddress(ctx, addr2)
	suite.app.AccountKeeper.SetAccount(ctx, acc2)
	suite.Require().NoError(suite.app.BankKeeper.SetBalances(ctx, addr2, sdk.NewCoins(sdk.NewInt64Coin("stake", 10000))))

	privKey := secp256k1.GenPrivKeyFromSecret([]byte("mySecret"))
	pubKey := privKey.PubKey()

	testCases := []struct {
		name      string
		msg       *types.MsgChangePubKey
		expectErr bool
	}{
		{
			name:      "try changing pubkey without enough fee balance",
			msg:       types.NewMsgChangePubKey(addr1, pubKey),
			expectErr: true,
		},
		{
			name:      "try changing pubkey with enough balance",
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
				// check remaining balance after successful run
				balance := suite.app.BankKeeper.GetBalance(ctx, tc.msg.GetAddress(), "stake")
				suite.Require().Equal(balance.Amount.Int64(), int64(5000))
			}
		})
	}
}

func TestHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(HandlerTestSuite))
}
