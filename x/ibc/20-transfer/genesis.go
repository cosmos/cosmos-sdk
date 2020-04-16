package transfer

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/ibc/20-transfer/types"
)

// GenesisState is currently only used to ensure that the InitGenesis gets run
// by the module manager
type GenesisState struct {
	PortID  string
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
}

func DefaultGenesis() GenesisState {
	return GenesisState{
		PortID:  types.PortID,
		Version: types.Version,
	}
}

// InitGenesis sets distribution information for genesis
func InitGenesis(ctx sdk.Context, keeper Keeper, state GenesisState) {
	// transfer module binds to the transfer port on InitChain
	// and claims the returned capability
	err := keeper.BindPort(ctx, state.PortID)
	if err != nil {
		panic(fmt.Sprintf("could not claim port capability: %v", err))
	}
	// check if the module account exists
	moduleAcc := keeper.GetTransferAccount(ctx)
	if moduleAcc == nil {
		panic(fmt.Sprintf("%s module account has not been set", types.GetModuleAccountName()))
	}
}
