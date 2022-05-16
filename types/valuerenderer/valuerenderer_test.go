package valuerenderer

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	bank "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/require"
)

func TestFormatInteger(t *testing.T) {
	type integerTest []string
	var testcases []integerTest
	raw, err := ioutil.ReadFile("./testdata/integers.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		output, err := formatInteger(tc[0])
		require.NoError(t, err)

		require.Equal(t, tc[1], output)
	}
}

func TestFormatDecimal(t *testing.T) {
	type decimalTest []string
	var testcases []decimalTest
	raw, err := ioutil.ReadFile("./testdata/decimals.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		output, err := formatDecimal(tc[0])
		require.NoError(t, err)

		require.Equal(t, tc[1], output)
	}
}

func TestFormatCoin(t *testing.T) {
	var testcases []coinTest
	raw, err := ioutil.ReadFile("./testdata/coins.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		output, err := formatCoin(tc.coin, bank.Metadata{
			Display:    tc.metadata.Denom,
			DenomUnits: []*bank.DenomUnit{{Denom: tc.coin.Denom, Exponent: 0}, {Denom: tc.metadata.Denom, Exponent: tc.metadata.Exponent}},
		})
		require.NoError(t, err)

		require.Equal(t, tc.expRes, output)
	}
}

func TestFormatCoins(t *testing.T) {
	var testcases []coinTest
	raw, err := ioutil.ReadFile("./testdata/coins.json")
	require.NoError(t, err)
	err = json.Unmarshal(raw, &testcases)
	require.NoError(t, err)

	for _, tc := range testcases {
		output, err := formatCoins(sdk.NewCoins(tc.coin), bank.Metadata{
			Display:    tc.metadata.Denom,
			DenomUnits: []*bank.DenomUnit{{Denom: tc.coin.Denom, Exponent: 0}, {Denom: tc.metadata.Denom, Exponent: tc.metadata.Exponent}},
		})
		require.NoError(t, err)

		require.Equal(t, tc.expRes, output)
	}
}

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
