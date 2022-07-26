package valuerenderer_test

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"strings"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestFormatCoins(t *testing.T) {
	var testcases []coinsTest
	raw, err := ioutil.ReadFile("../internal/testdata/coins.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		// Create a context.Context containing all coins metadata, to simulate
		// that they are in state.
		ctx := context.Background()
		for _, coin := range tc.coins {
			m := tc.metadataMap[coin.Denom]
			metadata := &bankv1beta1.Metadata{
				Display:    m.Denom,
				DenomUnits: []*bankv1beta1.DenomUnit{{Denom: coin.Denom, Exponent: 0}, {Denom: m.Denom, Exponent: m.Exponent}},
			}

			ctx = context.WithValue(ctx, mockCoinMetadataKey(coin.Denom), metadata)
		}

		r, err := valueRendererOf(tc.coins)
		require.NoError(t, err)
		b := new(strings.Builder)
		err = r.Format(ctx, protoreflect.ValueOf(tc.coins), b)
		require.NoError(t, err)

		require.Equal(t, tc.expRes, b.String())
	}
}

type coinsTest struct {
	coins       []*basev1beta1.Coin
	metadataMap map[string]coinTestMetadata
	expRes      string
}

func (t *coinsTest) UnmarshalJSON(b []byte) error {
	a := []interface{}{&t.coins, &t.metadataMap, &t.expRes}
	return json.Unmarshal(b, &a)
}
