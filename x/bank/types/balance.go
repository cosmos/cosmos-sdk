package types

import (
	"bytes"
	"encoding/json"
	"fmt"
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

type balanceByAddress struct {
	addresses []sdk.AccAddress
	balances  []Balance
}

func (b balanceByAddress) Len() int { return len(b.addresses) }
func (b balanceByAddress) Less(i, j int) bool {
	return bytes.Compare(b.addresses[i], b.addresses[j]) < 0
}
func (b balanceByAddress) Swap(i, j int) {
	b.addresses[i], b.addresses[j] = b.addresses[j], b.addresses[i]
	b.balances[i], b.balances[j] = b.balances[j], b.balances[i]
}

// SanitizeGenesisBalances sorts addresses and coin sets.
func SanitizeGenesisBalances(balances []Balance) []Balance {
	// Given that this function sorts balances, using the standard library's
	// Quicksort based algorithms, we have algorithmic complexities of:
	// * Best case: O(nlogn)
	// * Worst case: O(n^2)
	// The comparator used MUST be cheap to use lest we incur expenses like we had
	// before whereby sdk.AccAddressFromBech32, which is a very expensive operation
	// compared n * n elements yet discarded computations each time, as per:
	//  https://github.com/cosmos/cosmos-sdk/issues/7766#issuecomment-786671734

	// 1. Retrieve the address equivalents for each Balance's address.
	addresses := make([]sdk.AccAddress, len(balances))
	for i := range balances {
		addr, _ := sdk.AccAddressFromBech32(balances[i].Address)
		addresses[i] = addr
	}

	// 2. Sort balances.
	sort.Sort(balanceByAddress{addresses: addresses, balances: balances})

	return balances
}

// GenesisBalancesIterator implements genesis account iteration.
type GenesisBalancesIterator struct{}

// IterateGenesisBalances iterates over all the genesis balances found in
// appGenesis and invokes a callback on each genesis account. If any call
// returns true, iteration stops.
func (GenesisBalancesIterator) IterateGenesisBalances(
	cdc codec.JSONCodec, appState map[string]json.RawMessage, cb func(exported.GenesisBalance) (stop bool),
) {
	for _, balance := range GetGenesisStateFromAppState(cdc, appState).Balances {
		if cb(balance) {
			break
		}
	}
}
