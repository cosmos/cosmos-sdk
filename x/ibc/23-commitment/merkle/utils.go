package merkle

import (
	"errors"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/store/types"
	ics23 "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment"
)

func QueryMultiStore(cms types.CommitMultiStore, storeName string, prefix []byte, key []byte) ([]byte, Proof, error) {
	queryable, ok := cms.(types.Queryable)
	if !ok {
		panic("commitMultiStore not queryable")
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
		Data:  ics23.Join(prefix, key),
		Prove: true,
	}
}
