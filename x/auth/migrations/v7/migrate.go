package v7

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Migrate from v6 to v7. This includes the addition of the SigVerifyCostMlDsa65 field.
func Migrate(ctx context.Context, params collections.Item[types.Params]) error {
	p, err := params.Get(ctx)
	if err != nil {
		return err
	}

	// Do not call p.Validate() here: it also checks params added in later
	// versions, which are still unset when a chain runs this step on its way
	// to a newer version. The value set here is a valid constant.
	p.SigVerifyCostMlDsa65 = types.DefaultSigVerifyCostMlDsa65

	return params.Set(ctx, p)
}
