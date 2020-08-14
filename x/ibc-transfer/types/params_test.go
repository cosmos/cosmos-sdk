package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateParams(t *testing.T) {
	// TODO:

	params := DefaultParams()

	// default params have no error
	require.NoError(t, params.Validate())
}
