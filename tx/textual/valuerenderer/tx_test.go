package valuerenderer

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	_ "cosmossdk.io/api/cosmos/crypto/ed25519"
	_ "cosmossdk.io/api/cosmos/crypto/multisig"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	_ "cosmossdk.io/api/cosmos/gov/v1"
	textualv1 "cosmossdk.io/api/cosmos/msg/textual/v1"
	signingv1beta1 "cosmossdk.io/api/cosmos/tx/signing/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
)

// txJsonTestTx represents the type that in the JSON test
// cases `proto` field. The inner contents are protojson
// encoded, so we represent them as []byte here, and decode
// them inside the test.
type txJsonTestTx struct {
	Body       json.RawMessage
	AuthInfo   json.RawMessage `json:"auth_info"`
	SignerData json.RawMessage `json:"signer_data"`
}

type txJsonTest struct {
	Name    string
	Proto   txJsonTestTx
	Error   bool
	Screens []Screen
	Cbor    string
}

func TestTxJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/tx.json")
	require.NoError(t, err)

	var testcases []txJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			textualData := createTextualData(t, tc.Proto)

			tr := NewTextual(func(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) { return nil, nil })
			rend := NewTxValueRenderer(&tr)

			val := protoreflect.ValueOf(textualData.ProtoReflect())
			screens, err := rend.Format(context.Background(), val)
			if tc.Error {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.Screens, screens)

			var buf bytes.Buffer
			err = encode(screens, &buf)
			require.NoError(t, err)
			require.Equal(t, tc.Cbor, hex.EncodeToString(buf.Bytes()))

			// Round trip.
			// parsedVal, err := rend.Parse(context.Background(), screens)
			// require.NoError(t, err)
			// require.Equal(t, val.Interface(), parsedVal.Interface())
		})
	}
}

// createTextualData creates a Textual data give then JSON
// test case.
func createTextualData(t *testing.T, jsonTx txJsonTestTx) *textualv1.TextualData {
	txBody := &txv1beta1.TxBody{}
	txAuthInfo := &txv1beta1.AuthInfo{}
	signerData := &signingv1beta1.SignerData{}

	// We unmarshal from protojson to the protobuf types.
	err := protojson.Unmarshal(jsonTx.Body, txBody)
	require.NoError(t, err)
	err = protojson.Unmarshal(jsonTx.AuthInfo, txAuthInfo)
	require.NoError(t, err)
	err = protojson.Unmarshal(jsonTx.SignerData, signerData)
	require.NoError(t, err)

	// We marshal body and auth_info
	bodyBz, err := proto.Marshal(txBody)
	require.NoError(t, err)
	authInfoBz, err := proto.Marshal(txAuthInfo)
	require.NoError(t, err)

	return &textualv1.TextualData{
		BodyBytes:     bodyBz,
		AuthInfoBytes: authInfoBz,
		SignerData:    signerData,
	}
}
