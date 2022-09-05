package valuerenderer_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestFormatCoins(t *testing.T) {
	var testcases []coinsTest
	raw, err := os.ReadFile("../internal/testdata/coins.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	textual := valuerenderer.NewTextual(mockCoinMetadataQuerier)

	for _, tc := range testcases {
		t.Run(tc.expRes, func(t *testing.T) {
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

			r, err := textual.GetValueRenderer(fieldDescriptorFromName("COINS"))
			require.NoError(t, err)
			b := new(strings.Builder)
			listValue := NewGenericList(tc.coins)
			err = r.Format(ctx, protoreflect.ValueOf(listValue), b)
			require.NoError(t, err)

			require.Equal(t, tc.expRes, b.String())
		})
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
