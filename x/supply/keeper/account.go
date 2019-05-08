package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
	"github.com/tendermint/tendermint/crypto"
)

// GetAccountByName returns an Account based on the name
func (k Keeper) GetAccountByName(ctx sdk.Context, name string) (auth.Account, sdk.Error) {
	moduleAddress := sdk.AccAddress(crypto.AddressHash([]byte(name)))
	acc := k.ak.GetAccount(ctx, moduleAddress)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(fmt.Sprintf("module account %s does not exist", name))
	}

	return acc, nil
}

// GetPoolAccountByName returns a PoolAccount based on the name
func (k Keeper) GetPoolAccountByName(ctx sdk.Context, name string) (types.PoolAccount, sdk.Error) {
	acc, err := k.GetAccountByName(ctx, name)
	if err != nil {
		return nil, err
	}

	pacc, isPoolAccount := acc.(types.PoolAccount)
	if !isPoolAccount {
		return nil, sdk.ErrInvalidAddress(fmt.Sprintf("account %s is not a module account", name))
	}

	return pacc, nil
}

// SetPoolAccount sets the pool account to the auth account store
func (k Keeper) SetPoolAccount(ctx sdk.Context, pacc types.PoolAccount) {
	k.ak.SetAccount(ctx, pacc)
}
