package simulation

import (
	"fmt"
	"math/rand"
	"net/url"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// RandURIOfLength returns a random valid uri with a path of length: n and a host of length:  n - length(tld)
func RandURIOfLength(r *rand.Rand, n int) string {
	uri := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s.com", simtypes.RandStringOfLength(r, n)),
		Path:   simtypes.RandStringOfLength(r, n),
	}
	return uri.String()
}
