package v040

// DONTCOVER
// nolint

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	ModuleName = "bank"
)

var _ GenesisBalance = (*Balance)(nil)

type (
	GenesisBalance interface {
		GetAddress() sdk.AccAddress
		GetCoins() sdk.Coins
	}

	GenesisState struct {
		SendEnabled bool      `json:"send_enabled" yaml:"send_enabled"`
		Balances    []Balance `json:"balances" yaml:"balances"`
	}

	Balance struct {
		Address sdk.AccAddress `json:"address" yaml:"address"`
		Coins   sdk.Coins      `json:"coins" yaml:"coins"`
	}
)

func NewGenesisState(sendEnabled bool, balances []Balance) GenesisState {
	return GenesisState{SendEnabled: sendEnabled, Balances: balances}
}

func (b Balance) GetAddress() sdk.AccAddress {
	return b.Address
}

func (b Balance) GetCoins() sdk.Coins {
	return b.Coins
}
