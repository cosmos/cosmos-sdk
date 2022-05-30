package valuerenderer_test

import (
	"encoding/json"
	"io/ioutil"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/valuerenderer"
)

func TestFormatInteger(t *testing.T) {
	type integerTest []string
	var testcases []integerTest
	raw, err := ioutil.ReadFile("./internal/fixtures/integers.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	r := valuerenderer.NewADR050ValueRenderer()

	for _, tc := range testcases {
		b, err := strconv.ParseUint(tc[0], 10, 32)
		require.NoError(t, err)
		output, err := formatGoType(r, b)
		require.NoError(t, err)

		require.Equal(t, tc[1], output)
	}
}

func TestFormatDecimal(t *testing.T) {
	type decimalTest []string
	var testcases []decimalTest
	raw, err := ioutil.ReadFile("./internal/fixtures/decimals.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	r := valuerenderer.NewADR050ValueRenderer()

	for _, tc := range testcases {
		output, err := formatGoType(r, tc[0])
		require.NoError(t, err)

		require.Equal(t, tc[1], output)
	}
}

// func TestFormatCoin(t *testing.T) {
// 	var testcases []coinTest
// 	raw, err := ioutil.ReadFile("./internal/fixtures/coins.json")
// 	require.NoError(t, err)
// 	err = json.Unmarshal(raw, &testcases)
// 	require.NoError(t, err)

// 	r := valuerenderer.NewADR050ValueRenderer()

// 	for _, tc := range testcases {
// 		output, err := formatCoin(tc.coin, bank.Metadata{
// 			Display:    tc.metadata.Denom,
// 			DenomUnits: []*bank.DenomUnit{{Denom: tc.coin.Denom, Exponent: 0}, {Denom: tc.metadata.Denom, Exponent: tc.metadata.Exponent}},
// 		})
// 		require.NoError(t, err)

// 		require.Equal(t, tc.expRes, output)
// 	}
// }

// func TestFormatCoins(t *testing.T) {
// 	var testcases []coinTest
// 	raw, err := ioutil.ReadFile("./internal/fixtures/coins.json")
// 	require.NoError(t, err)
// 	err = json.Unmarshal(raw, &testcases)
// 	require.NoError(t, err)

// 	for _, tc := range testcases {
// 		output, err := formatCoins(sdk.NewCoins(tc.coin), bank.Metadata{
// 			Display:    tc.metadata.Denom,
// 			DenomUnits: []*bank.DenomUnit{{Denom: tc.coin.Denom, Exponent: 0}, {Denom: tc.metadata.Denom, Exponent: tc.metadata.Exponent}},
// 		})
// 		require.NoError(t, err)

// 		require.Equal(t, tc.expRes, output)
// 	}
// }

type coinTestMetadata struct {
	Denom    string `json:"denom"`
	Exponent uint32 `json:"exponent"`
}

type coinTest struct {
	coin     sdk.Coin
	metadata coinTestMetadata
	expRes   string
}

func (t *coinTest) UnmarshalJSON(b []byte) error {
	a := []interface{}{&t.coin, &t.metadata, &t.expRes}
	return json.Unmarshal(b, &a)
}
