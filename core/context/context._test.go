package context

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Sanity(t *testing.T) {
	require.NotEqual(t, CometInfoKey, ExecModeKey)
}
