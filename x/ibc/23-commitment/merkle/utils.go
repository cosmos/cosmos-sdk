package merkle

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store/types"
)

func (path Path) RequestQuery(key []byte) abci.RequestQuery {
	req := abci.RequestQuery{Path: "/store" + path.Path() + "/key", Data: key, Prove: true}
	return req
}

func (path Path) RequestQueryMultiStore(key []byte) abci.RequestQuery {
	return abci.RequestQuery{Path: path.Path() + "/key", Data: path.Key(key), Prove: true}
}

func (path Path) QueryMultiStore(cms types.CommitMultiStore, key []byte) (uint32, []byte, Proof) {
	qres := cms.(types.Queryable).Query(path.RequestQueryMultiStore(key))
	return qres.Code, qres.Value, Proof{Key: path.Key(key), Proof: qres.Proof}
}

func (path Path) Key(key []byte) []byte {
	return append(path.KeyPrefix, key...) // XXX: cloneAppend
}

func (path Path) Path() string {
	pathstr := ""
	for _, inter := range path.KeyPath {
		pathstr = pathstr + "/" + string(inter)
	}

	return pathstr
}
