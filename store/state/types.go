package state

import (
	"github.com/tendermint/tendermint/crypto/merkle"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type KVStore = sdk.KVStore
type Context = sdk.Context
type CLIContext = context.CLIContext
type Proof = merkle.Proof
type Codec = codec.Codec
