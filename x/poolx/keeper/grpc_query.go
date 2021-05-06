package keeper

import (
	"github.com/cosmos/cosmos-sdk/x/poolx/types"
)

var _ types.QueryServer = Keeper{}
