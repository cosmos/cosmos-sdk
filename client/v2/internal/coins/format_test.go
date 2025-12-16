package coins

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseCoin(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		amount string
		denom  string
		err    string
	}{
		{
			name:   "ok",
			input:  "1000stake",
			amount: "1000",
			denom:  "stake",
		},
		{
			name:  "empty",
			input: "",
			err:   "empty input when parsing coin",
		},
		{
			name:  "empty denom",
			input: "1000",
			err:   "invalid input format",
		},
		{
			name:  "empty amount",
			input: "stake",
			err:   "invalid input format",
		},
		{
			name:  "<denom><amount> format",
			input: "stake1000",
			err:   "invalid input format",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount, denom, err := parseCoin(tt.input)
			if tt.err != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.amount, amount)
				require.Equal(t, tt.denom, denom)
			}
		})
	}
}

func TestParseCoin(t *testing.T) {
	encodedCoin := "1000000000foo"
	coin, err := ParseCoin(encodedCoin)
	require.NoError(t, err)
	require.Equal(t, "1000000000", coin.Amount)
	require.Equal(t, "foo", coin.Denom)
}

func TestParseDecCoin(t *testing.T) {
	encodedCoin := "1000000000foo"
	coin, err := ParseDecCoin(encodedCoin)
	require.NoError(t, err)
	require.Equal(t, "1000000000", coin.Amount)
	require.Equal(t, "foo", coin.Denom)
}
