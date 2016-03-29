package types

import (
	"sort"
)

type AccountCache struct {
	state    AccountGetterSetter
	accounts map[string]*Account
}

func NewAccountCache(state AccountGetterSetter) *AccountCache {
	return &AccountCache{
		state:    state,
		accounts: make(map[string]*Account),
	}
}

func (cache *AccountCache) GetAccount(addr []byte) *Account {
	acc, ok := cache.accounts[string(addr)]
	if !ok {
		acc = cache.state.GetAccount(addr)
		cache.accounts[string(addr)] = acc
	}
	return acc
}

func (cache *AccountCache) SetAccount(addr []byte, acc *Account) {
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

	// Reset accounts
	cache.accounts = make(map[string]*Account)
}

//----------------------------------------

// NOT USED
type AccountCacher interface {
	GetAccount(addr []byte) *Account
	SetAccount(addr []byte, acc *Account)
	Sync()
}
