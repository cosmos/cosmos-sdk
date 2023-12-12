package offchain

import (
	"testing"


	"github.com/stretchr/testify/require"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func Test_verify(t *testing.T) {
	k := keyring.NewInMemory(getCodec())
	tests := []struct {
		name     string
		fromName string
		digest   string
		signMode apitxsigning.SignMode
		ctx      client.Context
		wantErr  bool
	}{
		{
			name:     "signMode direct",
			fromName: "directKey",
			digest:   "hello world",
			signMode: apitxsigning.SignMode_SIGN_MODE_DIRECT,
			ctx: client.Context{
				Keyring:      k,
				TxConfig:     MakeTestTxConfig(),
				Codec:        getCodec(),
				AddressCodec: address.NewBech32Codec("cosmos"),
			},
		},
		{
			name:     "signMode textual",
			fromName: "textualKey",
			digest:   "hello world",
			signMode: apitxsigning.SignMode_SIGN_MODE_TEXTUAL,
			ctx: client.Context{
				Keyring:      k,
				TxConfig:     MakeTestTxConfig(),
				Codec:        getCodec(),
				AddressCodec: address.NewBech32Codec("cosmos"),
			},
		},
		{
			name:     "signMode legacyAmino",
			fromName: "textualKey",
			digest:   "hello world",
			signMode: apitxsigning.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
			ctx: client.Context{
				Keyring:      k,
				TxConfig:     MakeTestTxConfig(),
				Codec:        getCodec(),
				AddressCodec: address.NewBech32Codec("cosmos"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := k.NewAccount(tt.fromName, testdata.TestMnemonic, tt.name, "m/44'/118'/0'/0/0", hd.Secp256k1)
			require.NoError(t, err)

			tx, err := sign(tt.ctx, tt.fromName, tt.digest, tt.signMode)
			require.NoError(t, err)

			err = verify(tt.ctx, tx)
			require.NoError(t, err)
		})
	}
}
