package offchain

import (
	"testing"

	"github.com/stretchr/testify/require"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

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
		TxConfig:     newTestConfig(t),
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
