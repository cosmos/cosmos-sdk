package merkle

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store/types"
)

func (path Path) RequestQuery(key []byte) abci.RequestQuery {
	pathstr := ""
	for _, inter := range path.KeyPath {
		pathstr = pathstr + "/" + string(inter)
	}
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
	return append(path.KeyPrefix, key...) // XXX: cloneAppend
}
