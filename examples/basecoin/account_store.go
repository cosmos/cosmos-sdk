package main

import (
	"encoding/json"
	"path"

	"github.com/cosmos/cosmos-sdk/types"
)

type AppAccountStore struct {
	kvStore types.KVStore
}

func newAccountStore(kvStore types.KVStore) types.AccountStore {
	return AppAccountStore{kvStore}
}

func (accStore AppAccountStore) NewAccountWithAddress(addr types.Address) types.Account {
	return &AppAccount{
		Address_: addr,
	}
}

func (accStore AppAccountStore) GetAccount(addr types.Address) types.Account {
	v := accStore.kvStore.Get(keyAccount(addr))

	if len(v) == 0 {
		return nil
	}

	acc := new(AppAccount)
	if err := json.Unmarshal(v, acc); err != nil {
		panic(err)
	}

	return acc
}

func (accStore AppAccountStore) SetAccount(acc types.Account) {
	b, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}

	appAcc, ok := acc.(*AppAccount)
	if !ok {
		panic("acc is not *AppAccount") // XXX
	}

	accStore.kvStore.Set(keyAccount(appAcc.Address_), b)
}

func keyAccount(addr types.Address) []byte {
	return []byte(path.Join("account", string(addr)))
}
