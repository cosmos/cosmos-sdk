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
	// Given that this function sorts balances, using the standard library's
	// Quicksort based algorithms, we have algorithmic complexities of:
	// * Best case: O(nlogn)
	// * Worst case: O(n^2)
	// The comparator used MUST be cheap to use lest we incur expenses like we had
	// before whereby sdk.AccAddressFromBech32, which is a very expensive operation
	// compared n * n elements yet discarded computations each time, as per:
	//  https://github.com/cosmos/cosmos-sdk/issues/7766#issuecomment-786671734
	// with this change the first step is to extract out and singly produce the values
	// that'll be used for comparisons and keep them cheap and fast.

	// 1. Retrieve the byte equivalents for each Balance's address and maintain a mapping of
	// its Balance, as the mapper will be used in sorting.
	type addrToBalance struct {
		// We use a pointer here to avoid averse effects of value copying
		// wasting RAM all around.
		balance *Balance
		accAddr []byte
	}
	adL := make([]*addrToBalance, 0, len(balances))
	for _, balance := range balances {
		balance := balance
		addr, _ := sdk.AccAddressFromBech32(balance.Address)
		adL = append(adL, &addrToBalance{
			balance: &balance,
			accAddr: []byte(addr),
		})
	}

	// 2. Sort with the cheap mapping, using the mapper's
	// already accAddr bytes values which is a cheaper operation.
	sort.Slice(adL, func(i, j int) bool {
		ai, aj := adL[i], adL[j]
		return bytes.Compare(ai.accAddr, aj.accAddr) < 0
	})

	// 3. To complete the sorting, we have to now just insert
	// back the balances in the order returned by the sort.
	for i, ad := range adL {
		balances[i] = *ad.balance
	}
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
