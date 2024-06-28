//go:build system_test

package systemtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestQueryTotalSupply(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	raw := cli.CustomQuery("q", "bank", "total-supply")

	exp := map[string]int64{
		"stake":     2000000190,
		"testtoken": 4000000000,
	}
	require.Len(t, gjson.Get(raw, "supply").Array(), len(exp), raw)

	for k, v := range exp {
		got := gjson.Get(raw, fmt.Sprintf("supply.#(denom==%q).amount", k)).Int()
		assert.Equal(t, v, got, raw)
	}
}
