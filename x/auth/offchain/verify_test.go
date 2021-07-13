package offchain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
)

type verifyMessageTestSuite struct {
	suite.Suite
	address   sdk.AccAddress
	validData []byte
}

func (ts *verifyMessageTestSuite) TestValidMessage() {
	err := verifyMessage(NewMsgSignData(ts.address, ts.validData))
	ts.Require().NoError(err, "message should be valid")
}

func (ts *verifyMessageTestSuite) TestInvalidMessageType() {
	err := verifyMessage(&types.MsgSend{})
	ts.Require().True(errors.Is(err, errInvalidType), "unexpected error: %s", err)
}

func (ts *verifyMessageTestSuite) TestInvalidRoute() {
	// err := verifyMessage()
	// ts.Require().True(errors.Is(err, errInvalidRoute), "unexpected error: %s", err)
}

type signatureVerifierSuite struct {
	suite.Suite
	verifier  SignatureVerifier
	txDecoder sdk.TxDecoder
	invalidTx []byte
}

func (ts *signatureVerifierSuite) SetupTest() {
	encConf := simapp.MakeTestEncodingConfig()
	RegisterInterfaces(encConf.InterfaceRegistry)
	RegisterLegacyAminoCodec(encConf.Amino)
	ts.txDecoder = encConf.TxConfig.TxJSONDecoder()
	ts.invalidTx = []byte(`{"body":{"messages":[{"@type":"/cosmos.offchain.v1alpha1.MsgSignData","signer":"cosmos1346fyal5a9xxwlygkqmkkqf7g3q3zwzpdmkam8","data":"ZGF0YQ=="}],"memo":"","timeout_height":"0","extension_options":[],"non_critical_extension_options":[]},"auth_info":{"signer_infos":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A+FkzsHk5mVRk8IkVq5p0XapCrqu1MFf8KT594BtN6ss"},"mode_info":{"single":{"mode":"SIGN_MODE_DIRECT"}},"sequence":"0"}],"fee":{"amount":[],"gas_limit":"0","payer":"","granter":""}},"signatures":["Ftbdnj79Hms/JCJS8TkyoESuO1t7M0+A3fDCzPTHpchpzTE5qHRw8L0dKZYTO81ewQO7emRSyB2OshAdPlApHQ=="]}`)
	ts.verifier = NewVerifier(encConf.TxConfig.SignModeHandler())
}

func (ts *signatureVerifierSuite) TestInvalidTxType() {
	err := ts.verifier.Verify((sdk.FeeTx)(nil))
	ts.Suite.True(errors.Is(err, sdkerrors.ErrInvalidRequest), "unexpected error: %s", err)
}

func (ts *signatureVerifierSuite) TestInvalidSignature() {
	tx, err := ts.txDecoder(ts.invalidTx)
	ts.Require().Nil(err, "decode of tx failed")
	err = ts.verifier.Verify(tx)
	ts.Suite.Require().True(errors.Is(sdkerrors.ErrUnauthorized, err), "unexpected error: %s", err)
}

func TestVerifyMessage(t *testing.T) {
	suite.Run(t, new(verifyMessageTestSuite))
}

func TestSignatureVerifier(t *testing.T) {
	suite.Run(t, new(signatureVerifierSuite))
}
