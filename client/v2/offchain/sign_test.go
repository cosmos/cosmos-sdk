package offchain

import (
	"testing"

	"github.com/stretchr/testify/require"

	clientcontext "cosmossdk.io/client/v2/context"

	"github.com/cosmos/cosmos-sdk/codec/address"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func TestSign(t *testing.T) {
	ac := address.NewBech32Codec("cosmos")
	vc := address.NewBech32Codec("cosmosvaloper")
	k := keyring.NewInMemory(getCodec())
	_, err := k.NewAccount("signVerify", mnemonic, "", "m/44'/118'/0'/0/0", hd.Secp256k1)
	require.NoError(t, err)

	autoKeyring, err := keyring.NewAutoCLIKeyring(k, ac)
	require.NoError(t, err)

	ctx := clientcontext.Context{
		AddressCodec:          ac,
		ValidatorAddressCodec: vc,
		Cdc:                   getCodec(),
		Keyring:               autoKeyring,
	}
	tests := []struct {
		name     string
		rawBytes []byte
		encoding string
		signMode string
		wantErr  bool
	}{
		{
			name:     "sign direct",
			rawBytes: []byte("hello world"),
			encoding: noEncoder,
			signMode: "direct",
		},
		{
			name:     "sign amino",
			rawBytes: []byte("hello world"),
			encoding: noEncoder,
			signMode: "amino-json",
		},
		{
			name:     "not supported sign mode",
			rawBytes: []byte("hello world"),
			encoding: noEncoder,
			signMode: "textual",
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Sign(ctx, tt.rawBytes, mockClientConn{}, "signVerify", tt.encoding, tt.signMode, "json")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}
