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

		err = keeper.SaveGrant(ctx, grantee, granter, authorization, entry.Expiration)
		if err != nil {
			panic(err)
		}
	}
}
