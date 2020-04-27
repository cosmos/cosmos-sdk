// +build cli_test

package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/cli/helpers"
)

func TestCLIKeysAddMultisig(t *testing.T) {
	t.Parallel()
	f := helpers.InitFixtures(t)

	// key names order does not matter
	f.KeysAdd("msig1", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", helpers.KeyBar, helpers.KeyBaz))
	ke1Address1 := f.KeysShow("msig1").Address
	f.KeysDelete("msig1")

	f.KeysAdd("msig2", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", helpers.KeyBaz, helpers.KeyBar))
	require.Equal(t, ke1Address1, f.KeysShow("msig2").Address)
	f.KeysDelete("msig2")

	f.KeysAdd("msig3", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", helpers.KeyBar, helpers.KeyBaz),
		"--nosort")
	f.KeysAdd("msig4", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", helpers.KeyBaz, helpers.KeyBar),
		"--nosort")
	require.NotEqual(t, f.KeysShow("msig3").Address, f.KeysShow("msig4").Address)

	// Cleanup testing directories
	f.Cleanup()
}
