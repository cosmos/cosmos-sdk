package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"

	_ "cosmossdk.io/api/cosmos/bank/v1beta1"
	_ "cosmossdk.io/api/cosmos/crypto/ed25519"
	_ "cosmossdk.io/api/cosmos/crypto/multisig"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/tx/signing"
	"cosmossdk.io/tx/textual/valuerenderer"
)

type txJsonSignerData struct {
	Address       string
	ChainId       string
	AccountNumber uint64
	Sequence      uint64
	PubKey        json.RawMessage
}

type txJsonTest struct {
	Proto      json.RawMessage
	SignerData txJsonSignerData
	Error      bool
	Screens    []valuerenderer.Screen
}

func TestTxJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/tx.json")
	require.NoError(t, err)

	var testcases []txJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for i, tc := range testcases {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {

			var pk anypb.Any
			err = protojson.Unmarshal(tc.SignerData.PubKey, &pk)
			require.NoError(t, err)

			signerData := signing.SignerData{
				Address:       tc.SignerData.Address,
				ChainID:       tc.SignerData.ChainId,
				AccountNumber: tc.SignerData.AccountNumber,
				Sequence:      tc.SignerData.Sequence,
				PubKey:        &pk,
			}

			tr := valuerenderer.NewTextual(EmptyCoinMetadataQuerier, signerData, []byte{02}, []byte{03})
			rend := valuerenderer.NewTxValueRenderer(&tr)

			var screens []valuerenderer.Screen
			if tc.Proto != nil {
				var protoTx txv1beta1.Tx
				protojson.Unmarshal(tc.Proto, &protoTx)
				screens, err = rend.Format(context.Background(), protoreflect.ValueOf(protoTx.ProtoReflect()))
				if tc.Error {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
				require.Equal(t, tc.Screens, screens)
			}

			// val, err := rend.Parse(context.Background(), screens)
			// if tc.Error {
			// 	require.Error(t, err)
			// 	return
			// }
			// require.NoError(t, err)
			// msg := val.Message().Interface()
			// require.IsType(t, &tspb.Timestamp{}, msg)
			// timestamp := msg.(*tspb.Timestamp)
			// require.True(t, proto.Equal(timestamp, tc.Proto))
		})
	}
}
