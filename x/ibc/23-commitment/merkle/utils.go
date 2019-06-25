package merkle

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store/types"
)

func (root Root) RequestQuery(key []byte) abci.RequestQuery {
	path := ""
	for _, inter := range root.KeyPath {
		path = path + "/" + string(inter)
	}
	path = path + "/key"

	data := append(root.KeyPrefix, key...)

	return abci.RequestQuery{Path: path, Data: data, Prove: true}
}

func (root Root) Query(cms types.CommitMultiStore, key []byte) (uint32, []byte, Proof) {
	qres := cms.(types.Queryable).Query(root.RequestQuery(key))
	return qres.Code, qres.Value, Proof{Key: key, Proof: qres.Proof}
}

func (root Root) Key(key []byte) []byte {
	return append(root.KeyPrefix, key...) // XXX: cloneAppend
}
