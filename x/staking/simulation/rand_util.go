package simulation

import (
	"fmt"
	"math/rand"
	"net/url"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// RandURIOfHostLength returns a random valid uri with hostname length n. If n = 0, returns an empty string.
func RandURIOfHostLength(r *rand.Rand, n int) string {
	if n == 0 {
		return ""
	}
	tld := ".com"
	hostLength := n - len(tld)
	uri := &url.URL{
		Scheme: "https",
		Host:   fmt.Sprintf("%s%s", simtypes.RandStringOfLength(r, hostLength), tld),
	}

	return uri.String()
}
