package offchain

import (
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/auth/tx"
	txmodule "cosmossdk.io/x/auth/tx/config"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing" // TODO: needed as textual is not enabled by default
)

func getCodec() codec.Codec {
	registry := testutil.CodecOptions{}.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)

	return codec.NewProtoCodec(registry)
}

func MakeTestTxConfig(t *testing.T) client.TxConfig {
	t.Helper()
	enabledSignModes := []signingtypes.SignMode{
		signingtypes.SignMode_SIGN_MODE_DIRECT,
		signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		signingtypes.SignMode_SIGN_MODE_TEXTUAL,
	}
	initClientCtx := client.Context{}
	txConfigOpts := tx.ConfigOptions{
		EnabledSignModes:           enabledSignModes,
		TextualCoinMetadataQueryFn: txmodule.NewGRPCCoinMetadataQueryFn(initClientCtx),
	}
	ir, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: signing.Options{
			AddressCodec:          address.NewBech32Codec("cosmos"),
			ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
		},
	})
	require.NoError(t, err)

	cryptocodec.RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)
	txConfig, err := tx.NewTxConfigWithOptions(cdc, txConfigOpts)
	require.NoError(t, err)

	return txConfig
}

func Test_getSignMode(t *testing.T) {
	tests := []struct {
		name        string
		signModeStr string
		want        apitxsigning.SignMode
	}{
		{
			name:        "direct",
			signModeStr: flags.SignModeDirect,
			want:        apitxsigning.SignMode_SIGN_MODE_DIRECT,
		},
		{
			name:        "legacy Amino JSON",
			signModeStr: flags.SignModeLegacyAminoJSON,
			want:        apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
		},
		{
			name:        "direct Aux",
			signModeStr: flags.SignModeDirectAux,
			want:        apitxsigning.SignMode_SIGN_MODE_DIRECT_AUX,
		},
		{
			name:        "textual",
			signModeStr: flags.SignModeTextual,
			want:        apitxsigning.SignMode_SIGN_MODE_TEXTUAL,
		},
		{
			name:        "unspecified",
			signModeStr: "",
			want:        apitxsigning.SignMode_SIGN_MODE_UNSPECIFIED,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getSignMode(tt.signModeStr)
			require.Equal(t, got, tt.want)
		})
	}
}

func Test_sign(t *testing.T) {
	k := keyring.NewInMemory(getCodec())

	ctx := client.Context{
		Keyring:      k,
		TxConfig:     MakeTestTxConfig(t),
		AddressCodec: address.NewBech32Codec("cosmos"),
	}

	type args struct {
		ctx      client.Context
		fromName string
		digest   string
		signMode apitxsigning.SignMode
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "signMode direct",
			args: args{
				ctx:      ctx,
				fromName: "direct",
				digest:   "Hello world!",
				signMode: apitxsigning.SignMode_SIGN_MODE_DIRECT,
			},
			wantErr: false,
		},
		{
			name: "signMode textual",
			args: args{
				ctx:      ctx,
				fromName: "textual",
				digest:   "Hello world!",
				signMode: apitxsigning.SignMode_SIGN_MODE_TEXTUAL,
			},
			wantErr: false,
		},
		{
			name: "signMode legacyAmino",
			args: args{
				ctx:      ctx,
				fromName: "legacy",
				digest:   "Hello world!",
				signMode: apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			},
			wantErr: false,
		},
		{
			name: "signMode direct aux",
			args: args{
				ctx:      ctx,
				fromName: "direct-aux",
				digest:   "Hello world!",
				signMode: apitxsigning.SignMode_SIGN_MODE_DIRECT_AUX,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := k.NewAccount(tt.args.fromName, mnemonic, tt.name, "m/44'/118'/0'/0/0", hd.Secp256k1)
			require.NoError(t, err)

			got, err := sign(tt.args.ctx, tt.args.fromName, tt.args.digest, tt.args.signMode)
			if !tt.wantErr {
				require.NoError(t, err)
				require.NotNil(t, got)
			} else {
				require.Error(t, err)
			}
		})
	}
}
