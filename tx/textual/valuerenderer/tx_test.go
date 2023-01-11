package valuerenderer_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/cosmos/cosmos-proto/any"
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
	"cosmossdk.io/api/cosmos/crypto/multisig"
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

type txJsonTest struct {
	Name       string
	Proto      txJsonTestTx
	SignerData json.RawMessage `json:"signer_data"`
	Metadata   *bankv1beta1.Metadata
	Error      bool
	Screens    []valuerenderer.Screen
}

func TestTxJsonTestcases(t *testing.T) {
	raw, err := os.ReadFile("../internal/testdata/tx.json")
	require.NoError(t, err)

	var testcases []txJsonTest
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Name, func(t *testing.T) {
			txBody, bodyBz, txAuthInfo, authInfoBz, signerData := createTextualData(t, tc.Proto, tc.SignerData)

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

			// Round trip.
			parsedVal, err := rend.Parse(ctx, screens)
			require.NoError(t, err)

			// We don't check that bodyBz and authInfoBz are equal, because
			// they don't need to be. Instead, we check that the semantic
			// proto objects are equal.
			parsedTextualData := parsedVal.Message().Interface().(*textualpb.TextualData)

			parsedBody := &txv1beta1.TxBody{}
			err = proto.Unmarshal(parsedTextualData.BodyBytes, parsedBody)
			require.NoError(t, err)
			diff := cmp.Diff(txBody, parsedBody, protocmp.Transform())
			require.Empty(t, diff)

			parsedAuthInfo := &txv1beta1.AuthInfo{}
			err = proto.Unmarshal(parsedTextualData.AuthInfoBytes, parsedAuthInfo)
			require.NoError(t, err)
			// Remove the non-parsable fields, i.e. the hashed bytes
			for i, si := range txAuthInfo.SignerInfos {
				txAuthInfo.SignerInfos[i].PublicKey = removePkKeys(t, si.PublicKey)
			}
			diff = cmp.Diff(txAuthInfo, parsedAuthInfo, protocmp.Transform())
			require.Empty(t, diff)

			// Remove the non-parsable fields, i.e. the hashed public key
			removePkKeys(t, signerData.PubKey)
			diff = cmp.Diff(
				signerData,
				signerDataFromProto(parsedTextualData.SignerData),
				protocmp.Transform(),
			)
			require.Empty(t, diff)
		})
	}
}

// createTextualData creates a Textual data give then JSON
// test case.
func createTextualData(t *testing.T, jsonTx txJsonTestTx, jsonSignerData json.RawMessage) (*txv1beta1.TxBody, []byte, *txv1beta1.AuthInfo, []byte, signing.SignerData) {
	body := &txv1beta1.TxBody{}
	authInfo := &txv1beta1.AuthInfo{}
	protoSignerData := &textualpb.SignerData{}

	// We unmarshal from protojson to the protobuf types.
	err := protojson.Unmarshal(jsonTx.Body, body)
	require.NoError(t, err)
	err = protojson.Unmarshal(jsonTx.AuthInfo, authInfo)
	require.NoError(t, err)
	err = protojson.Unmarshal(jsonSignerData, protoSignerData)
	require.NoError(t, err)

	// We marshal body and auth_info
	bodyBz, err := proto.Marshal(body)
	require.NoError(t, err)
	authInfoBz, err := proto.Marshal(authInfo)
	require.NoError(t, err)

	return body, bodyBz, authInfo, authInfoBz, signerDataFromProto(protoSignerData)
}

// signerDataFromProto converts a protobuf SignerData (internal) to a
// signing.SignerData (external).
func signerDataFromProto(d *textualpb.SignerData) signing.SignerData {
	return signing.SignerData{
		Address:       d.Address,
		ChainId:       d.ChainId,
		AccountNumber: d.AccountNumber,
		Sequence:      d.Sequence,
		PubKey:        d.PubKey,
	}
}

// removePkKeys takes a public key Any, decodes it, and recursively removes all
// the "key" fields (hashed by textual) inside it.
func removePkKeys(t *testing.T, pkAny *anypb.Any) *anypb.Any {
	pk, err := anypb.UnmarshalNew(pkAny, proto.UnmarshalOptions{})
	require.NoError(t, err)
	m := pk.ProtoReflect().Interface()
	switch m := m.(type) {
	case *multisig.LegacyAminoPubKey:
		newAnys := make([]*anypb.Any, len(m.PublicKeys))
		for i, any := range m.PublicKeys {
			newAnys[i] = removePkKeys(t, any)
		}

		m.PublicKeys = newAnys
		newMultisigAny, err := any.New(m)
		require.NoError(t, err)

		return newMultisigAny
	default:
		pkAny.Value = nil
		return pkAny
	}
}
