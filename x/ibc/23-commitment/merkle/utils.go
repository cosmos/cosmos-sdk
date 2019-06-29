package merkle

import (
	"bytes"
	"errors"
	"fmt"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/client/context"
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

// QueryMultiStore prefixes the root.KeyPrefix to the key
func (root Root) QueryMultiStore(cms types.CommitMultiStore, key []byte) (qres abci.ResponseQuery, proof Proof, err error) {
	qres = cms.(types.Queryable).Query(root.RequestQueryMultiStore(key))
	proof = Proof{Key: key, Proof: qres.Proof}
	return
}

// TrimPrefix is used when the input key is generate with prefix
// so need to be trimmed
func (root Root) trimPrefix(key []byte) ([]byte, error) {
	if !bytes.HasPrefix(key, root.KeyPrefix) {
		fmt.Println(string(key), key, string(root.KeyPrefix), root.KeyPrefix)
		return nil, errors.New("key not starting with root key prefix")
	}
	return bytes.TrimPrefix(key, root.KeyPrefix), nil
}

// QueryCLI does NOT prefixes the root.KeyPrefix to the key
func (root Root) QueryCLI(ctx context.CLIContext, key []byte) (qres abci.ResponseQuery, proof Proof, err error) {
	qres, err = ctx.QueryABCI(root.RequestQueryMultiStore(key))
	if err != nil {
		return
	}
	trimkey, err := root.trimPrefix(key)
	if err != nil {
		return
	}
	proof = Proof{Key: trimkey, Proof: qres.Proof}
	return
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
