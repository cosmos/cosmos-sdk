package main

import (
	"encoding/json"
	"path"

	crypto "github.com/tendermint/go-crypto"

	"github.com/cosmos/cosmos-sdk/types"
	acm "github.com/cosmos/cosmos-sdk/x/account"
	"github.com/cosmos/cosmos-sdk/x/sendtx"
	"github.com/cosmos/cosmos-sdk/x/store"
)

func txParser(txBytes []byte) (types.Tx, error) {
	var tx sendtx.SendTx
	err := json.Unmarshal(txBytes, &tx)
	return tx, err
}

//-----------------------------------------------------------------------------

type AccountStore struct {
	kvStore types.KVStore
}

func newAccountStore(kvStore types.KVStore) store.AccountStore {
	return AccountStore{kvStore}
}

func (accStore AccountStore) NewAccountWithAddress(addr crypto.Address) store.Account {
	return acm.NewBaseAccountWithAddress(addr)
}

func (accStore AccountStore) GetAccount(addr crypto.Address) store.Account {
	v := accStore.kvStore.Get(keyAccount(addr))

	if len(v) == 0 {
		return nil
	}

	acc := new(acm.BaseAccount)
	if err := json.Unmarshal(v, acc); err != nil {
		panic(err)
	}

	return acc
}

func (accStore AccountStore) SetAccount(acc store.Account) {
	b, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}

	appAcc, ok := acc.(*acm.BaseAccount)
	if !ok {
		panic("acc is not *acm.BaseAccount") // XXX
	}

	accStore.kvStore.Set(keyAccount(appAcc.Address()), b)
}

func keyAccount(addr crypto.Address) []byte {
	return []byte(path.Join("account", string(addr)))
}
