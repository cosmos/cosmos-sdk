package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
	"cosmossdk.io/tx/textual/valuerenderer"
)

// mockCoinMetadataKey is used in the mock coin metadata querier.
func mockCoinMetadataKey(denom string) string {
	return fmt.Sprintf("%s-%s", "coin-metadata", denom)
}

// mockCoinMetadataQuerier is a mock querier for coin metadata used for test
// purposes.
func mockCoinMetadataQuerier(ctx context.Context, denom string) (*bankv1beta1.Metadata, error) {
	v := ctx.Value(mockCoinMetadataKey(denom))
	if v == nil {
		return nil, nil
	}

	return v.(*bankv1beta1.Metadata), nil
}

func TestMetadataQuerier(t *testing.T) {
	// Errors on nil metadata querier
	textual := valuerenderer.NewTextual(nil)
	vr, err := textual.GetValueRenderer(fieldDescriptorFromName("COIN"))
	require.NoError(t, err)
	_, err = vr.Format(context.Background(), protoreflect.ValueOf((&basev1beta1.Coin{}).ProtoReflect()))
	require.Error(t, err)

	// Errors if metadata querier returns an error
	expErr := fmt.Errorf("mock error")
	textual = valuerenderer.NewTextual(func(_ context.Context, _ string) (*bankv1beta1.Metadata, error) {
		return nil, expErr
	})
	vr, err = textual.GetValueRenderer(fieldDescriptorFromName("COIN"))
	require.NoError(t, err)
	_, err = vr.Format(context.Background(), protoreflect.ValueOf((&basev1beta1.Coin{}).ProtoReflect()))
	require.ErrorIs(t, err, expErr)
	_, err = vr.Format(context.Background(), protoreflect.ValueOf(NewGenericList([]*basev1beta1.Coin{{}})))
	require.ErrorIs(t, err, expErr)
}

func TestCoinJsonTestcases(t *testing.T) {
	var testcases []coinJsonTest
	raw, err := os.ReadFile("../internal/testdata/coin.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	textual := valuerenderer.NewTextual(mockCoinMetadataQuerier)
	vr, err := textual.GetValueRenderer(fieldDescriptorFromName("COIN"))
	require.NoError(t, err)

	for _, tc := range testcases {
		t.Run(tc.Text, func(t *testing.T) {
			if tc.Proto != nil {
				ctx := context.WithValue(context.Background(), mockCoinMetadataKey(tc.Proto.Denom), tc.Metadata)
				screens, err := vr.Format(ctx, protoreflect.ValueOf(tc.Proto.ProtoReflect()))

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

// coinJsonTest is the type of test cases in the testdata file.
// If the test case has a Proto, try to Format() it. If Error is set, expect
// an error, otherwise match Text, then Parse() the text and expect it to
// match (via proto.Equals()) the original Proto. If the test case has no
// Proto, try to Parse() the Text and expect an error if Error is set.
type coinJsonTest struct {
	Proto    *basev1beta1.Coin
	Metadata *bankv1beta1.Metadata
	Error    bool
	Text     string
}
