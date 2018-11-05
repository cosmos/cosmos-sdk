package simulation

import (
	"fmt"
	"math/big"
	"math/rand"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
)

// shamelessly copied from https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang#31832326
// TODO we should probably move this to tendermint/libs/common/random.go

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Generate a random string of a particular length
func RandStringOfLength(r *rand.Rand, n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, r.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = r.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	return string(b)
}

// Pretty-print events as a table
func DisplayEvents(events map[string]uint) {
	var keys []string
	for key := range events {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	fmt.Printf("Event statistics: \n")
	for _, key := range keys {
		fmt.Printf("  % 60s => %d\n", key, events[key])
	}
}

// RandomAcc pick a random account from an array
func RandomAcc(r *rand.Rand, accs []Account) Account {
	return accs[r.Intn(
		len(accs),
	)]
}

// Generate a random amount
func RandomAmount(r *rand.Rand, max sdk.Int) sdk.Int {
	return sdk.NewInt(int64(r.Intn(int(max.Int64()))))
}

// RandomDecAmount generates a random decimal amount
func RandomDecAmount(r *rand.Rand, max sdk.Dec) sdk.Dec {
	randInt := big.NewInt(0).Rand(r, max.Int)
	return sdk.NewDecFromBigIntWithPrec(randInt, sdk.Precision)
}

// RandomAccounts generates n random accounts
func RandomAccounts(r *rand.Rand, n int) []Account {
	accs := make([]Account, n)
	for i := 0; i < n; i++ {
		// don't need that much entropy for simulation
		privkeySeed := make([]byte, 15)
		r.Read(privkeySeed)
		useSecp := r.Int63()%2 == 0
		if useSecp {
			accs[i].PrivKey = secp256k1.GenPrivKeySecp256k1(privkeySeed)
		} else {
			accs[i].PrivKey = ed25519.GenPrivKeyFromSecret(privkeySeed)
		}
		accs[i].PubKey = accs[i].PrivKey.PubKey()
		accs[i].Address = sdk.AccAddress(accs[i].PubKey.Address())
	}
	return accs
}

// Builds a function to add logs for this particular block
func addLogMessage(testingmode bool, blockLogBuilders []*strings.Builder, height int) func(string) {
	if testingmode {
		blockLogBuilders[height] = &strings.Builder{}
		return func(x string) {
			(*blockLogBuilders[height]).WriteString(x)
			(*blockLogBuilders[height]).WriteString("\n")
		}
	}
	return func(x string) {}
}

// assertAllInvariants asserts a list of provided invariants against application state
func assertAllInvariants(t *testing.T, app *baseapp.BaseApp,
	invariants []Invariant, where string, displayLogs func()) {

	for i := 0; i < len(invariants); i++ {
		err := invariants[i](app)
		if err != nil {
			fmt.Printf("Invariants broken after %s\n", where)
			fmt.Println(err.Error())
			displayLogs()
			t.Fatal()
		}
	}
}

// RandomSetGenesis wraps mock.RandomSetGenesis, but using simulation accounts
func RandomSetGenesis(r *rand.Rand, app *mock.App, accs []Account, denoms []string) {
	addrs := make([]sdk.AccAddress, len(accs))
	for i := 0; i < len(accs); i++ {
		addrs[i] = accs[i].Address
	}
	mock.RandomSetGenesis(r, app, addrs, denoms)
}

// Creates a function to print out the logs
func logPrinter(testingmode bool, logs []*strings.Builder) func() {
	if testingmode {
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
				fileName := fmt.Sprintf("simulation_log_%s.txt", time.Now().Format("2006-01-02 15:04:05"))
				fmt.Printf("Too many logs to display, instead writing to %s\n", fileName)
				f, _ = os.Create(fileName)
			}
			for i := 0; i < numLoggers; i++ {
				if f != nil {
					_, err := f.WriteString(fmt.Sprintf("Begin block %d\n", i+1))
					if err != nil {
						panic("Failed to write logs to file")
					}
					_, err = f.WriteString((*logs[i]).String())
					if err != nil {
						panic("Failed to write logs to file")
					}
				} else {
					fmt.Printf("Begin block %d\n", i+1)
					fmt.Println((*logs[i]).String())
				}
			}
		}
	}
	return func() {}
}
