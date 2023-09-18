package gov

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/stretchr/testify/require"
)

func failingHandler(_ sdk.Context, _ sdk.Msg) (*sdk.Result, error) {
	panic("test-fail")
}

func okHandler(_ sdk.Context, _ sdk.Msg) (*sdk.Result, error) {
	return new(sdk.Result), nil
}

func TestSafeExecuteHandler(t *testing.T) {
	t.Parallel()

	require := require.New(t)
	var ctx sdk.Context

	r, err := safeExecuteHandler(ctx, nil, failingHandler)
	require.ErrorContains(err, "test-fail")
	require.Nil(r)

	r, err = safeExecuteHandler(ctx, nil, okHandler)
	require.Nil(err)
	require.NotNil(r)
}
