package mock

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/stretchr/testify/require"
	// bap "github.com/cosmos/cosmos-sdk/baseapp"
	// "github.com/cosmos/cosmos-sdk/x/auth"
	// abci "github.com/tendermint/abci/types"
	"github.com/cosmos/cosmos-sdk/wire"
	crypto "github.com/tendermint/go-crypto"
	"math/rand"
	"testing"
)

// any operation that transforms state takes in RNG instance, and its keeper.
// returns msg to include in block, and meaningful message for log.
type Operation func(mod TestModule, r *rand.Rand, ctx sdk.Context, keeper interface{}) (sdk.Tx, string)

type KeeperStorage struct {
	Keeper interface{}
	Key    *sdk.KVStoreKey

	Other []interface{}
}

type TestModule interface {
	AssertInvariants(t *testing.T, log string, keeper interface{}) (bool, string)
	// Sets keeper and needed interfaces in KeeperStore. Don't mutate privKeyList
	RandomSetup(t *testing.T, r *rand.Rand, app *MockApp, size []int, privkeyList []crypto.PrivKey,
		keeperStore map[string]KeeperStorage) (name string)
	RandOperation(r *rand.Rand) Operation
	RegisterWire(cdc *wire.Codec)
	// Optional
	InitGenesis(ctx sdk.Context, keeperStore map[string]KeeperStorage) sdk.Error
}

var modules = []TestModule{}

func RegisterModule(module TestModule) {
	modules = append(modules, module)
}

func ClearModules() {
	modules = []TestModule{}
}
