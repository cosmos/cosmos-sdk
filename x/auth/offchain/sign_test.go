package offchain

import (
	"errors"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/suite"
	"testing"
)

type signerTestSuite struct {
	suite.Suite
	signer   Signer
	verifier SignatureVerifier
	privKey  cryptotypes.PrivKey
	address  sdk.AccAddress
}

func (ts *signerTestSuite) SetupTest() {
	encConf := simapp.MakeTestEncodingConfig()
	RegisterInterfaces(encConf.InterfaceRegistry)
	RegisterLegacyAminoCodec(encConf.Amino)
	ts.signer = NewSigner(encConf.TxConfig)
	ts.verifier = NewVerifier(encConf.TxConfig.SignModeHandler())
	ts.privKey = secp256k1.GenPrivKeyFromSecret(nil)
	ts.address = sdk.AccAddress(ts.privKey.PubKey().Address())
}

func (ts *signerTestSuite) TestEmptyMsgs() {
	_, err := ts.signer.Sign(ts.privKey, nil)
	ts.Require().True(errors.Is(err, sdkerrors.ErrInvalidRequest), "unexpected error: %s", err)
}

// tests sign verify cycle works
func (ts *signerTestSuite) TestVerifyCompatibility() {
	tx, err := ts.signer.Sign(ts.privKey, []sdk.Msg{
		NewMsgSignData(ts.address, []byte("data")),
	})
	ts.Require().NoError(err, "error while signing transaction")
	err = ts.verifier.Verify(tx)
	ts.Require().NoError(err, "valid transaction should be verified")
}

func TestSigner(t *testing.T) {
	suite.Run(t, new(signerTestSuite))
}
