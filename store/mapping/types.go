package mapping

import (
	"github.com/tendermint/tendermint/crypto/merkle"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KVStore = sdk.KVStore
type Context = sdk.Context

type KeyPath = merkle.KeyPath

const KeyEncodingHex = merkle.KeyEncodingHex
