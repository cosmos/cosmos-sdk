package circuit

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

type Keeper struct {
	space params.Subspace
}

func NewKeeper(space params.Subspace) Keeper {
	return Keeper{
		space: space.WithKeyTable(ParamKeyTable()),
	}
}
