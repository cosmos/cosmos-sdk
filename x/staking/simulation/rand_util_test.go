package simulation_test

import (
	"math/rand"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"cosmossdk.io/x/staking/simulation"
)

func TestRandURIOfHostLength(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"0-size", 0, 0},
		{"10-size", 10, 10},
		{"1_000_000-size", 1_000_000, 1_000_000},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			got := 0
			uri := simulation.RandURIOfHostLength(r, tc.n)
			if uri != "" {
				parsedUri, err := url.Parse(uri)
				require.NoError(t, err)
				got = len(parsedUri.Host)
			}
			require.Equal(t, tc.want, got)
		})
	}
}

func TestRandSocialHandleURIs(t *testing.T) {
	t.Parallel()
	r := rand.New(rand.NewSource(time.Now().Unix()))
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"0-handles", 0, 0},
		{"10-handles", 10, 10},
		{"100-handles", 100, 100},
	}
	for _, tt := range tests {
		tc := tt
		t.Run(tc.name, func(t *testing.T) {
			uris := simulation.RandSocialHandleURIs(r, tc.n, 10)
			require.Equal(t, tc.want, len(uris))
		})
	}
}
