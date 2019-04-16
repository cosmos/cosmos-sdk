package supply

import (
	"github.com/cosmos/cosmos-sdk/x/supply/keeper"
	"github.com/cosmos/cosmos-sdk/x/supply/types"
)

// nolint
type (
	Keeper      = keeper.Keeper
	Supplier    = types.Supplier
	TokenHolder = types.TokenHolder
)

// nolint
var (
	NewKeeper = keeper.NewKeeper

	GetTokenHolderKey = keeper.GetTokenHolderKey
	GetSupplier       = keeper.GetSupplier
	SetSupplier       = keeper.SetSupplier
	InflateSupply     = keeper.InflateSupply
	GetTokenHolders   = keeper.GetTokenHolders
	GetTokenHolder    = keeper.GetTokenHolder
	AddTokenHolder    = keeper.AddTokenHolder
	GetTokenHolder    = keeper.GetTokenHolder
	AddTokenHolder    = keeper.AddTokenHolder
	SetTokenHolder    = keeper.SetTokenHolder
	RequestTokens     = keeper.RequestTokens
	RelinquishTokens  = keeper.RelinquishTokens
)

// nolint
const (
	StoreKey     = keeper.StoreKey
	QuerierRoute = keeper.QuerierRoute
)
