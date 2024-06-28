//go:build system_test

package systemtests

import (
	"testing"
)

func TestQueryTotalSupply(t *testing.T) {
	sut.ResetChain(t)
	sut.StartChain(t)

	cli := NewCLIWrapper(t, sut, verbose)
	raw := cli.CustomQuery("q", "bank", "total-supply")
	t.Log("### got: " + raw)
}
