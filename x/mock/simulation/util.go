package simulation

import (
	"fmt"
	"math/rand"
	"sort"

	"github.com/tendermint/tendermint/crypto"

	sdk "github.com/cosmos/cosmos-sdk/types"
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

// Pick a random key from an array
func RandomKey(r *rand.Rand, keys []crypto.PrivKey) crypto.PrivKey {
	return keys[r.Intn(
		len(keys),
	)]
}

// Generate a random amount
func RandomAmount(r *rand.Rand, max sdk.Int) sdk.Int {
	return sdk.NewInt(int64(r.Intn(int(max.Int64()))))
}
