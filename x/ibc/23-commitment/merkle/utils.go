package merkle

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store/types"
)

// RequestQuery() constructs the abci.RequestQuery.
//
// RequestQuery.Path is a slash separated key list, ending with "/key"
//
// RequestQuery.Data is the concatanation of path.KeyPrefix and key argument
//
// RequestQuery.Prove is set to true
func (path Path) RequestQuery(key []byte) abci.RequestQuery {
	pathstr := ""
	for _, inter := range path.KeyPath {
		// The Queryable() stores uses slash-separated keypath format for querying
		pathstr = pathstr + "/" + string(inter)
	}
	// Suffixing pathstr with "/key".
	// iavl.Store.Query() switches over the last path element,
	// and performs key-value query only if it is "/key"
	pathstr = pathstr + "/key"

	data := append(path.KeyPrefix, key...)

	return abci.RequestQuery{Path: pathstr, Data: data, Prove: true}
}

func (path Path) Query(cms types.CommitMultiStore, key []byte) (uint32, []byte, Proof) {
	queryable, ok := cms.(types.Queryable)
	if !ok {
		panic("CommitMultiStore not queryable")
	}
	qres := queryable.Query(path.RequestQuery(key))
	return qres.Code, qres.Value, Proof{Key: key, Proof: qres.Proof}
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
