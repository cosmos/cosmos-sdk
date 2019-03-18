package crisis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Params are the params used for the crisis module
type Params struct {
	ConstantFee sdk.Coin
}

// Params - create a new Params object
func NewParams(constantFee sdk.Coin) Params {
	return Params{
		ConstantFee: constantFee,
	}
}

// Default parameter namespace
const (
	DefaultParamspace = ModuleName
)

var (
	// key for constant fee parameter
	ParamStoreKeyConstantFee = []byte("ConstantFee")
)

// type declaration for parameters
func ParamKeyTable() params.KeyTable {
	return params.NewKeyTable(
		ParamStoreKeyConstantFee, sdk.Coin{},
	)
}

// GetConstantFee get's the constant fee from the paramSpace
func (k Keeper) GetConstantFee(ctx sdk.Context) sdk.Coin {
	var constantFee sdk.Coin
	k.paramSpace.Get(ctx, ParamStoreKeyConstantFee, &constantFee)
	return constantFee
}

// GetConstantFee set's the constant fee in the paramSpace
func (k Keeper) SetConstantFee(ctx sdk.Context, constantFee sdk.Dec) {
	k.paramSpace.Set(ctx, ParamStoreKeyConstantFee, &constantFee)
}
