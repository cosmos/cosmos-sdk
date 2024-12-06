package prompt

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/core/address"

	address2 "github.com/cosmos/cosmos-sdk/codec/address"
)

func TestValidatePromptNotEmpty(t *testing.T) {
	require := require.New(t)

	require.NoError(ValidatePromptNotEmpty("foo"))
	require.ErrorContains(ValidatePromptNotEmpty(""), "input cannot be empty")
}

func TestValidateAddress(t *testing.T) {
	tests := []struct {
		name string
		ac   address.Codec
		addr string
	}{
		{
			name: "address",
			ac:   address2.NewBech32Codec("cosmos"),
			addr: "cosmos129lxcu2n3hx54fdxlwsahqkjr3sp32cxm00zlm",
		},
		{
			name: "validator address",
			ac:   address2.NewBech32Codec("cosmosvaloper"),
			addr: "cosmosvaloper1tnh2q55v8wyygtt9srz5safamzdengsn9dsd7z",
		},
		{
			name: "consensus address",
			ac:   address2.NewBech32Codec("cosmosvalcons"),
			addr: "cosmosvalcons136uu5rj23kdr3jjcmjt7aw5qpugjjat2klgrus",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.ac)(tt.addr)
			require.NoError(t, err)
		})
	}
}

func TestValidatePromptURL(t *testing.T) {
	require := require.New(t)

	require.NoError(ValidatePromptURL("https://example.com"))
	require.ErrorContains(ValidatePromptURL("foo"), "invalid URL")
}
