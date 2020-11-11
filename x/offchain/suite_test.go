package offchain

import (
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	"testing"
)

type packageTestSuite struct {
	suite.Suite
	signer   Signer
	verifier SignatureVerifier
	privKey  cryptotypes.PrivKey
	address  sdk.AccAddress
}

func (ts *packageTestSuite) SetupTest() {
	encConf := params.MakeTestEncodingConfig()
	ts.signer = NewSigner(encConf.TxConfig)
	ts.verifier = NewVerifier(encConf.TxConfig.SignModeHandler())
	ts.privKey = secp256k1.GenPrivKeyFromSecret(nil)
	ts.address = sdk.AccAddress(ts.privKey.PubKey().Address())
}

// tests sign verify cycle works
func (ts *packageTestSuite) TestSignVerify() {
	tx, err := ts.signer.Sign(ts.privKey, []msg{
		NewMsgSignData(ts.address, []byte("data")),
	})
	ts.Require().NoError(err, "error while signing transaction")

	err = ts.verifier.Verify(tx)
	ts.Require().NoError(err, "valid transaction should be verified")
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(packageTestSuite))
}
