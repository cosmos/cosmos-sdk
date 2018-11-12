package simulation

import (
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
func RandomAmount(r *rand.Rand, max sdk.Int) sdk.Int {
	return sdk.NewInt(int64(r.Intn(int(max.Int64()))))
}

// RandomDecAmount generates a random decimal amount
func RandomDecAmount(r *rand.Rand, max sdk.Dec) sdk.Dec {
	randInt := big.NewInt(0).Rand(r, max.Int)
	return sdk.NewDecFromBigIntWithPrec(randInt, sdk.Precision)
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
