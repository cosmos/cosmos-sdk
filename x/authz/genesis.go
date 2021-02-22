package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz/exported"
	"github.com/cosmos/cosmos-sdk/x/authz/keeper"
	"github.com/cosmos/cosmos-sdk/x/authz/types"
)

// InitGenesis new authz genesis
func InitGenesis(ctx sdk.Context, keeper keeper.Keeper, data *types.GenesisState) {
	for _, entry := range data.Authorization {
		grantee, err := sdk.AccAddressFromBech32(entry.Grantee)
		if err != nil {
			panic(err)
		}
		granter, err := sdk.AccAddressFromBech32(entry.Granter)
		if err != nil {
			panic(err)
		}
		authorization, ok := entry.Authorization.GetCachedValue().(exported.Authorization)
		if !ok {
			panic("expected authorization")
		}

		err = keeper.Grant(ctx, grantee, granter, authorization, entry.Expiration)
		if err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	var entries []types.GrantAuthorization
	keeper.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant types.AuthorizationGrant) bool {
		exp := grant.Expiration
		entries = append(entries, types.GrantAuthorization{
			Granter:       granter.String(),
			Grantee:       grantee.String(),
			Expiration:    exp,
			Authorization: grant.Authorization,
		})
		return false
	})

	return types.NewGenesisState(entries)
}
