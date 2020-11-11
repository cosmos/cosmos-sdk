package offchain

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"os"
	"testing"
)

var testSigner Signer
var testVerifier SignatureVerifier
var testPrivKey cryptotypes.PrivKey
var testAddress sdk.AccAddress

func TestMain(m *testing.M) {
	err := initTest()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to init test: %s", err)
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func initTest() error {
	encConf := params.MakeTestEncodingConfig()
	testSigner = NewSigner(encConf.TxConfig)
	testVerifier = NewVerifier(encConf.TxConfig.SignModeHandler())
	testPrivKey = secp256k1.GenPrivKeyFromSecret(nil)
	testAddress = sdk.AccAddress(testPrivKey.PubKey().Address())
	return nil
}
