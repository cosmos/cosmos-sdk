package types

import sdk "github.com/cosmos/cosmos-sdk/types"

func (m *Fee) GetGas() uint64 {
	return m.GasLimit
}

func (m *Fee) SetGas(u uint64) {
	m.GasLimit = u
}

func (m *Fee) SetAmount(coins sdk.Coins) {
	m.Amount = coins
}
