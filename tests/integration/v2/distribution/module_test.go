package distribution

import (
	"testing"

	"cosmossdk.io/x/distribution/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"gotest.tools/v3/assert"
)

func TestItCreatesModuleAccountOnInitBlock(t *testing.T) {
	f := createTestFixture(t)
	acc := f.authKeeper.GetAccount(f.ctx, authtypes.NewModuleAddress(types.ModuleName))
	assert.Assert(t, acc != nil)
}
