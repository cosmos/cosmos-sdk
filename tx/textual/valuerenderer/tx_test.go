package valuerenderer_test

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/anypb"

	_ "cosmossdk.io/api/cosmos/auth/v1beta1"
	_ "cosmossdk.io/api/cosmos/authz/v1beta1"
	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	_ "cosmossdk.io/api/cosmos/crypto/ed25519"
	_ "cosmossdk.io/api/cosmos/crypto/multisig"
	_ "cosmossdk.io/api/cosmos/crypto/secp256k1"
	_ "cosmossdk.io/api/cosmos/gov/v1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/tx/signing"
	"cosmossdk.io/tx/textual/internal/textualpb"
	"cosmossdk.io/tx/textual/valuerenderer"
)

// txJsonTestTx represents the type that in the JSON test
// cases `proto` field. The inner contents are protojson
// encoded, so we represent them as []byte here, and decode
// them inside the test.
type txJsonTestTx struct {
	Body     json.RawMessage
	AuthInfo json.RawMessage `json:"auth_info"`
}

type txJsonSignerData struct {
	Address       string
	AccountNumber uint64          `json:"account_number"`
	ChainID       string          `json:"chain_id"`
	PubKey        json.RawMessage `json:"pub_key"`
	Sequence      uint64
}

type txJsonTest struct {
	Name       string
	Proto      txJsonTestTx
	SignerData txJsonSignerData `json:"signer_data"`
	Metadata   *bankv1beta1.Metadata
	Error      bool
	Screens    []valuerenderer.Screen
	Cbor       string
}

func TestTxJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/tx.json")
	require.NoError(t, err)

	var testcases []txJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			expBody, bodyBz, expAuthInfo, authInfoBz, signerData := createTextualData(t, tc.Proto, tc.SignerData)

			tr := valuerenderer.NewTextual(mockCoinMetadataQuerier)
			rend := valuerenderer.NewTxValueRenderer(&tr)
			ctx := addMetadataToContext(context.Background(), tc.Metadata)

			data := &textualpb.TextualData{
				BodyBytes:     bodyBz,
				AuthInfoBytes: authInfoBz,
				SignerData: &textualpb.SignerData{
					Address:       signerData.Address,
					ChainId:       signerData.ChainId,
					AccountNumber: signerData.AccountNumber,
					Sequence:      signerData.Sequence,
					PubKey:        signerData.PubKey,
				},
			}

			// Make sure the screens match.
			val := protoreflect.ValueOf(data.ProtoReflect())
			screens, err := rend.Format(ctx, val)
			if tc.Error {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.Screens, screens)

			// Make sure the CBOR bytes match.
			bz, err := tr.GetSignBytes(ctx, bodyBz, authInfoBz, signerData)
			require.NoError(t, err)
			require.Equal(t, tc.Cbor, hex.EncodeToString(bz))

			// Round trip.
			parsedVal, err := rend.Parse(ctx, screens)
			require.NoError(t, err)

			// We don't check that bodyBz and authInfoBz are equal, because
			// they don't need to be. Instead, we check that the semantic
			// proto objects are equal.
			parsedTextualData := parsedVal.Message().Interface().(*textualpb.TextualData)

			parsedBody := &txv1beta1.TxBody{}
			err = proto.Unmarshal(bodyBz, expBody)
			require.NoError(t, err)
			err = proto.Unmarshal(parsedTextualData.BodyBytes, parsedBody)
			require.NoError(t, err)
			diff := cmp.Diff(expBody, parsedBody, protocmp.Transform())
			require.Empty(t, diff)

			parsedAuthInfo := &txv1beta1.AuthInfo{}
			err = proto.Unmarshal(authInfoBz, expAuthInfo)
			require.NoError(t, err)
			err = proto.Unmarshal(parsedTextualData.AuthInfoBytes, parsedAuthInfo)
			require.NoError(t, err)
			diff = cmp.Diff(expAuthInfo, parsedAuthInfo, protocmp.Transform())
			require.Empty(t, diff)

			require.Equal(t, signerData.AccountNumber, parsedTextualData.SignerData.AccountNumber)
			require.Equal(t, signerData.Sequence, parsedTextualData.SignerData.Sequence)
			require.Equal(t, signerData.ChainId, parsedTextualData.SignerData.ChainId)
			require.Equal(t, signerData.PubKey, parsedTextualData.SignerData.PubKey)
		})
	}
}

// createTextualData creates a Textual data give then JSON
// test case.
func createTextualData(t *testing.T, jsonTx txJsonTestTx, jsonSignerData txJsonSignerData) (*txv1beta1.TxBody, []byte, *txv1beta1.AuthInfo, []byte, signing.SignerData) {
	body := &txv1beta1.TxBody{}
	authInfo := &txv1beta1.AuthInfo{}

	// We unmarshal from protojson to the protobuf types.
	err := protojson.Unmarshal(jsonTx.Body, body)
	require.NoError(t, err)
	err = protojson.Unmarshal(jsonTx.AuthInfo, authInfo)
	require.NoError(t, err)

	// Unmarshal the pubkey
	anyPubKey := &anypb.Any{}
	err = protojson.Unmarshal(jsonSignerData.PubKey, anyPubKey)
	require.NoError(t, err)

	signerData := signing.SignerData{
		Address:       jsonSignerData.Address,
		ChainId:       jsonSignerData.ChainID,
		AccountNumber: jsonSignerData.AccountNumber,
		Sequence:      jsonSignerData.Sequence,
		PubKey:        anyPubKey,
	}

	// We marshal body and auth_info
	bodyBz, err := proto.Marshal(body)
	require.NoError(t, err)
	authInfoBz, err := proto.Marshal(authInfo)
	require.NoError(t, err)

	return body, bodyBz, authInfo, authInfoBz, signerData
}
