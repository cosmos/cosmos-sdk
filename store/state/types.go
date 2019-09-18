package state

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	stypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KVStore = sdk.KVStore
type Context = sdk.Context
type Proof = merkle.Proof
type Codec = codec.Codec

type ABCIQuerier interface {
	Query(storeName string, key []byte) (abci.ResponseQuery, error)
}

var _ ABCIQuerier = CLIQuerier{}

type CLIQuerier struct {
	ctx context.CLIContext
}

func NewCLIQuerier(ctx context.CLIContext) CLIQuerier {
	return CLIQuerier{ctx}
}

func (q CLIQuerier) Query(storeName string, key []byte) (abci.ResponseQuery, error) {
	req := abci.RequestQuery{
		Path:  "/store/" + storeName + "/key",
		Data:  key,
		Prove: true,
	}

	return q.ctx.QueryABCI(req)
}

var _ ABCIQuerier = StoreQuerier{}

type StoreQuerier struct {
	store stypes.Queryable
}

func NewStoreQuerier(store stypes.Queryable) StoreQuerier {
	return StoreQuerier{store}
}

func (q StoreQuerier) Query(storeName string, key []byte) (abci.ResponseQuery, error) {
	req := abci.RequestQuery{
		Path:  "/" + storeName + "/key",
		Data:  key,
		Prove: true,
	}

	return q.store.Query(req), nil

}
