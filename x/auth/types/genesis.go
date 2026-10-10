package types

import (
	"encoding/json"
	"fmt"
	"slices"
	"sort"

	proto "github.com/cosmos/gogoproto/proto"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
)

var _ types.UnpackInterfacesMessage = GenesisState{}

// RandomGenesisAccountsFn defines the function required to generate custom account types
type RandomGenesisAccountsFn func(simState *module.SimulationState) GenesisAccounts

// NewGenesisState - Create a new genesis state
func NewGenesisState(params Params, accounts GenesisAccounts) *GenesisState {
	genAccounts, err := PackAccounts(accounts)
	if err != nil {
		panic(err)
	}
	return &GenesisState{
		Params:   params,
		Accounts: genAccounts,
	}
}

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (g GenesisState) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	for _, any := range g.Accounts {
		var account GenesisAccount
		err := unpacker.UnpackAny(any, &account)
		if err != nil {
			return err
		}
	}
	for i := range g.PubKeyHistory {
		if err := g.PubKeyHistory[i].UnpackInterfaces(unpacker); err != nil {
			return err
		}
	}
	return nil
}

// DefaultGenesisState - Return a default genesis state
func DefaultGenesisState() *GenesisState {
	return NewGenesisState(DefaultParams(), GenesisAccounts{})
}

// GetGenesisStateFromAppState returns x/auth GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.Codec, appState map[string]json.RawMessage) GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return genesisState
}

// ValidateGenesis performs basic validation of auth genesis data returning an
// error for any failed validation criteria.
func ValidateGenesis(data GenesisState) error {
	if err := data.Params.Validate(); err != nil {
		return err
	}

	genAccs, err := UnpackAccounts(data.Accounts)
	if err != nil {
		return err
	}

	if err := ValidateGenAccounts(genAccs); err != nil {
		return err
	}

	accsByAddr := make(map[string]GenesisAccount, len(genAccs))
	for _, acc := range genAccs {
		accsByAddr[acc.GetAddress().String()] = acc
	}

	rekeyed, err := validatePubKeyHistory(data.PubKeyHistory, accsByAddr)
	if err != nil {
		return err
	}

	// An account whose pubkey does not hash to its address must have been
	// rekeyed, so it must have a pubkey history.
	for _, acc := range genAccs {
		pk := acc.GetPubKey()
		if pk == nil {
			continue
		}
		if _, ok := pk.(*ModuleCredential); ok {
			continue
		}
		addr := acc.GetAddress()
		pkAddr, err := pubKeyAddress(pk)
		if err != nil {
			return fmt.Errorf("invalid account found in genesis state; address: %s, error: %w", addr, err)
		}
		if pkAddr.Equals(addr) {
			continue
		}
		if !rekeyed[addr.String()] {
			return fmt.Errorf("invalid account found in genesis state; address: %s, error: pubkey address does not match and account has no pubkey history", addr)
		}
	}

	return nil
}

// validatePubKeyHistory checks the genesis pubkey history against the genesis
// accounts (keyed by bech32 address) and returns the set of addresses (bech32)
// that have history.
//
// Each history must belong to an account in genesis whose stored pubkey is
// neither nil nor a ModuleCredential, since only such accounts can rekey. No
// entry's pubkey may be a ModuleCredential, and every pubkey must be well
// formed enough to derive an address. The first entry's pubkey must hash to the
// account address, since it is the account's original key. The entries must
// form a chain: each entry's new_key_address is the natural address of the next
// entry's pubkey, and the last entry's new_key_address is the natural address
// of the account's current pubkey.
func validatePubKeyHistory(history []GenesisPubKeyHistory, accsByAddr map[string]GenesisAccount) (map[string]bool, error) {
	seen := make(map[string]bool, len(history))
	for _, h := range history {
		addr, err := sdk.AccAddressFromBech32(h.Address)
		if err != nil {
			return nil, fmt.Errorf("invalid pubkey history address %q: %w", h.Address, err)
		}
		addrStr := addr.String()
		if seen[addrStr] {
			return nil, fmt.Errorf("duplicate pubkey history found in genesis state; address: %s", addrStr)
		}
		seen[addrStr] = true

		if len(h.Entries) == 0 {
			return nil, fmt.Errorf("pubkey history for %s has no entries", addrStr)
		}

		acc, ok := accsByAddr[addrStr]
		if !ok {
			return nil, fmt.Errorf("pubkey history for %s has no account in genesis state", addrStr)
		}
		currentPk := acc.GetPubKey()
		if currentPk == nil {
			return nil, fmt.Errorf("pubkey history for %s but the account has no pubkey", addrStr)
		}
		if _, ok := currentPk.(*ModuleCredential); ok {
			return nil, fmt.Errorf("pubkey history for %s but the account's pubkey is a ModuleCredential", addrStr)
		}
		currentPkAddr, err := pubKeyAddress(currentPk)
		if err != nil {
			return nil, fmt.Errorf("pubkey history for %s: account pubkey: %w", addrStr, err)
		}

		// Entries are in rotation order. Heights may repeat (two rotations in
		// one block) or decrease (rotations after a zero-height export).
		newKeyAddrs := make([]sdk.AccAddress, len(h.Entries))
		pubKeyAddrs := make([]sdk.AccAddress, len(h.Entries))
		for i, entry := range h.Entries {
			if entry.ReplacedAtHeight < 0 {
				return nil, fmt.Errorf("pubkey history for %s has negative height %d", addrStr, entry.ReplacedAtHeight)
			}
			if entry.PubKey == nil {
				return nil, fmt.Errorf("pubkey history for %s entry %d has no pubkey", addrStr, i)
			}
			pk, ok := entry.PubKey.GetCachedValue().(cryptotypes.PubKey)
			if !ok || pk == nil {
				return nil, fmt.Errorf("pubkey history for %s entry %d has no pubkey: cannot unpack %s", addrStr, i, entry.PubKey.TypeUrl)
			}
			if _, ok := pk.(*ModuleCredential); ok {
				return nil, fmt.Errorf("pubkey history for %s entry %d is a ModuleCredential", addrStr, i)
			}
			pubKeyAddrs[i], err = pubKeyAddress(pk)
			if err != nil {
				return nil, fmt.Errorf("pubkey history for %s entry %d: %w", addrStr, i, err)
			}
			newKeyAddrs[i], err = sdk.AccAddressFromBech32(entry.NewKeyAddress)
			if err != nil {
				return nil, fmt.Errorf("pubkey history for %s entry %d has invalid new key address %q: %w", addrStr, i, entry.NewKeyAddress, err)
			}
		}

		// The first entry holds the account's original key, which
		// SetPubKeyDecorator only stores if it hashes to the account address.
		if !pubKeyAddrs[0].Equals(addr) {
			return nil, fmt.Errorf("pubkey history for %s entry 0 pubkey address %s does not match the account address", addrStr, pubKeyAddrs[0])
		}

		for i, newKeyAddr := range newKeyAddrs {
			if i+1 < len(pubKeyAddrs) {
				if !newKeyAddr.Equals(pubKeyAddrs[i+1]) {
					return nil, fmt.Errorf("pubkey history for %s entry %d new key address %s does not match the next entry's pubkey", addrStr, i, newKeyAddr)
				}
				continue
			}
			if !newKeyAddr.Equals(currentPkAddr) {
				return nil, fmt.Errorf("pubkey history for %s entry %d new key address %s does not match the account's current pubkey", addrStr, i, newKeyAddr)
			}
		}
	}
	return seen, nil
}

// pubKeyAddress returns pk's natural address. Address() panics on some
// malformed keys, such as a secp256k1 key of the wrong length, which a
// hand-edited genesis can contain, so the panic is returned as an error.
func pubKeyAddress(pk cryptotypes.PubKey) (addr sdk.AccAddress, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("malformed pubkey %T: %v", pk, r)
		}
	}()
	return sdk.AccAddress(pk.Address()), nil
}

// SanitizeGenesisAccounts sorts accounts and coin sets.
func SanitizeGenesisAccounts(genAccs GenesisAccounts) GenesisAccounts {
	// Make sure there aren't any duplicated account numbers by fixing the duplicates with the lowest unused values.
	// seenAccNum = easy lookup for used account numbers.
	seenAccNum := map[uint64]bool{}
	// dupAccNum = a map of account number to accounts with duplicate account numbers (excluding the 1st one seen).
	dupAccNum := map[uint64]GenesisAccounts{}
	for _, acc := range genAccs {
		num := acc.GetAccountNumber()
		if !seenAccNum[num] {
			seenAccNum[num] = true
		} else {
			dupAccNum[num] = append(dupAccNum[num], acc)
		}
	}

	// dupAccNums a sorted list of the account numbers with duplicates.
	var dupAccNums []uint64
	for num := range dupAccNum {
		dupAccNums = append(dupAccNums, num)
	}
	slices.Sort(dupAccNums)

	// Change the account number of the duplicated ones to the first unused value.
	globalNum := uint64(0)
	for _, dupNum := range dupAccNums {
		accs := dupAccNum[dupNum]
		for _, acc := range accs {
			for seenAccNum[globalNum] {
				globalNum++
			}
			if err := acc.SetAccountNumber(globalNum); err != nil {
				panic(err)
			}
			seenAccNum[globalNum] = true
		}
	}

	// Then sort them all by account number.
	sort.Slice(genAccs, func(i, j int) bool {
		return genAccs[i].GetAccountNumber() < genAccs[j].GetAccountNumber()
	})
	return genAccs
}

// ValidateGenAccounts validates an array of GenesisAccounts and checks for duplicates
func ValidateGenAccounts(accounts GenesisAccounts) error {
	addrMap := make(map[string]bool, len(accounts))

	for _, acc := range accounts {
		// check for duplicated accounts
		addrStr := acc.GetAddress().String()
		if _, ok := addrMap[addrStr]; ok {
			return fmt.Errorf("duplicate account found in genesis state; address: %s", addrStr)
		}

		addrMap[addrStr] = true

		// check account specific validation
		if err := acc.Validate(); err != nil {
			return fmt.Errorf("invalid account found in genesis state; address: %s, error: %s", addrStr, err.Error())
		}
	}
	return nil
}

// GenesisAccountIterator implements genesis account iteration.
type GenesisAccountIterator struct{}

// IterateGenesisAccounts iterates over all the genesis accounts found in
// appGenesis and invokes a callback on each genesis account. If any call
// returns true, iteration stops.
func (GenesisAccountIterator) IterateGenesisAccounts(
	cdc codec.Codec, appGenesis map[string]json.RawMessage, cb func(sdk.AccountI) (stop bool),
) {
	for _, genAcc := range GetGenesisStateFromAppState(cdc, appGenesis).Accounts {
		acc, ok := genAcc.GetCachedValue().(sdk.AccountI)
		if !ok {
			panic("expected account")
		}
		if cb(acc) {
			break
		}
	}
}

// PackAccounts converts GenesisAccounts to Any slice
func PackAccounts(accounts GenesisAccounts) ([]*types.Any, error) {
	accountsAny := make([]*types.Any, len(accounts))
	for i, acc := range accounts {
		msg, ok := acc.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("cannot proto marshal %T", acc)
		}
		any, err := types.NewAnyWithValue(msg)
		if err != nil {
			return nil, err
		}
		accountsAny[i] = any
	}

	return accountsAny, nil
}

// UnpackAccounts converts Any slice to GenesisAccounts
func UnpackAccounts(accountsAny []*types.Any) (GenesisAccounts, error) {
	accounts := make(GenesisAccounts, len(accountsAny))
	for i, any := range accountsAny {
		acc, ok := any.GetCachedValue().(GenesisAccount)
		if !ok {
			return nil, fmt.Errorf("expected genesis account")
		}
		accounts[i] = acc
	}

	return accounts, nil
}
