package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func TestFormatInvariant(t *testing.T) {
	t.Parallel()
	require.Equal(t, ":  invariant\n\n", sdk.FormatInvariant("", "", ""))
	require.Equal(t, "module: name invariant\nmsg\n", sdk.FormatInvariant("module", "name", "msg"))
}
