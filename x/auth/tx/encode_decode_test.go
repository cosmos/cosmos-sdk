package tx

import (
	"fmt"
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/types/tx"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
)

func TestDefaultTxDecoderError(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	encoder := DefaultTxEncoder()
	decoder := DefaultTxDecoder(cdc)

	builder := newBuilder()
	err := builder.SetMsgs(testdata.NewTestMsg())
	require.NoError(t, err)

	txBz, err := encoder(builder.GetTx())
	require.NoError(t, err)

	_, err = decoder(txBz)
	require.EqualError(t, err, "unable to resolve type URL /testdata.TestMsg: tx parse error")

	testdata.RegisterInterfaces(registry)
	_, err = decoder(txBz)
	require.NoError(t, err)
}

func TestUnknownFields(t *testing.T) {
	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)
	decoder := DefaultTxDecoder(cdc)

	tests := []struct {
		name           string
		body           *testdata.TestUpdatedTxBody
		authInfo       *testdata.TestUpdatedAuthInfo
		shouldErr      bool
		shouldAminoErr string
	}{
		{
			name: "no new fields should pass",
			body: &testdata.TestUpdatedTxBody{
				Memo: "foo",
			},
			authInfo:  &testdata.TestUpdatedAuthInfo{},
			shouldErr: false,
		},
		{
			name: "non-critical fields in TxBody should not error on decode, but should error with amino",
			body: &testdata.TestUpdatedTxBody{
				Memo:                         "foo",
				SomeNewFieldNonCriticalField: "blah",
			},
			authInfo:       &testdata.TestUpdatedAuthInfo{},
			shouldErr:      false,
			shouldAminoErr: fmt.Sprintf("%s: %s", aminoNonCriticalFieldsError, sdkerrors.ErrInvalidRequest.Error()),
		},
		{
			name: "critical fields in TxBody should error on decode",
			body: &testdata.TestUpdatedTxBody{
				Memo:         "foo",
				SomeNewField: 10,
			},
			authInfo:  &testdata.TestUpdatedAuthInfo{},
			shouldErr: true,
		},
		{
			name: "critical fields in AuthInfo should error on decode",
			body: &testdata.TestUpdatedTxBody{
				Memo: "foo",
			},
			authInfo: &testdata.TestUpdatedAuthInfo{
				NewField_3: []byte("xyz"),
			},
			shouldErr: true,
		},
		{
			name: "non-critical fields in AuthInfo should error on decode",
			body: &testdata.TestUpdatedTxBody{
				Memo: "foo",
			},
			authInfo: &testdata.TestUpdatedAuthInfo{
				NewField_1024: []byte("xyz"),
			},
			shouldErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bodyBz, err := tt.body.Marshal()
			require.NoError(t, err)

			authInfoBz, err := tt.authInfo.Marshal()
			require.NoError(t, err)

			txRaw := &tx.TxRaw{
				BodyBytes:     bodyBz,
				AuthInfoBytes: authInfoBz,
			}
			txBz, err := txRaw.Marshal()
			require.NoError(t, err)

			_, err = decoder(txBz)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			if tt.shouldAminoErr != "" {
				handler := signModeLegacyAminoJSONHandler{}
				decoder := DefaultTxDecoder(codec.NewProtoCodec(codectypes.NewInterfaceRegistry()))
				theTx, err := decoder(txBz)
				require.NoError(t, err)
				_, err = handler.GetSignBytes(signingtypes.SignMode_SIGN_MODE_LEGACY_AMINO_JSON, signing.SignerData{}, theTx)
				require.EqualError(t, err, tt.shouldAminoErr)
			}
		})
	}

	t.Log("test TxRaw no new fields, should succeed")
	txRaw := &testdata.TestUpdatedTxRaw{}
	txBz, err := txRaw.Marshal()
	require.NoError(t, err)
	_, err = decoder(txBz)
	require.NoError(t, err)

	t.Log("new field in TxRaw should fail")
	txRaw = &testdata.TestUpdatedTxRaw{
		NewField_5: []byte("abc"),
	}
	txBz, err = txRaw.Marshal()
	require.NoError(t, err)
	_, err = decoder(txBz)
	require.Error(t, err)

	//
	t.Log("new \"non-critical\" field in TxRaw should fail")
	txRaw = &testdata.TestUpdatedTxRaw{
		NewField_1024: []byte("abc"),
	}
	txBz, err = txRaw.Marshal()
	require.NoError(t, err)
	_, err = decoder(txBz)
	require.Error(t, err)
}
