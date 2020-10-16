package msg_authorization

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/keeper"
	"github.com/cosmos/cosmos-sdk/x/msg_authorization/types"
)

// InitGenesis new msg_authorization genesis
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
		authorization, ok := entry.Authorization.GetCachedValue().(types.Authorization)
		if !ok {
			panic("expected authorization")
		}

		keeper.Grant(ctx, grantee, granter, authorization, entry.Expiration)
	}
}

// ExportGenesis returns a GenesisState for a given context and keeper.
func ExportGenesis(ctx sdk.Context, keeper keeper.Keeper) *types.GenesisState {
	var entries []types.MsgGrantAuthorization
	keeper.IterateGrants(ctx, func(granter, grantee sdk.AccAddress, grant types.AuthorizationGrant) bool {
		exp := time.Unix(grant.Expiration, 0)
		entries = append(entries, types.MsgGrantAuthorization{
			Granter:       granter.String(),
			Grantee:       grantee.String(),
			Expiration:    exp,
			Authorization: grant.Authorization,
		})
		return false
	})

	return types.NewGenesisState(entries)
}
