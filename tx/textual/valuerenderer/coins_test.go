package valuerenderer_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
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
	vr, err := textual.GetValueRenderer(fieldDescriptorFromName("COINS"))
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Text, func(t *testing.T) {
			if tc.Proto != nil {
				// Create a context.Context containing all coins metadata, to simulate
				// that they are in state.
				ctx := context.Background()
				for _, coin := range tc.Proto {
					ctx = context.WithValue(ctx, mockCoinMetadataKey(coin.Denom), tc.Metadata[coin.Denom])
				}

				listValue := NewGenericList(tc.Proto)
				screens, err := vr.Format(ctx, protoreflect.ValueOf(listValue))

				if tc.Error {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				require.Equal(t, 1, len(screens))
				require.Equal(t, tc.Text, screens[0].Text)
			}

			// TODO Add parsing tests
			// https://github.com/cosmos/cosmos-sdk/issues/13153
		})
	}
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
