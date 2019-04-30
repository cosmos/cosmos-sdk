package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CoinsSupply defines relevant supply information from all the coins tracked by the current chain.
// It is the format exported to clients.
type CoinsSupply struct {
	Circulating    sdk.Coins `json:"circulating"`     // supply held by accounts that's not vesting; circulating = total - vesting
	InitialVesting sdk.Coins `json:"initial_vesting"` // initial locked supply held by vesting accounts
	Modules        sdk.Coins `json:"modules"`         // supply held by modules acccounts
	Liquid         sdk.Coins `json:"liquid"`          // sum of account spendable coins at a given time
	Total          sdk.Coins `json:"total"`           // total supply of tokens on the chain
}

// NewCoinsSupply creates a supply instance for all the coins
func NewCoinsSupply(circulating, vesting, modules, liquid, total sdk.Coins) CoinsSupply {
	return CoinsSupply{
		Circulating:    circulating,
		InitialVesting: vesting,
		Modules:        modules,
		Liquid:         liquid,
		Total:          total,
	}
}

// NewCoinsSupplyFromSupplier creates CoinsSupply instance from a given Supplier
func NewCoinsSupplyFromSupplier(supplier Supplier) CoinsSupply {
	return NewCoinsSupply(supplier.CirculatingSupply,
		supplier.InitialVestingSupply, supplier.ModulesSupply,
		supplier.TotalSupply.Sub(supplier.InitialVestingSupply), supplier.TotalSupply)
}

// CoinSupply defines the supply information for a single coin on the current chain.
// It is the format exported to clients when an individual denom is queried.
type CoinSupply struct {
	Circulating    sdk.Int `json:"circulating"`     // supply held by accounts that's not vesting; circulating = total - vesting
	InitialVesting sdk.Int `json:"initial_vesting"` // initial locked supply held by vesting accounts
	Modules        sdk.Int `json:"modules"`         // supply held by modules acccounts
	Liquid         sdk.Int `json:"liquid"`          // sum of account spendable coins at a given time
	Total          sdk.Int `json:"total"`           // total supply of tokens on the chain
}

// NewCoinSupply creates a supply instance for a single coin denom
func NewCoinSupply(circulating, vesting, modules, liquid, total sdk.Int) CoinSupply {
	return CoinSupply{
		Circulating:    circulating,
		InitialVesting: vesting,
		Modules:        modules,
		Liquid:         liquid,
		Total:          total,
	}
}

// NewCoinSupplyFromSupplier creates CoinSupply instance from a given Supplier
func NewCoinSupplyFromSupplier(denom string, supplier Supplier) CoinSupply {
	return NewCoinSupply(
		supplier.CirculatingSupply.AmountOf(denom),
		supplier.InitialVestingSupply.AmountOf(denom),
		supplier.ModulesSupply.AmountOf(denom),
		supplier.TotalSupply.Sub(supplier.InitialVestingSupply).AmountOf(denom),
		supplier.TotalSupply.AmountOf(denom))
}
