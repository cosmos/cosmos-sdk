package account

import (
	"encoding/json"
	"path"

	"github.com/cosmos/cosmos-sdk/types"
	crypto "github.com/tendermint/go-crypto"
)

type AccountStore struct {
	kvStore types.KVStore
}

func NewAccountStore(kvStore types.KVStore) types.AccountStore {
	return AccountStore{kvStore}
}

func (accStore AccountStore) NewAccountWithAddress(addr crypto.Address) types.Account {
	return NewBaseAccountWithAddress(addr)
}

func (accStore AccountStore) GetAccount(addr crypto.Address) types.Account {
	v := accStore.kvStore.Get(keyAccount(addr))

	if len(v) == 0 {
		return nil
	}

	acc := new(BaseAccount)
	if err := json.Unmarshal(v, acc); err != nil {
		panic(err)
	}

	return acc
}

func (accStore AccountStore) SetAccount(acc types.Account) {
	b, err := json.Marshal(acc)
	if err != nil {
		panic(err)
	}

	appAcc, ok := acc.(*BaseAccount)
	if !ok {
		panic("acc is not *BaseAccount") // XXX
	}

	accStore.kvStore.Set(keyAccount(appAcc.Address()), b)
}

func keyAccount(addr crypto.Address) []byte {
	return []byte(path.Join("account", string(addr)))
}
