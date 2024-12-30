package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	f := createTestSuite(t)
	ctx := f.ctx
	acc := f.authKeeper.GetAccount(ctx, types.NewModuleAddress(types.FeeCollectorName))
	require.NotNil(t, acc)
}
