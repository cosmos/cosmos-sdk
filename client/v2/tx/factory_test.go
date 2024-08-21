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
					address: addr,
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
				chainID: "demo",
				AccountConfig: AccountConfig{
					address: addr,
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
				chainID: "demo",
				AccountConfig: AccountConfig{
					address: addr,
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
			got, err := f.BuildUnsignedTx(tt.msgs...)
			if tt.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, got)
				builder, ok := got.(*txBuilder)
				require.True(t, ok)
				require.Nil(t, builder.signatures)
				require.Nil(t, builder.signerInfos)
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
				chainID: "demo",
				AccountConfig: AccountConfig{
					address: addr,
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
		{
			name: "offline mode",
			txParams: TxParameters{
				chainID: "demo",
				AccountConfig: AccountConfig{
					address: addr,
				},
				ExecutionOptions: ExecutionOptions{
					offline: true,
				},
			},
			msgs: []transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: signer,
					Count:  0,
				},
			},
			error: true,
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
				chainID: "demo",
				AccountConfig: AccountConfig{
					address: addr,
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
				chainID: "demo",
				AccountConfig: AccountConfig{
					address: addr,
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
				chainID: "demo",
				AccountConfig: AccountConfig{
					fromName: "alice",
					address:  addr,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f, err := NewFactory(setKeyring(), cdc, mockAccountRetriever{}, txConf, ac, mockClientConn{}, tt.txParams)
			require.NoError(t, err)
			require.NotNil(t, f)

			builder, err := f.BuildUnsignedTx([]transaction.Msg{
				&countertypes.MsgIncreaseCounter{
					Signer: signer,
					Count:  0,
				},
			}...)
			require.NoError(t, err)
			require.NotNil(t, builder)

			builderTx, ok := builder.(*txBuilder)
			require.True(t, ok)
			require.Nil(t, builderTx.signatures)
			require.Nil(t, builderTx.signerInfos)

			tx, err := f.sign(context.Background(), builder, true)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				sigs, err := tx.GetSignatures()
				require.NoError(t, err)
				require.NotNil(t, sigs)
				require.NotNil(t, builderTx.signerInfos)
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
				chainID:  "demo",
				signMode: apitxsigning.SignMode_SIGN_MODE_DIRECT,
				AccountConfig: AccountConfig{
					address: addr,
				},
			},
		},
		{
			name: "signMode not specified",
			txParams: TxParameters{
				chainID: "demo",
				AccountConfig: AccountConfig{
					address: addr,
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

			txb, err := f.BuildUnsignedTx([]transaction.Msg{
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
				ChainID:       f.txParams.chainID,
				AccountNumber: 0,
				Sequence:      0,
				PubKey: &anypb.Any{
					TypeUrl: codectypes.MsgTypeURL(pk),
					Value:   pk.Bytes(),
				},
			}

			got, err := f.getSignBytesAdapter(context.Background(), signerData, txb)
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
					address: addr,
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
					address: addr,
				},
			},
			withFunc: func(f *Factory) {
				f.WithSequence(10)
			},
			checkFunc: func(f *Factory) bool {
				return f.txParams.AccountConfig.sequence == 10
			},
		},
		{
			name: "with account number",
			txParams: TxParameters{
				AccountConfig: AccountConfig{
					address: addr,
				},
			},
			withFunc: func(f *Factory) {
				f.WithAccountNumber(123)
			},
			checkFunc: func(f *Factory) bool {
				return f.txParams.AccountConfig.accountNumber == 123
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
