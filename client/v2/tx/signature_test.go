package tx

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	apimultisig "cosmossdk.io/api/cosmos/crypto/multisig/v1beta1"
	apisigning "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
)

func TestSignatureDataToModeInfoAndSig(t *testing.T) {
	tests := []struct {
		name      string
		data      SignatureData
		mIResult  *apitx.ModeInfo
		sigResult []byte
	}{
		{
			name: "single signature",
			data: &SingleSignatureData{
				SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
				Signature: []byte("signature"),
			},
			mIResult: &apitx.ModeInfo{
				Sum: &apitx.ModeInfo_Single_{
					Single: &apitx.ModeInfo_Single{Mode: apisigning.SignMode_SIGN_MODE_DIRECT},
				},
			},
			sigResult: []byte("signature"),
		},
		{
			name: "multi signature",
			data: &MultiSignatureData{
				BitArray: nil,
				Signatures: []SignatureData{
					&SingleSignatureData{
						SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
						Signature: []byte("signature"),
					},
				},
			},
			mIResult: &apitx.ModeInfo{
				Sum: &apitx.ModeInfo_Multi_{
					Multi: &apitx.ModeInfo_Multi{
						Bitarray: nil,
						ModeInfos: []*apitx.ModeInfo{
							{
								Sum: &apitx.ModeInfo_Single_{
									Single: &apitx.ModeInfo_Single{Mode: apisigning.SignMode_SIGN_MODE_DIRECT},
								},
							},
						},
					},
				},
			},
			sigResult: []byte("\n\tsignature"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modeInfo, signature := signatureDataToModeInfoAndSig(tt.data)
			require.Equal(t, tt.mIResult, modeInfo)
			require.Equal(t, tt.sigResult, signature)
		})
	}
}

func TestModeInfoAndSigToSignatureData(t *testing.T) {
	type args struct {
		modeInfo func() *apitx.ModeInfo
		sig      []byte
	}
	tests := []struct {
		name    string
		args    args
		want    SignatureData
		wantErr bool
	}{
		{
			name: "to SingleSignatureData",
			args: args{
				modeInfo: func() *apitx.ModeInfo {
					return &apitx.ModeInfo{
						Sum: &apitx.ModeInfo_Single_{
							Single: &apitx.ModeInfo_Single{Mode: apisigning.SignMode_SIGN_MODE_DIRECT},
						},
					}
				},
				sig: []byte("signature"),
			},
			want: &SingleSignatureData{
				SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
				Signature: []byte("signature"),
			},
		},
		{
			name: "to MultiSignatureData",
			args: args{
				modeInfo: func() *apitx.ModeInfo {
					return &apitx.ModeInfo{
						Sum: &apitx.ModeInfo_Multi_{
							Multi: &apitx.ModeInfo_Multi{
								Bitarray: &apimultisig.CompactBitArray{},
								ModeInfos: []*apitx.ModeInfo{
									{
										Sum: &apitx.ModeInfo_Single_{
											Single: &apitx.ModeInfo_Single{Mode: apisigning.SignMode_SIGN_MODE_DIRECT},
										},
									},
								},
							},
						},
					}
				},
				sig: []byte("\n\tsignature"),
			},
			want: &MultiSignatureData{ // Changed from SingleSignatureData to MultiSignatureData
				BitArray: &apimultisig.CompactBitArray{},
				Signatures: []SignatureData{
					&SingleSignatureData{
						SignMode:  apisigning.SignMode_SIGN_MODE_DIRECT,
						Signature: []byte("signature"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := modeInfoAndSigToSignatureData(tt.args.modeInfo(), tt.args.sig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ModeInfoAndSigToSignatureData() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ModeInfoAndSigToSignatureData() got = %v, want %v", got, tt.want)
			}
		})
	}
}
