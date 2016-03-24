package state

import (
	"github.com/tendermint/basecoin/types"
	"sort"
)

type AccountCache struct {
	state    *State
	accounts map[string]*types.Account
}

func NewAccountCache(state *State) *AccountCache {
	return &AccountCache{
		state:    state,
		accounts: make(map[string]*types.Account),
	}
}

func (cache *AccountCache) GetAccount(addr []byte) *types.Account {
	acc, ok := cache.accounts[string(addr)]
	if !ok {
		acc = cache.state.GetAccount(addr)
		cache.accounts[string(addr)] = acc
	}
	return acc
}

func (cache *AccountCache) SetAccount(addr []byte, acc *types.Account) {
	cache.accounts[string(addr)] = acc
}

func (cache *AccountCache) Sync() {
	// MUST BE DETERMINISTIC
	// First, order the addrs.
	addrs := []string{}
	for addr := range cache.accounts {
		addrs = append(addrs, string(addr))
	}
	sort.Strings(addrs)

	// Set the accounts in order.
	for _, addr := range addrs {
		cache.state.SetAccount([]byte(addr), cache.accounts[addr])
	}
}
