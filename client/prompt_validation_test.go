package client_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/client"
)

func TestValidatePromptURL(t *testing.T) {
	require := require.New(t)

	require.NoError(client.ValidatePromptURL("https://example.com"))
	require.ErrorContains(client.ValidatePromptURL("foo"), "invalid URL")
}
