package adr036

import (
	"context"
	"testing"

	"github.com/cosmos/gogoproto/proto"
	"github.com/stretchr/testify/require"

	sig "cosmossdk.io/x/tx/signing"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/address"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
)

func MakeTestTxConfig() client.TxConfig {
	accAddressPrefix := "cosmos"
	valAddressPrefix := "cosmosvaloper"
	enabledSignModes := []signingtypes.SignMode{
		signingtypes.SignMode_SIGN_MODE_DIRECT,
		signingtypes.SignMode_SIGN_MODE_DIRECT_AUX,
		signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
	}
	//txConfigOpts := tx.ConfigOptions{
	//	EnabledSignModes: enabledSignModes,
	//}
	ir, err := codectypes.NewInterfaceRegistryWithOptions(codectypes.InterfaceRegistryOptions{
		ProtoFiles: proto.HybridResolver,
		SigningOptions: sig.Options{
			AddressCodec:          address.NewBech32Codec(accAddressPrefix),
			ValidatorAddressCodec: address.NewBech32Codec(valAddressPrefix),
		},
	})
	if err != nil {
		panic(err)
	}
	RegisterInterfaces(ir)
	cdc := codec.NewProtoCodec(ir)
	txConfig := tx.NewTxConfig(cdc, enabledSignModes)
	return txConfig
}

func TestOffChainVerifier_Verify(t *testing.T) {
	type fields struct {
		txConfig client.TxConfig
	}
	tests := []struct {
		name      string
		fields    fields
		createKey func(k keyring.Keyring, uid string) (*keyring.Record, error)
		msgs      []sdk.Msg
		signMode  signingtypes.SignMode
		wantErr   bool
	}{
		{
			name:   "Verify",
			fields: fields{txConfig: MakeTestTxConfig()},
			createKey: func(k keyring.Keyring, uid string) (*keyring.Record, error) {
				return k.NewAccount(uid, testdata.TestMnemonic, "", sdk.GetConfig().GetFullBIP44Path(), hd.Secp256k1)
			},
			msgs: []sdk.Msg{
				&MsgSignArbitraryData{
					Signer: "cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h", // TestMnemonic address
					Data:   []byte("Hello"),
				},
			},
			signMode: signingtypes.SignMode_SIGN_MODE_DIRECT,
		},
		{
			name:   "signed with different Signer",
			fields: fields{txConfig: MakeTestTxConfig()},
			createKey: func(k keyring.Keyring, uid string) (*keyring.Record, error) {
				r, _, err := k.NewMnemonic(uid, keyring.English, sdk.FullFundraiserPath, keyring.DefaultBIP39Passphrase, hd.Secp256k1)
				return r, err
			},
			msgs: []sdk.Msg{
				&MsgSignArbitraryData{
					Signer: "cosmos1w34k53py5v5xyluazqpq65agyajavep2rflq6h", // TestMnemonic address
					Data:   []byte("Hello"),
				},
			},
			signMode: signingtypes.SignMode_SIGN_MODE_DIRECT,
			wantErr:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			k := keyring.NewInMemory(getCodec())
			_, err := tt.createKey(k, "verifyKey")
			require.NoError(t, err)
			signer, err := newKeyRingWrapper("verifyKey", k)
			require.NoError(t, err)
			v := OffChainVerifier{
				txConfig: tt.fields.txConfig,
			}
			s := OffChainSigner{
				signer:   signer,
				txConfig: tt.fields.txConfig,
			}
			signedTx, err := s.Sign(context.Background(), tt.msgs, tt.signMode)
			require.NoError(t, err)
			require.NotNil(t, signedTx)
			if err := v.Verify(context.Background(), signedTx); (err != nil) != tt.wantErr {
				t.Errorf("Verify() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
