package types

import (
	"bytes"
	"encoding/json"
	fmt "fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/exported"
)

var _ exported.GenesisBalance = (*Balance)(nil)

// GetAddress returns the account address of the Balance object.
func (b Balance) GetAddress() sdk.AccAddress {
	addr, err := sdk.AccAddressFromBech32(b.Address)
	if err != nil {
		panic(fmt.Errorf("couldn't convert %q to account address: %v", b.Address, err))
	}

	return addr
}

// GetCoins returns the account coins of the Balance object.
func (b Balance) GetCoins() sdk.Coins {
	return b.Coins
}

// Validate checks for address and coins correctness.
func (b Balance) Validate() error {
	_, err := sdk.AccAddressFromBech32(b.Address)
	if err != nil {
		return err
	}
	seenDenoms := make(map[string]bool)

	// NOTE: we perform a custom validation since the coins.Validate function
	// errors on zero balance coins
	for _, coin := range b.Coins {
		if seenDenoms[coin.Denom] {
			return fmt.Errorf("duplicate denomination %s", coin.Denom)
		}

		if err := sdk.ValidateDenom(coin.Denom); err != nil {
			return err
		}

		if coin.IsNegative() {
			return fmt.Errorf("coin %s amount is cannot be negative", coin.Denom)
		}

		seenDenoms[coin.Denom] = true
	}

	// sort the coins post validation
	b.Coins = b.Coins.Sort()

	return nil
}

// SanitizeGenesisBalances sorts addresses and coin sets.
func SanitizeGenesisBalances(balances []Balance) []Balance {
	sort.Slice(balances, func(i, j int) bool {
		addr1, _ := sdk.AccAddressFromBech32(balances[i].Address)
		addr2, _ := sdk.AccAddressFromBech32(balances[j].Address)
		return bytes.Compare(addr1.Bytes(), addr2.Bytes()) < 0
	})

	for _, balance := range balances {
		balance.Coins = balance.Coins.Sort()
	}

	return balances
}

// GenesisBalancesIterator implements genesis account iteration.
type GenesisBalancesIterator struct{}

// IterateGenesisBalances iterates over all the genesis balances found in
// appGenesis and invokes a callback on each genesis account. If any call
// returns true, iteration stops.
func (GenesisBalancesIterator) IterateGenesisBalances(
	cdc codec.JSONMarshaler, appState map[string]json.RawMessage, cb func(exported.GenesisBalance) (stop bool),
) {
	for _, balance := range GetGenesisStateFromAppState(cdc, appState).Balances {
		if cb(balance) {
			break
		}
	}
}
