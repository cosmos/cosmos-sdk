package mock

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
	// "github.com/stretchr/testify/require"
	// "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	abci "github.com/tendermint/abci/types"
)

func SingleModuleTest(t *testing.T, module TestModule, keeper interface{}, size []int) {
	maxSize := getMaxValue(size)
	keys, addrs := GenerateNPrivKeyAddressPairs(maxSize)
	blocksize := 20 // Arbitrary. Later replace this with some sort of normal distribution
	numblocks := 20 // Same as above
	numRuns := 10
	cdc := wire.NewCodec()
	wire.RegisterCrypto(cdc)
	module.RegisterWire(cdc)

	for i := 0; i < numRuns; i++ {
		app, cleanup, err := setupMockApp()
		app.Cdc = cdc
		assert.Nil(t, err)
		time := time.Now().UnixNano()
		log := "Starting SingleModuleTest with randomness created with seed " + strconv.Itoa(int(time))
		r := rand.New(rand.NewSource(time))

		keeperStore := make(map[string]KeeperStorage)
		moduleKey := module.RandomSetup(t, r, app, size, keys, keeperStore)

		app.MountStoresIAVL(app.KeyMain, app.KeyAccountStore, keeperStore[moduleKey].Key)
		// TODO: Eventually abstract the following to support more ante handlers.
		app.SetAnteHandler(auth.NewAnteHandler(app.AccountMapper, app.FeeCollectionKeeper))
		// TODO: Randomize number of denominatioons
		app.SetInitChainer(createInitChain(r, app, addrs, []string{"FooCoin"},
			[]TestModule{module}, keeperStore))

		// sanity check. Must allow no-transition invariant assertion
		module.AssertInvariants(t, log, keeper)

		header := abci.Header{
			AppHash: []byte("apphash"),
			Height:  0,
		}
		for i := 0; i < numblocks; i++ {
			app.BeginBlock(abci.RequestBeginBlock{})

			ctx := app.NewContext(false, header)
			for j := 0; j < blocksize; j++ {
				tx, newLog := module.RandOperation(r)(module, r, ctx, keeperStore[moduleKey].Keeper)
				app.Deliver(tx)
				module.AssertInvariants(t, log, keeper)
				log += newLog + "\n"
			}
			app.EndBlock(abci.RequestEndBlock{})
			app.Commit()
			incrementHeight(header)

			// sanity check. Must allow no-transition invariant assertion
			module.AssertInvariants(t, log, keeper)
		}

		cleanup()
	}
}

func createInitChain(r *rand.Rand, app *MockApp, addrs []sdk.Address, coinDenoms []string,
	mods []TestModule, keeperStore map[string]KeeperStorage) sdk.InitChainer {
	return func(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
		accts := CreateRandomGenesisAccounts(r, addrs, coinDenoms)
		for _, act := range accts {
			app.AccountMapper.SetAccount(ctx, &act)
		}

		// Application specific genesis handling
		for _, mod := range mods {
			err := mod.InitGenesis(ctx, keeperStore)
			if err != nil {
				panic(err)
			}
		}
		return abci.ResponseInitChain{}
	}
}

func MasterTest(t *testing.T, modules []TestModule, sizes [][]int, coinDenoms []string, numValidators int) {

}

// func initChain(bapp *baseapp.BaseApp, accs [])

func incrementHeight(head abci.Header) {
	head.Height++
}

func getMaxValue(vals []int) int {
	m := vals[0]
	for _, e := range vals {
		if e < m {
			m = e
		}
	}
	return m
}
