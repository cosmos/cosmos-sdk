package signing_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	authsign "cosmossdk.io/x/auth/signing"

	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
)

func TestGetSignBytesAdapterNoPublicKey(t *testing.T) {
	encodingConfig := moduletestutil.MakeTestEncodingConfig(codectestutil.CodecOptions{})
	txConfig := encodingConfig.TxConfig
	_, _, addr := testdata.KeyTestPubAddr()
	addrStr, err := codectestutil.CodecOptions{}.GetAddressCodec().BytesToString(addr)
	require.NoError(t, err)
	signerData := authsign.SignerData{
		Address:       addrStr,
		ChainID:       "test-chain",
		AccountNumber: 11,
		Sequence:      15,
	}
	w := txConfig.NewTxBuilder()
	_, err = authsign.GetSignBytesAdapter(
		context.Background(),
		txConfig.SignModeHandler(),
		signing.SignMode_SIGN_MODE_DIRECT,
		signerData,
		w.GetTx())
	require.NoError(t, err)
}
