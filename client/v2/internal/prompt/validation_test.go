package prompt_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/client/v2/internal/prompt"
)

func TestValidatePromptNotEmpty(t *testing.T) {
	require := require.New(t)

	require.NoError(prompt.ValidatePromptNotEmpty("foo"))
	require.ErrorContains(prompt.ValidatePromptNotEmpty(""), "input cannot be empty")
}

func TestValidatePromptURL(t *testing.T) {
	require := require.New(t)

	require.NoError(prompt.ValidatePromptURL("https://example.com"))
	require.ErrorContains(prompt.ValidatePromptURL("foo"), "invalid URL")
}

func TestValidatePromptAddress(t *testing.T) {
	require := require.New(t)

	require.NoError(prompt.ValidatePromptAddress("cosmos1huydeevpz37sd9snkgul6070mstupukw00xkw9"))
	require.NoError(prompt.ValidatePromptAddress("cosmosvaloper1sjllsnramtg3ewxqwwrwjxfgc4n4ef9u2lcnj0"))
	require.NoError(prompt.ValidatePromptAddress("cosmosvalcons1ntk8eualewuprz0gamh8hnvcem2nrcdsgz563h"))
	require.ErrorContains(prompt.ValidatePromptAddress("foo"), "invalid address")
}

func TestValidatePromptCoins(t *testing.T) {
	require := require.New(t)

	require.NoError(prompt.ValidatePromptCoins("100stake"))
	require.ErrorContains(prompt.ValidatePromptCoins("foo"), "invalid coins")
}
