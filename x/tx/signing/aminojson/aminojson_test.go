package aminojson_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	txv1beta1 "cosmossdk.io/api/cosmos/tx/v1beta1"
	"cosmossdk.io/x/tx/signing/aminojson"
	"cosmossdk.io/x/tx/signing/testutil"
)

func TestAminoJsonSignMode(t *testing.T) {
	handlerOptions := testutil.HandlerArgumentOptions{
		ChainId: "test-chain",
		Memo:    "sometestmemo",
		Tip:     &txv1beta1.Tip{Tipper: "tipper", Amount: []*basev1beta1.Coin{{Denom: "Tip-token", Amount: "10"}}},
		Msg: &bankv1beta1.MsgSend{
			FromAddress: "foo",
			ToAddress:   "bar",
			Amount:      []*basev1beta1.Coin{{Denom: "demon", Amount: "100"}},
		},
		AccNum:        1,
		AccSeq:        2,
		SignerAddress: "signerAddress",
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
			name: "nil tip",
			malleate: func(opts testutil.HandlerArgumentOptions) testutil.HandlerArgumentOptions {
				opts.Tip = nil
				return opts
			},
		},
		{
			name: "empty tipper",
			malleate: func(opts testutil.HandlerArgumentOptions) testutil.HandlerArgumentOptions {
				opts.Tip.Tipper = ""
				return opts
			},
			error: "tipper cannot be empty",
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

func TestAminoJsonSignMode_GetSignBytes(t *testing.T) {

}
