package gov

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

func (keeper Keeper) SupplyKeeper() SupplyKeeper {
	return keeper.supplyKeeper
}

func (keeper Keeper) ParamSpace() params.Subspace {
	return keeper.paramSpace
}

func (keeper Keeper) StoreKey() sdk.StoreKey {
	return keeper.storeKey
}

func (keeper Keeper) Cdc() *codec.Codec {
	return keeper.cdc
}

func (keeper Keeper) Router() Router {
	return keeper.router
}

func (keeper Keeper) Codespace() sdk.CodespaceType {
	return keeper.codespace
}

type (
	ValidatorGovInfo = validatorGovInfo
)

var (
	NewValidatorGovInfo = newValidatorGovInfo
	IsAlphaNumeric      = isAlphaNumeric
)
