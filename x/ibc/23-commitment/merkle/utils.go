package merkle

import (
	"errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store/types"
)

func QueryMultiStore(cms types.CommitMultiStore, storeName string, prefix []byte, key []byte) ([]byte, Proof, error) {
	queryable, ok := cms.(types.Queryable)
	if !ok {
		panic("CommitMultiStore not queryable")
	}
	qres := queryable.Query(RequestQueryMultiStore(storeName, prefix, key))
	if !qres.IsOK() {
		return nil, Proof{}, errors.New(qres.Log)
	}

	return qres.Value, Proof{Key: key, Proof: qres.Proof}, nil
}

func RequestQueryMultiStore(storeName string, prefix []byte, key []byte) abci.RequestQuery {
	// Suffixing path with "/key".
	// iavl.Store.Query() switches over the last path element,
	// and performs key-value query only if it is "/key"
	return abci.RequestQuery{
		Path:  "/" + storeName + "/key",
		Data:  join(prefix, key),
		Prove: true,
	}
}

func (path Path) Key(key []byte) []byte {
	return join(path.KeyPrefix, key)
}

func join(a, b []byte) (res []byte) {
	res = make([]byte, len(a)+len(b))
	copy(res, a)
	copy(res[len(a):], b)
	return
}
