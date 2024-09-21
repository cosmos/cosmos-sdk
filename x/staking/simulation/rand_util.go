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

// RandSocialHandleURIs returns a string array of length num with uris.
func RandSocialHandleURIs(r *rand.Rand, num, uriHostLength int) []string {
	if num == 0 {
		return []string{}
	}
	var socialHandles []string
	for i := 0; i < num; i++ {
		socialHandles = append(socialHandles, RandURIOfHostLength(r, uriHostLength))
	}
	return socialHandles
}
