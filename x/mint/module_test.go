package mint_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testapp "github.com/cosmos/cosmos-sdk/testutil/testapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	ta := testapp.Setup(t)
	ctx := ta.NewContext(false)
	acc := ta.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.ModuleName))
	require.NotNil(t, acc)
}
