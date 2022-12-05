package valuerenderer_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/math"
	"cosmossdk.io/tx/textual/valuerenderer"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func TestCoinsJsonTestcases(t *testing.T) {
	var testcases []coinsJsonTest
	raw, err := os.ReadFile("../internal/testdata/coins.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	textual := valuerenderer.NewTextual(mockCoinMetadataQuerier)
	vr, err := textual.GetFieldValueRenderer(fieldDescriptorFromName("COINS"))
	vrr := vr.(valuerenderer.RepeatedValueRenderer)
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Text, func(t *testing.T) {
			if tc.Proto != nil {
				// Create a context.Context containing all coins metadata, to simulate
				// that they are in state.
				ctx := context.Background()

				for _, v := range tc.Metadata {
					ctx = context.WithValue(ctx, mockCoinMetadataKey(v.Base), v)
					ctx = context.WithValue(ctx, mockCoinMetadataKey(v.Display), v)
				}

				listValue := NewGenericList(tc.Proto)
				screens, err := vrr.FormatRepeated(ctx, protoreflect.ValueOf(listValue))

				require.NoError(t, err)
				require.Equal(t, 1, len(screens))
				require.Equal(t, tc.Text, screens[0].Text)

				// Round trip.
				parsedValue := NewGenericList([]*basev1beta1.Coin{})
				err = vrr.ParseRepeated(ctx, screens, parsedValue)
				if tc.Error {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				checkCoinsEqual(t, listValue, parsedValue)
			}
		})
	}
}

// checkCoinsEqual checks that the 2 lists of Coins contain the same
// **set** of coins. It does not check that the order of coins are
// equal, because in Textual, we sort the coins alphabetically after
// rendering, so we lose initial Coins ordering. Instead, we just check
// set equality using a map.
func checkCoinsEqual(t *testing.T, l1, l2 protoreflect.List) {
	require.Equal(t, l1.Len(), l2.Len())
	var coinsMap = make(map[string]*basev1beta1.Coin, l1.Len())

	for i := 0; i < l1.Len(); i++ {
		coin, ok := l1.Get(i).Message().Interface().(*basev1beta1.Coin)
		require.True(t, ok)
		coinsMap[coin.Denom] = coin
	}

	for i := 0; i < l2.Len(); i++ {
		coin, ok := l2.Get(i).Message().Interface().(*basev1beta1.Coin)
		require.True(t, ok)

		coin1 := coinsMap[coin.Denom]
		checkCoinEqual(t, coin, coin1)
	}
}

func checkCoinEqual(t *testing.T, coin, coin1 *basev1beta1.Coin) {
	require.Equal(t, coin1.Denom, coin.Denom)
	v, err := math.LegacyNewDecFromStr(coin.Amount)
	require.NoError(t, err)
	v1, err := math.LegacyNewDecFromStr(coin1.Amount)
	require.NoError(t, err)
	require.True(t, v.Equal(v1))
}

// coinsJsonTest is the type of test cases in the testdata file.
// If the test case has a Proto, try to Format() it. If Error is set, expect
// an error, otherwise match Text, then Parse() the text and expect it to
// match (via proto.Equals()) the original Proto. If the test case has no
// Proto, try to Parse() the Text and expect an error if Error is set.
type coinsJsonTest struct {
	Proto    []*basev1beta1.Coin
	Metadata map[string]*bankv1beta1.Metadata
	Text     string
	Error    bool
}
