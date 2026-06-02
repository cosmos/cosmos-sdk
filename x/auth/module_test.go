package auth_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testapp "github.com/cosmos/cosmos-sdk/testutil/testapp"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	ta := testapp.Setup(t)
	ctx := ta.NewContext(false)
	acc := ta.AccountKeeper.GetAccount(ctx, types.NewModuleAddress(types.FeeCollectorName))
	require.NotNil(t, acc)
}
