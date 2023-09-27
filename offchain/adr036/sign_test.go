package adr036

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

func getCodec() codec.Codec {
	registry := codectypes.NewInterfaceRegistry()
	cryptocodec.RegisterInterfaces(registry)
	return codec.NewProtoCodec(registry)
}

func TestNewOfflineSigner(t *testing.T) {
	type args struct {
		txConfig client.TxConfig
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "offChainSigner constructor",
			args: args{
				txConfig: MakeTestTxConfig(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := keyring.NewInMemory(getCodec())
			_, err := k.NewAccount("uid", testdata.TestMnemonic, "", sdk.GetConfig().GetFullBIP44Path(), hd.Secp256k1)
			signer, err := newKeyRingWrapper("uid", k)
			require.NoError(t, err)
			got := NewOffChainSigner(signer, tt.args.txConfig)
			require.NotNil(t, got)
			require.NotNil(t, got.signer)
			require.NotNil(t, got.txConfig)
		})
	}
}

func TestOfflineSigner_Sign(t *testing.T) {
	type fields struct {
		txConfig client.TxConfig
	}
	type args struct {
		ctx      context.Context
		msgs     []sdk.Msg
		uid      string
		signMode signing.SignMode
	}
	tests := []struct {
		name      string
		fields    fields
		createKey func(k keyring.Keyring, uid string) (*keyring.Record, error)
		args      args
		want      authsigning.SigVerifiableTx
	}{
		{
			name: "DIRECT signing",
			fields: fields{
				txConfig: MakeTestTxConfig(),
			},
			createKey: func(k keyring.Keyring, uid string) (*keyring.Record, error) {
				return k.NewAccount(uid, testdata.TestMnemonic, "", sdk.GetConfig().GetFullBIP44Path(), hd.Secp256k1)
			},
			args: args{
				ctx: context.Background(),
				msgs: []sdk.Msg{
					&MsgSignArbitraryData{
						Signer: "cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h",
						Data:   []byte("Hello"),
					},
				},
				uid:      "correctSigning",
				signMode: signing.SignMode_SIGN_MODE_DIRECT,
			},
		},
		{
			name: "DIRECT_AUX signing",
			fields: fields{
				txConfig: MakeTestTxConfig(),
			},
			createKey: func(k keyring.Keyring, uid string) (*keyring.Record, error) {
				r, _, err := k.NewMnemonic(uid, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
				return r, err
			},
			args: args{
				ctx: context.Background(),
				msgs: []sdk.Msg{
					&MsgSignArbitraryData{
						Signer: "cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h",
						Data:   []byte("Hello"),
					},
				},
				uid:      "correctSigning",
				signMode: signing.SignMode_SIGN_MODE_DIRECT_AUX,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := keyring.NewInMemory(getCodec())
			_, err := tt.createKey(k, tt.args.uid)
			require.NoError(t, err)
			signer, err := newKeyRingWrapper(tt.args.uid, k)
			require.NoError(t, err)
			s := OffChainSigner{
				signer:   signer,
				txConfig: tt.fields.txConfig,
			}
			got, err := s.Sign(tt.args.ctx, tt.args.msgs, tt.args.uid, tt.args.signMode)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func Test_validateMsgs(t *testing.T) {
	type args struct {
		msgs []sdk.Msg
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid message",
			args: args{
				msgs: []sdk.Msg{
					&MsgSignArbitraryData{
						Signer: "cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h",
						Data:   []byte("Hello"),
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "no msgf",
			args:    args{msgs: []sdk.Msg{}},
			wantErr: true,
		},
		{
			name: "empty signer",
			args: args{
				msgs: []sdk.Msg{
					&MsgSignArbitraryData{
						Signer: "",
						Data:   []byte("Hello"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty signer",
			args: args{
				msgs: []sdk.Msg{
					&MsgSignArbitraryData{
						Signer: "",
						Data:   []byte("Hello"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid signer",
			args: args{
				msgs: []sdk.Msg{
					&MsgSignArbitraryData{
						Signer: "Imsigning",
						Data:   []byte("Hello"),
					},
				},
			},
			wantErr: true,
		},
		{
			name: "empty data",
			args: args{
				msgs: []sdk.Msg{
					&MsgSignArbitraryData{
						Signer: "cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h",
						Data:   []byte{},
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateMsgs(tt.args.msgs); (err != nil) != tt.wantErr {
				t.Errorf("validateMsgs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
