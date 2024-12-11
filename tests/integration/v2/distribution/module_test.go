package distribution

import (
	"testing"

	"gotest.tools/v3/assert"

	"cosmossdk.io/x/distribution/types"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	f := createTestFixture(t)
	acc := f.authKeeper.GetAccount(f.ctx, authtypes.NewModuleAddress(types.ModuleName))
	assert.Assert(t, acc != nil)
}
