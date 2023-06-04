package keeper

import (
	"cosmossdk.io/x/circuit/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k *Keeper) ExportGenesis(ctx sdk.Context) (data *types.GenesisState) {
	var (
		permissions  []*types.GenesisAccountPermissions
		disabledMsgs []string
	)

	k.IteratePermissions(ctx, func(address []byte, perm types.Permissions) (stop bool) {
		add, err := k.addressCodec.BytesToString(address)
		if err != nil {
			panic(err)
		}
		// Convert the Permissions struct to a GenesisAccountPermissions struct
		// and add it to the permissions slice
		permissions = append(permissions, &types.GenesisAccountPermissions{
			Address:     add,
			Permissions: &perm,
		})
		return false
	})

	k.IterateDisableLists(ctx, func(url string) (stop bool) {
		disabledMsgs = append(disabledMsgs, url)
		return false
	})

	return &types.GenesisState{
		AccountPermissions: permissions,
		DisabledTypeUrls:   disabledMsgs,
	}
}

// InitGenesis initializes the circuit module's state from a given genesis state.
func (k *Keeper) InitGenesis(ctx sdk.Context, genState *types.GenesisState) {
	for _, accounts := range genState.AccountPermissions {
		add, err := k.addressCodec.StringToBytes(accounts.Address)
		if err != nil {
			panic(err)
		}

		// Set the permissions for the account
		if err := k.SetPermissions(ctx, add, accounts.Permissions); err != nil {
			panic(err)
		}
	}
	for _, url := range genState.DisabledTypeUrls {
		// Set the disabled type urls
		k.DisableMsg(ctx, url)
	}
}
