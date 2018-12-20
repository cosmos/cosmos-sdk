package simulation

import (
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

func getTestingMode(tb testing.TB) (testingMode bool, t *testing.T, b *testing.B) {
	testingMode = false
	if _t, ok := tb.(*testing.T); ok {
		t = _t
		testingMode = true
	} else {
		b = tb.(*testing.B)
	}
	return
}

// Builds a function to add logs for this particular block
func addLogMessage(testingmode bool,
	blockLogBuilders []*strings.Builder, height int) func(string) {

	if !testingmode {
		return func(_ string) {}
	}

	blockLogBuilders[height] = &strings.Builder{}
	return func(x string) {
		(*blockLogBuilders[height]).WriteString(x)
		(*blockLogBuilders[height]).WriteString("\n")
	}
}

// Creates a function to print out the logs
func logPrinter(testingmode bool, logs []*strings.Builder) func() {
	if !testingmode {
		return func() {}
	}

	return func() {
		numLoggers := 0
		for i := 0; i < len(logs); i++ {
			// We're passed the last created block
			if logs[i] == nil {
				numLoggers = i
				break
			}
		}

		var f *os.File
		if numLoggers > 10 {
			fileName := fmt.Sprintf("simulation_log_%s.txt",
				time.Now().Format("2006-01-02 15:04:05"))
			fmt.Printf("Too many logs to display, instead writing to %s\n",
				fileName)
			f, _ = os.Create(fileName)
		}

		for i := 0; i < numLoggers; i++ {
			if f == nil {
				fmt.Printf("Begin block %d\n", i+1)
				fmt.Println((*logs[i]).String())
				continue
			}

			_, err := f.WriteString(fmt.Sprintf("Begin block %d\n", i+1))
			if err != nil {
				panic("Failed to write logs to file")
			}

			_, err = f.WriteString((*logs[i]).String())
			if err != nil {
				panic("Failed to write logs to file")
			}
		}
	}
}

// getBlockSize returns a block size as determined from the transition matrix.
// It targets making average block size the provided parameter. The three
// states it moves between are:
//  - "over stuffed" blocks with average size of 2 * avgblocksize,
//  - normal sized blocks, hitting avgBlocksize on average,
//  - and empty blocks, with no txs / only txs scheduled from the past.
func getBlockSize(r *rand.Rand, params Params,
	lastBlockSizeState, avgBlockSize int) (state, blocksize int) {

	// TODO: Make default blocksize transition matrix actually make the average
	// blocksize equal to avgBlockSize.
	state = params.BlockSizeTransitionMatrix.NextState(r, lastBlockSizeState)
	switch state {
	case 0:
		blocksize = r.Intn(avgBlockSize * 4)
	case 1:
		blocksize = r.Intn(avgBlockSize * 2)
	default:
		blocksize = 0
	}
	return state, blocksize
}

// PeriodicInvariant returns an Invariant function closure that asserts a given
// invariant if the mock application's last block modulo the given period is
// congruent to the given offset.
//
// NOTE this function is intended to be used manually used while running
// computationally heavy simulations.
// TODO reference this function in the codebase probably through use of a switch
func PeriodicInvariant(invariant Invariant, period int, offset int) Invariant {
	return func(ctx sdk.Context) error {
		if int(ctx.BlockHeight())%period == offset {
			return invariant(ctx)
		}
		return nil
	}
}
