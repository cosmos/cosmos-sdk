package mock

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	db "github.com/tendermint/tm-db"
)

type ParamStore struct {
	Db db.DB
}

func (ps *ParamStore) Set(_ sdk.Context, key []byte, value interface{}) {
	bz, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}

	ps.Db.Set(key, bz)
}

func (ps *ParamStore) Has(_ sdk.Context, key []byte) bool {
	ok, err := ps.Db.Has(key)
	if err != nil {
		panic(err)
	}

	return ok
}

func (ps *ParamStore) Get(_ sdk.Context, key []byte, ptr interface{}) {
	bz, err := ps.Db.Get(key)
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
