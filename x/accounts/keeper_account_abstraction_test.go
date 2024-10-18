package accounts

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	aa_interface_v1 "cosmossdk.io/x/accounts/interfaces/account_abstraction/v1"
	txdecode "cosmossdk.io/x/tx/decode"
)

func TestVerifyAndExtractAaXtFromTx(t *testing.T) {
	currentTime := time.Now()
	currentBlock := uint64(1000)

	validXt := &aa_interface_v1.TxExtension{
		AuthenticationGasLimit: 100,
		BundlerPaymentMessages: nil,
		BundlerPaymentGasLimit: 0,
		ExecutionGasLimit:      1000,
	}
	validXtBytes, err := validXt.Marshal()
	require.NoError(t, err)

	tests := []struct {
		name         string
		bundledTx    *txdecode.DecodedTx
		currentBlock uint64
		currentTime  time.Time
		wantExt      *aa_interface_v1.TxExtension
		wantErr      string
	}{
		{
			name: "Valid transaction",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1")},
				Tx: &txv1beta1.Tx{
					AuthInfo: &txv1beta1.AuthInfo{
						SignerInfos: []*txv1beta1.SignerInfo{
							{ModeInfo: &txv1beta1.ModeInfo{Sum: &txv1beta1.ModeInfo_Single_{
								Single: &txv1beta1.ModeInfo_Single{Mode: 1},
							}}},
						},
					},
					Body: &txv1beta1.TxBody{
						ExtensionOptions: []*anypb.Any{
							{TypeUrl: aaXtName, Value: validXtBytes},
						},
					},
				},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      validXt,
			wantErr:      "",
		},
		{
			name: "Multiple signers",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1"), []byte("signer2")},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      nil,
			wantErr:      "account abstraction bundled txs can only have one signer, got: 2",
		},
		{
			name: "Multiple signer infos",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1")},
				Tx: &txv1beta1.Tx{
					AuthInfo: &txv1beta1.AuthInfo{
						SignerInfos: []*txv1beta1.SignerInfo{{}, {}},
					},
				},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      nil,
			wantErr:      "account abstraction tx must have one signer info",
		},
		{
			name: "Invalid mode info",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1")},
				Tx: &txv1beta1.Tx{
					AuthInfo: &txv1beta1.AuthInfo{
						SignerInfos: []*txv1beta1.SignerInfo{
							{ModeInfo: &txv1beta1.ModeInfo{}},
						},
					},
				},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      nil,
			wantErr:      "account abstraction mode info must be single",
		},
		{
			name: "Fee set",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1")},
				Tx: &txv1beta1.Tx{
					AuthInfo: &txv1beta1.AuthInfo{
						SignerInfos: []*txv1beta1.SignerInfo{
							{ModeInfo: &txv1beta1.ModeInfo{Sum: &txv1beta1.ModeInfo_Single_{
								Single: &txv1beta1.ModeInfo_Single{Mode: 1},
							}}},
						},
						Fee: &txv1beta1.Fee{},
					},
				},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      nil,
			wantErr:      "account abstraction tx must not have the Fee field set",
		},
		{
			name: "Timeout timestamp exceeded",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1")},
				Tx: &txv1beta1.Tx{
					AuthInfo: &txv1beta1.AuthInfo{
						SignerInfos: []*txv1beta1.SignerInfo{
							{ModeInfo: &txv1beta1.ModeInfo{Sum: &txv1beta1.ModeInfo_Single_{
								Single: &txv1beta1.ModeInfo_Single{Mode: 1},
							}}},
						},
					},
					Body: &txv1beta1.TxBody{
						TimeoutTimestamp: timestamppb.New(currentTime.Add(-1 * time.Hour)),
					},
				},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      nil,
			wantErr:      "block time is after tx timeout timestamp",
		},
		{
			name: "Timeout height exceeded",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1")},
				Tx: &txv1beta1.Tx{
					AuthInfo: &txv1beta1.AuthInfo{
						SignerInfos: []*txv1beta1.SignerInfo{
							{ModeInfo: &txv1beta1.ModeInfo{Sum: &txv1beta1.ModeInfo_Single_{
								Single: &txv1beta1.ModeInfo_Single{Mode: 1},
							}}},
						},
					},
					Body: &txv1beta1.TxBody{
						TimeoutHeight: currentBlock - 1,
					},
				},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      nil,
			wantErr:      "block height is after tx timeout height",
		},
		{
			name: "Multiple AA extensions",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1")},
				Tx: &txv1beta1.Tx{
					AuthInfo: &txv1beta1.AuthInfo{
						SignerInfos: []*txv1beta1.SignerInfo{
							{ModeInfo: &txv1beta1.ModeInfo{Sum: &txv1beta1.ModeInfo_Single_{
								Single: &txv1beta1.ModeInfo_Single{Mode: 1},
							}}},
						},
					},
					Body: &txv1beta1.TxBody{
						ExtensionOptions: []*anypb.Any{
							{TypeUrl: aaXtName, Value: validXtBytes},
							{TypeUrl: aaXtName, Value: validXtBytes},
						},
					},
				},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      nil,
			wantErr:      "multiple aa extensions on the same tx",
		},
		{
			name: "Missing AA extension",
			bundledTx: &txdecode.DecodedTx{
				Signers: [][]byte{[]byte("signer1")},
				Tx: &txv1beta1.Tx{
					AuthInfo: &txv1beta1.AuthInfo{
						SignerInfos: []*txv1beta1.SignerInfo{
							{ModeInfo: &txv1beta1.ModeInfo{Sum: &txv1beta1.ModeInfo_Single_{
								Single: &txv1beta1.ModeInfo_Single{Mode: 1},
							}}},
						},
					},
					Body: &txv1beta1.TxBody{
						ExtensionOptions: []*anypb.Any{},
					},
				},
			},
			currentBlock: currentBlock,
			currentTime:  currentTime,
			wantExt:      nil,
			wantErr:      "did not have AA extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotExt, err := verifyAndExtractAaXtFromTx(tt.bundledTx, tt.currentBlock, tt.currentTime)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantExt, gotExt)
			}
		})
	}
}
