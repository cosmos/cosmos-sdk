package mock

import (
	// "github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/abci/types"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// RandomizedTesting tests application by sending random messages.
func (app *App) RandomizedTesting(t *testing.T, ops []TestAndRunTx, setups []RandSetup,
	invariants []AssertInvariants, numKeys int, numBlocks int, blockSize int) {

	time := time.Now().UnixNano()
	log := "Starting SingleModuleTest with randomness created with seed " + strconv.Itoa(int(time))
	keys, addrs := GenerateNPrivKeyAddressPairs(numKeys)
	r := rand.New(rand.NewSource(time))
	for i := 0; i < len(setups); i++ {
		setups[i](r, keys)
	}
	RandomSetGenesis(r, app, addrs, []string{"foocoin"})
	header := abci.Header{Height: 0}
	for i := 0; i < numBlocks; i++ {
		// Make sure invariants hold at begnning of block / when nothing was done.
		app.assertAllInvariants(t, invariants, log)
		app.BeginBlock(abci.RequestBeginBlock{})
		ctx := app.NewContext(false, header)
		for j := 0; j < blockSize; j++ {
			logUpdate, err := ops[r.Intn(len(ops))](r, app, ctx)
			// Add this to log
			log += "\n" + logUpdate
			require.Nil(t, err, log)
			app.assertAllInvariants(t, invariants, log)
		}
		app.EndBlock(abci.RequestEndBlock{})
		header.Height++
	}
}

func (app *App) assertAllInvariants(t *testing.T, tests []AssertInvariants, log string) {
	for i := 0; i < len(tests); i++ {
		tests[i](t, app, log)
	}
}
