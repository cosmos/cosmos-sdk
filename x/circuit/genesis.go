package circuit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params"
)

type GenesisState struct {
	CircuitBreakTypes []string `json:"circuit-break-types"`
}

func InitGenesis(ctx sdk.Context, space params.Subspace, data GenesisState) {
	for _, ty := range data.CircuitBreakTypes {
		space.SetWithSubkey(ctx, MsgTypeKey, []byte(ty), true)
	}
}
