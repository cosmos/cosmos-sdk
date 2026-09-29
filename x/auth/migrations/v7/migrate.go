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

	p.SigVerifyCostMlDsa65 = types.DefaultSigVerifyCostMlDsa65
	if err = p.Validate(); err != nil {
		return err
	}

	if err = params.Set(ctx, p); err != nil {
		return err
	}

	return nil
}
