package accountstd

import (
	"context"

	"cosmossdk.io/collections/colltest"
	"cosmossdk.io/core/store"
	"cosmossdk.io/x/accounts/internal/implementation"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewMockContext(
	accNumber uint64,
	accountAddr []byte,
	sender []byte,
	funds sdk.Coins,
	moduleExec implementation.ModuleExecFunc,
	moduleExecUntyped implementation.ModuleExecUntypedFunc,
	moduleQuery implementation.ModuleQueryFunc,
) (context.Context, store.KVStoreService) {
	ss, ctx := colltest.MockStore()

	return implementation.MakeAccountContext(
		ctx, ss, accNumber, accountAddr, sender, funds, moduleExec, moduleExecUntyped, moduleQuery,
	), ss
}
