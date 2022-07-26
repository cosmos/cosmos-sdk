package valuerenderer_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/reflect/protoreflect"

	bankv1beta1 "cosmossdk.io/api/cosmos/bank/v1beta1"
	basev1beta1 "cosmossdk.io/api/cosmos/base/v1beta1"
)

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

func TestFormatCoin(t *testing.T) {
	var testcases []coinTest
	raw, err := ioutil.ReadFile("../internal/testdata/coin.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		metadata := &bankv1beta1.Metadata{
			Display:    tc.metadata.Denom,
			DenomUnits: []*bankv1beta1.DenomUnit{{Denom: tc.coin.Denom, Exponent: 0}, {Denom: tc.metadata.Denom, Exponent: tc.metadata.Exponent}},
		}
		ctx := context.WithValue(context.Background(), mockCoinMetadataKey(tc.coin.Denom), metadata)

		r, err := valueRendererOf(tc.coin)
		require.NoError(t, err)
		b := new(strings.Builder)
		err = r.Format(ctx, protoreflect.ValueOf(tc.coin.ProtoReflect()), b)
		require.NoError(t, err)

		require.Equal(t, tc.expRes, b.String())
	}
}

type coinTestMetadata struct {
	Denom    string `json:"denom"`
	Exponent uint32 `json:"exponent"`
}

type coinTest struct {
	coin     *basev1beta1.Coin
	metadata coinTestMetadata
	expRes   string
}

func (t *coinTest) UnmarshalJSON(b []byte) error {
	a := []interface{}{&t.coin, &t.metadata, &t.expRes}
	return json.Unmarshal(b, &a)
}
