package state

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KVStore = sdk.KVStore
type Context = sdk.Context
type Proof = merkle.Proof
type Codec = codec.Codec

type ABCIQuerier interface {
	QueryABCI(req abci.RequestQuery) (abci.ResponseQuery, error)
}
