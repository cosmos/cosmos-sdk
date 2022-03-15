package keeper

import (
	"github.com/cosmos/cosmos-sdk/x/tieredfee/types"
)

var _ types.QueryServer = Keeper{}
