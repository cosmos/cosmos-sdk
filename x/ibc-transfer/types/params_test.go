package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateParams(t *testing.T) {
	params := DefaultParams()
	require.NoError(t, params.Validate())
}
