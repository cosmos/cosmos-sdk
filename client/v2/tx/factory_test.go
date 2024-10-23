package tx

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"

	base "cosmossdk.io/api/cosmos/base/v1beta1"
	apitxsigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	"cosmossdk.io/core/transaction"
	"cosmossdk.io/x/tx/signing"

	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/multisig"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	countertypes "github.com/cosmos/cosmos-sdk/testutil/x/counter/types"
)

var (
	signer  = "cosmos1zglwfu6xjzvzagqcmvzewyzjp9xwqw5qwrr8n9"
	addr, _ = ac.StringToBytes(signer)
)

func TestFactory_prepareTxParams(t *testing.T) {
	tests := []struct {
		name     string
		txParams TxParameters
		error    bool
	}{
		{
			name: "no error",
			txParams: TxParameters{
				AccountConfig: AccountConfig{
					Address: addr,
				},
			},
		},
		{
			name:     "without account",
			txParams: TxParameters{},
			error:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			tt.txParams, err = prepareTxParams(tt.txParams, mockAccountRetriever{}, false)
			if (err != nil) != tt.error {
				t.Errorf("Prepare() error = %v, wantErr %v", err, tt.error)
			}
		})
	}
}

func TestFactory_BuildUnsignedTx(t *testing.T) {
	tests := []struct {
		name     string
		txParams TxParameters
		msgs     []transaction.Msg
		error    bool
	}{
		{
			name: "no error",
			txParams: TxParameters{
				ChainID: "demo",
				AccountConfig: AccountConfig{
					Address: addr,
				},
			},
			msgs: []transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: signer,
					Count:  0,
				},
			},
		},
		{
			name: "fees and gas price provided",
			txParams: TxParameters{
				ChainID: "demo",
				AccountConfig: AccountConfig{
					Address: addr,
				},
				GasConfig: GasConfig{
					gasPrices: []*base.DecCoin{
						{
							Amount: "1000",
							Denom:  "stake",
						},
					},
				},
				FeeConfig: FeeConfig{
					fees: []*base.Coin{
						{
							Amount: "1000",
							Denom:  "stake",
						},
					},
				},
			},
			msgs:  []transaction.Msg{},
			error: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, tt.txParams)
			require.NoError(t, err)
			require.NotNil(t, f)
			err = f.BuildUnsignedTx(tt.msgs...)
			if tt.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Nil(t, f.tx.signatures)
				require.Nil(t, f.tx.signerInfos)
			}
		})
	}
}

func TestFactory_calculateGas(t *testing.T) {
	tests := []struct {
		name     string
		txParams TxParameters
		msgs     []transaction.Msg
		error    bool
	}{
		{
			name: "no error",
			txParams: TxParameters{
				ChainID: "demo",
				AccountConfig: AccountConfig{
					Address: addr,
				},
				GasConfig: GasConfig{
					gasAdjustment: 1,
				},
			},
			msgs: []transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: signer,
					Count:  0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, tt.txParams)
			require.NoError(t, err)
			require.NotNil(t, f)
			err = f.calculateGas(tt.msgs...)
			if tt.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotZero(t, f.txParams.GasConfig)
			}
		})
	}
}

func TestFactory_Simulate(t *testing.T) {
	tests := []struct {
		name     string
		txParams TxParameters
		msgs     []transaction.Msg
		error    bool
	}{
		{
			name: "no error",
			txParams: TxParameters{
				ChainID: "demo",
				AccountConfig: AccountConfig{
					Address: addr,
				},
				GasConfig: GasConfig{
					gasAdjustment: 1,
				},
			},
			msgs: []transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: signer,
					Count:  0,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, tt.txParams)
			require.NoError(t, err)
			require.NotNil(t, f)
			got, got1, err := f.Simulate(tt.msgs...)
			if tt.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				require.NotZero(t, got1)
			}
		})
	}
}

func TestFactory_BuildSimTx(t *testing.T) {
	tests := []struct {
		name     string
		txParams TxParameters
		msgs     []transaction.Msg
		want     []byte
		error    bool
	}{
		{
			name: "no error",
			txParams: TxParameters{
				ChainID: "demo",
				AccountConfig: AccountConfig{
					Address: addr,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, tt.txParams)
			require.NoError(t, err)
			require.NotNil(t, f)
			got, err := f.BuildSimTx(tt.msgs...)
			if tt.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func TestFactory_Sign(t *testing.T) {
	tests := []struct {
		name     string
		txParams TxParameters
		wantErr  bool
	}{
		{
			name: "no error",
			txParams: TxParameters{
				ChainID: "demo",
				AccountConfig: AccountConfig{
					FromName: "alice",
					Address:  addr,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(setKeyring(), cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, tt.txParams)
			require.NoError(t, err)
			require.NotNil(t, f)

			err = f.BuildUnsignedTx([]transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: signer,
					Count:  0,
				},
			}...)
			require.NoError(t, err)

			require.Nil(t, f.tx.signatures)
			require.Nil(t, f.tx.signerInfos)

			tx, err := f.sign(context.Background(), true)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				sigs, err := tx.GetSignatures()
				require.NoError(t, err)
				require.NotNil(t, sigs)
				require.NotNil(t, f.tx.signerInfos)
			}
		})
	}
}

func TestFactory_getSignBytesAdapter(t *testing.T) {
	tests := []struct {
		name     string
		txParams TxParameters
		error    bool
	}{
		{
			name: "no error",
			txParams: TxParameters{
				ChainID:  "demo",
				SignMode: apitxsigning.SignMode_SIGN_MODE_DIRECT,
				AccountConfig: AccountConfig{
					Address: addr,
				},
			},
		},
		{
			name: "signMode not specified",
			txParams: TxParameters{
				ChainID: "demo",
				AccountConfig: AccountConfig{
					Address: addr,
				},
			},
			error: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(setKeyring(), cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, tt.txParams)
			require.NoError(t, err)
			require.NotNil(t, f)

			err = f.BuildUnsignedTx([]transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: signer,
					Count:  0,
				},
			}...)
			require.NoError(t, err)

			pk, err := f.keybase.GetPubKey("alice")
			require.NoError(t, err)
			require.NotNil(t, pk)

			addr, err := f.ac.BytesToString(pk.Address())
			require.NoError(t, err)
			require.NotNil(t, addr)

			signerData := signing.SignerData{
				Address:       addr,
				ChainID:       f.txParams.ChainID,
				AccountNumber: 0,
				Sequence:      0,
				PubKey: &anypb.Any{
					TypeUrl: codectypes.MsgTypeURL(pk),
					Value:   pk.Bytes(),
				},
			}

			got, err := f.getSignBytesAdapter(context.Background(), signerData)
			if tt.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
			}
		})
	}
}

func Test_validateMemo(t *testing.T) {
	tests := []struct {
		name    string
		memo    string
		wantErr bool
	}{
		{
			name: "empty memo",
			memo: "",
		},
		{
			name: "valid memo",
			memo: "11245",
		},
		{
			name:    "invalid Memo",
			memo:    "echo echo echo echo echo echo echo echo echo echo echo echo echo echo echo",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateMemo(tt.memo); (err != nil) != tt.wantErr {
				t.Errorf("validateMemo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFactory_WithFunctions(t *testing.T) {
	tests := []struct {
		name      string
		txParams  TxParameters
		withFunc  func(*Factory)
		checkFunc func(*Factory) bool
	}{
		{
			name: "with gas",
			txParams: TxParameters{
				AccountConfig: AccountConfig{
					Address: addr,
				},
			},
			withFunc: func(f *Factory) {
				f.WithGas(1000)
			},
			checkFunc: func(f *Factory) bool {
				return f.txParams.GasConfig.gas == 1000
			},
		},
		{
			name: "with sequence",
			txParams: TxParameters{
				AccountConfig: AccountConfig{
					Address: addr,
				},
			},
			withFunc: func(f *Factory) {
				f.WithSequence(10)
			},
			checkFunc: func(f *Factory) bool {
				return f.txParams.AccountConfig.Sequence == 10
			},
		},
		{
			name: "with account number",
			txParams: TxParameters{
				AccountConfig: AccountConfig{
					Address: addr,
				},
			},
			withFunc: func(f *Factory) {
				f.WithAccountNumber(123)
			},
			checkFunc: func(f *Factory) bool {
				return f.txParams.AccountConfig.AccountNumber == 123
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(setKeyring(), cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, tt.txParams)
			require.NoError(t, err)
			require.NotNil(t, f)

			tt.withFunc(&f)
			require.True(t, tt.checkFunc(&f))
		})
	}
}

func TestFactory_getTx(t *testing.T) {
	tests := []struct {
		name        string
		txSetter    func(f *Factory)
		checkResult func(Tx)
	}{
		{
			name: "empty tx",
			txSetter: func(f *Factory) {
			},
			checkResult: func(tx Tx) {
				wTx, ok := tx.(*wrappedTx)
				require.True(t, ok)
				// require.Equal(t, []*anypb.Any(nil), wTx.Tx.Body.Messages)
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
			txSetter: func(f *Factory) {
				pk := secp256k1.GenPrivKey().PubKey()
				addr, _ := f.ac.BytesToString(pk.Address())

				f.tx.msgs = []transaction.Msg{&countertypes.MsgIncreaseCounter{
					Signer: addr,
					Count:  0,
				}}

				err := f.setFeePayer(addr)
				require.NoError(t, err)

				f.tx.fees = []*base.Coin{{
					Denom:  "cosmos",
					Amount: "1000",
				}}

				err = f.setSignatures([]Signature{{
					PubKey: pk,
					Data: &SingleSignatureData{
						SignMode:  apitxsigning.SignMode_SIGN_MODE_DIRECT,
						Signature: nil,
					},
					Sequence: 0,
				}}...)
				require.NoError(t, err)
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
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			tt.txSetter(&f)
			got, err := f.getTx()
			require.NoError(t, err)
			require.NotNil(t, got)
			tt.checkResult(got)
		})
	}
}

func TestFactory_getFee(t *testing.T) {
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
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			f.tx.fees = tt.feeAmount
			err = f.setFeeGranter(tt.feeGranter)
			require.NoError(t, err)
			err = f.setFeePayer(tt.feePayer)
			require.NoError(t, err)

			fee, err := f.getFee()
			require.NoError(t, err)
			require.NotNil(t, fee)

			require.Equal(t, fee.Amount, tt.feeAmount)
			require.Equal(t, fee.Granter, tt.feeGranter)
			require.Equal(t, fee.Payer, tt.feePayer)
		})
	}
}

func TestFactory_getSigningTxData(t *testing.T) {
	tests := []struct {
		name     string
		txSetter func(f *Factory)
	}{
		{
			name:     "empty tx",
			txSetter: func(f *Factory) {},
		},
		{
			name: "full tx",
			txSetter: func(f *Factory) {
				pk := secp256k1.GenPrivKey().PubKey()
				addr, _ := ac.BytesToString(pk.Address())

				f.tx.msgs = []transaction.Msg{&countertypes.MsgIncreaseCounter{
					Signer: addr,
					Count:  0,
				}}

				err := f.setFeePayer(addr)
				require.NoError(t, err)

				f.tx.fees = []*base.Coin{{
					Denom:  "cosmos",
					Amount: "1000",
				}}

				err = f.setSignatures([]Signature{{
					PubKey: pk,
					Data: &SingleSignatureData{
						SignMode:  apitxsigning.SignMode_SIGN_MODE_DIRECT,
						Signature: []byte("signature"),
					},
					Sequence: 0,
				}}...)
				require.NoError(t, err)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			tt.txSetter(&f)
			got, err := f.getSigningTxData()
			require.NoError(t, err)
			require.NotNil(t, got)
		})
	}
}

func TestFactoryr_setMsgs(t *testing.T) {
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
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			f.tx.msgs = tt.msgs
			require.NoError(t, err)
			require.Equal(t, len(tt.msgs), len(f.tx.msgs))

			for i, msg := range tt.msgs {
				require.Equal(t, msg, f.tx.msgs[i])
			}
		})
	}
}

func TestFactory_SetMemo(t *testing.T) {
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
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			f.tx.memo = tt.memo
			require.Equal(t, f.tx.memo, tt.memo)
		})
	}
}

func TestFactory_SetFeeAmount(t *testing.T) {
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
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			f.tx.fees = tt.coins
			require.Equal(t, len(tt.coins), len(f.tx.fees))

			for i, coin := range tt.coins {
				require.Equal(t, coin.Amount, f.tx.fees[i].Amount)
			}
		})
	}
}

func TestFactory_SetGasLimit(t *testing.T) {
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
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			f.tx.gasLimit = tt.gasLimit
			require.Equal(t, f.tx.gasLimit, tt.gasLimit)
		})
	}
}

func TestFactory_SetUnordered(t *testing.T) {
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
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			f.tx.unordered = tt.unordered
			require.Equal(t, f.tx.unordered, tt.unordered)
		})
	}
}

func TestFactory_setSignatures(t *testing.T) {
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
						SignMode:  apitxsigning.SignMode_SIGN_MODE_DIRECT,
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
						SignMode:  apitxsigning.SignMode_SIGN_MODE_DIRECT,
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
								SignMode:  apitxsigning.SignMode_SIGN_MODE_DIRECT,
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
								SignMode:  apitxsigning.SignMode_SIGN_MODE_DIRECT,
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
			f, err := NewFactory(keybase, cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, TxParameters{})
			require.NoError(t, err)
			sigs := tt.signatures()
			err = f.setSignatures(sigs...)
			require.NoError(t, err)
			tx, err := f.getTx()
			require.NoError(t, err)
			signatures, err := tx.GetSignatures()
			require.NoError(t, err)
			require.Equal(t, len(sigs), len(signatures))
			for i := range signatures {
				require.Equal(t, sigs[i].PubKey, signatures[i].PubKey)
			}
		})
	}
}

///////////////////////

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
