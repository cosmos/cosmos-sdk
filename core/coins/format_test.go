package coins_test

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/core/coins"
)

// coinsJsonTest is the type of test cases in the coin.json file.
type coinJSONTest struct {
	Proto    *basev1beta1.Coin
	Metadata *bankv1beta1.Metadata
	Text     string
	Error    bool
}

// coinsJSONTest is the type of test cases in the coins.json file.
type coinsJSONTest struct {
	Proto    []*basev1beta1.Coin
	Metadata map[string]*bankv1beta1.Metadata
	Text     string
	Error    bool
}

func TestFormatCoin(t *testing.T) {
	var testcases []coinJSONTest
	raw, err := os.ReadFile("../../x/tx/signing/textual/internal/testdata/coin.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Text, func(t *testing.T) {
			if tc.Proto != nil {
				out, err := coins.FormatCoins([]*basev1beta1.Coin{tc.Proto}, []*bankv1beta1.Metadata{tc.Metadata})

				if tc.Error {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.Text, out)
			}
		})
	}
}

func TestFormatCoins(t *testing.T) {
	var testcases []coinsJSONTest
	raw, err := os.ReadFile("../../x/tx/signing/textual/internal/testdata/coins.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Text, func(t *testing.T) {
			if tc.Proto != nil {
				metadata := make([]*bankv1beta1.Metadata, len(tc.Proto))
				for i, coin := range tc.Proto {
					metadata[i] = tc.Metadata[coin.Denom]
				}

				out, err := coins.FormatCoins(tc.Proto, metadata)

				if tc.Error {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.Text, out)
			}
		})
	}
}

func TestDecodeCoin(t *testing.T) {
	encodedCoin := "1000000000foo"
	coin, err := coins.ParseCoin(encodedCoin)
	require.NoError(t, err)
	require.Equal(t, "1000000000", coin.Amount)
	require.Equal(t, "foo", coin.Denom)
}
