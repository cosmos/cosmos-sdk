package merkle

import (
	"errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
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
	req := path.RequestQueryMultiStore(key)
	// BaseApp switches over the first path element,
	// and performs KVStore query only if it is "/store"
	req.Path = "/store" + req.Path
	return req
}

func (path Path) RequestQueryMultiStore(key []byte) abci.RequestQuery {
	// Suffixing path with "/key".
	// iavl.Store.Query() switches over the last path element,
	// and performs key-value query only if it is "/key"
	return abci.RequestQuery{Path: path.Path() + "/key", Data: path.Key(key), Prove: true}
}

func (path Path) Query(ctx context.CLIContext, key []byte) (code uint32, value []byte, proof Proof, err error) {
	resp, err := ctx.QueryABCI(path.RequestQuery(key))
	if err != nil {
		return code, value, proof, err
	}
	if !resp.IsOK() {
		return resp.Code, value, proof, errors.New(resp.Log)
	}
	return resp.Code, resp.Value, Proof{Key: key, Proof: resp.Proof}, nil
}

func (path Path) QueryValue(ctx context.CLIContext, cdc *codec.Codec, key []byte, ptr interface{}) (Proof, error) {
	_, value, proof, err := path.Query(ctx, key)
	if err != nil {
		return Proof{}, err
	}
	err = cdc.UnmarshalBinaryBare(value, ptr) // TODO
	return proof, err
}

func (path Path) QueryMultiStore(cms types.CommitMultiStore, key []byte) (uint32, []byte, Proof) {
	queryable, ok := cms.(types.Queryable)
	if !ok {
		panic("CommitMultiStore not queryable")
	}
	qres := queryable.Query(path.RequestQueryMultiStore(key))
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

func (path Path) Path() string {
	pathstr := ""
	for _, inter := range path.KeyPath {
		// The Queryable() stores uses slash-separated keypath format for querying
		pathstr = pathstr + "/" + string(inter)
	}

	return pathstr
}
