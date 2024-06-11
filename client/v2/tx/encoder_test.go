package tx

import (
	apitx "cosmossdk.io/api/cosmos/tx/v1beta1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/anypb"
	"testing"
)

func Test_txEncoder_txDecoder(t *testing.T) {
	tests := []struct {
		name string
		tx   *apitx.Tx
	}{
		{
			name: "encode and tx",
			tx: &apitx.Tx{
				Body: &apitx.TxBody{
					Messages: []*anypb.Any{{
						TypeUrl: "/test/decode",
						Value:   []byte("foo"),
					}},
					Memo:                        "memo",
					TimeoutHeight:               1,
					Unordered:                   false,
					ExtensionOptions:            nil,
					NonCriticalExtensionOptions: nil,
				},
				AuthInfo: &apitx.AuthInfo{
					SignerInfos: []*apitx.SignerInfo{
						{
							PublicKey: &anypb.Any{
								TypeUrl: "customKey",
								Value:   []byte("key"),
							},
							Sequence: 1,
						},
					},
					Fee: nil,
				},
				Signatures: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encodedTx, err := txEncoder(tt.tx)
			require.NoError(t, err)
			require.NotNil(t, encodedTx)

			isDeterministic, err := txEncoder(tt.tx)
			require.NoError(t, err)
			require.NotNil(t, encodedTx)
			require.Equal(t, encodedTx, isDeterministic)

			decodedTx, err := txDecoder(encodedTx)
			require.NoError(t, err)
			require.NotNil(t, decodedTx)
		})
	}
}

func Test_txJsonEncoder_txJsonDecoder(t *testing.T) {
	tests := []struct {
		name string
		tx   *apitx.Tx
	}{
		{
			name: "json encode and decode tx",
			tx: &apitx.Tx{
				Body: &apitx.TxBody{
					Messages:                    []*anypb.Any{},
					Memo:                        "memo",
					TimeoutHeight:               1,
					Unordered:                   false,
					ExtensionOptions:            nil,
					NonCriticalExtensionOptions: nil,
				},
				AuthInfo: &apitx.AuthInfo{
					SignerInfos: []*apitx.SignerInfo{
						{
							PublicKey: &anypb.Any{},
							Sequence:  1,
						},
					},
					Fee: nil,
				},
				Signatures: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encodedTx, err := txJsonEncoder(tt.tx)
			require.NoError(t, err)
			require.NotNil(t, encodedTx)

			decodedTx, err := txJsonDecoder(encodedTx)
			require.NoError(t, err)
			require.NotNil(t, decodedTx)
		})
	}
}
