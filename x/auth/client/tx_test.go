package client_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	clienttx "github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func TestParseQueryResponse(t *testing.T) {
	simRes := &sdk.SimulationResponse{
		GasInfo: sdk.GasInfo{GasUsed: 10, GasWanted: 20},
		Result:  &sdk.Result{Data: []byte("tx data"), Log: "log"},
	}

	bz, err := codec.ProtoMarshalJSON(simRes, nil)
	require.NoError(t, err)

	res, err := authclient.ParseQueryResponse(bz)
	require.NoError(t, err)
	require.Equal(t, 10, int(res.GasUsed))
	require.NotNil(t, res.Result)

	res, err = authclient.ParseQueryResponse([]byte("fuzzy"))
	require.Error(t, err)
}

func TestReadTxFromFile(t *testing.T) {
	t.Parallel()

	encodingConfig := moduletestutil.MakeTestEncodingConfig()
	interfaceRegistry := encodingConfig.InterfaceRegistry
	txConfig := encodingConfig.TxConfig

	clientCtx := client.Context{}
	clientCtx = clientCtx.WithInterfaceRegistry(interfaceRegistry)
	clientCtx = clientCtx.WithTxConfig(txConfig)

	feeAmount := sdk.Coins{sdk.NewInt64Coin("atom", 150)}
	gasLimit := uint64(50000)
	memo := "foomemo"

	txBuilder := txConfig.NewTxBuilder()
	txBuilder.SetFeeAmount(feeAmount)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetMemo(memo)

	// Write it to the file
	encodedTx, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)

	jsonTxFile := testutil.WriteToNewTempFile(t, string(encodedTx))
	// Read it back
	decodedTx, err := authclient.ReadTxFromFile(clientCtx, jsonTxFile.Name())
	require.NoError(t, err)
	txBldr, err := txConfig.WrapTxBuilder(decodedTx)
	require.NoError(t, err)
	t.Log(txBuilder.GetTx())
	t.Log(txBldr.GetTx())
	require.Equal(t, txBuilder.GetTx().GetMemo(), txBldr.GetTx().GetMemo())
	require.Equal(t, txBuilder.GetTx().GetFee(), txBldr.GetTx().GetFee())
}

func TestBatchScanner_Scan(t *testing.T) {
	t.Parallel()

	encodingConfig := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	txConfig := encodingConfig.TxConfig

	clientCtx := client.Context{}
	clientCtx = clientCtx.WithTxConfig(txConfig)

	// generate some tx JSON
	bldr := txConfig.NewTxBuilder()
	bldr.SetGasLimit(50000)
	bldr.SetFeeAmount(sdk.NewCoins(sdk.NewInt64Coin("atom", 150)))
	bldr.SetMemo("foomemo")
	txJSON, err := txConfig.TxJSONEncoder()(bldr.GetTx())
	require.NoError(t, err)

	// use the tx JSON to generate some tx batches (it doesn't matter that we use the same JSON because we don't care about the actual context)
	goodBatchOf3Txs := fmt.Sprintf("%s\n%s\n%s\n", txJSON, txJSON, txJSON)
	malformedBatch := fmt.Sprintf("%s\nmalformed\n%s\n", txJSON, txJSON)
	batchOf2TxsWithNoNewline := fmt.Sprintf("%s\n%s", txJSON, txJSON)
	batchWithEmptyLine := fmt.Sprintf("%s\n\n%s", txJSON, txJSON)

	tests := []struct {
		name               string
		batch              string
		wantScannerError   bool
		wantUnmarshalError bool
		numTxs             int
	}{
		{"good batch", goodBatchOf3Txs, false, false, 3},
		{"malformed", malformedBatch, false, true, 1},
		{"missing trailing newline", batchOf2TxsWithNoNewline, false, false, 2},
		{"empty line", batchWithEmptyLine, false, true, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scanner, i := authclient.NewBatchScanner(clientCtx.TxConfig, strings.NewReader(tt.batch)), 0
			for scanner.Scan() {
				_ = scanner.Tx()
				i++
			}
			require.Equal(t, tt.wantScannerError, scanner.Err() != nil)
			require.Equal(t, tt.wantUnmarshalError, scanner.UnmarshalErr() != nil)
			require.Equal(t, tt.numTxs, i)
		})
	}
}

// recordingAccountRetriever records the address SignTx looks up.
type recordingAccountRetriever struct {
	client.MockAccountRetriever
	queried *sdk.AccAddress
}

func (r recordingAccountRetriever) GetAccountNumberSequence(_ client.Context, addr sdk.AccAddress) (uint64, uint64, error) {
	*r.queried = addr
	return 7, 3, nil
}

func TestSignTx_WithSignerAddress(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(auth.AppModuleBasic{})
	kb := keyring.NewInMemory(encCfg.Codec)
	k1, _, err := kb.NewMnemonic("k1", keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
	require.NoError(t, err)
	keyAddr, err := k1.GetAddress()
	require.NoError(t, err)

	// A is a rekeyed account whose stored pubkey is k1, so addr(k1) != A.
	accountAddr := sdk.AccAddress("rekeyed_account_addr")
	require.NotEqual(t, keyAddr, accountAddr)

	newTx := func(t *testing.T, txf clienttx.Factory) client.TxBuilder {
		t.Helper()
		msg := banktypes.NewMsgSend(accountAddr, keyAddr, sdk.NewCoins(sdk.NewInt64Coin("stake", 1)))
		txb, err := txf.BuildUnsignedTx(msg)
		require.NoError(t, err)
		return txb
	}
	baseFactory := clienttx.Factory{}.
		WithTxConfig(encCfg.TxConfig).
		WithKeybase(kb).
		WithChainID("test-chain").
		WithSignMode(signing.SignMode_SIGN_MODE_DIRECT)

	t.Run("signer address set", func(t *testing.T) {
		var queried sdk.AccAddress
		clientCtx := client.Context{}.
			WithTxConfig(encCfg.TxConfig).
			WithAccountRetriever(recordingAccountRetriever{queried: &queried})
		txf := baseFactory.WithSignerAddress(accountAddr)
		txb := newTx(t, txf)

		require.NoError(t, authclient.SignTx(txf, clientCtx, "k1", txb, false, true))
		require.Equal(t, accountAddr, queried, "account number and sequence must be looked up for the signer address")

		sigs, err := txb.GetTx().GetSignaturesV2()
		require.NoError(t, err)
		require.Len(t, sigs, 1)
		require.Equal(t, uint64(3), sigs[0].Sequence)
	})

	t.Run("no signer address", func(t *testing.T) {
		clientCtx := client.Context{}.WithTxConfig(encCfg.TxConfig)
		txb := newTx(t, baseFactory)

		err := authclient.SignTx(baseFactory, clientCtx, "k1", txb, true, true)
		require.ErrorIs(t, err, errors.ErrorInvalidSigner)
	})
}
