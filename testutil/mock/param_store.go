package mock

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/db"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type ParamStore struct {
	Txn db.ReadWriter
}

func NewParamStore(db db.Connection) *ParamStore {
	return &ParamStore{Txn: db.ReadWriter()}
}

func (ps *ParamStore) Set(_ sdk.Context, key []byte, value interface{}) {
	bz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	ps.Txn.Set(key, bz)
}

func (ps *ParamStore) Has(_ sdk.Context, key []byte) bool {
	ok, err := ps.Txn.Has(key)
	if err != nil {
		panic(err)
	}

	return ok
}

func (ps *ParamStore) Get(_ sdk.Context, key []byte, ptr interface{}) {
	bz, err := ps.Txn.Get(key)
	if err != nil {
		panic(err)
	}

	if len(bz) == 0 {
		return
	}

	if err := json.Unmarshal(bz, ptr); err != nil {
		panic(err)
	}
}
