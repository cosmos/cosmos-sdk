package staking

import (
	"testing"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/staking/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	f := initFixture(t, false)
	acc := f.accountKeeper.GetAccount(f.ctx, authtypes.NewModuleAddress(types.BondedPoolName))
	require.NotNil(t, acc)

	acc = f.accountKeeper.GetAccount(f.ctx, authtypes.NewModuleAddress(types.NotBondedPoolName))
	require.NotNil(t, acc)
}
