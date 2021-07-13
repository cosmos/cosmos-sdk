package offchain

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

type signerTestSuite struct {
	suite.Suite
	signer   Signer
	verifier SignatureVerifier
	address  sdk.AccAddress
}

func (ts *signerTestSuite) SetupTest() {
	encConf := simapp.MakeTestEncodingConfig()
	privKey := secp256k1.GenPrivKeyFromSecret(nil)

	RegisterInterfaces(encConf.InterfaceRegistry)
	RegisterLegacyAminoCodec(encConf.Amino)
	ts.signer = NewSigner(encConf.TxConfig, privKey)
	ts.verifier = NewVerifier(encConf.TxConfig.SignModeHandler())
	ts.address = sdk.AccAddress(privKey.PubKey().Address())
}

func (ts *signerTestSuite) TestEmptyMsgs() {
	_, err := ts.signer.Sign(nil)
	ts.Require().True(errors.Is(err, sdkerrors.ErrInvalidRequest), "unexpected error: %s", err)
}

// tests sign verify cycle works
func (ts *signerTestSuite) TestVerifyCompatibility() {
	tx, err := ts.signer.Sign([]sdk.Msg{
		NewMsgSignData(ts.address, []byte("data")),
	})
	ts.Require().NoError(err, "error while signing transaction")
	err = ts.verifier.Verify(tx)
	ts.Require().NoError(err, "valid transaction should be verified")
}

func TestSigner(t *testing.T) {
	suite.Run(t, new(signerTestSuite))
}
