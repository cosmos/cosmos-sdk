package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// keeper of the stake store
type Keeper struct {
	storeKey            sdk.StoreKey
	cdc                 *codec.Codec
	paramSpace          params.Subspace
	bankKeeper          types.BankKeeper
	stakeKeeper         types.StakeKeeper
	feeCollectionKeeper types.FeeCollectionKeeper

	// codespace
	codespace sdk.CodespaceType
}

// create a new keeper
func NewKeeper(cdc *codec.Codec, key sdk.StoreKey, paramSpace params.Subspace, ck types.BankKeeper,
	sk types.StakeKeeper, fck types.FeeCollectionKeeper, codespace sdk.CodespaceType) Keeper {
	keeper := Keeper{
		storeKey:            key,
		cdc:                 cdc,
		paramSpace:          paramSpace.WithTypeTable(ParamTypeTable()),
		bankKeeper:          ck,
		stakeKeeper:         sk,
		feeCollectionKeeper: fck,
		codespace:           codespace,
	}
	return keeper
}
