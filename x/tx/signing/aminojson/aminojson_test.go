package aminojson_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/known/timestamppb"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"

	"github.com/cosmos/cosmos-sdk/x/tx/signing/aminojson"
	"github.com/cosmos/cosmos-sdk/x/tx/signing/testutil"
)

func TestAminoJsonSignMode(t *testing.T) {
	fee := &txv1beta1.Fee{
		Amount: []*basev1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
	}
	handlerOptions := testutil.HandlerArgumentOptions{
		ChainID: "test-chain",
		Memo:    "sometestmemo",
		Msg: &bankv1beta1.MsgSend{
			FromAddress: "foo",
			ToAddress:   "bar",
			Amount:      []*basev1beta1.Coin{{Denom: "demon", Amount: "100"}},
		},
		AccNum:        1,
		AccSeq:        2,
		SignerAddress: "signerAddress",
		Fee:           fee,
	}

	testCases := []struct {
		name     string
		malleate func(opts testutil.HandlerArgumentOptions) testutil.HandlerArgumentOptions
		error    string
	}{
		{
			name: "happy path",
			malleate: func(opts testutil.HandlerArgumentOptions) testutil.HandlerArgumentOptions {
				return opts
			},
		},
		{
			name: "empty signer",
			malleate: func(opts testutil.HandlerArgumentOptions) testutil.HandlerArgumentOptions {
				opts.SignerAddress = ""
				return opts
			},
			error: "got empty address in SIGN_MODE_LEGACY_AMINO_JSON handler: invalid request",
		},
		{
			name: "nil fee",
			malleate: func(opts testutil.HandlerArgumentOptions) testutil.HandlerArgumentOptions {
				opts.Fee = nil
				return opts
			},
			error: "fee cannot be nil",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opts := tc.malleate(handlerOptions)
			signerData, txData, err := testutil.MakeHandlerArguments(opts)
			require.NoError(t, err)

			handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
			_, err = handler.GetSignBytes(context.Background(), signerData, txData)
			if tc.error != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.error)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestUnorderedTimeoutCompat(t *testing.T) {
	fee := &txv1beta1.Fee{
		Amount: []*basev1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
	}

	now := time.Now()
	tests := []struct {
		name             string
		unordered        bool
		timeout          *timestamppb.Timestamp
		wantUnordered    any
		wantTimeoutEmpty bool
	}{
		{
			name:             "empty unordered and timeout",
			unordered:        false,
			timeout:          nil,
			wantUnordered:    nil,
			wantTimeoutEmpty: true,
		},
		{
			name:             "business unordered and timeout",
			unordered:        true,
			timeout:          func() *timestamppb.Timestamp { t := now.Add(3 * time.Hour); return timestamppb.New(t) }(),
			wantUnordered:    true,
			wantTimeoutEmpty: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opts := testutil.HandlerArgumentOptions{
				ChainID: "test-chain",
				Memo:    "sometestmemo",
				Msg: &bankv1beta1.MsgSend{
					FromAddress: "foo",
					ToAddress:   "bar",
					Amount:      []*basev1beta1.Coin{{Denom: "demon", Amount: "100"}},
				},
				AccNum:           1,
				AccSeq:           2,
				SignerAddress:    "signerAddress",
				Fee:              fee,
				Unordered:        tc.unordered,
				Timeouttimestamp: tc.timeout,
			}

			signerData, txData, err := testutil.MakeHandlerArguments(opts)
			require.NoError(t, err)

			handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
			bz, err := handler.GetSignBytes(context.Background(), signerData, txData)
			require.NoError(t, err)

			var result map[string]any
			require.NoError(t, json.Unmarshal(bz, &result))

			require.Equal(t, tc.wantUnordered, result["unordered"])

			if tc.wantTimeoutEmpty {
				require.Empty(t, result["timeout_timestamp"])
			} else {
				require.NotEmpty(t, result["timeout_timestamp"])
				gotTimeStr := result["timeout_timestamp"].(string)
				gotTime, err := time.Parse("2006-01-02T15:04:05.999999Z", gotTimeStr)
				require.NoError(t, err)
				require.True(t, gotTime.Equal(tc.timeout.AsTime()))
			}
		})
	}
}

func TestUnorderedEmpty(t *testing.T) {
	fee := &txv1beta1.Fee{
		Amount: []*basev1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
	}

	opts := testutil.HandlerArgumentOptions{
		ChainID: "test-chain",
		Memo:    "sometestmemo",
		Msg: &bankv1beta1.MsgSend{
			FromAddress: "foo",
			ToAddress:   "bar",
			Amount:      []*basev1beta1.Coin{{Denom: "demon", Amount: "100"}},
		},
		AccNum:        1,
		AccSeq:        2,
		SignerAddress: "signerAddress",
		Fee:           fee,
	}
	signerData, txData, err := testutil.MakeHandlerArguments(opts)
	require.NoError(t, err)

	handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
	bz, err := handler.GetSignBytes(context.Background(), signerData, txData)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Empty(t, result["unordered"])
	require.Empty(t, result["timeout_timestamp"])
}

func TestUnorderedBusiness(t *testing.T) {
	fee := &txv1beta1.Fee{
		Amount: []*basev1beta1.Coin{{Denom: "uatom", Amount: "1000"}},
	}

	timeout := time.Now().Add(3 * time.Hour)
	opts := testutil.HandlerArgumentOptions{
		ChainID: "test-chain",
		Memo:    "sometestmemo",
		Msg: &bankv1beta1.MsgSend{
			FromAddress: "foo",
			ToAddress:   "bar",
			Amount:      []*basev1beta1.Coin{{Denom: "demon", Amount: "100"}},
		},
		AccNum:           1,
		AccSeq:           2,
		SignerAddress:    "signerAddress",
		Fee:              fee,
		Unordered:        true,
		Timeouttimestamp: timestamppb.New(timeout),
	}
	signerData, txData, err := testutil.MakeHandlerArguments(opts)
	require.NoError(t, err)

	handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
	bz, err := handler.GetSignBytes(context.Background(), signerData, txData)
	require.NoError(t, err)

	var result map[string]any
	require.NoError(t, json.Unmarshal(bz, &result))
	require.Equal(t, true, result["unordered"])
	require.NotEmpty(t, result["timeout_timestamp"])
	gotTimeStr := result["timeout_timestamp"].(string)
	gotTime, err := time.Parse("2006-01-02T15:04:05.999999Z", gotTimeStr)
	require.NoError(t, err)
	require.True(t, gotTime.Equal(timeout))
}

func TestNewSignModeHandler(t *testing.T) {
	handler := aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{})
	require.NotNil(t, handler)
	aj := aminojson.NewEncoder(aminojson.EncoderOptions{})
	handler = aminojson.NewSignModeHandler(aminojson.SignModeHandlerOptions{
		FileResolver: protoregistry.GlobalFiles,
		TypeResolver: protoregistry.GlobalTypes,
		Encoder:      &aj,
	})
	require.NotNil(t, handler)
}

func TestNullSliceAsEmptyEncoder(t *testing.T) {
	encoder := aminojson.NewEncoder(aminojson.EncoderOptions{})

	testCases := []struct {
		name        string
		amount      []*basev1beta1.Coin
		wantJSON    string
		description string
	}{
		{
			name:        "empty slice encodes as empty array",
			amount:      []*basev1beta1.Coin{},
			wantJSON:    `[]`,
			description: "Empty slice should be encoded as [] not null",
		},
		{
			name:        "nil slice encodes as empty array",
			amount:      nil,
			wantJSON:    `[]`,
			description: "Nil slice should be encoded as [] not null",
		},
		{
			name: "non-empty slice encodes normally",
			amount: []*basev1beta1.Coin{
				{Denom: "uatom", Amount: "1000"},
				{Denom: "stake", Amount: "500"},
			},
			wantJSON:    "", // Will check content instead of exact match
			description: "Non-empty slice should encode normally",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fee := &txv1beta1.Fee{
				Amount: tc.amount,
			}

			// Test that the encoder works with legacy_coins (which uses NullSliceAsEmptyEncoder)
			bz, err := encoder.Marshal(fee)
			require.NoError(t, err)

			var result map[string]interface{}
			require.NoError(t, json.Unmarshal(bz, &result))

			amountJSON, err := json.Marshal(result["amount"])
			require.NoError(t, err)
			if tc.wantJSON == "" {
				// For non-empty slices, just verify it's a valid array with expected content
				require.Contains(t, string(amountJSON), "uatom")
				require.Contains(t, string(amountJSON), "stake")
				require.Contains(t, string(amountJSON), "1000")
				require.Contains(t, string(amountJSON), "500")
			} else {
				require.Equal(t, tc.wantJSON, string(amountJSON), tc.description)
			}
		})
	}
}

func TestNullSliceAsEmptyEncoderDirect(t *testing.T) {
	encoder := aminojson.NewEncoder(aminojson.EncoderOptions{})

	// Test direct usage of NullSliceAsEmptyEncoder with a custom field encoding
	customEncoder := encoder.DefineFieldEncoding("test_field", aminojson.NullSliceAsEmptyEncoder)
	require.NotNil(t, customEncoder)

	// Create a Fee message with an empty list (Fee uses legacy_coins which uses NullSliceAsEmptyEncoder)
	fee := &txv1beta1.Fee{
		Amount: []*basev1beta1.Coin{}, // empty slice
	}

	// Marshal using the encoder
	bz, err := encoder.Marshal(fee)
	require.NoError(t, err)

	var result map[string]interface{}
	require.NoError(t, json.Unmarshal(bz, &result))

	// Verify that amount field exists and is an empty array (not null)
	amount, ok := result["amount"]
	require.True(t, ok, "amount field should exist")
	amountJSON, err := json.Marshal(amount)
	require.NoError(t, err)
	require.Equal(t, `[]`, string(amountJSON), "empty slice should be encoded as [] not null")
}
