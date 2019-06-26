package merkle

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store/types"
)

func (root Root) RequestQuery(key []byte) abci.RequestQuery {
	req := root.RequestQueryMultiStore(key)
	req.Path = "/store" + req.Path
	return req
}

func (root Root) RequestQueryMultiStore(key []byte) abci.RequestQuery {
	return abci.RequestQuery{Path: root.Path() + "/key", Data: root.Key(key), Prove: true}
}

func (root Root) QueryMultiStore(cms types.CommitMultiStore, key []byte) (uint32, []byte, Proof) {
	qres := cms.(types.Queryable).Query(root.RequestQueryMultiStore(key))
	return qres.Code, qres.Value, Proof{Key: key, Proof: qres.Proof}
}

func (root Root) Key(key []byte) []byte {
	return append(root.KeyPrefix, key...) // XXX: cloneAppend
}

func (root Root) Path() string {
	path := ""
	for _, inter := range root.KeyPath {
		path = path + "/" + string(inter)
	}

	return path
}
