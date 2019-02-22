package simulation

import (
	"fmt"
	"math/big"
	"math/rand"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/mock"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// shamelessly copied from
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang#31832326
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

// Generate a random amount
// Note: The range of RandomAmount includes max, and is, in fact, biased to return max.
func RandomAmount(r *rand.Rand, max sdk.Int) sdk.Int {
	// return sdk.NewInt(int64(r.Intn(int(max.Int64()))))
	max2 := big.NewInt(0).Mul(max.BigInt(), big.NewInt(2))
	randInt := big.NewInt(0).Rand(r, max2)
	if randInt.Cmp(max.BigInt()) > 0 {
		randInt = max.BigInt()
	}
	result := sdk.NewIntFromBigInt(randInt)
	// Sanity
	if result.GT(max) {
		panic(fmt.Sprintf("%v > %v", result, max))
	}
	return result
}

// RandomDecAmount generates a random decimal amount
// Note: The range of RandomDecAmount includes max, and is, in fact, biased to return max.
func RandomDecAmount(r *rand.Rand, max sdk.Dec) sdk.Dec {
	// randInt := big.NewInt(0).Rand(r, max.Int)
	max2 := big.NewInt(0).Mul(max.Int, big.NewInt(2))
	randInt := big.NewInt(0).Rand(r, max2)
	if randInt.Cmp(max.Int) > 0 {
		randInt = max.Int
	}
	result := sdk.NewDecFromBigIntWithPrec(randInt, sdk.Precision)
	// Sanity
	if result.GT(max) {
		panic(fmt.Sprintf("%v > %v", result, max))
	}
	return result
}

// RandomSetGenesis wraps mock.RandomSetGenesis, but using simulation accounts
func RandomSetGenesis(r *rand.Rand, app *mock.App, accs []Account, denoms []string) {
	addrs := make([]sdk.AccAddress, len(accs))
	for i := 0; i < len(accs); i++ {
		addrs[i] = accs[i].Address
	}
	mock.RandomSetGenesis(r, app, addrs, denoms)
}

// RandTimestamp generates a random timestamp
func RandTimestamp(r *rand.Rand) time.Time {
	// json.Marshal breaks for timestamps greater with year greater than 9999
	unixTime := r.Int63n(253373529600)
	return time.Unix(unixTime, 0)
}

// Derive a new rand deterministically from a rand.
// Unlike rand.New(rand.NewSource(seed)), the result is "more random"
// depending on the source and state of r.
// NOTE: not crypto safe.
func DeriveRand(r *rand.Rand) *rand.Rand {
	const num = 8 // TODO what's a good number?  Too large is too slow.
	ms := multiSource(make([]rand.Source, num))
	for i := 0; i < num; i++ {
		ms[i] = rand.NewSource(r.Int63())
	}
	return rand.New(ms)
}

type multiSource []rand.Source

func (ms multiSource) Int63() (r int64) {
	for _, source := range ms {
		r ^= source.Int63()
	}
	return r
}

func (ms multiSource) Seed(seed int64) {
	panic("multiSource Seed should not be called")
}
