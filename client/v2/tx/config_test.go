package tx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apicrypto "cosmossdk.io/api/cosmos/crypto/multisig/v1beta1"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	codec2 "github.com/cosmos/cosmos-sdk/crypto/codec"
	kmultisig "github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
)

type mockModeHandler struct{}

func (t mockModeHandler) Mode() apitxsigning.SignMode {
	return apitxsigning.SignMode_SIGN_MODE_DIRECT
}

func (t mockModeHandler) GetSignBytes(_ context.Context, _ signing.SignerData, _ signing.TxData) ([]byte, error) {
	return []byte{}, nil
}

func TestConfigOptions_validate(t *testing.T) {
	tests := []struct {
		name    string
		opts    ConfigOptions
		wantErr bool
	}{
		{
			name: "valid options",
			opts: ConfigOptions{
				AddressCodec:          address.NewBech32Codec("cosmos"),
				Decoder:               decoder,
				Cdc:                   cdc,
				ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
			},
		},
		{
			name: "missing address codec",
			opts: ConfigOptions{
				Decoder:               decoder,
				Cdc:                   cdc,
				ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
			},
			wantErr: true,
		},
		{
			name: "missing decoder",
			opts: ConfigOptions{
				AddressCodec:          address.NewBech32Codec("cosmos"),
				Cdc:                   cdc,
				ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
			},
		},
		{
			name: "missing codec",
			opts: ConfigOptions{
				AddressCodec:          address.NewBech32Codec("cosmos"),
				Decoder:               decoder,
				ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
			},
			wantErr: true,
		},
		{
			name: "missing validator address codec",
			opts: ConfigOptions{
				AddressCodec: address.NewBech32Codec("cosmos"),
				Decoder:      decoder,
				Cdc:          cdc,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.opts.validate(); (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_newHandlerMap(t *testing.T) {
	tests := []struct {
		name string
		opts ConfigOptions
	}{
		{
			name: "handler map with default sign modes",
			opts: ConfigOptions{
				AddressCodec:          address.NewBech32Codec("cosmos"),
				Decoder:               decoder,
				Cdc:                   cdc,
				ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
			},
		},
		{
			name: "handler map with just one sign mode",
			opts: ConfigOptions{
				AddressCodec:          address.NewBech32Codec("cosmos"),
				Decoder:               decoder,
				Cdc:                   cdc,
				ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
				EnablesSignModes:      []apitxsigning.SignMode{apitxsigning.SignMode_SIGN_MODE_DIRECT},
			},
		},
		{
			name: "handler map with custom sign modes",
			opts: ConfigOptions{
				AddressCodec:          address.NewBech32Codec("cosmos"),
				Decoder:               decoder,
				Cdc:                   cdc,
				ValidatorAddressCodec: address.NewBech32Codec("cosmosvaloper"),
				CustomSignModes:       []signing.SignModeHandler{mockModeHandler{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.opts.validate()
			require.NoError(t, err)

			signingCtx, err := newSigningContext(tt.opts)
			require.NoError(t, err)

			handlerMap, err := newHandlerMap(tt.opts, signingCtx)
			require.NoError(t, err)
			require.NotNil(t, handlerMap)
			require.Equal(t, len(handlerMap.SupportedModes()), len(tt.opts.EnablesSignModes)+len(tt.opts.CustomSignModes))
		})
	}
}

func TestNewTxConfig(t *testing.T) {
	tests := []struct {
		name    string
		options ConfigOptions
		wantErr bool
	}{
		{
			name: "valid options",
			options: ConfigOptions{
				AddressCodec:          ac,
				Cdc:                   cdc,
				ValidatorAddressCodec: valCodec,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewTxConfig(tt.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewTxConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.NotNil(t, got)
		})
	}
}

func Test_defaultTxSigningConfig_MarshalSignatureJSON(t *testing.T) {
	tests := []struct {
		name       string
		options    ConfigOptions
		signatures func(t *testing.T) []Signature
	}{
		{
			name: "single signature",
			options: ConfigOptions{
				AddressCodec:          ac,
				Cdc:                   cdc,
				ValidatorAddressCodec: valCodec,
			},
			signatures: func(t *testing.T) []Signature {
				t.Helper()

				k := setKeyring()
				pk, err := k.GetPubKey("alice")
				require.NoError(t, err)
				signature, err := k.Sign("alice", make([]byte, 10), apitxsigning.SignMode_SIGN_MODE_DIRECT)
				require.NoError(t, err)
				return []Signature{
					{
						PubKey: pk,
						Data: &SingleSignatureData{
							SignMode:  apitxsigning.SignMode_SIGN_MODE_DIRECT,
							Signature: signature,
						},
					},
				}
			},
		},
		{
			name: "multisig signatures",
			options: ConfigOptions{
				AddressCodec:          ac,
				Cdc:                   cdc,
				ValidatorAddressCodec: valCodec,
			},
			signatures: func(t *testing.T) []Signature {
				t.Helper()

				n := 2
				pubKeys := make([]cryptotypes.PubKey, n)
				sigs := make([]SignatureData, n)
				for i := 0; i < n; i++ {
					sk := secp256k1.GenPrivKey()
					pubKeys[i] = sk.PubKey()
					msg, err := sk.Sign(make([]byte, 10))
					require.NoError(t, err)
					sigs[i] = &SingleSignatureData{
						SignMode:  apitxsigning.SignMode_SIGN_MODE_DIRECT,
						Signature: msg,
					}
				}
				bitArray := cryptotypes.NewCompactBitArray(n)
				mKey := kmultisig.NewLegacyAminoPubKey(n, pubKeys)
				return []Signature{
					{
						PubKey: mKey,
						Data: &MultiSignatureData{
							BitArray: &apicrypto.CompactBitArray{
								ExtraBitsStored: bitArray.ExtraBitsStored,
								Elems:           bitArray.Elems,
							},
							Signatures: sigs,
						},
					},
				}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewTxConfig(tt.options)
			require.NoError(t, err)

			got, err := config.MarshalSignatureJSON(tt.signatures(t))
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func Test_defaultTxSigningConfig_UnmarshalSignatureJSON(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	codec2.RegisterInterfaces(registry)
	cdc := codec.NewProtoCodec(registry)
	tests := []struct {
		name    string
		options ConfigOptions
		bz      []byte
	}{
		{
			name: "single signature",
			options: ConfigOptions{
				AddressCodec:          ac,
				Cdc:                   cdc,
				ValidatorAddressCodec: valCodec,
			},
			bz: []byte(`{"signatures":[{"public_key":{"@type":"/cosmos.crypto.secp256k1.PubKey", "key":"A0/vnNfExjWI07A/61KBudIyy6NNbz1xruWSEf+/4f6H"}, "data":{"single":{"mode":"SIGN_MODE_DIRECT", "signature":"usUTJwdc4PWPuox0Y0G/RuHoxyj+QpUcBGvXyNdDX1FOdoVj0tg4TGKT2NnM3QP6wCNbubjHuMOhTtqfW8SkYg=="}}}]}`),
		},
		{
			name: "multisig signatures",
			options: ConfigOptions{
				AddressCodec:          ac,
				Cdc:                   cdc,
				ValidatorAddressCodec: valCodec,
			},
			bz: []byte(`{"signatures":[{"public_key":{"@type":"/cosmos.crypto.multisig.LegacyAminoPubKey","threshold":2,"public_keys":[{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"A4Bs9huvS/COpZNhVhTnhgc8YR6VrSQ8hLQIHgnA+m3w"},{"@type":"/cosmos.crypto.secp256k1.PubKey","key":"AuNz2lFkLn3sKNjC5r4OWhgkWg5DZpGUiR9OdpzXspnp"}]},"data":{"multi":{"bitarray":{"extra_bits_stored":2,"elems":"AA=="},"signatures":[{"single":{"mode":"SIGN_MODE_DIRECT","signature":"vng4IlPzLH3fDFpikM5y1SfXFGny4BcLGwIFU0Ty4yoWjIxjTS4m6fgDB61sxEkV5DK/CD7gUwenGuEpzJ2IGw=="}},{"single":{"mode":"SIGN_MODE_DIRECT","signature":"2dsGmr13bq/mPxbk9AgqcFpuvk4beszWu6uxkx+EhTMdVGp4J8FtjZc8xs/Pp3oTWY4ScAORYQHxwqN4qwMXGg=="}}]}}}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewTxConfig(tt.options)
			require.NoError(t, err)

			got, err := config.UnmarshalSignatureJSON(tt.bz)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}
