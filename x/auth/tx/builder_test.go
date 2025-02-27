package tx

import (
	"testing"

	any "github.com/cosmos/gogoproto/types/any"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	codectestutil "github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authclient "github.com/cosmos/cosmos-sdk/x/auth/client"
)

func TestIntoV2SignerInfo(t *testing.T) {
	require.NotNil(t, intoV2SignerInfo([]*tx.SignerInfo{{}}))
	require.NotNil(t, intoV2SignerInfo([]*tx.SignerInfo{{PublicKey: &any.Any{}}}))
}

func TestBuilderWithTimeoutTimestamp(t *testing.T) {
	cdc := codectestutil.CodecOptions{}.NewCodec()
	interfaceRegistry := cdc.InterfaceRegistry()
	signingCtx := interfaceRegistry.SigningContext()
	txConfig := NewTxConfig(cdc, signingCtx.AddressCodec(), signingCtx.ValidatorAddressCodec(), DefaultSignModes)
	txBuilder := txConfig.NewTxBuilder()
	encodedTx, err := txConfig.TxJSONEncoder()(txBuilder.GetTx())
	require.NoError(t, err)
	file := testutil.WriteToNewTempFile(t, string(encodedTx))
	clientCtx := client.Context{InterfaceRegistry: interfaceRegistry, TxConfig: txConfig}
	decodedTx, err := authclient.ReadTxFromFile(clientCtx, file.Name())
	require.NoError(t, err)
	txBldr, err := txConfig.WrapTxBuilder(decodedTx)
	require.NoError(t, err)
	b := txBldr.(*builder)
	require.False(t, b.timeoutTimestamp.IsZero())
}
