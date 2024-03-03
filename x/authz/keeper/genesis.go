package keeper

import (
	"context"

	"cosmossdk.io/x/authz"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes new authz genesis
func (k Keeper) InitGenesis(ctx context.Context, data *authz.GenesisState) error {
	now := k.environment.HeaderService.GetHeaderInfo(ctx).Time
	for _, entry := range data.Authorization {
		// ignore expired authorizations
		if entry.Expiration != nil && entry.Expiration.Before(now) {
			continue
		}

		grantee, err := k.authKeeper.AddressCodec().StringToBytes(entry.Grantee)
		if err != nil {
			return err
		}
		granter, err := k.authKeeper.AddressCodec().StringToBytes(entry.Granter)
		if err != nil {
			return err
		}

		a, ok := entry.Authorization.GetCachedValue().(authz.Authorization)
		if !ok {
			panic("expected authorization")
		}

		err = k.SaveGrant(ctx, grantee, granter, a, entry.Expiration)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis returns a GenesisState for a given context.
func (k Keeper) ExportGenesis(ctx context.Context) *authz.GenesisState {
	var entries []authz.GrantAuthorization
	k.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant authz.Grant) bool {
		entries = append(entries, authz.GrantAuthorization{
			Granter:       granter.String(),
			Grantee:       grantee.String(),
			Expiration:    grant.Expiration,
			Authorization: grant.Authorization,
		})
		return false
	})

	return authz.NewGenesisState(entries)
}
