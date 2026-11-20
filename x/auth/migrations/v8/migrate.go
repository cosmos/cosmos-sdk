package v8

import (
	"context"

	"cosmossdk.io/collections"

	"github.com/cosmos/cosmos-sdk/x/auth/types"
)

// Migrate from v7 to v8. This adds the PubKeyChangeEnabled and
// PubKeyChangeCost params. Account rekeying starts disabled.
func Migrate(ctx context.Context, params collections.Item[types.Params]) error {
	p, err := params.Get(ctx)
	if err != nil {
		return err
	}

	// Do not call p.Validate() here: it would also check params added in later
	// versions, which are still unset when a chain runs this step on its way
	// to a newer version. The values set here are valid constants.
	p.PubKeyChangeEnabled = types.DefaultPubKeyChangeEnabled
	p.PubKeyChangeCost = types.DefaultPubKeyChangeCost

	return params.Set(ctx, p)
}
