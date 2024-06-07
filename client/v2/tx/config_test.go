package tx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/codec/address"
)

type testModeHandler struct{}

func (t testModeHandler) Mode() apitxsigning.SignMode {
	return apitxsigning.SignMode_SIGN_MODE_DIRECT
}

func (t testModeHandler) GetSignBytes(_ context.Context, _ signing.SignerData, _ signing.TxData) ([]byte, error) {
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
			wantErr: true,
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
				CustomSignModes:       []signing.SignModeHandler{testModeHandler{}},
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
