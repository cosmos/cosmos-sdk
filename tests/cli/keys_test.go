// +build cli_test

package cli_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/cosmos/cosmos-sdk/tests/cli"
)

func TestCLIKeysAddMultisig(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	// key names order does not matter
	f.KeysAdd("msig1", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", cli.KeyBar, cli.KeyBaz))
	ke1Address1 := f.KeysShow("msig1").Address
	f.KeysDelete("msig1")

	f.KeysAdd("msig2", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", cli.KeyBaz, cli.KeyBar))
	require.Equal(t, ke1Address1, f.KeysShow("msig2").Address)
	f.KeysDelete("msig2")

	f.KeysAdd("msig3", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", cli.KeyBar, cli.KeyBaz),
		"--nosort")
	f.KeysAdd("msig4", "--multisig-threshold=2",
		fmt.Sprintf("--multisig=%s,%s", cli.KeyBaz, cli.KeyBar),
		"--nosort")
	require.NotEqual(t, f.KeysShow("msig3").Address, f.KeysShow("msig4").Address)

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIKeysAddRecover(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	exitSuccess, _, _ := f.KeysAddRecover("empty-mnemonic", "")
	require.False(t, exitSuccess)

	exitSuccess, _, _ = f.KeysAddRecover("test-recover", "dentist task convince chimney quality leave banana trade firm crawl eternal easily")
	require.True(t, exitSuccess)
	require.Equal(t, "cosmos1qcfdf69js922qrdr4yaww3ax7gjml6pdds46f4", f.KeyAddress("test-recover").String())

	// Cleanup testing directories
	f.Cleanup()
}

func TestCLIKeysAddRecoverHDPath(t *testing.T) {
	t.Parallel()
	f := cli.InitFixtures(t)

	f.KeysAddRecoverHDPath("test-recoverHD1", "dentist task convince chimney quality leave banana trade firm crawl eternal easily", 0, 0)
	require.Equal(t, "cosmos1qcfdf69js922qrdr4yaww3ax7gjml6pdds46f4", f.KeyAddress("test-recoverHD1").String())

	f.KeysAddRecoverHDPath("test-recoverH2", "dentist task convince chimney quality leave banana trade firm crawl eternal easily", 1, 5)
	require.Equal(t, "cosmos1pdfav2cjhry9k79nu6r8kgknnjtq6a7rykmafy", f.KeyAddress("test-recoverH2").String())

	f.KeysAddRecoverHDPath("test-recoverH3", "dentist task convince chimney quality leave banana trade firm crawl eternal easily", 1, 17)
	require.Equal(t, "cosmos1909k354n6wl8ujzu6kmh49w4d02ax7qvlkv4sn", f.KeyAddress("test-recoverH3").String())

	f.KeysAddRecoverHDPath("test-recoverH4", "dentist task convince chimney quality leave banana trade firm crawl eternal easily", 2, 17)
	require.Equal(t, "cosmos1v9plmhvyhgxk3th9ydacm7j4z357s3nhtwsjat", f.KeyAddress("test-recoverH4").String())

	// Cleanup testing directories
	f.Cleanup()
}
