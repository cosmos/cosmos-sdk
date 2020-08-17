package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateParams(t *testing.T) {
	require.NoError(t, DefaultParams().Validate())
	require.NoError(t, NewParams(true, false).Validate())
}
