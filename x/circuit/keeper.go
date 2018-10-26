package circuit

import (
	"github.com/cosmos/cosmos-sdk/x/params"
)

// Keeper stores keytable-initialized params.Subspace for circuit breaking
type Keeper struct {
	space params.Subspace
}

// NewKeeper constructs new keeper
func NewKeeper(space params.Subspace) Keeper {
	return Keeper{
		space: space.WithKeyTable(ParamKeyTable()),
	}
}
