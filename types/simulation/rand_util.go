package simulation

import (
	"errors"
	"math/big"
	"math/rand"
	"time"
	"unsafe"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// shamelessly copied from
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-golang#31832326

// RandStringOfLength generates a random string of a particular length
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

	return *(*string)(unsafe.Pointer(&b))
}

// RandPositiveInt get a rand positive sdk.Int
func RandPositiveInt(r *rand.Rand, max sdk.Int) (sdk.Int, error) {
	if !max.GTE(sdk.OneInt()) {
		return sdk.Int{}, errors.New("max too small")
	}

	max = max.Sub(sdk.OneInt())

	return sdk.NewIntFromBigInt(new(big.Int).Rand(r, max.BigInt())).Add(sdk.OneInt()), nil
}

// RandomAmount generates a random amount
// Note: The range of RandomAmount includes max, and is, in fact, biased to return max as well as 0.
func RandomAmount(r *rand.Rand, max sdk.Int) sdk.Int {
	var randInt = big.NewInt(0)

	switch r.Intn(10) {
	case 0:
		// randInt = big.NewInt(0)
	case 1:
		randInt = max.BigInt()
	default: // NOTE: there are 10 total cases.
		randInt = big.NewInt(0).Rand(r, max.BigInt()) // up to max - 1
	}

	return sdk.NewIntFromBigInt(randInt)
}

// RandomDecAmount generates a random decimal amount
// Note: The range of RandomDecAmount includes max, and is, in fact, biased to return max as well as 0.
func RandomDecAmount(r *rand.Rand, max sdk.Dec) sdk.Dec {
	var randInt = big.NewInt(0)

	switch r.Intn(10) {
	case 0:
		// randInt = big.NewInt(0)
	case 1:
		randInt = max.BigInt() // the underlying big int with all precision bits.
	default: // NOTE: there are 10 total cases.
		randInt = big.NewInt(0).Rand(r, max.BigInt())
	}

	return sdk.NewDecFromBigIntWithPrec(randInt, sdk.Precision)
}

// RandTimestamp generates a random timestamp
func RandTimestamp(r *rand.Rand) time.Time {
	// json.Marshal breaks for timestamps greater with year greater than 9999
	unixTime := r.Int63n(253373529600)
	return time.Unix(unixTime, 0)
}

// RandIntBetween returns a random int between two numbers inclusively.
func RandIntBetween(r *rand.Rand, min, max int) int {
	return r.Intn(max-min) + min
}

// returns random subset of the provided coins
// will return at least one coin unless coins argument is empty or malformed
// i.e. 0 amt in coins
func RandSubsetCoins(r *rand.Rand, coins sdk.Coins) sdk.Coins {
	if len(coins) == 0 {
		return sdk.Coins{}
	}
	// make sure at least one coin added
	denomIdx := r.Intn(len(coins))
	coin := coins[denomIdx]
	amt, err := RandPositiveInt(r, coin.Amount)
	// malformed coin. 0 amt in coins
	if err != nil {
		return sdk.Coins{}
	}

	subset := sdk.Coins{sdk.NewCoin(coin.Denom, amt)}

	for i, c := range coins {
		// skip denom that we already chose earlier
		if i == denomIdx {
			continue
		}
		// coin flip if multiple coins
		// if there is single coin then return random amount of it
		if r.Intn(2) == 0 && len(coins) != 1 {
			continue
		}

		amt, err := RandPositiveInt(r, c.Amount)
		// ignore errors and try another denom
		if err != nil {
			continue
		}

		subset = append(subset, sdk.NewCoin(c.Denom, amt))
	}

	return subset.Sort()
}

// DeriveRand derives a new Rand deterministically from another random source.
// Unlike rand.New(rand.NewSource(seed)), the result is "more random"
// depending on the source and state of r.
//
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
