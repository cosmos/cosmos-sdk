package tx_test

import (
	"context"
	"io"
	"testing"

	gogoproto "github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	protov2 "google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	txtestutil "github.com/cosmos/cosmos-sdk/x/auth/tx/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	txsigning "github.com/cosmos/cosmos-sdk/x/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/tx/signing/aminojson"
)

func TestGenerator(t *testing.T) {
	interfaceRegistry := testutil.CodecOptions{}.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	interfaceRegistry.RegisterImplementations((*sdk.Msg)(nil), &testdata.TestMsg{})
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	suite.Run(t, txtestutil.NewTxConfigTestSuite(tx.NewTxConfig(protoCodec, tx.DefaultSignModes)))
}

func TestConfigOptions(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	protoCodec := codec.NewProtoCodec(interfaceRegistry)
	configOptions := tx.ConfigOptions{}
	txConfig, err := tx.NewTxConfigWithOptions(protoCodec, configOptions)
	require.NoError(t, err)
	require.NotNil(t, txConfig)
	handler := txConfig.SignModeHandler()
	require.NotNil(t, handler)
}

// TestNewTxConfigPropagatesCustomGetSigners verifies that when the codec's
// interface registry was built with CustomGetSigners, those custom signer
// functions are reachable through txConfig.SigningContext(). Regression test
// for https://github.com/cosmos/cosmos-sdk/issues/22200.
func TestNewTxConfigPropagatesCustomGetSigners(t *testing.T) {
	addrBytes := []byte("test-signer-addr-bytes")
	customSigner := func(_ protov2.Message) ([][]byte, error) {
		return [][]byte{addrBytes}, nil
	}

	signingOpts := txsigning.Options{
		AddressCodec:          address.NewBech32Codec("cosmos"),
		ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
	}
	signingOpts.DefineCustomGetSigners(protoreflect.FullName(gogoproto.MessageName(&testdata.TestMsg{})), customSigner)

	interfaceRegistry, err := types.NewInterfaceRegistryWithOptions(types.InterfaceRegistryOptions{
		ProtoFiles:     gogoproto.HybridResolver,
		SigningOptions: signingOpts,
	})
	require.NoError(t, err)
	protoCodec := codec.NewProtoCodec(interfaceRegistry)

	txConfig := tx.NewTxConfig(protoCodec, tx.DefaultSignModes)

	// Sanity: txConfig must reuse the interface registry's signing context so
	// custom signers carry through. With the pre-fix code, NewTxConfig built a
	// fresh context from defaults and the lookup below failed.
	require.Same(t, interfaceRegistry.SigningContext(), txConfig.SigningContext())
}

// TestNewTxConfigWithExplicitSigningOptions covers the branch where the caller
// supplies SigningOptions directly. The function must build a fresh signing
// context from those options (not reuse the codec's interface-registry
// context) and must default FileResolver to the interface registry when
// unset.
func TestNewTxConfigWithExplicitSigningOptions(t *testing.T) {
	interfaceRegistry := types.NewInterfaceRegistry()
	protoCodec := codec.NewProtoCodec(interfaceRegistry)

	signingOpts := &txsigning.Options{
		AddressCodec:          address.NewBech32Codec("cosmos"),
		ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
	}

	txConfig, err := tx.NewTxConfigWithOptions(protoCodec, tx.ConfigOptions{
		SigningOptions: signingOpts,
	})
	require.NoError(t, err)
	require.NotNil(t, txConfig.SigningContext())
	// A new context was built from the supplied SigningOptions, so it must not
	// be the interface registry's own signing context.
	require.NotSame(t, interfaceRegistry.SigningContext(), txConfig.SigningContext())
	// FileResolver default came from the interface registry.
	require.NotNil(t, signingOpts.FileResolver)
}

// TestConfigOptionsAminoJSONEncoder verifies that a custom aminojson.Encoder
// supplied via ConfigOptions.AminoJSONEncoder is wired into the
// SIGN_MODE_LEGACY_AMINO_JSON handler, so applications can configure custom
// field encodings without replicating the SDK's HandlerMap construction.
// Regression test for https://github.com/cosmos/cosmos-sdk/issues/25221.
func TestConfigOptionsAminoJSONEncoder(t *testing.T) {
	interfaceRegistry := testutil.CodecOptions{}.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	protoCodec := codec.NewProtoCodec(interfaceRegistry)

	// Sentinel emitted by a custom encoder for the "legacy_coins" field encoding,
	// which bank MsgSend amounts and tx fees are tagged with.
	const sentinel = `"custom-amino-coins-encoding"`
	customCoinsEncoder := func(_ *aminojson.Encoder, _ protoreflect.Value, w io.Writer) error {
		_, err := w.Write([]byte(sentinel))
		return err
	}
	enc := aminojson.NewEncoder(aminojson.EncoderOptions{
		FileResolver: interfaceRegistry,
	}).DefineFieldEncoding("legacy_coins", customCoinsEncoder)

	signModes := []signingtypes.SignMode{signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON}

	// With the custom encoder, the sentinel must appear in the amino JSON sign bytes.
	customConfig, err := tx.NewTxConfigWithOptions(protoCodec, tx.ConfigOptions{
		EnabledSignModes: signModes,
		AminoJSONEncoder: &enc,
	})
	require.NoError(t, err)
	require.Contains(t, string(getAminoJSONSignBytes(t, customConfig)), sentinel)

	// Without the option the default encoder is used, so the sentinel is absent.
	defaultConfig, err := tx.NewTxConfigWithOptions(protoCodec, tx.ConfigOptions{
		EnabledSignModes: signModes,
	})
	require.NoError(t, err)
	require.NotContains(t, string(getAminoJSONSignBytes(t, defaultConfig)), sentinel)
}

// getAminoJSONSignBytes builds a bank MsgSend transaction with the given
// TxConfig and returns its SIGN_MODE_LEGACY_AMINO_JSON sign bytes.
func getAminoJSONSignBytes(t *testing.T, txConfig client.TxConfig) []byte {
	t.Helper()

	_, pubKey, fromAddr := testdata.KeyTestPubAddr()
	_, _, toAddr := testdata.KeyTestPubAddr()
	coins := sdk.NewCoins(sdk.NewInt64Coin("stake", 10))

	builder := txConfig.NewTxBuilder()
	require.NoError(t, builder.SetMsgs(banktypes.NewMsgSend(fromAddr, toAddr, coins)))
	builder.SetFeeAmount(coins)
	builder.SetGasLimit(10000)
	builder.SetMemo("memo")

	signerData := authsigning.SignerData{
		Address:       fromAddr.String(),
		ChainID:       "test-chain",
		AccountNumber: 1,
		Sequence:      1,
		PubKey:        pubKey,
	}
	signBz, err := authsigning.GetSignBytesAdapter(
		context.Background(),
		txConfig.SignModeHandler(),
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		signerData,
		builder.GetTx(),
	)
	require.NoError(t, err)
	return signBz
}
