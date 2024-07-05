package tx

import (
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"testing"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/transaction"
	txdecode "cosmossdk.io/x/tx/decode"
	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	countertypes "github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
)

func TestNewBuilderProvider(t *testing.T) {
	type args struct {
		addressCodec address.Codec
		decoder      *txdecode.Decoder
		codec        codec.BinaryCodec
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "create new builder provider",
			args: args{
				addressCodec: ac,
				decoder:      decoder,
				codec:        cdc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewBuilderProvider(tt.args.addressCodec, tt.args.decoder, tt.args.codec)
			require.NotNil(t, got)
		})
	}
}

func TestBuilderProvider_NewTxBuilder(t *testing.T) {
	type fields struct {
		addressCodec address.Codec
		decoder      *txdecode.Decoder
		codec        codec.BinaryCodec
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "New txBuilder",
			fields: fields{
				addressCodec: ac,
				decoder:      decoder,
				codec:        cdc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := BuilderProvider{
				addressCodec: tt.fields.addressCodec,
				decoder:      tt.fields.decoder,
				codec:        tt.fields.codec,
			}
			got := b.NewTxBuilder()
			require.NotNil(t, got)
		})
	}
}

func Test_newTxBuilder(t *testing.T) {
	type args struct {
		addressCodec address.Codec
		decoder      *txdecode.Decoder
		codec        codec.BinaryCodec
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "new txBuilder",
			args: args{
				addressCodec: ac,
				decoder:      decoder,
				codec:        cdc,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newTxBuilder(tt.args.addressCodec, tt.args.decoder, tt.args.codec)
			require.NotNil(t, got)
			require.Equal(t, got.addressCodec, tt.args.addressCodec)
			require.Equal(t, got.decoder, tt.args.decoder)
			require.Equal(t, got.codec, tt.args.codec)
		})
	}
}

func Test_txBuilder_GetTx(t *testing.T) {
	tests := []struct {
		name        string
		txSetter    func() *txBuilder
		checkResult func(Tx)
	}{
		{
			name: "empty tx",
			txSetter: func() *txBuilder {
				return newTxBuilder(ac, decoder, cdc)
			},
			checkResult: func(tx Tx) {
				wTx, ok := tx.(*wrappedTx)
				require.True(t, ok)
				//require.Equal(t, []*anypb.Any(nil), wTx.Tx.Body.Messages)
				require.Nil(t, wTx.Tx.Body.Messages)
				require.Empty(t, wTx.Tx.Body.Memo)
				require.Equal(t, uint64(0), wTx.Tx.Body.TimeoutHeight)
				require.Equal(t, wTx.Tx.Body.Unordered, false)
				require.Nil(t, wTx.Tx.Body.ExtensionOptions)
				require.Nil(t, wTx.Tx.Body.NonCriticalExtensionOptions)

				require.Nil(t, wTx.Tx.AuthInfo.SignerInfos)
				require.Nil(t, wTx.Tx.AuthInfo.Fee.Amount)
				require.Equal(t, uint64(0), wTx.Tx.AuthInfo.Fee.GasLimit)
				require.Empty(t, wTx.Tx.AuthInfo.Fee.Payer)
				require.Empty(t, wTx.Tx.AuthInfo.Fee.Granter)

				require.Nil(t, wTx.Tx.Signatures)
			},
		},
		{
			name: "full tx",
			txSetter: func() *txBuilder {
				pk := secp256k1.GenPrivKey().PubKey()
				addr, _ := ac.BytesToString(pk.Address())
				b := newTxBuilder(ac, decoder, cdc)

				err := b.SetMsgs([]transaction.Msg{&countertypes.MsgIncreaseCounter{
					Signer: addr,
					Count:  0,
				}}...)
				require.NoError(t, err)

				err = b.SetFeePayer(addr)
				require.NoError(t, err)

				b.SetFeeAmount([]*base.Coin{{
					Denom:  "cosmos",
					Amount: "1000",
				}})

				err = b.SetSignatures([]Signature{{
					PubKey: pk,
					Data: &SingleSignatureData{
						SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
						Signature: nil,
					},
					Sequence: 0,
				}}...)
				require.NoError(t, err)
				return b
			},
			checkResult: func(tx Tx) {
				wTx, ok := tx.(*wrappedTx)
				require.True(t, ok)
				require.True(t, len(wTx.Tx.Body.Messages) == 1)

				require.NotNil(t, wTx.Tx.AuthInfo.SignerInfos)
				require.NotNil(t, wTx.Tx.AuthInfo.Fee.Amount)

				require.NotNil(t, wTx.Tx.Signatures)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.txSetter()
			got, err := b.GetTx()
			require.NoError(t, err)
			require.NotNil(t, got)
			tt.checkResult(got)
		})
	}
}

func Test_msgsV1toAnyV2(t *testing.T) {
	tests := []struct {
		name string
		msgs []transaction.Msg
	}{
		{
			name: "convert msgV1 to V2",
			msgs: []transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: "cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
					Count:  0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := msgsV1toAnyV2(tt.msgs)
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func Test_intoAnyV2(t *testing.T) {
	tests := []struct {
		name string
		msgs []*codectypes.Any
	}{
		{
			name: "any to v2",
			msgs: []*codectypes.Any{
				{
					TypeUrl: "/random/msg",
					Value:   []byte("random message"),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := intoAnyV2(tt.msgs)
			require.NotNil(t, got)
			require.Equal(t, len(got), len(tt.msgs))
			for i, msg := range got {
				require.Equal(t, msg.TypeUrl, tt.msgs[i].TypeUrl)
				require.Equal(t, msg.Value, tt.msgs[i].Value)
			}
		})
	}
}

func Test_txBuilder_getFee(t *testing.T) {
	tests := []struct {
		name       string
		feeAmount  []*base.Coin
		feeGranter string
		feePayer   string
	}{
		{
			name: "get fee with payer",
			feeAmount: []*base.Coin{
				{
					Denom:  "cosmos",
					Amount: "1000",
				},
			},
			feeGranter: "",
			feePayer:   "cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
		},
		{
			name: "get fee with granter",
			feeAmount: []*base.Coin{
				{
					Denom:  "cosmos",
					Amount: "1000",
				},
			},
			feeGranter: "cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
			feePayer:   "",
		},
		{
			name: "get fee with granter and granter",
			feeAmount: []*base.Coin{
				{
					Denom:  "cosmos",
					Amount: "1000",
				},
			},
			feeGranter: "cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
			feePayer:   "cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTxBuilder(ac, decoder, cdc)
			b.SetFeeAmount(tt.feeAmount)
			err := b.SetFeeGranter(tt.feeGranter)
			require.NoError(t, err)
			err = b.SetFeePayer(tt.feePayer)
			require.NoError(t, err)

			fee, err := b.getFee()
			require.NoError(t, err)
			require.NotNil(t, fee)

			require.Equal(t, fee.Amount, tt.feeAmount)
			require.Equal(t, fee.Granter, tt.feeGranter)
			require.Equal(t, fee.Payer, tt.feePayer)
		})
	}
}

func Test_txBuilder_GetSigningTxData(t *testing.T) {
	tests := []struct {
		name     string
		txSetter func() *txBuilder
	}{
		{
			name: "empty tx",
			txSetter: func() *txBuilder {
				return newTxBuilder(ac, decoder, cdc)
			},
		},
		{
			name: "full tx",
			txSetter: func() *txBuilder {
				pk := secp256k1.GenPrivKey().PubKey()
				addr, _ := ac.BytesToString(pk.Address())
				b := newTxBuilder(ac, decoder, cdc)

				err := b.SetMsgs([]transaction.Msg{&countertypes.MsgIncreaseCounter{
					Signer: addr,
					Count:  0,
				}}...)
				require.NoError(t, err)

				err = b.SetFeePayer(addr)
				require.NoError(t, err)

				b.SetFeeAmount([]*base.Coin{{
					Denom:  "cosmos",
					Amount: "1000",
				}})

				err = b.SetSignatures([]Signature{{
					PubKey: pk,
					Data: &SingleSignatureData{
						SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
						Signature: []byte("signature"),
					},
					Sequence: 0,
				}}...)
				require.NoError(t, err)

				return b
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.txSetter()
			got, err := b.GetSigningTxData()
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func Test_txBuilder_SetMsgs(t *testing.T) {
	tests := []struct {
		name    string
		msgs    []transaction.Msg
		wantErr bool
	}{
		{
			name: "set msgs",
			msgs: []transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: "cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
					Count:  0,
				},
				&countertypes.MsgIncreaseCounter{
					Signer: "cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9",
					Count:  1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTxBuilder(ac, decoder, cdc)
			err := b.SetMsgs(tt.msgs...)
			require.NoError(t, err)
			require.Equal(t, len(tt.msgs), len(tt.msgs))

			for i, msg := range tt.msgs {
				require.Equal(t, msg, tt.msgs[i])
			}
		})
	}
}

func Test_txBuilder_SetMemo(t *testing.T) {
	tests := []struct {
		name string
		memo string
	}{
		{
			name: "set memo",
			memo: "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTxBuilder(ac, decoder, cdc)
			b.SetMemo(tt.memo)
			require.Equal(t, b.memo, tt.memo)
		})
	}
}

func Test_txBuilder_SetFeeAmount(t *testing.T) {
	tests := []struct {
		name  string
		coins []*base.Coin
	}{
		{
			name: "set coins",
			coins: []*base.Coin{
				{
					Denom:  "cosmos",
					Amount: "1000",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTxBuilder(ac, decoder, cdc)
			b.SetFeeAmount(tt.coins)
			require.Equal(t, len(tt.coins), len(tt.coins))

			for i, coin := range tt.coins {
				require.Equal(t, coin.Amount, tt.coins[i].Amount)
			}
		})
	}
}

func Test_txBuilder_SetGasLimit(t *testing.T) {
	tests := []struct {
		name     string
		gasLimit uint64
	}{
		{
			name:     "set gas limit",
			gasLimit: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTxBuilder(ac, decoder, cdc)
			b.SetGasLimit(tt.gasLimit)
			require.Equal(t, b.gasLimit, tt.gasLimit)
		})
	}
}

func Test_txBuilder_SetUnordered(t *testing.T) {
	tests := []struct {
		name      string
		unordered bool
	}{
		{
			name:      "unordered",
			unordered: true,
		},
		{
			name:      "not unordered",
			unordered: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := newTxBuilder(ac, decoder, cdc)
			b.SetUnordered(tt.unordered)
			require.Equal(t, b.unordered, tt.unordered)
		})
	}
}

func Test_txBuilder_SetSignatures(t *testing.T) {
	tests := []struct {
		name       string
		signatures func() []Signature
	}{
		{
			name: "set empty single signature",
			signatures: func() []Signature {
				return []Signature{{
					PubKey: secp256k1.GenPrivKey().PubKey(),
					Data: &SingleSignatureData{
						SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
						Signature: nil,
					},
					Sequence: 0,
				}}
			},
		},
		{
			name: "set single signature",
			signatures: func() []Signature {
				return []Signature{{
					PubKey: secp256k1.GenPrivKey().PubKey(),
					Data: &SingleSignatureData{
						SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
						Signature: []byte("signature"),
					},
					Sequence: 0,
				}}
			},
		},
		{
			name: "set empty multi signature",
			signatures: func() []Signature {
				return []Signature{{
					PubKey: multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{secp256k1.GenPrivKey().PubKey()}),
					Data: &MultiSignatureData{
						BitArray: nil,
						Signatures: []SignatureData{
							&SingleSignatureData{
								SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
								Signature: nil,
							},
						},
					},
					Sequence: 0,
				}}
			},
		},
		{
			name: "set multi signature",
			signatures: func() []Signature {
				return []Signature{{
					PubKey: multisig.NewLegacyAminoPubKey(1, []cryptotypes.PubKey{secp256k1.GenPrivKey().PubKey()}),
					Data: &MultiSignatureData{
						BitArray: nil,
						Signatures: []SignatureData{
							&SingleSignatureData{
								SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
								Signature: []byte("signature"),
							},
						},
					},
					Sequence: 0,
				}}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cryptocodec.RegisterInterfaces(cdc.InterfaceRegistry())
			b := newTxBuilder(ac, decoder, cdc)
			sigs := tt.signatures()
			err := b.SetSignatures(sigs...)
			require.NoError(t, err)
			tx, err := b.GetTx()
			require.NoError(t, err)
			signatures, err := tx.GetSignatures()
			require.NoError(t, err)
			require.Equal(t, len(sigs), len(signatures))
		})
	}
}
