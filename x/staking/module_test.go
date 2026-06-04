package staking_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testapp "github.com/cosmos/cosmos-sdk/testutil/testapp"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/staking/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	ta := testapp.Setup(t)
	ctx := testapp.NewContext(ta)

	acc := ta.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.BondedPoolName))
	require.NotNil(t, acc)

	acc = ta.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.NotBondedPoolName))
	require.NotNil(t, acc)

	acc = ta.AccountKeeper.GetAccount(ctx, authtypes.NewModuleAddress(types.KeyRotationFeePoolName))
	require.NotNil(t, acc)
}
